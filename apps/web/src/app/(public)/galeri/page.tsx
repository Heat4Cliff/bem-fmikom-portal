import type { Metadata } from "next";
import { getPublicGalleryCategory, getPublicGalleryPage } from "@/features/content/api";
import { ContentPagination, GalleryCard, PublicContentHeading } from "@/features/content/public-content";

export const metadata: Metadata = {
  title: "Galeri Kegiatan",
  description: "Dokumentasi kegiatan BEM FMIKOM UNUGHA.",
};

export default async function GalleryPage({ searchParams }: { searchParams: Promise<{ page?: string }> }) {
  const params = await searchParams;
  const page = Math.max(1, Number.parseInt(params.page ?? "1", 10) || 1);
  const result = await getPublicGalleryPage(page, 12);
  const categories = await Promise.all(
    result.items.map(async (category) => {
      try {
        return await getPublicGalleryCategory(category.slug);
      } catch {
        return category;
      }
    }),
  );
  return (
    <div>
      <PublicContentHeading
        eyebrow="Arsip dokumentasi kegiatan"
        title="Lensa & Galeri Aktivitas FMIKOM"
        description="Jelajahi dokumentasi kegiatan menurut kategori dan buka album resmi Google Photos."
      />
      <main className="px-4 py-8 md:px-8 lg:px-12 lg:py-12">
        <div className="mx-auto max-w-[var(--container-public)]">
          {categories.length ? (
            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
              {categories.map((category) => (
                <GalleryCard key={category.id} category={category} />
              ))}
            </div>
          ) : (
            <p className="rounded-xl border border-border bg-white p-8 text-sm text-muted">
              Dokumentasi kegiatan belum tersedia.
            </p>
          )}
          <ContentPagination
            pathname="/galeri"
            page={result.meta.page}
            totalPages={result.meta.total_pages}
          />
        </div>
      </main>
    </div>
  );
}
