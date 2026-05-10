package ingestion

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	qdrant "github.com/qdrant/go-client/qdrant"

	"github.com/dibyochakraborty/notebooklm/internal/chunker"
	"github.com/dibyochakraborty/notebooklm/internal/embedder"
	"github.com/dibyochakraborty/notebooklm/internal/parser"
	"github.com/dibyochakraborty/notebooklm/internal/store"
)

type Pipeline struct {
	docStore       *store.DocumentStore
	embedder       *embedder.Embedder
	qdrantClient   qdrant.PointsClient
	collectionName string
	chunkSize      int
	chunkOverlap   int
	embedWorkers   int
}

func New(
	docStore *store.DocumentStore,
	emb *embedder.Embedder,
	qc qdrant.PointsClient,
	collectionName string,
	chunkSize, chunkOverlap, embedWorkers int,
) *Pipeline {
	return &Pipeline{
		docStore:       docStore,
		embedder:       emb,
		qdrantClient:   qc,
		collectionName: collectionName,
		chunkSize:      chunkSize,
		chunkOverlap:   chunkOverlap,
		embedWorkers:   embedWorkers,
	}
}

// Run processes the document: parse → chunk → embed → store. Designed to run in a goroutine.
func (p *Pipeline) Run(ctx context.Context, docID, filePath, mimeType string) {
	if err := p.docStore.SetStatus(docID, "indexing", ""); err != nil {
		log.Printf("ingestion: set status indexing: %v", err)
		return
	}

	if err := p.process(ctx, docID, filePath, mimeType); err != nil {
		log.Printf("ingestion: process %s: %v", docID, err)
		p.docStore.SetStatus(docID, "failed", err.Error())
		return
	}

	p.docStore.SetStatus(docID, "ready", "")
	log.Printf("ingestion: document %s ready", docID)
}

func (p *Pipeline) process(ctx context.Context, docID, filePath, mimeType string) error {
	// Stage 1: Parse
	text, err := parser.Parse(filePath, mimeType)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(text) == 0 {
		return fmt.Errorf("document is empty after parsing")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Stage 2: Chunk
	chunks := chunker.Chunk(text, p.chunkSize, p.chunkOverlap)
	if len(chunks) == 0 {
		return fmt.Errorf("no chunks produced")
	}
	log.Printf("ingestion: %d chunks for doc %s", len(chunks), docID)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Stage 3: Embed (concurrent worker pool)
	embedded, err := p.embedder.EmbedChunks(ctx, chunks, p.embedWorkers)
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Stage 4: Store in Qdrant + SQLite
	return p.store(ctx, docID, embedded)
}

func (p *Pipeline) store(ctx context.Context, docID string, embedded []embedder.EmbeddedChunk) error {
	now := time.Now().UnixMilli()

	// Build Qdrant points
	points := make([]*qdrant.PointStruct, len(embedded))
	dbChunks := make([]*store.Chunk, len(embedded))

	for i, e := range embedded {
		pointID := uuid.NewString()
		points[i] = &qdrant.PointStruct{
			Id: &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: pointID}},
			Vectors: &qdrant.Vectors{VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{Data: e.Embedding},
			}},
			Payload: map[string]*qdrant.Value{
				"document_id":  {Kind: &qdrant.Value_StringValue{StringValue: docID}},
				"chunk_id":     {Kind: &qdrant.Value_StringValue{StringValue: pointID}},
				"chunk_index":  {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(e.Chunk.ChunkIndex)}},
				"text":         {Kind: &qdrant.Value_StringValue{StringValue: e.Chunk.Text}},
				"start_char":   {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(e.Chunk.StartChar)}},
				"end_char":     {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(e.Chunk.EndChar)}},
			},
		}
		dbChunks[i] = &store.Chunk{
			ID:            pointID,
			DocumentID:    docID,
			ChunkIndex:    e.Chunk.ChunkIndex,
			Text:          e.Chunk.Text,
			StartChar:     e.Chunk.StartChar,
			EndChar:       e.Chunk.EndChar,
			QdrantPointID: pointID,
			CreatedAt:     now,
		}
	}

	// Upsert into Qdrant
	_, err := p.qdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: p.collectionName,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("qdrant upsert: %w", err)
	}

	// Insert chunk metadata into SQLite
	if err := p.docStore.InsertChunks(dbChunks); err != nil {
		return fmt.Errorf("sqlite insert chunks: %w", err)
	}

	return nil
}
