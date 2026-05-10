import type { Document, Session, Message, QueryResponse, QueryLog, DayMetrics } from "./types";

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    cache: "no-store",
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error ?? "request failed");
  return json.data as T;
}

// Documents
export const listDocuments = () => request<Document[]>("/documents");
export const getDocument = (id: string) =>
  request<{ document: Document; chunk_count: number }>(`/documents/${id}`);
export const deleteDocument = (id: string) =>
  request<{ deleted: boolean }>(`/documents/${id}`, { method: "DELETE" });

export async function uploadDocument(file: File): Promise<Document> {
  const form = new FormData();
  form.append("file", file);
  const res = await fetch(`${BASE}/documents`, { method: "POST", body: form });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error ?? "upload failed");
  return json.data as Document;
}

// Sessions
export const createSession = (documentId: string, title?: string) =>
  request<Session>("/sessions", {
    method: "POST",
    body: JSON.stringify({ document_id: documentId, title }),
  });
export const listSessions = (documentId: string) =>
  request<Session[]>(`/sessions?document_id=${documentId}`);
export const getSession = (id: string) =>
  request<{ session: Session; messages: Message[] }>(`/sessions/${id}`);
export const updateSessionTitle = (id: string, title: string) =>
  request<{ updated: boolean }>(`/sessions/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ title }),
  });
export const deleteSession = (id: string) =>
  request<{ deleted: boolean }>(`/sessions/${id}`, { method: "DELETE" });

// Query
export const sendQuery = (sessionId: string, query: string, topK?: number) =>
  request<QueryResponse>(`/sessions/${sessionId}/query`, {
    method: "POST",
    body: JSON.stringify({ query, top_k: topK }),
  });

// Metrics
export const getGlobalMetrics = () => request<DayMetrics[]>("/metrics/global");
export const getSessionMetrics = (id: string) =>
  request<QueryLog[]>(`/metrics/sessions/${id}`);
export const getDocumentMetrics = (id: string) =>
  request<DayMetrics[]>(`/metrics/documents/${id}`);
