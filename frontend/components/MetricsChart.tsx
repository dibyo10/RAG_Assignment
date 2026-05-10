"use client";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer,
} from "recharts";
import type { DayMetrics } from "@/lib/types";

interface Props {
  data: DayMetrics[];
}

export function MetricsChart({ data }: Props) {
  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 border border-dashed border-rule">
        <p className="text-sm font-display italic text-ink-faint">
          The index awaits its first inscription.
        </p>
      </div>
    );
  }
  return (
    <ResponsiveContainer width="100%" height={320}>
      <LineChart data={data} margin={{ top: 10, right: 24, bottom: 5, left: 0 }}>
        <CartesianGrid strokeDasharray="2 4" stroke="#c9bfa9" opacity={0.4} vertical={false} />
        <XAxis
          dataKey="day"
          tick={{ fontSize: 10, fontFamily: "var(--font-mono-stack)", fill: "#6b7770" }}
          axisLine={{ stroke: "#c9bfa9" }}
          tickLine={false}
        />
        <YAxis
          domain={[0, 1]}
          tick={{ fontSize: 10, fontFamily: "var(--font-mono-stack)", fill: "#6b7770" }}
          axisLine={false}
          tickLine={false}
        />
        <Tooltip
          contentStyle={{
            background: "#1a2421",
            border: "none",
            borderRadius: 0,
            color: "#f4efe6",
            fontFamily: "var(--font-mono-stack)",
            fontSize: 11,
          }}
          itemStyle={{ color: "#f4efe6" }}
          labelStyle={{ color: "#c9bfa9", fontSize: 10, textTransform: "uppercase", letterSpacing: "0.15em" }}
          formatter={(v) => (typeof v === "number" ? v.toFixed(3) : v)}
        />
        <Legend
          iconType="plainline"
          wrapperStyle={{ fontFamily: "var(--font-mono-stack)", fontSize: 11, color: "#3d4a45", paddingTop: 12 }}
        />
        <Line type="monotone" dataKey="avg_mrr"    name="MRR"      stroke="#2d4a3e" strokeWidth={1.5} dot={{ r: 2.5, fill: "#2d4a3e" }} activeDot={{ r: 4 }} />
        <Line type="monotone" dataKey="avg_recall" name="Recall@K" stroke="#b8541f" strokeWidth={1.5} dot={{ r: 2.5, fill: "#b8541f" }} activeDot={{ r: 4 }} />
        <Line type="monotone" dataKey="avg_ndcg"   name="NDCG"     stroke="#b89548" strokeWidth={1.5} dot={{ r: 2.5, fill: "#b89548" }} activeDot={{ r: 4 }} />
        <Line type="monotone" dataKey="avg_score"  name="μ score"  stroke="#1a2421" strokeWidth={1.5} strokeDasharray="3 3" dot={false} />
      </LineChart>
    </ResponsiveContainer>
  );
}
