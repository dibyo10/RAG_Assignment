"use client";
import type { Chunk } from "@/lib/types";

interface Props {
  chunks: Chunk[];
  selectedChunkId: string | null;
  onSelect: (id: string) => void;
}

function scoreColor(score: number) {
  if (score >= 0.8) return "text-accent";
  if (score >= 0.6) return "text-accent-gold";
  return "text-accent-warm";
}

export function ChunkViewer({ chunks, selectedChunkId, onSelect }: Props) {
  if (!chunks.length) {
    return (
      <p className="text-xs font-display italic text-ink-faint p-4">
        Passages will appear here once you ask a question.
      </p>
    );
  }
  return (
    <div className="flex flex-col">
      {chunks.map((chunk, i) => (
        <button
          key={chunk.chunk_id}
          onClick={() => onSelect(chunk.chunk_id)}
          className={`text-left p-4 border-b border-rule-soft last:border-0 transition-colors group ${
            selectedChunkId === chunk.chunk_id
              ? "bg-paper-deep"
              : "hover:bg-paper-soft"
          }`}
        >
          <div className="flex items-baseline justify-between mb-1.5">
            <div className="flex items-baseline gap-2">
              <sup className="font-mono text-[11px] text-accent-warm font-semibold">{i + 1}</sup>
              <span className="text-[10px] uppercase tracking-[0.2em] font-mono text-ink-faint">
                §{chunk.chunk_index}
              </span>
            </div>
            <span className={`text-xs font-mono tabular-nums ${scoreColor(chunk.score)}`}>
              {(chunk.score * 100).toFixed(1)}
            </span>
          </div>
          <p className="text-xs text-ink-soft leading-relaxed line-clamp-4 font-light">
            {chunk.text}
          </p>
          <p className="text-[10px] font-mono text-ink-faint mt-2 tabular-nums">
            chars {chunk.start_char}–{chunk.end_char}
          </p>
        </button>
      ))}
    </div>
  );
}
