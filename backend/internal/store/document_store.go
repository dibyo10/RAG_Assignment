package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	Status    string `json:"status"`
	ErrorMsg  string `json:"error_msg"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type Chunk struct {
	ID            string `json:"id"`
	DocumentID    string `json:"document_id"`
	ChunkIndex    int    `json:"chunk_index"`
	Text          string `json:"text"`
	StartChar     int    `json:"start_char"`
	EndChar       int    `json:"end_char"`
	QdrantPointID string `json:"qdrant_point_id"`
	CreatedAt     int64  `json:"created_at"`
}

type DocumentStore struct {
	db *sql.DB
}

func NewDocumentStore(db *sql.DB) *DocumentStore {
	return &DocumentStore{db: db}
}

func (s *DocumentStore) Create(name, mimeType string) (*Document, error) {
	now := time.Now().UnixMilli()
	doc := &Document{
		ID:        uuid.NewString(),
		Name:      name,
		MimeType:  mimeType,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := s.db.Exec(
		`INSERT INTO documents (id, name, mime_type, status, created_at, updated_at) VALUES (?,?,?,?,?,?)`,
		doc.ID, doc.Name, doc.MimeType, doc.Status, doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}
	return doc, nil
}

func (s *DocumentStore) Get(id string) (*Document, error) {
	row := s.db.QueryRow(`SELECT id, name, mime_type, status, COALESCE(error_msg,''), created_at, updated_at FROM documents WHERE id=?`, id)
	var d Document
	if err := row.Scan(&d.ID, &d.Name, &d.MimeType, &d.Status, &d.ErrorMsg, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return &d, nil
}

func (s *DocumentStore) List() ([]*Document, error) {
	rows, err := s.db.Query(`SELECT id, name, mime_type, status, COALESCE(error_msg,''), created_at, updated_at FROM documents ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var docs []*Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.Name, &d.MimeType, &d.Status, &d.ErrorMsg, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, &d)
	}
	return docs, rows.Err()
}

func (s *DocumentStore) SetStatus(id, status, errMsg string) error {
	_, err := s.db.Exec(`UPDATE documents SET status=?, error_msg=?, updated_at=? WHERE id=?`,
		status, errMsg, time.Now().UnixMilli(), id)
	return err
}

func (s *DocumentStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM documents WHERE id=?`, id)
	return err
}

func (s *DocumentStore) InsertChunks(chunks []*Chunk) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO chunks (id, document_id, chunk_index, text, start_char, end_char, qdrant_point_id, created_at) VALUES (?,?,?,?,?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, c := range chunks {
		if _, err := stmt.Exec(c.ID, c.DocumentID, c.ChunkIndex, c.Text, c.StartChar, c.EndChar, c.QdrantPointID, c.CreatedAt); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert chunk %d: %w", c.ChunkIndex, err)
		}
	}
	return tx.Commit()
}

func (s *DocumentStore) GetChunks(documentID string) ([]*Chunk, error) {
	rows, err := s.db.Query(`SELECT id, document_id, chunk_index, text, start_char, end_char, qdrant_point_id, created_at FROM chunks WHERE document_id=? ORDER BY chunk_index`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chunks []*Chunk
	for rows.Next() {
		var c Chunk
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.ChunkIndex, &c.Text, &c.StartChar, &c.EndChar, &c.QdrantPointID, &c.CreatedAt); err != nil {
			return nil, err
		}
		chunks = append(chunks, &c)
	}
	return chunks, rows.Err()
}

func (s *DocumentStore) GetChunksByIDs(ids []string) ([]*Chunk, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}
	rows, err := s.db.Query(`SELECT id, document_id, chunk_index, text, start_char, end_char, qdrant_point_id, created_at FROM chunks WHERE id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chunks []*Chunk
	for rows.Next() {
		var c Chunk
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.ChunkIndex, &c.Text, &c.StartChar, &c.EndChar, &c.QdrantPointID, &c.CreatedAt); err != nil {
			return nil, err
		}
		chunks = append(chunks, &c)
	}
	return chunks, rows.Err()
}

func (s *DocumentStore) CountChunks(documentID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM chunks WHERE document_id=?`, documentID).Scan(&count)
	return count, err
}
