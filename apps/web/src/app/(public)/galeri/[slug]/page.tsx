import { ArrowUpRight, CalendarDays } from "lucide-react";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MediaImage } from "@/components/ui/media-image";
import { getPublicGalleryCategory } from "@/features/content/api";
import { BackLink, formatDate, PublicContentHeading } from "@/features/content/public-content";
import type { GalleryCategory } from "@/features/content/types";

type Props = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  try {
    const { slug } = await params;
    const category = await getPublicGalleryCategory(slug);
    return { title: category.title, description: category.description ?? `Dokumentasi ${category.title}` };
  } catch {
    return { title: "Galeri tidak ditemukan" };
  }
}

export default async function GalleryDetailPage({ params }: Props) {
  const { slug } = await params;
  let category: GalleryCategory;
  try {
    category = await getPublicGalleryCategory(slug);
  } catch {
    notFound();
  }
  return (
    <div>
      <PublicContentHeading
        eyebrow="Dokumentasi kegiatan"
        title={category.title}
        description={category.description ?? "Dokumentasi kegiatan BEM FMIKOM UNUGHA."}
      />
      <main className="px-4 py-8 md:px-8 lg:px-12 lg:py-12">
        <div className="mx-auto max-w-[var(--container-public)]">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <BackLink href="/galeri">Semua galeri</BackLink>
            {category.event_date ? (
              <span className="inline-flex items-center gap-2 text-sm text-muted">
                <CalendarDays aria-hidden="true" className="size-4" />
                {formatDate(category.event_date)}
              </span>
            ) : null}
          </div>
          {category.items?.length ? (
            <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {category.items.map((item) => (
                <article
                  key={item.id}
                  className="overflow-hidden rounded-xl border border-border/70 bg-surface shadow-card"
                >
                  <div className="relative grid aspect-[4/3] place-items-center bg-surface-muted">
                    <MediaImage
                      value={item.preview_image_url}
                      alt={item.title}
                      fill
                      sizes="(max-width: 768px) 100vw, 33vw"
                      className="object-contain object-center"
                      fallbackClassName="absolute inset-0 grid place-items-center px-4 text-center text-sm text-muted"
                      fallbackText="Pratinjau foto belum tersedia"
                    />
                  </div>
                  <div className="p-4">
                    <h2 className="font-heading text-base font-bold text-primary">{item.title}</h2>
                    {item.description ? (
                      <p className="mt-2 text-sm leading-6 text-muted">{item.description}</p>
                    ) : null}
                    <a
                      href={item.google_photos_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="mt-3 inline-flex min-h-10 items-center gap-2 text-sm font-semibold text-primary"
                    >
                      Buka album Google Photos <ArrowUpRight aria-hidden="true" className="size-4" />
                    </a>
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <p className="mt-6 rounded-xl border border-border bg-surface p-8 text-sm text-muted">
              Belum ada item dokumentasi untuk kategori ini.
            </p>
          )}
        </div>
      </main>
    </div>
  );
}
