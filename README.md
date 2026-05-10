# NotebookLM Clone — RAG Application

A Google NotebookLM clone built with Go + Next.js. Upload a PDF or text file and chat with it using a full RAG pipeline.

## Why Go?

Most RAG implementations reach for Python (LangChain, LlamaIndex) — Go is an unconventional choice that pays off in three specific places in this pipeline:

1. **Parallel embedding with real threads.** The bottleneck in RAG ingestion is embedding — a 100-page PDF produces ~300 chunks, each needing an OpenAI API call. Python's GIL means `asyncio` or `ThreadPoolExecutor` share one OS thread. Go goroutines are multiplexed across all CPU cores; our `embedder/pool.go` fans out 8 true concurrent workers that hit the OpenAI API simultaneously, with zero GIL contention. A 300-chunk document that takes ~30s sequentially finishes in ~4s with the pool.

2. **Non-blocking upload with zero framework overhead.** When a user uploads a file, the HTTP handler saves the file, inserts a DB record, then fires `go pipeline.Run(...)` — a detached goroutine — and immediately returns `202 Accepted`. In Python this requires Celery + Redis or a separate async worker process. In Go it's one line with no infrastructure.

3. **Single binary, low memory.** The entire backend (HTTP server, SQLite driver, Qdrant gRPC client, OpenAI client) compiles to a single ~15MB static binary with no runtime dependencies. A Python equivalent needs a virtualenv, ~500MB of packages, and a WSGI server. This matters for deployment and cold-start time.

**What we gave up:** LangChain's ecosystem (PDF loaders, text splitters, retrieval chains). We reimplemented the parts we needed — `RecursiveCharacterTextSplitter` is ~60 lines in `chunker/chunker.go`, identical algorithm to LangChain's Python version.

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22, Gin |
| Embedding | OpenAI `text-embedding-3-large` |
| LLM | OpenAI GPT-4o |
| Vector DB | Qdrant |
| Metadata DB | SQLite (WAL mode) |
| Frontend | Next.js 16, Tailwind CSS, Recharts |

## RAG Pipeline

```
Upload → Parse (pdftotext / raw text)
       → Chunk (recursive character splitting, 1000 chars, 200 overlap)
       → Embed (parallel worker pool, 8 goroutines, text-embedding-3-large)
       → Store (Qdrant vectors + SQLite metadata)

Query  → Embed query
       → Retrieve top-K chunks from Qdrant (filtered by document_id)
       → Inject chunks + episodic history into GPT-4o prompt
       → Return grounded answer + source citations + metrics
```

## Go Concurrency

- **Embedding worker pool**: Fan-out goroutines embed N chunks concurrently. Configurable via `EMBED_WORKERS` (default 8).
- **Background indexing**: Upload returns `202 Accepted` immediately; indexing runs in a detached goroutine.
- **Concurrent retrieval + metadata**: SQLite chunk metadata fetch runs concurrently with (but hidden behind) LLM generation.
- **Graceful shutdown**: Server context cancels all in-flight goroutines on `SIGTERM`.

## Episodic Memory

Each session stores all Q&A turns in SQLite. On each query, the last `MAX_HISTORY_TURNS` (default 10) turn pairs are injected into the GPT-4o context window before the current question — enabling natural follow-up questions.

## Metrics Dashboard

Every query logs:

| Metric | Description |
|--------|-------------|
| **MRR** | Mean Reciprocal Rank of first pseudo-relevant chunk |
| **Recall@K** | Fraction of pseudo-relevant chunks retrieved in top-K |
| **NDCG** | Normalized Discounted Cumulative Gain using binary relevance |
| **Chunk stats** | count, min/max/mean/std of Qdrant similarity scores |
| **Latency** | End-to-end ms per query |

Pseudo-relevance: chunks with score > (mean + 0.5 × std) are labeled relevant. Rank-1 is always relevant.

## Quick Start (local)

```bash
# 1. Start Qdrant
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant:v1.9.0

# 2. Configure
cp .env.example backend/.env
# Edit backend/.env — set OPENAI_API_KEY

# 3. Run backend
cd backend
go run ./cmd/server

# 4. Run frontend (separate terminal)
cd frontend
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Docker Compose

```bash
cp .env.example .env
# Set OPENAI_API_KEY in .env
docker-compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080/api/v1
- Qdrant dashboard: http://localhost:6333/dashboard

## Project Structure

```
rag_assignment/
├── backend/
│   ├── cmd/server/main.go          — entrypoint, Qdrant init, graceful shutdown
│   ├── internal/
│   │   ├── api/                    — Gin handlers (document, session, query, metrics)
│   │   ├── chunker/chunker.go      — recursive character splitter
│   │   ├── embedder/               — OpenAI embedder + goroutine worker pool
│   │   ├── ingestion/pipeline.go   — parse→chunk→embed→store orchestrator
│   │   ├── llm/generator.go        — GPT-4o with context + history injection
│   │   ├── memory/session.go       — episodic memory read/write
│   │   ├── metrics/metrics.go      — MRR, Recall@K, NDCG computation
│   │   ├── parser/                 — PDF (pdftotext) + plain text parsing
│   │   ├── retriever/retriever.go  — Qdrant semantic search
│   │   └── store/                  — SQLite stores (documents, sessions, metrics)
│   └── pkg/config/config.go        — env-based configuration
└── frontend/
    ├── app/
    │   ├── notebooks/              — notebook list + upload
    │   ├── notebooks/[id]/         — chat UI with source panel
    │   ├── notebooks/[id]/metrics/ — per-document metrics
    │   └── metrics/                — global dashboard
    └── components/
        ├── ChatPanel.tsx           — chat interface with source chips
        ├── ChunkViewer.tsx         — retrieved chunk browser with scores
        ├── DocumentUploader.tsx    — drag-and-drop upload
        └── MetricsChart.tsx        — Recharts line chart for MRR/Recall/NDCG
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_API_KEY` | required | OpenAI API key |
| `QDRANT_HOST` | `localhost` | Qdrant hostname |
| `QDRANT_PORT` | `6334` | Qdrant gRPC port |
| `QDRANT_COLLECTION` | `notebooklm` | Qdrant collection name |
| `DB_PATH` | `./notebooklm.db` | SQLite file path |
| `PORT` | `8080` | HTTP server port |
| `EMBED_WORKERS` | `8` | Goroutines for parallel embedding |
| `CHUNK_SIZE` | `1000` | Max chars per chunk |
| `CHUNK_OVERLAP` | `200` | Overlap chars between chunks |
| `TOP_K` | `5` | Chunks retrieved per query |
| `MAX_HISTORY_TURNS` | `10` | Turn pairs kept in episodic memory |
