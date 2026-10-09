import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft, MessageSquareHeart } from "lucide-react";

export const metadata: Metadata = {
  title: "Aspirasi",
  description: "Informasi kesiapan layanan aspirasi BEM FMIKOM UNUGHA.",
  robots: { index: false, follow: false },
};

export default function AspirationPlaceholderPage() {
  return (
    <section className="mx-auto grid min-h-[65vh] w-full max-w-[var(--container-public)] content-center px-4 py-16 md:px-8 lg:px-12">
      <div className="mx-auto max-w-xl rounded-2xl border border-border bg-white p-6 text-center shadow-card sm:p-10">
        <span className="mx-auto grid size-12 place-items-center rounded-xl bg-surface-muted text-primary">
          <MessageSquareHeart aria-hidden="true" className="size-6" />
        </span>
        <p className="mt-5 text-xs font-bold uppercase tracking-[0.16em] text-secondary">Layanan aspirasi</p>
        <h1 className="mt-2 font-heading text-2xl font-bold text-primary sm:text-3xl">
          Formulir aspirasi belum tersedia
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          Halaman ini adalah placeholder navigasi Phase 3. Belum ada formulir, pengiriman, atau penyimpanan
          aspirasi.
        </p>
        <Link
          href="/"
          className="mt-6 inline-flex min-h-11 items-center gap-2 rounded-md bg-primary px-4 text-sm font-semibold text-white hover:bg-primary/90 focus-visible:outline-2 focus-visible:outline-secondary"
        >
          <ArrowLeft aria-hidden="true" className="size-4" /> Kembali ke beranda
        </Link>
      </div>
    </section>
  );
}
