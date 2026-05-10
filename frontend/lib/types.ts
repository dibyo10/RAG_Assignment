export interface Document {
  id: string;
  name: string;
  mime_type: string;
  status: "pending" | "indexing" | "ready" | "failed";
  error_msg: string;
  created_at: number;
  updated_at: number;
}

export interface Session {
  id: string;
  document_id: string;
  title: string;
  created_at: number;
  updated_at: number;
}

export interface Message {
  id: string;
  session_id: string;
  turn_index: number;
  role: "user" | "assistant";
  content: string;
  created_at: number;
}

export interface Chunk {
  chunk_id: string;
  text: string;
  score: number;
  rank: number;
  chunk_index: number;
  start_char: number;
  end_char: number;
}

export interface QueryMetrics {
  mrr: number;
  recall_at_k: number;
  ndcg: number;
  score_min: number;
  score_max: number;
  score_mean: number;
  score_std: number;
  count: number;
}

export interface QueryResponse {
  answer: string;
  chunks: Chunk[];
  latency_ms: number;
  metrics: QueryMetrics;
}

export interface QueryLog {
  id: string;
  session_id: string;
  document_id: string;
  query_text: string;
  top_k: number;
  chunks_retrieved: number;
  score_min: number;
  score_max: number;
  score_mean: number;
  score_std: number;
  mrr: number;
  recall_at_k: number;
  ndcg: number;
  latency_ms: number;
  created_at: number;
}

export interface DayMetrics {
  day: string;
  avg_mrr: number;
  avg_recall: number;
  avg_ndcg: number;
  avg_score: number;
  query_count: number;
}
