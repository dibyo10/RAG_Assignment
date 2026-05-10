"use client";
import { useState, useRef, useEffect } from "react";
import ReactMarkdown from "react-markdown";
import { sendQuery } from "@/lib/api";
import type { Chunk, Message, QueryMetrics } from "@/lib/types";

interface ChatMessage {
  role: "user" | "assistant";
  content: string;
  chunks?: Chunk[];
  metrics?: QueryMetrics;
  latency_ms?: number;
}

interface Props {
  sessionId: string;
  documentName: string;
  initialMessages: Message[];
  onChunksChange: (chunks: Chunk[]) => void;
}

export function ChatPanel({ sessionId, documentName, initialMessages, onChunksChange }: Props) {
  const [messages, setMessages] = useState<ChatMessage[]>(
    initialMessages.map((m) => ({ role: m.role, content: m.content }))
  );
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, loading]);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = Math.min(textareaRef.current.scrollHeight, 200) + "px";
    }
  }, [input]);

  async function handleSend() {
    const query = input.trim();
    if (!query || loading) return;
    setInput("");
    setError(null);
    setMessages((prev) => [...prev, { role: "user", content: query }]);
    setLoading(true);
    try {
      const res = await sendQuery(sessionId, query);
      setMessages((prev) => [
        ...prev,
        {
          role: "assistant",
          content: res.answer,
          chunks: res.chunks,
          metrics: res.metrics,
          latency_ms: res.latency_ms,
        },
      ]);
      if (res.chunks?.length) onChunksChange(res.chunks);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Query failed");
    } finally {
      setLoading(false);
    }
  }

  const isEmpty = messages.length === 0;

  return (
    <div className="flex flex-col h-full relative">
      {/* Scrollable thread */}
      <div className="flex-1 overflow-y-auto">
        <div className="max-w-3xl mx-auto px-6 lg:px-10 py-10">
          {isEmpty ? (
            <div className="flex flex-col items-center justify-center min-h-[60vh] text-center animate-rise">
              <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint mb-3">
                Now Reading
              </p>
              <h1 className="font-display text-4xl lg:text-5xl text-ink font-light italic mb-3 leading-tight">
                {documentName}
              </h1>
              <p className="font-display text-lg text-ink-soft italic max-w-md">
                Ask anything. The answers will be drawn from these pages.
              </p>
              <div className="mt-10 flex flex-col gap-2 text-sm text-ink-soft items-center">
                {[
                  "What is the central argument?",
                  "Summarise the methodology.",
                  "Quote the most striking passage.",
                ].map((s) => (
                  <button
                    key={s}
                    onClick={() => setInput(s)}
                    className="text-left font-display italic text-base text-ink-soft hover:text-accent transition-colors"
                  >
                    &ldquo;{s}&rdquo;
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-10">
              {messages.map((msg, i) => (
                <div key={i} className="animate-msg">
                  {msg.role === "user" ? (
                    <UserBubble content={msg.content} />
                  ) : (
                    <AssistantBlock
                      content={msg.content}
                      chunks={msg.chunks}
                      metrics={msg.metrics}
                      latency={msg.latency_ms}
                      onCite={() => msg.chunks && onChunksChange(msg.chunks)}
                    />
                  )}
                </div>
              ))}
              {loading && (
                <div className="animate-msg">
                  <ThinkingIndicator />
                </div>
              )}
              {error && (
                <p className="text-sm font-mono text-accent-warm italic">! {error}</p>
              )}
            </div>
          )}
          <div ref={bottomRef} />
        </div>
      </div>

      {/* Floating composer */}
      <div className="border-t border-rule bg-paper/90 backdrop-blur sticky bottom-0">
        <div className="max-w-3xl mx-auto px-6 lg:px-10 py-5">
          <div className={`relative rounded-sm border bg-paper-soft transition-all ${
            input ? "border-ink/40 shadow-[0_4px_16px_-8px_rgba(26,36,33,0.15)]" : "border-rule"
          }`}>
            <textarea
              ref={textareaRef}
              rows={1}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  handleSend();
                }
              }}
              placeholder="Inscribe your question…"
              disabled={loading}
              className="w-full resize-none bg-transparent px-5 py-4 pr-16 text-base text-ink placeholder:text-ink-faint placeholder:italic placeholder:font-display focus:outline-none disabled:opacity-50"
            />
            <button
              onClick={handleSend}
              disabled={loading || !input.trim()}
              aria-label="Send"
              className="absolute right-3 bottom-3 w-9 h-9 rounded-sm bg-accent text-paper hover:bg-ink disabled:opacity-30 disabled:cursor-not-allowed transition-colors flex items-center justify-center"
            >
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M7 12V2M7 2L2.5 6.5M7 2l4.5 4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </button>
          </div>
          <div className="flex items-center justify-between mt-2 text-[10px] font-mono uppercase tracking-[0.2em] text-ink-faint">
            <span>↵ to send · ⇧↵ for newline</span>
            <span>{messages.length > 0 && `${messages.length} exchange${messages.length > 1 ? "s" : ""}`}</span>
          </div>
        </div>
      </div>
    </div>
  );
}

