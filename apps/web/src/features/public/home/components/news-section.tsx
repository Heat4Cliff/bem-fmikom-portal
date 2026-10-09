import { ArrowRight, Newspaper } from "lucide-react";
import Link from "next/link";
import type { NewsPreview } from "../types";
import { EmptyPreview, SectionHeading } from "./section-primitives";

export function NewsSection({ news, demoMode }: { news: NewsPreview[]; demoMode: boolean }) {
  const visualTones = ["public-media-art--blue", "public-media-art--amber", "public-media-art--slate"];

  return (
    <section id="berita" className="scroll-mt-24 bg-surface-muted px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow="Kabar fakultas"
          title="Berita & Aktivitas Mahasiswa"
          description="Informasi dan cerita kegiatan BEM FMIKOM UNUGHA akan hadir di sini."
          action={
            <Link
              href="#berita"
              className="inline-flex min-h-10 items-center gap-2 rounded-md px-3 text-sm font-semibold text-primary hover:bg-white focus-visible:outline-2 focus-visible:outline-secondary"
            >
              Semua berita <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          }
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-amber-900">
            Contoh visual · kartu berita bukan publikasi resmi
          </p>
        ) : null}
        {news.length ? (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {news.map((item, index) => (
              <article
                key={item.id}
                className="group overflow-hidden rounded-xl border border-border/70 bg-white shadow-card transition-shadow hover:shadow-raised"
              >
                <div
                  role="img"
                  aria-label={item.visualLabel}
                  className={`grid aspect-[16/10] place-items-center ${visualTones[index % visualTones.length]}`}
                >
                  <Newspaper aria-hidden="true" className="size-10 text-primary/70" />
                </div>
                <div className="p-5">
                  <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-semibold text-muted">
                    <span className="rounded-full bg-surface-muted px-2.5 py-1 text-primary">
                      {item.category}
                    </span>
                    <time>{item.date}</time>
                  </div>
                  <h3 className="mt-3 font-heading text-lg font-bold leading-snug text-primary">
                    <Link
                      className="rounded-sm after:absolute after:inset-0 focus-visible:outline-2 focus-visible:outline-secondary"
                      href={item.href}
                    >
                      {item.title}
                    </Link>
                  </h3>
                  <p className="mt-2 line-clamp-3 text-sm leading-6 text-muted">{item.excerpt}</p>
                  <Link
                    href={item.href}
                    aria-label={`Baca ${item.title}`}
                    className="relative z-10 mt-4 inline-flex min-h-10 items-center gap-2 text-sm font-semibold text-primary hover:text-secondary focus-visible:outline-2 focus-visible:outline-secondary"
                  >
                    Baca pratinjau <ArrowRight aria-hidden="true" className="size-4" />
                  </Link>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <EmptyPreview title="Belum ada berita untuk ditampilkan">
            Berita akan muncul setelah konten resmi tersedia. Tidak ada data contoh yang ditampilkan pada mode
            produksi.
          </EmptyPreview>
        )}
      </div>
    </section>
  );
}
