"use client";
import { use, useEffect, useState } from "react";
import Link from "next/link";
import { getDocumentMetrics, getDocument } from "@/lib/api";
import type { DayMetrics, Document } from "@/lib/types";
import { MetricsChart } from "@/components/MetricsChart";

export default function DocumentMetricsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [doc, setDoc] = useState<Document | null>(null);
  const [data, setData] = useState<DayMetrics[]>([]);

  useEffect(() => {
    getDocument(id).then((d) => setDoc(d.document)).catch(() => {});
    getDocumentMetrics(id).then(setData).catch(() => {});
  }, [id]);

  return (
    <div className="max-w-5xl mx-auto px-8 py-12 lg:py-16">
      <header className="mb-10 animate-rise">
        <Link
          href={`/notebooks/${id}`}
          className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint hover:text-ink transition-colors inline-flex items-center gap-1 group"
        >
          <span className="transition-transform group-hover:-translate-x-0.5">←</span>
          Return to text
        </Link>
        <div className="border-b border-ink pb-4 mt-4">
          <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint mb-2">
            Marginalia · Index
          </p>
          <h1 className="font-display text-4xl lg:text-5xl font-light text-ink leading-tight italic">
            {doc?.name ? doc.name.replace(/\.[^.]+$/, "") : id}
          </h1>
        </div>
        <p className="font-display italic text-lg text-ink-soft mt-5 max-w-xl leading-relaxed">
          Retrieval quality, measured per query, aggregated daily.
        </p>
      </header>

      <section className="bg-paper-soft border border-rule p-6 lg:p-8 mb-10 animate-rise" style={{ animationDelay: "0.1s" }}>
        <div className="flex items-baseline justify-between mb-6">
          <h2 className="font-display text-2xl text-ink italic">MRR · Recall@K · NDCG</h2>
          <span className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint">
            Daily aggregate
          </span>
        </div>
        <MetricsChart data={data} />
      </section>

      {data.length > 0 && (
        <section className="animate-rise" style={{ animationDelay: "0.2s" }}>
          <div className="flex items-center gap-4 mb-4">
            <h2 className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint">Ledger</h2>
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