function UserBubble({ content }: { content: string }) {
  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-center gap-2">
        <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint">
          You ask
        </span>
        <div className="flex-1 h-px bg-rule" />
      </div>
      <p className="font-display text-xl lg:text-2xl text-ink italic leading-relaxed">
        {content}
      </p>
    </div>
  );
}

function AssistantBlock({
  content,
  chunks,
  metrics,
  latency,
  onCite,
}: {
  content: string;
  chunks?: Chunk[];
  metrics?: QueryMetrics;
  latency?: number;
  onCite: () => void;
}) {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-accent">
          The text replies
        </span>
        <div className="flex-1 h-px bg-rule" />
      </div>

      <div className="prose-editorial text-[15px] lg:text-base text-ink">
        <ReactMarkdown>{content}</ReactMarkdown>
      </div>

      {chunks && chunks.length > 0 && (
        <div className="mt-3 pt-3 border-t border-rule-soft">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint">
              Cited from {chunks.length} passage{chunks.length > 1 ? "s" : ""}
            </span>
          </div>
          <div className="flex flex-wrap gap-1.5">
            {chunks.map((c, i) => (
              <button
                key={c.chunk_id}
                onClick={onCite}
                className="group flex items-baseline gap-1.5 text-xs font-mono text-ink-soft hover:text-accent transition-colors px-2 py-1 border border-rule-soft hover:border-accent rounded-sm bg-paper-soft"
              >
                <sup className="text-accent-warm font-semibold">{i + 1}</sup>
                <span>§{c.chunk_index}</span>
                <span className="text-ink-faint">·</span>
                <span className="tabular-nums">{(c.score * 100).toFixed(0)}%</span>
              </button>
            ))}
          </div>
        </div>
      )}

      {metrics && (
        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 mt-1 text-[10px] font-mono text-ink-faint uppercase tracking-[0.15em]">
          <Stat label="MRR" value={metrics.mrr.toFixed(2)} />
          <Stat label="R@K" value={metrics.recall_at_k.toFixed(2)} />
          <Stat label="NDCG" value={metrics.ndcg.toFixed(2)} />
          <Stat label="μ" value={metrics.score_mean.toFixed(2)} />
          {latency !== undefined && <Stat label="t" value={`${latency}ms`} />}
        </div>
      )}
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <span className="flex items-baseline gap-1">
      <span className="text-ink-faint">{label}</span>
      <span className="text-ink-soft tabular-nums">{value}</span>
    </span>
  );
}

function ThinkingIndicator() {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-accent">
          Consulting the text
        </span>
        <div className="flex-1 h-px bg-rule" />
      </div>
      <div className="flex items-center gap-1 py-2">
        <span className="dot w-1.5 h-1.5 rounded-full bg-ink-faint" />
        <span className="dot w-1.5 h-1.5 rounded-full bg-ink-faint" />
        <span className="dot w-1.5 h-1.5 rounded-full bg-ink-faint" />
      </div>
    </div>
  );
}
