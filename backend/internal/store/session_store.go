package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type Message struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	TurnIndex int    `json:"turn_index"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

func (s *SessionStore) Create(documentID, title string) (*Session, error) {
	if title == "" {
		title = "New Session"
	}
	now := time.Now().UnixMilli()
	sess := &Session{
		ID:         uuid.NewString(),
		DocumentID: documentID,
		Title:      title,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := s.db.Exec(`INSERT INTO sessions (id, document_id, title, created_at, updated_at) VALUES (?,?,?,?,?)`,
		sess.ID, sess.DocumentID, sess.Title, sess.CreatedAt, sess.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return sess, nil
}

func (s *SessionStore) Get(id string) (*Session, error) {
	row := s.db.QueryRow(`SELECT id, document_id, title, created_at, updated_at FROM sessions WHERE id=?`, id)
	var sess Session
	if err := row.Scan(&sess.ID, &sess.DocumentID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &sess, nil
}

func (s *SessionStore) ListByDocument(documentID string) ([]*Session, error) {
	rows, err := s.db.Query(`SELECT id, document_id, title, created_at, updated_at FROM sessions WHERE document_id=? ORDER BY created_at DESC`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.DocumentID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, &sess)
	}
	return sessions, rows.Err()
}

func (s *SessionStore) UpdateTitle(id, title string) error {
	_, err := s.db.Exec(`UPDATE sessions SET title=?, updated_at=? WHERE id=?`, title, time.Now().UnixMilli(), id)
	return err
}

func (s *SessionStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id=?`, id)
	return err
}

func (s *SessionStore) AppendMessage(sessionID, role, content string) (*Message, error) {
	var maxTurn sql.NullInt64
	s.db.QueryRow(`SELECT MAX(turn_index) FROM messages WHERE session_id=?`, sessionID).Scan(&maxTurn)
	nextTurn := 0
	if maxTurn.Valid {
		nextTurn = int(maxTurn.Int64) + 1
	}
	msg := &Message{
		ID:        uuid.NewString(),
		SessionID: sessionID,
		TurnIndex: nextTurn,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UnixMilli(),
	}
	_, err := s.db.Exec(`INSERT INTO messages (id, session_id, turn_index, role, content, created_at) VALUES (?,?,?,?,?,?)`,
		msg.ID, msg.SessionID, msg.TurnIndex, msg.Role, msg.Content, msg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("append message: %w", err)
	}
	return msg, nil
}

func (s *SessionStore) GetMessages(sessionID string) ([]*Message, error) {
	rows, err := s.db.Query(`SELECT id, session_id, turn_index, role, content, created_at FROM messages WHERE session_id=? ORDER BY turn_index ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []*Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.TurnIndex, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, rows.Err()
}
