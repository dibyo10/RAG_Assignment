CREATE TABLE IF NOT EXISTS documents (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    mime_type   TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    error_msg   TEXT,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS chunks (
    id               TEXT PRIMARY KEY,
    document_id      TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index      INTEGER NOT NULL,
    text             TEXT NOT NULL,
    start_char       INTEGER NOT NULL,
    end_char         INTEGER NOT NULL,
    qdrant_point_id  TEXT NOT NULL,
    created_at       INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON chunks(document_id);
