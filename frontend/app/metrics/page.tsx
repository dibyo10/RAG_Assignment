"use client";
import { useEffect, useState } from "react";
import { getGlobalMetrics } from "@/lib/api";
import type { DayMetrics } from "@/lib/types";
import { MetricsChart } from "@/components/MetricsChart";

export default function GlobalMetricsPage() {
  const [data, setData] = useState<DayMetrics[]>([]);

  useEffect(() => {
    getGlobalMetrics().then(setData).catch(() => {});
  }, []);

  const total = data.reduce((s, d) => s + d.query_count, 0);
  const avgMRR = data.length ? data.reduce((s, d) => s + d.avg_mrr, 0) / data.length : 0;
  const avgRecall = data.length ? data.reduce((s, d) => s + d.avg_recall, 0) / data.length : 0;
  const avgNDCG = data.length ? data.reduce((s, d) => s + d.avg_ndcg, 0) / data.length : 0;

  return (
    <div className="max-w-5xl mx-auto px-8 py-12 lg:py-16">
      {/* Header */}
      <header className="mb-10 animate-rise">
        <div className="flex items-baseline justify-between border-b border-ink pb-4">
          <div>
            <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint mb-2">
              The Index
            </p>
            <h1 className="font-display text-5xl lg:text-6xl font-light text-ink leading-none">
              Quality<span className="text-accent-warm">.</span>
            </h1>
          </div>
          <div className="text-right">
            <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint">
              {total.toString().padStart(4, "0")} retrievals
            </p>
          </div>
        </div>
        <p className="font-display italic text-lg text-ink-soft mt-5 max-w-xl leading-relaxed">
          Retrieval quality across the corpus. Higher is better, save for latency.
        </p>
      </header>

      {/* Stat cards — editorial newspaper style */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-px bg-rule mb-12 animate-rise" style={{ animationDelay: "0.1s" }}>
        <Stat label="Total queries" value={total.toString()} accent="ink" />
        <Stat label="Mean MRR" value={avgMRR.toFixed(3)} accent="accent" />
        <Stat label="Mean Recall@K" value={avgRecall.toFixed(3)} accent="warm" />
        <Stat label="Mean NDCG" value={avgNDCG.toFixed(3)} accent="gold" />
      </section>

      {/* Chart */}
      <section className="bg-paper-soft border border-rule p-6 lg:p-8 mb-10 animate-rise" style={{ animationDelay: "0.2s" }}>
        <div className="flex items-baseline justify-between mb-6">
          <h2 className="font-display text-2xl text-ink italic">Retrieval over time</h2>
          <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint">
            Daily aggregate
          </span>
        </div>
        <MetricsChart data={data} />
      </section>

      {/* Table */}
      {data.length > 0 && (
        <section className="animate-rise" style={{ animationDelay: "0.3s" }}>
          <div className="flex items-center gap-4 mb-4">
            <h2 className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint">
              Daily ledger
            </h2>
            <div className="flex-1 h-px bg-rule" />
          </div>
          <table className="w-full text-sm border border-rule bg-paper">
            <thead>
              <tr className="bg-paper-soft border-b border-rule">
                {["Day", "Queries", "MRR", "Recall@K", "NDCG", "μ score"].map((h) => (
                  <th key={h} className="px-4 py-3 text-left text-[10px] uppercase tracking-[0.2em] font-mono text-ink-faint font-normal">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.map((row, i) => (
                <tr key={i} className="border-b border-rule-soft last:border-0 hover:bg-paper-soft/50 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-ink-soft">{row.day}</td>
                  <td className="px-4 py-3 font-mono text-xs text-ink tabular-nums">{row.query_count}</td>
                  <td className="px-4 py-3 font-mono text-xs text-accent tabular-nums">{row.avg_mrr.toFixed(3)}</td>
                  <td className="px-4 py-3 font-mono text-xs text-accent-warm tabular-nums">{row.avg_recall.toFixed(3)}</td>
                  <td className="px-4 py-3 font-mono text-xs text-accent-gold tabular-nums">{row.avg_ndcg.toFixed(3)}</td>
                  <td className="px-4 py-3 font-mono text-xs text-ink tabular-nums">{row.avg_score.toFixed(3)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}
    </div>
  );
}

function Stat({ label, value, accent }: { label: string; value: string; accent: "ink" | "accent" | "warm" | "gold" }) {
  const colors = {
    ink: "text-ink",
    accent: "text-accent",
    warm: "text-accent-warm",
    gold: "text-accent-gold",
  };
  return (
    <div className="bg-paper p-6">
      <p className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint mb-3">
        {label}
      </p>
      <p className={`font-display text-4xl lg:text-5xl font-light tabular-nums ${colors[accent]}`}>
        {value}
      </p>
    </div>
  );
}
