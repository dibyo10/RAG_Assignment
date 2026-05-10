package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type QueryLog struct {
	ID              string  `json:"id"`
	SessionID       string  `json:"session_id"`
	DocumentID      string  `json:"document_id"`
	QueryText       string  `json:"query_text"`
	TopK            int     `json:"top_k"`
	ChunksRetrieved int     `json:"chunks_retrieved"`
	ScoreMin        float64 `json:"score_min"`
	ScoreMax        float64 `json:"score_max"`
	ScoreMean       float64 `json:"score_mean"`
	ScoreStd        float64 `json:"score_std"`
	MRR             float64 `json:"mrr"`
	RecallAtK       float64 `json:"recall_at_k"`
	NDCG            float64 `json:"ndcg"`
	LatencyMs       int64   `json:"latency_ms"`
	CreatedAt       int64   `json:"created_at"`
}

type QueryChunkHit struct {
	QueryLogID string
	ChunkID    string
	Rank       int
	Score      float64
}

type DayMetrics struct {
	Day        string  `json:"day"`
	AvgMRR     float64 `json:"avg_mrr"`
	AvgRecall  float64 `json:"avg_recall"`
	AvgNDCG    float64 `json:"avg_ndcg"`
	AvgScore   float64 `json:"avg_score"`
	QueryCount int     `json:"query_count"`
}

type MetricsStore struct {
	db *sql.DB
}

func NewMetricsStore(db *sql.DB) *MetricsStore {
	return &MetricsStore{db: db}
}

func (s *MetricsStore) InsertQueryLog(log *QueryLog, hits []*QueryChunkHit) error {
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	if log.CreatedAt == 0 {
		log.CreatedAt = time.Now().UnixMilli()
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO query_logs (id, session_id, document_id, query_text, top_k, chunks_retrieved, score_min, score_max, score_mean, score_std, mrr, recall_at_k, ndcg, latency_ms, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		log.ID, log.SessionID, log.DocumentID, log.QueryText, log.TopK, log.ChunksRetrieved,
		log.ScoreMin, log.ScoreMax, log.ScoreMean, log.ScoreStd,
		log.MRR, log.RecallAtK, log.NDCG, log.LatencyMs, log.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insert query log: %w", err)
	}
	for _, h := range hits {
		_, err = tx.Exec(`INSERT INTO query_chunk_hits (query_log_id, chunk_id, rank, score) VALUES (?,?,?,?)`,
			log.ID, h.ChunkID, h.Rank, h.Score)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert chunk hit: %w", err)
		}
	}
	return tx.Commit()
}

func (s *MetricsStore) GetSessionMetrics(sessionID string) ([]*QueryLog, error) {
	rows, err := s.db.Query(`SELECT id, session_id, document_id, query_text, top_k, chunks_retrieved, score_min, score_max, score_mean, score_std, mrr, recall_at_k, ndcg, latency_ms, created_at FROM query_logs WHERE session_id=? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*QueryLog
	for rows.Next() {
		var l QueryLog
		if err := rows.Scan(&l.ID, &l.SessionID, &l.DocumentID, &l.QueryText, &l.TopK, &l.ChunksRetrieved,
			&l.ScoreMin, &l.ScoreMax, &l.ScoreMean, &l.ScoreStd, &l.MRR, &l.RecallAtK, &l.NDCG, &l.LatencyMs, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}

func (s *MetricsStore) GetGlobalMetrics() ([]*DayMetrics, error) {
	rows, err := s.db.Query(`
		SELECT
			date(created_at/1000, 'unixepoch') as day,
			AVG(mrr)         as avg_mrr,
			AVG(recall_at_k) as avg_recall,
			AVG(ndcg)        as avg_ndcg,
			AVG(score_mean)  as avg_score,
			COUNT(*)         as query_count
		FROM query_logs
		GROUP BY day
		ORDER BY day ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*DayMetrics
	for rows.Next() {
		var d DayMetrics
		if err := rows.Scan(&d.Day, &d.AvgMRR, &d.AvgRecall, &d.AvgNDCG, &d.AvgScore, &d.QueryCount); err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	return results, rows.Err()
}

func (s *MetricsStore) GetDocumentMetrics(documentID string) ([]*DayMetrics, error) {
	rows, err := s.db.Query(`
		SELECT
			date(created_at/1000, 'unixepoch') as day,
			AVG(mrr)         as avg_mrr,
			AVG(recall_at_k) as avg_recall,
			AVG(ndcg)        as avg_ndcg,
			AVG(score_mean)  as avg_score,
			COUNT(*)         as query_count
		FROM query_logs
		WHERE document_id=?
		GROUP BY day
		ORDER BY day ASC
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*DayMetrics
	for rows.Next() {
		var d DayMetrics
		if err := rows.Scan(&d.Day, &d.AvgMRR, &d.AvgRecall, &d.AvgNDCG, &d.AvgScore, &d.QueryCount); err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	return results, rows.Err()
}
