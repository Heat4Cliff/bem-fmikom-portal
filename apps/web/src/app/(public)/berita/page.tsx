import type { Metadata } from "next";
import { getPublicArticlePage } from "@/features/content/api";
import { ArticleCard, ContentPagination, PublicContentHeading } from "@/features/content/public-content";

export const metadata: Metadata = {
  title: "Berita & Informasi",
  description: "Berita dan informasi resmi BEM FMIKOM UNUGHA.",
};

export default async function ArticlesPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; search?: string }>;
}) {
  const params = await searchParams;
  const page = Math.max(1, Number.parseInt(params.page ?? "1", 10) || 1);
  const search = (params.search ?? "").slice(0, 120);
  const result = await getPublicArticlePage(page, 12, search);
  return (
    <div>
      <PublicContentHeading
        eyebrow="Warta resmi civitas akademika FMIKOM UNUGHA"
        title="Warta, Kiprah & Agenda Terkini FMIKOM"
        description="Liputan kegiatan, program kerja, dan informasi resmi BEM FMIKOM UNUGHA."
      />
      <main className="px-4 py-8 md:px-8 lg:px-12 lg:py-12">
        <div className="mx-auto max-w-[var(--container-public)]">
          <form
            action="/berita"
            className="mb-6 flex flex-col gap-3 rounded-xl border border-border/70 bg-white p-4 sm:flex-row"
          >
            <label className="sr-only" htmlFor="article-search">
              Cari berita
            </label>
            <input
              id="article-search"
              name="search"
              defaultValue={search}
              placeholder="Cari judul berita..."
              className="min-h-11 min-w-0 flex-1 rounded-md border border-border bg-surface-muted px-4 text-sm"
            />
            <button
              type="submit"
              className="min-h-11 rounded-md bg-primary px-5 text-sm font-semibold text-white"
            >
              Cari berita
            </button>
          </form>
          {result.items.length ? (
            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
              {result.items.map((article) => (
                <ArticleCard key={article.id} article={article} />
              ))}
            </div>
          ) : (
            <p className="rounded-xl border border-border bg-white p-8 text-sm text-muted">
              Belum ada berita terpublikasi yang sesuai.
            </p>
          )}
          <ContentPagination
            pathname="/berita"
            page={result.meta.page}
            totalPages={result.meta.total_pages}
            search={search}
          />
        </div>
      </main>
    </div>
  );
}
