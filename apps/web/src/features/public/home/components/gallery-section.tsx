import { ArrowRight, Images } from "lucide-react";
import Link from "next/link";
import { MediaImage } from "@/components/ui/media-image";
import type { GalleryPreview } from "../types";
import { EmptyPreview, SectionHeading } from "./section-primitives";

const tones: Record<GalleryPreview["tone"], string> = {
  navy: "public-media-art--navy",
  blue: "public-media-art--blue",
  amber: "public-media-art--amber",
  slate: "public-media-art--slate",
};

export function GallerySection({ gallery, demoMode }: { gallery: GalleryPreview[]; demoMode: boolean }) {
  return (
    <section id="galeri" className="scroll-mt-24 bg-background px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow="Dokumentasi"
          title="Galeri Kegiatan Fakultas"
          description="Cuplikan dokumentasi kegiatan akan muncul setelah sumber galeri resmi tersedia."
          action={
            <Link
              href="/galeri"
              className="inline-flex min-h-10 items-center gap-2 rounded-md px-3 text-sm font-semibold text-primary hover:bg-surface-muted focus-visible:outline-2 focus-visible:outline-secondary"
            >
              Lihat galeri <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          }
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-amber-900">
            Contoh visual · dokumentasi bukan foto kegiatan
          </p>
        ) : null}
        {gallery.length ? (
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            {gallery.map((item) => (
              <article key={item.id}>
                <Link
                  href={item.href ?? "/galeri"}
                  role="img"
                  aria-label={item.visualLabel}
                  className={`relative grid aspect-[4/3] place-items-center overflow-hidden rounded-xl border border-border/50 ${tones[item.tone]}`}
                >
                  <MediaImage
                    value={item.imageUrl}
                    alt=""
                    fill
                    sizes="(max-width: 768px) 50vw, 25vw"
                    className="object-contain object-center"
                    fallbackClassName={`absolute inset-0 grid place-items-center ${tones[item.tone]}`}
                    fallback={<Images aria-hidden="true" className="size-9 text-white/85" />}
                  />
                </Link>
                <h3 className="mt-3 font-heading text-sm font-bold leading-snug text-primary sm:text-base">
                  {item.title}
                </h3>
                <p className="mt-1 text-xs text-muted">
                  {[item.dateLabel, item.division].filter(Boolean).join(" · ")}
                </p>
              </article>
            ))}
          </div>
        ) : (
          <EmptyPreview title="Dokumentasi belum tersedia">
            Galeri akan menampilkan cuplikan kegiatan ketika daftar dokumentasi resmi sudah dikelola.
          </EmptyPreview>
        )}
      </div>
    </section>
  );
}
