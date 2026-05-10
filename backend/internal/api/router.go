package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	qdrant "github.com/qdrant/go-client/qdrant"

	"github.com/dibyochakraborty/notebooklm/internal/embedder"
	"github.com/dibyochakraborty/notebooklm/internal/ingestion"
	"github.com/dibyochakraborty/notebooklm/internal/llm"
	"github.com/dibyochakraborty/notebooklm/internal/memory"
	"github.com/dibyochakraborty/notebooklm/internal/retriever"
	"github.com/dibyochakraborty/notebooklm/internal/store"
	"github.com/dibyochakraborty/notebooklm/pkg/config"
)

func NewRouter(cfg *config.Config, db *sql.DB, qdrantPoints qdrant.PointsClient, qdrantCollections qdrant.CollectionsClient, emb *embedder.Embedder, gen *llm.Generator) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	docStore := store.NewDocumentStore(db)
	sessStore := store.NewSessionStore(db)
	metricsStore := store.NewMetricsStore(db)

	ret := retriever.New(emb, qdrantPoints, cfg.CollectionName)
	mem := memory.New(sessStore, cfg.MaxHistoryTurns)
	pipeline := ingestion.New(docStore, emb, qdrantPoints, cfg.CollectionName, cfg.ChunkSize, cfg.ChunkOverlap, cfg.EmbedWorkers)

	dh := &documentHandler{docStore: docStore, pipeline: pipeline, uploadDir: cfg.UploadDir}
	sh := &sessionHandler{sessStore: sessStore, docStore: docStore}
	qh := &queryHandler{
		sessStore:    sessStore,
		docStore:     docStore,
		metricsStore: metricsStore,
		retriever:    ret,
		memory:       mem,
		generator:    gen,
		topK:         cfg.TopK,
	}
	mh := &metricsHandler{metricsStore: metricsStore}

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			if err := db.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db unavailable"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok", "qdrant_host": cfg.QdrantHost})
		})

		v1.POST("/documents", dh.Upload)
		v1.GET("/documents", dh.List)
		v1.GET("/documents/:id", dh.Get)
		v1.DELETE("/documents/:id", dh.Delete)

		v1.POST("/sessions", sh.Create)
		v1.GET("/sessions", sh.List)
		v1.GET("/sessions/:id", sh.Get)
		v1.PATCH("/sessions/:id", sh.UpdateTitle)
		v1.DELETE("/sessions/:id", sh.Delete)

		v1.POST("/sessions/:id/query", qh.Query)

		v1.GET("/metrics/global", mh.Global)
		v1.GET("/metrics/sessions/:id", mh.Session)
		v1.GET("/metrics/documents/:id", mh.Document)
	}

	_ = qdrantCollections // used during server init, not per-request

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
