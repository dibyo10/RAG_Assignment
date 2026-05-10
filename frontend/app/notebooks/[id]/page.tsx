"use client";
import { use, useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  getDocument, listSessions, getSession, createSession, deleteSession,
} from "@/lib/api";
import type { Document, Session, Message, Chunk } from "@/lib/types";
import { ChatPanel } from "@/components/ChatPanel";
import { ChunkViewer } from "@/components/ChunkViewer";
import { StatusBadge } from "@/components/StatusBadge";

export default function NotebookPage({ params }: { params: Promise<{ id: string }> }) {
  const { id: docId } = use(params);

  const [doc, setDoc] = useState<Document | null>(null);
  const [chunkCount, setChunkCount] = useState(0);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [activeMessages, setActiveMessages] = useState<Message[]>([]);
  const [activeChunks, setActiveChunks] = useState<Chunk[]>([]);
  const [selectedChunkId, setSelectedChunkId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [sourcesOpen, setSourcesOpen] = useState(false);

  const loadDoc = useCallback(async () => {
    const data = await getDocument(docId);
    setDoc(data.document);
    setChunkCount(data.chunk_count);
  }, [docId]);

  const loadSessions = useCallback(async () => {
    const data = await listSessions(docId);
    setSessions(data ?? []);
  }, [docId]);

  useEffect(() => {
    Promise.all([loadDoc(), loadSessions()]).finally(() => setLoading(false));
  }, [loadDoc, loadSessions]);

  // Auto-activate first session if any exist, otherwise create one
  useEffect(() => {
    if (loading) return;
    if (!activeSessionId && sessions.length > 0) {
      activateSession(sessions[0].id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, sessions]);

  const activateSession = useCallback(async (sid: string) => {
    setActiveSessionId(sid);
    setActiveChunks([]);
    setSelectedChunkId(null);
    setSourcesOpen(false);
    const data = await getSession(sid);
    setActiveMessages(data.messages ?? []);
  }, []);

  const handleNewSession = async () => {
    const sess = await createSession(docId);
    setSessions((prev) => [sess, ...prev]);
    await activateSession(sess.id);
  };

  const handleDeleteSession = async (sid: string) => {
    setSessions((prev) => prev.filter((s) => s.id !== sid));
    if (activeSessionId === sid) {
      setActiveSessionId(null);
      setActiveMessages([]);
    }
    deleteSession(sid).catch(() => {});
  };

  const handleChunksChange = useCallback((chunks: Chunk[]) => {
    setActiveChunks(chunks);
    setSelectedChunkId(null);
    setSourcesOpen(true);
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-65px)]">
        <p className="font-display italic text-ink-faint">Opening volume…</p>
      </div>
    );
  }
  if (!doc) {
    return <div className="p-12 text-accent-warm font-mono">Document not found.</div>;
  }

  return (
    <div className="flex h-[calc(100vh-65px)] overflow-hidden">
      {/* === Left sidebar: sessions === */}
      <aside
        className={`shrink-0 border-r border-rule bg-paper-soft flex flex-col transition-all duration-300 ${
          sidebarCollapsed ? "w-12" : "w-64"
        }`}
      >
        {/* Collapse toggle */}
        <button
          onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
          className="border-b border-rule h-12 flex items-center justify-center text-ink-faint hover:text-ink hover:bg-paper-deep/50 transition-colors"
          aria-label="Toggle sidebar"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" className={`transition-transform ${sidebarCollapsed ? "rotate-180" : ""}`}>
            <path d="M10 4L6 8l4 4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>

        {!sidebarCollapsed && (
          <>
            {/* Document header */}
            <div className="p-4 border-b border-rule">
              <Link href="/notebooks" className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint hover:text-ink transition-colors">
                ← Library
              </Link>
              <h2 className="font-display text-lg text-ink mt-2 leading-tight line-clamp-2">
                {(doc.name ?? "Untitled").replace(/\.[^.]+$/, "")}
              </h2>
              <div className="flex items-center gap-2 mt-2">
                <StatusBadge status={doc.status} />
                <span className="text-[10px] font-mono text-ink-faint tabular-nums">
                  {chunkCount} §
                </span>
              </div>
              <Link
                href={`/notebooks/${docId}/metrics`}
                className="text-xs font-display italic text-accent hover:text-ink transition-colors mt-3 inline-flex items-center gap-1 group"
              >
                Retrieval metrics
                <span className="transition-transform group-hover:translate-x-0.5">→</span>
              </Link>
            </div>

            {/* New session button */}
            <button
              onClick={handleNewSession}
              className="m-3 px-3 py-2.5 border border-ink text-ink hover:bg-ink hover:text-paper transition-colors text-sm font-display italic flex items-center justify-center gap-2 group"
            >
              <span className="text-base leading-none">+</span>
              <span>New Conversation</span>
            </button>

            {/* Session list */}
            <div className="flex-1 overflow-y-auto px-2 pb-3">
              {sessions.length === 0 && (
                <p className="text-xs font-display italic text-ink-faint px-2 py-3">
                  No conversations yet.
                </p>
              )}
              {sessions.map((sess) => (
                <div
                  key={sess.id}
                  onClick={() => activateSession(sess.id)}
                  className={`group flex items-center justify-between px-3 py-2.5 cursor-pointer rounded-sm transition-colors ${
                    activeSessionId === sess.id
                      ? "bg-paper-deep text-ink"
                      : "hover:bg-paper-deep/50 text-ink-soft"
                  }`}
                >
                  <span className="text-sm truncate font-light">{sess.title}</span>
                  <button
                    onClick={(e) => { e.stopPropagation(); handleDeleteSession(sess.id); }}
                    className="text-ink-faint hover:text-accent-warm text-xs ml-2 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
                    aria-label="Delete session"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          </>
        )}
      </aside>

      {/* === Main reading column === */}
      <div className="flex-1 flex flex-col min-w-0 bg-paper">
        {activeSessionId ? (
          <>
            {/* Top bar with sources toggle */}
            {activeChunks.length > 0 && (
              <div className="border-b border-rule px-6 py-2.5 flex items-center justify-between">
                <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint">
                  In dialogue
                </span>
                <button
                  onClick={() => setSourcesOpen(!sourcesOpen)}
                  className={`text-xs font-display italic transition-colors ${
                    sourcesOpen ? "text-ink" : "text-ink-soft hover:text-ink"
                  }`}
                >
                  {sourcesOpen ? "Hide" : "Show"} passages ({activeChunks.length})
                </button>
              </div>
            )}
            <ChatPanel
              key={activeSessionId}
              sessionId={activeSessionId}
              documentName={(doc.name ?? "Untitled").replace(/\.[^.]+$/, "")}
              initialMessages={activeMessages}
              onChunksChange={handleChunksChange}
            />
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center px-6">
            <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint mb-3">
              Begin
            </p>
            <h2 className="font-display text-3xl text-ink italic mb-6">
              No conversation in progress.
            </h2>
            <button
              onClick={handleNewSession}
              className="px-6 py-3 bg-ink text-paper hover:bg-accent transition-colors font-display italic text-base"
            >
              Start a new dialogue
            </button>
          </div>
        )}
      </div>

      {/* === Right slide-out: passages === */}
      <aside
        className={`shrink-0 border-l border-rule bg-paper-soft transition-all duration-300 overflow-hidden ${
          sourcesOpen && activeChunks.length > 0 ? "w-80" : "w-0"
        }`}
      >
        <div className="w-80 h-full flex flex-col">
          <div className="border-b border-rule px-4 h-12 flex items-center justify-between shrink-0">
            <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint">
              Retrieved Passages
            </span>
            <button
              onClick={() => setSourcesOpen(false)}
              className="text-ink-faint hover:text-ink text-sm"
              aria-label="Close passages"
            >
              ✕
            </button>
          </div>
          <div className="flex-1 overflow-y-auto">
            <ChunkViewer
              chunks={activeChunks}
              selectedChunkId={selectedChunkId}
              onSelect={setSelectedChunkId}
            />
          </div>
        </div>
      </aside>
    </div>
  );
}
