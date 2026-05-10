CREATE TABLE IF NOT EXISTS messages (
    id           TEXT PRIMARY KEY,
    session_id   TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    turn_index   INTEGER NOT NULL,
    role         TEXT NOT NULL CHECK(role IN ('user','assistant')),
    content      TEXT NOT NULL,
    created_at   INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_session_id_turn ON messages(session_id, turn_index);
