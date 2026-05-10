package memory

import (
	"github.com/dibyochakraborty/notebooklm/internal/store"
)

type Memory struct {
	sessionStore    *store.SessionStore
	maxHistoryTurns int
}

func New(ss *store.SessionStore, maxHistoryTurns int) *Memory {
	return &Memory{sessionStore: ss, maxHistoryTurns: maxHistoryTurns}
}

// GetHistory returns the last maxHistoryTurns pairs of messages (oldest first).
func (m *Memory) GetHistory(sessionID string) ([]*store.Message, error) {
	msgs, err := m.sessionStore.GetMessages(sessionID)
	if err != nil {
		return nil, err
	}
	// Each "turn pair" = 1 user + 1 assistant message = 2 messages
	maxMessages := m.maxHistoryTurns * 2
	if len(msgs) > maxMessages {
		msgs = msgs[len(msgs)-maxMessages:]
	}
	return msgs, nil
}

func (m *Memory) AppendTurn(sessionID, userContent, assistantContent string) error {
	if _, err := m.sessionStore.AppendMessage(sessionID, "user", userContent); err != nil {
		return err
	}
	if _, err := m.sessionStore.AppendMessage(sessionID, "assistant", assistantContent); err != nil {
		return err
	}
	return nil
}
