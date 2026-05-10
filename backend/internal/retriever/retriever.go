package retriever

import (
	"context"
	"fmt"

	qdrant "github.com/qdrant/go-client/qdrant"

	"github.com/dibyochakraborty/notebooklm/internal/embedder"
)

type ScoredChunk struct {
	ChunkID    string
	DocumentID string
	Text       string
	Score      float64
	ChunkIndex int
	StartChar  int
	EndChar    int
	Rank       int
}

type Retriever struct {
	embedder       *embedder.Embedder
	qdrantClient   qdrant.PointsClient
	collectionName string
}

func New(emb *embedder.Embedder, qc qdrant.PointsClient, collectionName string) *Retriever {
	return &Retriever{embedder: emb, qdrantClient: qc, collectionName: collectionName}
}

func (r *Retriever) Retrieve(ctx context.Context, query, documentID string, topK int) ([]*ScoredChunk, error) {
	// Embed query
	vecs, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	// Filter by document_id
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			{
				ConditionOneOf: &qdrant.Condition_Field{
					Field: &qdrant.FieldCondition{
						Key: "document_id",
						Match: &qdrant.Match{
							MatchValue: &qdrant.Match_Keyword{Keyword: documentID},
						},
					},
				},
			},
		},
	}

	withPayload := qdrant.NewWithPayload(true)
	resp, err := r.qdrantClient.Search(ctx, &qdrant.SearchPoints{
		CollectionName: r.collectionName,
		Vector:         vecs[0],
		Filter:         filter,
		Limit:          uint64(topK),
		WithPayload:    withPayload,
	})
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}

	chunks := make([]*ScoredChunk, 0, len(resp.Result))
	for i, hit := range resp.Result {
		payload := hit.Payload
		sc := &ScoredChunk{
			Score: float64(hit.Score),
			Rank:  i + 1,
		}
		if v, ok := payload["chunk_id"]; ok {
			sc.ChunkID = v.GetStringValue()
		}
		if v, ok := payload["document_id"]; ok {
			sc.DocumentID = v.GetStringValue()
		}
		if v, ok := payload["text"]; ok {
			sc.Text = v.GetStringValue()
		}
		if v, ok := payload["chunk_index"]; ok {
			sc.ChunkIndex = int(v.GetIntegerValue())
		}
		if v, ok := payload["start_char"]; ok {
			sc.StartChar = int(v.GetIntegerValue())
		}
		if v, ok := payload["end_char"]; ok {
			sc.EndChar = int(v.GetIntegerValue())
		}
		chunks = append(chunks, sc)
	}
	return chunks, nil
}
