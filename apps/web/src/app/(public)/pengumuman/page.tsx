import type { Metadata } from "next";
import { getPublicAnnouncementPage } from "@/features/content/api";
import { AnnouncementCard, ContentPagination, PublicContentHeading } from "@/features/content/public-content";

export const metadata: Metadata = {
  title: "Pengumuman",
  description: "Arsip pengumuman resmi BEM FMIKOM UNUGHA.",
};

export default async function AnnouncementsPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; search?: string }>;
}) {
  const params = await searchParams;
  const page = Math.max(1, Number.parseInt(params.page ?? "1", 10) || 1);
  const search = (params.search ?? "").slice(0, 120);
  const result = await getPublicAnnouncementPage(page, 12, search);
  return (
    <div>
      <PublicContentHeading
        eyebrow="Informasi resmi"
        title="Pengumuman BEM FMIKOM"
        description="Informasi aktif, terjadwal, dan arsip pengumuman yang dipublikasikan organisasi."
      />
      <main className="px-4 py-8 md:px-8 lg:px-12 lg:py-12">
        <div className="mx-auto max-w-[var(--container-public)]">
          <form
            action="/pengumuman"
            className="mb-6 flex flex-col gap-3 rounded-xl border border-border/70 bg-surface p-4 sm:flex-row"
          >
            <label className="sr-only" htmlFor="announcement-search">
              Cari pengumuman
            </label>
            <input
              id="announcement-search"
              name="search"
              defaultValue={search}
              placeholder="Cari judul pengumuman..."
              className="min-h-11 min-w-0 flex-1 rounded-md border border-border bg-surface-muted px-4 text-sm"
            />
            <button
              type="submit"
              className="min-h-11 rounded-md bg-primary px-5 text-sm font-semibold text-primary-foreground"
            >
              Cari pengumuman
            </button>
          </form>
          {result.items.length ? (
            <div className="grid gap-4">
              {result.items.map((announcement) => (
                <AnnouncementCard key={announcement.id} announcement={announcement} />
              ))}
            </div>
          ) : (
            <p className="rounded-xl border border-border bg-surface p-8 text-sm text-muted">
              Belum ada pengumuman yang dipublikasikan.
            </p>
          )}
          <ContentPagination
            pathname="/pengumuman"
            page={result.meta.page}
            totalPages={result.meta.total_pages}
            search={search}
          />
        </div>
      </main>
    </div>
  );
}
