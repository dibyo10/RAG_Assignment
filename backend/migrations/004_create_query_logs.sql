CREATE TABLE IF NOT EXISTS query_logs (
    id               TEXT PRIMARY KEY,
    session_id       TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    document_id      TEXT NOT NULL,
    query_text       TEXT NOT NULL,
    top_k            INTEGER NOT NULL,
    chunks_retrieved INTEGER NOT NULL,
    score_min        REAL,
    score_max        REAL,
    score_mean       REAL,
    score_std        REAL,
    mrr              REAL,
    recall_at_k      REAL,
    ndcg             REAL,
    latency_ms       INTEGER NOT NULL,
    created_at       INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS query_chunk_hits (
    query_log_id  TEXT NOT NULL REFERENCES query_logs(id) ON DELETE CASCADE,
    chunk_id      TEXT NOT NULL,
    rank          INTEGER NOT NULL,
    score         REAL NOT NULL,
    PRIMARY KEY (query_log_id, chunk_id)
);

CREATE INDEX IF NOT EXISTS idx_query_logs_session    ON query_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_query_logs_document   ON query_logs(document_id);
CREATE INDEX IF NOT EXISTS idx_query_logs_created    ON query_logs(created_at);
