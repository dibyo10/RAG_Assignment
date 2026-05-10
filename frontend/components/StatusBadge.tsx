export function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, { bg: string; dot: string; label: string }> = {
    pending:  { bg: "bg-paper-deep text-ink-soft",                dot: "bg-accent-gold",            label: "queued" },
    indexing: { bg: "bg-accent-gold/15 text-[#7a5e1f]",          dot: "bg-accent-gold animate-pulse", label: "indexing" },
    ready:    { bg: "bg-accent/10 text-accent",                  dot: "bg-accent",                  label: "ready" },
    failed:   { bg: "bg-accent-warm/10 text-accent-warm",        dot: "bg-accent-warm",            label: "failed" },
  };
  const s = styles[status] ?? styles.pending;
  return (
    <span className={`inline-flex items-center gap-1.5 text-[10px] uppercase tracking-[0.15em] font-mono px-2 py-0.5 rounded-sm ${s.bg}`}>
      <span className={`w-1 h-1 rounded-full ${s.dot}`} />
      {s.label}
    </span>
  );
}
