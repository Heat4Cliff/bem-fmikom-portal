import { ArrowLeft, ArrowRight, CalendarDays, Images, Newspaper } from "lucide-react";
import Link from "next/link";
import { MediaImage } from "@/components/ui/media-image";
import type { Announcement, Article, GalleryCategory } from "./types";

export function PublicContentHeading({
  eyebrow,
  title,
  description,
}: {
  eyebrow: string;
  title: string;
  description: string;
}) {
  return (
    <header className="border-b border-border/70 bg-primary-container/5 px-4 py-12 md:px-8 lg:px-12 lg:py-16">
      <div className="mx-auto max-w-[var(--container-public)]">
        <p className="mb-3 inline-flex rounded-full bg-primary-container/10 px-3 py-1 text-xs font-semibold text-primary">
          {eyebrow}
        </p>
        <h1 className="max-w-3xl font-heading text-3xl font-bold leading-tight text-primary md:text-5xl">
          {title}
        </h1>
        <p className="mt-4 max-w-2xl text-sm leading-7 text-muted md:text-base">{description}</p>
      </div>
    </header>
  );
}

export function ArticleCard({ article }: { article: Article }) {
  const summary =
    article.excerpt ||
    article.content_html
      .replace(/<[^>]*>/g, " ")
      .replace(/\s+/g, " ")
      .trim()
      .slice(0, 180);
  return (
    <article className="group overflow-hidden rounded-xl border border-border/70 bg-white shadow-card transition-shadow hover:shadow-raised">
      <Link
        href={`/berita/${encodeURIComponent(article.slug)}`}
        className="relative grid aspect-[16/9] place-items-center overflow-hidden bg-surface-muted"
        aria-label={`Baca ${article.title}`}
      >
        <MediaImage
          value={article.thumbnail_url}
          alt=""
          fill
          sizes="(max-width: 768px) 100vw, 33vw"
          className="object-contain object-center"
          fallbackClassName="absolute inset-0 grid place-items-center bg-primary-container/10 text-primary/70"
          fallback={<Newspaper aria-hidden="true" className="size-10" />}
        />
      </Link>
      <div className="p-5">
        <div className="flex flex-wrap items-center gap-2 text-xs text-muted">
          {article.categories.slice(0, 2).map((category) => (
            <span
              key={category.id}
              className="rounded-full bg-surface-muted px-2.5 py-1 font-semibold text-primary"
            >
              {category.name}
            </span>
          ))}
          {article.published_at ? (
            <time dateTime={article.published_at}>{formatDate(article.published_at)}</time>
          ) : null}
        </div>
        <h2 className="mt-3 font-heading text-lg font-bold leading-snug text-primary">
          <Link href={`/berita/${encodeURIComponent(article.slug)}`} className="hover:text-secondary">
            {article.title}
          </Link>
        </h2>
        <p className="mt-2 line-clamp-3 text-sm leading-6 text-muted">{summary}</p>
        <Link
          href={`/berita/${encodeURIComponent(article.slug)}`}
          className="mt-4 inline-flex min-h-10 items-center gap-2 text-sm font-semibold text-primary"
        >
          Baca selengkapnya <ArrowRight aria-hidden="true" className="size-4" />
        </Link>
      </div>
    </article>
  );
}

export function AnnouncementCard({ announcement }: { announcement: Announcement }) {
  return (
    <article className="grid gap-5 rounded-xl border border-border/70 bg-white p-5 shadow-card sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center sm:p-6">
      <div>
        <p className="flex items-center gap-2 text-xs font-semibold text-muted">
          <CalendarDays aria-hidden="true" className="size-4 text-secondary" />
          <time dateTime={announcement.start_at}>{formatDate(announcement.start_at)}</time>
        </p>
        <h2 className="mt-2 font-heading text-xl font-bold text-primary">
          <Link
            href={`/pengumuman/${encodeURIComponent(announcement.slug)}`}
            className="hover:text-secondary"
          >
            {announcement.title}
          </Link>
        </h2>
        <p className="mt-2 line-clamp-2 text-sm leading-6 text-muted">
          {announcement.content_html
            .replace(/<[^>]*>/g, " ")
            .replace(/\s+/g, " ")
            .trim()}
        </p>
      </div>
      <Link
        href={`/pengumuman/${encodeURIComponent(announcement.slug)}`}
        className="inline-flex min-h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-semibold text-white hover:bg-primary/90"
      >
        Baca pengumuman <ArrowRight aria-hidden="true" className="size-4" />
      </Link>
    </article>
  );
}

