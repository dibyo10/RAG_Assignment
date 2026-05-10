CREATE TABLE IF NOT EXISTS sessions (
    id           TEXT PRIMARY KEY,
    document_id  TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    title        TEXT NOT NULL DEFAULT 'New Session',
    created_at   INTEGER NOT NULL,
    updated_at   INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_document_id ON sessions(document_id);
