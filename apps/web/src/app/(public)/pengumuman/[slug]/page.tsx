import type { Metadata } from "next";
import Image from "next/image";
import { notFound } from "next/navigation";
import { getPublicAnnouncement } from "@/features/content/api";
import { BackLink, formatDate, SanitizedRichText } from "@/features/content/public-content";
import type { Announcement } from "@/features/content/types";

type Props = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  try {
    const { slug } = await params;
    const item = await getPublicAnnouncement(slug);
    return {
      title: item.title,
      description: item.title,
      openGraph: { title: item.title, images: item.image_url ? [item.image_url] : [] },
    };
  } catch {
    return { title: "Pengumuman tidak ditemukan" };
  }
}

export default async function AnnouncementDetailPage({ params }: Props) {
  const { slug } = await params;
  let announcement: Announcement;
  try {
    announcement = await getPublicAnnouncement(slug);
  } catch {
    notFound();
  }
  return (
    <main className="px-4 py-8 md:px-8 lg:px-12 lg:py-12">
      <article className="mx-auto max-w-4xl">
        <BackLink href="/pengumuman">Arsip pengumuman</BackLink>
        <div className="mt-6 overflow-hidden rounded-2xl border border-border/70 bg-white shadow-card">
          {announcement.image_url ? (
            <div className="relative aspect-[16/7] max-h-[26rem] bg-surface-muted">
              <Image
                src={announcement.image_url}
                alt=""
                fill
                sizes="(max-width: 768px) 100vw, 896px"
                className="object-cover"
                unoptimized
              />
            </div>
          ) : null}
          <div className="p-5 sm:p-8 lg:p-10">
            <p className="inline-flex rounded-full bg-success/10 px-3 py-1 text-xs font-semibold text-success">
              {announcement.status === "ACTIVE"
                ? "Pengumuman aktif"
                : announcement.status === "SCHEDULED"
                  ? "Terjadwal"
                  : announcement.status === "EXPIRED"
                    ? "Arsip"
                    : "Tidak aktif"}
            </p>
            <h1 className="mt-4 font-heading text-3xl font-bold leading-tight text-primary md:text-5xl">
              {announcement.title}
            </h1>
            <p className="mt-4 text-sm text-muted">
              <time dateTime={announcement.start_at}>{formatDate(announcement.start_at)}</time>
              {announcement.end_at ? ` — ${formatDate(announcement.end_at)}` : ""}
            </p>
            <SanitizedRichText html={announcement.content_html} />
          </div>
        </div>
      </article>
    </main>
  );
}