export function GalleryCard({ category }: { category: GalleryCategory }) {
  const cover = category.items?.find((item) => item.preview_image_url)?.preview_image_url;
  return (
    <article className="overflow-hidden rounded-xl border border-border/70 bg-white shadow-card">
      <Link
        href={`/galeri/${encodeURIComponent(category.slug)}`}
        className="relative grid aspect-[4/3] place-items-center overflow-hidden bg-surface-muted"
        aria-label={`Lihat galeri ${category.title}`}
      >
        <MediaImage
          value={cover}
          alt=""
          fill
          sizes="(max-width: 768px) 100vw, 33vw"
          className="object-contain object-center"
          fallbackClassName="absolute inset-0 grid place-items-center bg-primary-container/10 text-primary/70"
          fallback={<Images aria-hidden="true" className="size-10" />}
        />
      </Link>
      <div className="p-5">
        <p className="text-xs text-muted">
          {[category.event_date ? formatDate(category.event_date) : "", category.division_name ?? ""]
            .filter(Boolean)
            .join(" · ")}
        </p>
        <h2 className="mt-2 font-heading text-lg font-bold text-primary">
          <Link href={`/galeri/${encodeURIComponent(category.slug)}`}>{category.title}</Link>
        </h2>
        {category.description ? (
          <p className="mt-2 line-clamp-2 text-sm leading-6 text-muted">{category.description}</p>
        ) : null}
        <p className="mt-3 text-xs font-semibold text-muted">
          {category.item_count ?? category.items?.length ?? 0} item dokumentasi
        </p>
      </div>
    </article>
  );
}

export function ContentPagination({
  pathname,
  page,
  totalPages,
  search = "",
}: {
  pathname: string;
  page: number;
  totalPages: number;
  search?: string;
}) {
  if (totalPages < 2) return null;
  const href = (next: number) =>
    `${pathname}?page=${next}${search ? `&search=${encodeURIComponent(search)}` : ""}`;
  return (
    <nav
      aria-label="Halaman daftar"
      className="mt-8 flex items-center justify-between rounded-lg border border-border bg-white p-3 text-sm"
    >
      {page > 1 ? (
        <Link
          className="inline-flex min-h-10 items-center gap-2 px-2 font-semibold text-primary"
          href={href(page - 1)}
        >
          <ArrowLeft aria-hidden="true" className="size-4" /> Sebelumnya
        </Link>
      ) : (
        <span className="px-2 text-muted">Halaman pertama</span>
      )}
      <span className="text-muted">
        Halaman {page} dari {totalPages}
      </span>
      {page < totalPages ? (
        <Link
          className="inline-flex min-h-10 items-center gap-2 px-2 font-semibold text-primary"
          href={href(page + 1)}
        >
          Selanjutnya <ArrowRight aria-hidden="true" className="size-4" />
        </Link>
      ) : (
        <span className="px-2 text-muted">Halaman terakhir</span>
      )}
    </nav>
  );
}

export function SanitizedRichText({ html }: { html: string }) {
  // API rich text is stored and returned only after backend allowlist sanitization.
  // biome-ignore lint/security/noDangerouslySetInnerHtml: sanitized API HTML uses the documented formatting allowlist
  return <div className="content-rich-text" dangerouslySetInnerHTML={{ __html: html }} />;
}

export function BackLink({ href, children }: { href: string; children: string }) {
  return (
    <Link href={href} className="inline-flex min-h-10 items-center gap-2 text-sm font-semibold text-primary">
      <ArrowLeft aria-hidden="true" className="size-4" />
      {children}
    </Link>
  );
}

export function formatDate(value: string) {
  return new Date(value).toLocaleDateString("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "Asia/Jakarta",
  });
}
