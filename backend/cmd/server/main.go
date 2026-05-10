package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dibyochakraborty/notebooklm/internal/api"
	"github.com/dibyochakraborty/notebooklm/internal/embedder"
	"github.com/dibyochakraborty/notebooklm/internal/llm"
	"github.com/dibyochakraborty/notebooklm/internal/store"
	"github.com/dibyochakraborty/notebooklm/pkg/config"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY, name TEXT NOT NULL, mime_type TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending', error_msg TEXT,
		created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS chunks (
		id TEXT PRIMARY KEY, document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
		chunk_index INTEGER NOT NULL, text TEXT NOT NULL,
		start_char INTEGER NOT NULL, end_char INTEGER NOT NULL,
		qdrant_point_id TEXT NOT NULL, created_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON chunks(document_id);`,

	`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY, document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
		title TEXT NOT NULL DEFAULT 'New Session', created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_document_id ON sessions(document_id);`,

	`CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY, session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		turn_index INTEGER NOT NULL, role TEXT NOT NULL CHECK(role IN ('user','assistant')),
		content TEXT NOT NULL, created_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_messages_session_id_turn ON messages(session_id, turn_index);`,

	`CREATE TABLE IF NOT EXISTS query_logs (
		id TEXT PRIMARY KEY, session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		document_id TEXT NOT NULL, query_text TEXT NOT NULL,
		top_k INTEGER NOT NULL, chunks_retrieved INTEGER NOT NULL,
		score_min REAL, score_max REAL, score_mean REAL, score_std REAL,
		mrr REAL, recall_at_k REAL, ndcg REAL, latency_ms INTEGER NOT NULL, created_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS query_chunk_hits (
		query_log_id TEXT NOT NULL REFERENCES query_logs(id) ON DELETE CASCADE,
		chunk_id TEXT NOT NULL, rank INTEGER NOT NULL, score REAL NOT NULL,
		PRIMARY KEY (query_log_id, chunk_id)
	);
	CREATE INDEX IF NOT EXISTS idx_query_logs_session  ON query_logs(session_id);
	CREATE INDEX IF NOT EXISTS idx_query_logs_document ON query_logs(document_id);
	CREATE INDEX IF NOT EXISTS idx_query_logs_created  ON query_logs(created_at);`,
}

func main() {
	cfg := config.Load()

	// Open SQLite
	db, err := store.Open(cfg.DBPath, migrations)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Connect to Qdrant via gRPC
	qdrantAddr := fmt.Sprintf("%s:%d", cfg.QdrantHost, cfg.QdrantPort)
	conn, err := grpc.NewClient(qdrantAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect qdrant: %v", err)
	}
	defer conn.Close()

	collectionsClient := qdrant.NewCollectionsClient(conn)
	pointsClient := qdrant.NewPointsClient(conn)

	// Ensure collection exists
	if err := ensureCollection(collectionsClient, cfg.CollectionName); err != nil {
		log.Fatalf("ensure qdrant collection: %v", err)
	}

	// Build embedder and LLM generator (both use Gemini API key)
	emb, err := embedder.New(cfg.GeminiKey)
	if err != nil {
		log.Fatalf("create embedder: %v", err)
	}
	gen, err := llm.New(cfg.GeminiKey)
	if err != nil {
		log.Fatalf("create generator: %v", err)
	}

	// Build and start HTTP server
	router := api.NewRouter(cfg, db, pointsClient, collectionsClient, emb, gen)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("stopped")
}

// Vector size for gemini-embedding-2 is 3072
const embeddingDim = 3072

func ensureCollection(client qdrant.CollectionsClient, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Get(ctx, &qdrant.GetCollectionInfoRequest{CollectionName: name})
	if err == nil {
		return nil // already exists
	}

	_, err = client.Create(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     embeddingDim,
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})
	return err
}
