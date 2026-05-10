"use client";
import { useState, useEffect, useCallback, useRef } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { listDocuments, deleteDocument } from "@/lib/api";
import type { Document } from "@/lib/types";
import { DocumentUploader } from "@/components/DocumentUploader";
import { StatusBadge } from "@/components/StatusBadge";

export default function NotebooksPage() {
  const router = useRouter();
  const [docs, setDocs] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [autoOpenId, setAutoOpenId] = useState<string | null>(null);
  const autoOpenTriggered = useRef(false);

  const load = useCallback(async () => {
    try {
      const data = await listDocuments();
      setDocs(data ?? []);
    } catch {}
    setLoading(false);
  }, []);

  useEffect(() => { load(); }, [load]);

  // Poll while any doc is still indexing — fast (500ms) so the UX feels instant
  useEffect(() => {
    const needsPoll = docs.some((d) => d.status === "pending" || d.status === "indexing");
    if (!needsPoll) return;
    const t = setInterval(load, 500);
    return () => clearInterval(t);
  }, [docs, load]);

  // Auto-redirect when the freshly-uploaded doc becomes ready
  useEffect(() => {
    if (!autoOpenId || autoOpenTriggered.current) return;
    const doc = docs.find((d) => d.id === autoOpenId);
    if (doc?.status === "ready") {
      autoOpenTriggered.current = true;
      router.push(`/notebooks/${doc.id}`);
    }
  }, [docs, autoOpenId, router]);

  const handleUploaded = (doc: Document) => {
    autoOpenTriggered.current = false;
    setAutoOpenId(doc.id);
    setDocs((prev) => {
      const existing = prev.find((d) => d.id === doc.id);
      return existing ? prev : [doc, ...prev];
    });
    // Immediately fetch fresh state — embedding usually finishes in <1s
    setTimeout(load, 300);
    setTimeout(load, 800);
    setTimeout(load, 1500);
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Remove this manuscript and all its sessions?")) return;
    setDocs((prev) => prev.filter((d) => d.id !== id));
    deleteDocument(id).catch(() => {});
  };

  const formatDate = (ms: number) => {
    const d = new Date(ms);
    if (isNaN(d.getTime())) return "—";
    return d.toLocaleDateString("en-US", { month: "long", day: "numeric", year: "numeric" });
  };

  // Track the doc currently being indexed for the in-progress banner
  const inFlight = autoOpenId ? docs.find((d) => d.id === autoOpenId) : null;
  const inFlightActive =
    inFlight && (inFlight.status === "pending" || inFlight.status === "indexing");

  return (
    <div className="max-w-5xl mx-auto px-8 py-12 lg:py-16">
      {/* Editorial header */}
      <header className="mb-10 animate-rise">
        <div className="flex items-baseline justify-between border-b border-ink pb-4">
          <div>
            <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint mb-2">
              The Reading Room
            </p>
            <h1 className="font-display text-5xl lg:text-6xl font-light text-ink leading-none">
              Library<span className="text-accent-warm">.</span>
            </h1>
          </div>
          <div className="text-right">
            <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint">
              Volume {docs.length.toString().padStart(2, "0")}
            </p>
            <p className="text-xs font-mono text-ink-soft mt-1">
              {new Date().toLocaleDateString("en-US", { weekday: "long" })}
            </p>
          </div>
        </div>
        <p className="font-display italic text-lg text-ink-soft mt-5 max-w-xl leading-relaxed">
          A private archive of documents, queryable through retrieval-augmented dialogue.
        </p>
      </header>

      <div className="animate-rise" style={{ animationDelay: "0.1s" }}>
        <DocumentUploader onUploaded={handleUploaded} />
      </div>

      {/* In-flight indexing banner with redirect promise */}
      {inFlightActive && (
        <div className="mt-6 border border-accent/30 bg-accent/5 px-5 py-4 flex items-center gap-4 animate-rise">
          <div className="flex gap-1">
            <span className="dot w-1.5 h-1.5 rounded-full bg-accent" />
            <span className="dot w-1.5 h-1.5 rounded-full bg-accent" />
            <span className="dot w-1.5 h-1.5 rounded-full bg-accent" />
          </div>
          <div className="flex-1">
            <p className="font-display italic text-base text-ink leading-tight">
              Reading <em className="not-italic font-medium">{(inFlight!.name ?? "manuscript").replace(/\.[^.]+$/, "")}</em>
            </p>
            <p className="text-[10px] uppercase tracking-[0.25em] font-mono text-ink-faint mt-1">
              Chunking · embedding · indexing — opens automatically
            </p>
          </div>
        </div>
      )}

      {/* Documents list — newspaper-style entries */}
      <section className="mt-14">
        <div className="flex items-center gap-4 mb-8">
          <h2 className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint">
            Catalogued · {docs.length}
          </h2>
          <div className="flex-1 h-px bg-rule" />
        </div>

        {loading && (
          <p className="text-sm font-mono text-ink-faint italic">Consulting the index…</p>
        )}
        {!loading && docs.length === 0 && (
          <p className="text-sm font-display italic text-ink-faint">
            No manuscripts on file. Begin by depositing one above.
          </p>
        )}

        <div className="divide-y divide-rule">
          {docs.map((doc, i) => (
            <article
              key={doc.id ?? `idx-${i}`}
              className="group py-7 flex items-baseline justify-between gap-6 animate-rise hover:bg-paper-soft/50 -mx-4 px-4 transition-colors duration-300 rounded-sm"
              style={{ animationDelay: `${0.1 + i * 0.05}s` }}
            >
              {/* Index number */}
              <div className="font-mono text-xs text-ink-faint pt-1 w-12 shrink-0 tabular-nums">
                №{(i + 1).toString().padStart(3, "0")}
              </div>

              {/* Title block — clickable when ready */}
              <div className="flex-1 min-w-0">
                {doc.status === "ready" ? (
                  <Link href={`/notebooks/${doc.id}`} className="block group/title">
                    <div className="flex items-center gap-3 flex-wrap mb-1.5">
                      <h3 className="font-display text-2xl text-ink truncate font-medium group-hover/title:text-accent transition-colors">
                        {(doc.name ?? "Untitled").replace(/\.[^.]+$/, "")}
                      </h3>
                      <StatusBadge status={doc.status} />
                    </div>
                  </Link>
                ) : (
                  <div className="flex items-center gap-3 flex-wrap mb-1.5">
                    <h3 className="font-display text-2xl text-ink truncate font-medium">
                      {(doc.name ?? "Untitled").replace(/\.[^.]+$/, "")}
                    </h3>
                    <StatusBadge status={doc.status} />
                  </div>
                )}
                <div className="flex items-center gap-3 text-xs font-mono text-ink-faint">
                  <span>{formatDate(doc.created_at)}</span>
                  <span>·</span>
                  <span className="uppercase tracking-wider">{doc.mime_type?.split("/")[1] ?? "doc"}</span>
                </div>
                {doc.error_msg && (
                  <p className="text-xs text-accent-warm font-mono mt-2 italic">
                    {doc.error_msg}
                  </p>
                )}
              </div>

              {/* Actions */}
              <div className="flex items-center gap-5 shrink-0">
                {doc.status === "ready" && (
                  <Link
                    href={`/notebooks/${doc.id}`}
                    className="group/link relative font-display italic text-base text-accent hover:text-ink transition-colors"
                  >
                    Open
                    <span className="ml-1 inline-block transition-transform group-hover/link:translate-x-1">→</span>
                  </Link>
                )}
                {doc.id && (
                  <button
                    onClick={() => handleDelete(doc.id)}
                    className="text-xs font-mono uppercase tracking-wider text-ink-faint hover:text-accent-warm transition-colors opacity-0 group-hover:opacity-100"
                  >
                    Remove
                  </button>
                )}
              </div>
            </article>
          ))}
        </div>
      </section>

      {/* Footer flourish */}
      <footer className="mt-20 pt-8 border-t border-rule flex items-center justify-between text-xs font-mono text-ink-faint">
        <span>Set in Fraunces & IBM Plex.</span>
        <span className="italic font-display">Marginalia</span>
      </footer>
    </div>
  );
}
