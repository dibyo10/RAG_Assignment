import type { Metadata } from "next";
import { Fraunces, IBM_Plex_Sans, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import Link from "next/link";

const fraunces = Fraunces({
  subsets: ["latin"],
  variable: "--font-display-stack",
  display: "swap",
  axes: ["opsz", "SOFT", "WONK"],
});

const plex = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600"],
  variable: "--font-sans-stack",
  display: "swap",
});

const mono = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-mono-stack",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Marginalia — Notebook RAG",
  description: "Read between the lines of your documents.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html
      lang="en"
      className={`${fraunces.variable} ${plex.variable} ${mono.variable} h-full`}
    >
      <body className="min-h-full flex flex-col relative">
        <div className="relative z-10 border-b border-rule bg-amber-50 text-amber-900 text-xs">
          <div className="max-w-6xl mx-auto px-8 py-2 text-center">
            Heads up: the backend runs on Render&rsquo;s free tier. The first request after a quiet spell may take ~1 minute to wake up.
          </div>
        </div>
        <header className="relative z-10 border-b border-rule bg-paper/60 backdrop-blur">
          <div className="max-w-6xl mx-auto px-8 py-4 flex items-baseline justify-between">
            <Link href="/notebooks" className="flex items-baseline gap-3">
              <span className="font-display text-2xl font-medium tracking-tight text-ink">
                Marginalia
              </span>
              <span className="text-[10px] uppercase tracking-[0.2em] text-ink-faint font-mono">
                · vol. i
              </span>
            </Link>

            <nav className="flex items-center gap-7 text-sm">
              <Link href="/notebooks" className="text-ink-soft hover:text-ink transition-colors relative group">
                Library
                <span className="absolute -bottom-1 left-0 w-0 h-px bg-accent transition-all group-hover:w-full" />
              </Link>
              <Link href="/metrics" className="text-ink-soft hover:text-ink transition-colors relative group">
                Index
                <span className="absolute -bottom-1 left-0 w-0 h-px bg-accent transition-all group-hover:w-full" />
              </Link>
            </nav>
          </div>
        </header>
        <main className="flex-1 relative z-10">{children}</main>
      </body>
    </html>
  );
}
