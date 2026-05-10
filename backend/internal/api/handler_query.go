package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dibyochakraborty/notebooklm/internal/llm"
	"github.com/dibyochakraborty/notebooklm/internal/memory"
	"github.com/dibyochakraborty/notebooklm/internal/metrics"
	"github.com/dibyochakraborty/notebooklm/internal/retriever"
	"github.com/dibyochakraborty/notebooklm/internal/store"
)

type queryHandler struct {
	sessStore    *store.SessionStore
	docStore     *store.DocumentStore
	metricsStore *store.MetricsStore
	retriever    *retriever.Retriever
	memory       *memory.Memory
	generator    *llm.Generator
	topK         int
}

type chunkResponse struct {
	ChunkID    string  `json:"chunk_id"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	Rank       int     `json:"rank"`
	ChunkIndex int     `json:"chunk_index"`
	StartChar  int     `json:"start_char"`
	EndChar    int     `json:"end_char"`
}

func (h *queryHandler) Query(c *gin.Context) {
	sessionID := c.Param("id")

	var body struct {
		Query string `json:"query" binding:"required"`
		TopK  int    `json:"top_k"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topK := h.topK
	if body.TopK > 0 {
		topK = body.TopK
	}

	sess, err := h.sessStore.Get(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	start := time.Now()
	ctx := c.Request.Context()

	// Retrieve relevant chunks
	scored, err := h.retriever.Retrieve(ctx, body.Query, sess.DocumentID, topK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "retrieval failed: " + err.Error()})
		return
	}

	// Concurrently: fetch history + (later) LLM call
	// History fetch is fast (<10ms); LLM is slow. We fetch history first since we need it for LLM.
	history, err := h.memory.GetHistory(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "history fetch failed: " + err.Error()})
		return
	}

	// Extract chunk texts for LLM context
	chunkTexts := make([]string, len(scored))
	for i, sc := range scored {
		chunkTexts[i] = sc.Text
	}

	// Generate answer
	answer, err := h.generator.Generate(ctx, chunkTexts, history, body.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generation failed: " + err.Error()})
		return
	}

	latencyMs := time.Since(start).Milliseconds()

	// Compute metrics
	m := metrics.Compute(scored)

	// Build chunk hits for storage
	hits := make([]*store.QueryChunkHit, len(scored))
	for i, sc := range scored {
		hits[i] = &store.QueryChunkHit{
			ChunkID: sc.ChunkID,
			Rank:    sc.Rank,
			Score:   sc.Score,
		}
	}

	// Store query log
	qlog := &store.QueryLog{
		SessionID:       sessionID,
		DocumentID:      sess.DocumentID,
		QueryText:       body.Query,
		TopK:            topK,
		ChunksRetrieved: len(scored),
		ScoreMin:        m.ScoreMin,
		ScoreMax:        m.ScoreMax,
		ScoreMean:       m.ScoreMean,
		ScoreStd:        m.ScoreStd,
		MRR:             m.MRR,
		RecallAtK:       m.RecallAtK,
		NDCG:            m.NDCG,
		LatencyMs:       latencyMs,
	}
	go h.metricsStore.InsertQueryLog(qlog, hits)

	// Persist conversation turn (background to not block response)
	go h.memory.AppendTurn(sessionID, body.Query, answer)

	// Auto-title session after first query
	if len(history) == 0 {
		go autoTitle(h.sessStore, sessionID, body.Query)
	}

	// Build response
	chunkResp := make([]chunkResponse, len(scored))
	for i, sc := range scored {
		chunkResp[i] = chunkResponse{
			ChunkID:    sc.ChunkID,
			Text:       sc.Text,
			Score:      sc.Score,
			Rank:       sc.Rank,
			ChunkIndex: sc.ChunkIndex,
			StartChar:  sc.StartChar,
			EndChar:    sc.EndChar,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"answer":     answer,
			"chunks":     chunkResp,
			"latency_ms": latencyMs,
			"metrics": gin.H{
				"mrr":        m.MRR,
				"recall_at_k": m.RecallAtK,
				"ndcg":       m.NDCG,
				"score_min":  m.ScoreMin,
				"score_max":  m.ScoreMax,
				"score_mean": m.ScoreMean,
				"score_std":  m.ScoreStd,
				"count":      m.Count,
			},
		},
	})
}

func autoTitle(ss *store.SessionStore, sessionID, query string) {
	title := query
	if len(title) > 60 {
		title = title[:57] + "..."
	}
	ss.UpdateTitle(sessionID, title)
}
