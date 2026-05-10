"use client";
import { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";
import { uploadDocument } from "@/lib/api";
import type { Document } from "@/lib/types";

interface Props {
  onUploaded: (doc: Document) => void;
}

export function DocumentUploader({ onUploaded }: Props) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onDrop = useCallback(
    async (files: File[]) => {
      if (!files[0]) return;
      setError(null);
      setUploading(true);
      try {
        const doc = await uploadDocument(files[0]);
        onUploaded(doc);
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "Upload failed");
      } finally {
        setUploading(false);
      }
    },
    [onUploaded]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { "application/pdf": [".pdf"], "text/plain": [".txt"], "text/markdown": [".md"] },
    maxFiles: 1,
    disabled: uploading,
  });

  return (
    <div
      {...getRootProps()}
      className={`relative border border-dashed rounded-sm cursor-pointer transition-all duration-300 ${
        isDragActive
          ? "border-accent bg-accent/5 scale-[1.005]"
          : "border-rule hover:border-ink-soft bg-paper-soft hover:bg-paper-deep/50"
      } ${uploading ? "opacity-60 cursor-wait" : ""}`}
    >
      <input {...getInputProps()} />
      <div className="px-10 py-12 text-center">
        {/* Decorative tick marks at corners */}
        <div className="absolute top-2 left-2 w-3 h-3 border-t border-l border-rule" />
        <div className="absolute top-2 right-2 w-3 h-3 border-t border-r border-rule" />
        <div className="absolute bottom-2 left-2 w-3 h-3 border-b border-l border-rule" />
        <div className="absolute bottom-2 right-2 w-3 h-3 border-b border-r border-rule" />

        {uploading ? (
          <div className="flex flex-col items-center gap-2">
            <span className="text-[10px] uppercase tracking-[0.3em] font-mono text-accent">
              Indexing manuscript
            </span>
            <div className="flex gap-1">
              <span className="dot w-1.5 h-1.5 rounded-full bg-accent" />
              <span className="dot w-1.5 h-1.5 rounded-full bg-accent" />
              <span className="dot w-1.5 h-1.5 rounded-full bg-accent" />
            </div>
          </div>
        ) : isDragActive ? (
          <p className="font-display text-xl text-accent italic">Release to deposit</p>
        ) : (
          <div className="flex flex-col items-center gap-1.5">
            <p className="font-display text-2xl text-ink italic">
              Deposit a manuscript
            </p>
            <p className="text-[10px] uppercase tracking-[0.3em] font-mono text-ink-faint">
              · drop or browse · pdf · txt · md ·
            </p>
          </div>
        )}
        {error && (
          <p className="mt-3 text-xs text-accent-warm font-mono">! {error}</p>
        )}
      </div>
    </div>
  );
}
