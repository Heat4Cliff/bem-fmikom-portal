import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MediaImage } from "@/components/ui/media-image";
import { getPublicArticle } from "@/features/content/api";
import { BackLink, formatDate, SanitizedRichText } from "@/features/content/public-content";
import type { Article } from "@/features/content/types";

type Props = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  try {
    const { slug } = await params;
    const article = await getPublicArticle(slug);
    return {
      title: article.title,
      description: article.excerpt ?? article.title,
      openGraph: {
        title: article.title,
        description: article.excerpt ?? article.title,
        images: article.thumbnail_url ? [article.thumbnail_url] : [],
      },
    };
  } catch {
    return { title: "Berita tidak ditemukan" };
  }
}

export default async function ArticleDetailPage({ params }: Props) {
  const { slug } = await params;
  let article: Article;
  try {
    article = await getPublicArticle(slug);
  } catch {
    notFound();
  }
  return (
    <main className="px-4 py-8 md:px-8 lg:px-12 lg:py-12">
      <article className="mx-auto max-w-4xl">
        <BackLink href="/berita">Daftar berita</BackLink>
        <div className="mt-6 rounded-2xl border border-border/70 bg-surface p-5 shadow-card sm:p-8 lg:p-10">
          <div className="flex flex-wrap gap-2">
            {article.categories.map((category) => (
              <span
                key={category.id}
                className="rounded-full bg-primary-container/10 px-3 py-1 text-xs font-semibold text-primary"
              >
                {category.name}
              </span>
            ))}
          </div>
          <h1 className="mt-4 font-heading text-3xl font-bold leading-tight text-primary md:text-5xl">
            {article.title}
          </h1>
          <p className="mt-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted">
            <time dateTime={article.published_at ?? article.created_at}>
              {formatDate(article.published_at ?? article.created_at)}
            </time>
            {article.division_name ? <span>· {article.division_name}</span> : null}
          </p>
          {article.thumbnail_url ? (
            <div className="relative mt-7 aspect-[16/9] overflow-hidden rounded-xl bg-surface-muted">
              <MediaImage
                value={article.thumbnail_url}
                alt=""
                fill
                sizes="(max-width: 768px) 100vw, 896px"
                className="object-contain object-center"
                fallbackClassName="absolute inset-0 grid place-items-center bg-surface-muted"
              />
            </div>
          ) : null}
          {article.excerpt ? (
            <p className="mt-7 border-l-4 border-secondary bg-surface-muted px-5 py-4 text-base leading-7 text-body">
              {article.excerpt}
            </p>
          ) : null}
          <SanitizedRichText html={article.content_html} />
        </div>
      </article>
    </main>
  );
}
