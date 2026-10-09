import type { Metadata } from "next";
import { getLandingPageContent, isDemoContentEnabled } from "@/features/public/home/data/home-content";
import { PublicHome } from "@/features/public/home/public-home";
import { getPublicSiteOrigin } from "@/lib/seo/site-url";
import {
  getPublicArticles,
  getActiveAnnouncements,
  getPublicGallery,
  getPublicOrganization,
} from "@/features/content/api";
import type { Announcement, Article, GalleryCategory, PublicOrganization } from "@/features/content/types";

export const dynamic = "force-dynamic";

const publicSiteOrigin = getPublicSiteOrigin();

export const metadata: Metadata = {
  title: "Beranda",
  description:
    "Portal publik BEM FMIKOM UNUGHA untuk informasi organisasi, berita, kegiatan, pengumuman, dan layanan mahasiswa.",
  openGraph: {
    title: "BEM FMIKOM UNUGHA",
    description: "Portal informasi publik dan kegiatan BEM Fakultas Matematika dan Ilmu Komputer UNUGHA.",
    type: "website",
    locale: "id_ID",
  },
  robots: { index: true, follow: true },
  alternates: publicSiteOrigin ? { canonical: new URL("/", publicSiteOrigin).toString() } : undefined,
};

function plainText(html: string) {
  return html
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function formatDate(value?: string | null) {
  if (!value) return "";
  return new Date(value).toLocaleDateString("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "Asia/Jakarta",
  });
}

export default async function PublicHomePage() {
  const content = getLandingPageContent(isDemoContentEnabled());
  const [articleResult, announcementResult, galleryResult, organizationResult] = await Promise.allSettled([
    getPublicArticles(1, 3),
    getActiveAnnouncements(1, 1),
    getPublicGallery(1, 4),
    getPublicOrganization(),
  ]);
  const articles = articleResult.status === "fulfilled" ? articleResult.value : [];
  const announcements = announcementResult.status === "fulfilled" ? announcementResult.value : [];
  const galleries = galleryResult.status === "fulfilled" ? galleryResult.value : [];
  const organization: PublicOrganization | null =
    organizationResult.status === "fulfilled" ? organizationResult.value : null;
  if (organization) {
    const toPreview = (member: PublicOrganization["leadership"][number], division?: string) => ({
      id: member.id,
      name: member.name,
      role: member.position,
      nim: member.nim,
      photoUrl: member.photo_url,
      division,
    });
    content.organization = {
      period: organization.period,
      periodLabel: organization.period?.name,
      leadership: organization.leadership.map((member) => toPreview(member, "BPH")),
      divisions: organization.divisions.map((division) => ({
        id: division.id,
        name: division.name,
        head: division.head ? toPreview(division.head, division.name) : null,
        members: division.members.map((member) => toPreview(member, division.name)),
      })),
      members: [
        ...organization.leadership.map((member) => toPreview(member, "BPH")),
        ...organization.divisions.flatMap((division) => [
          ...(division.head ? [toPreview(division.head, division.name)] : []),
          ...division.members.map((member) => toPreview(member, division.name)),
        ]),
      ],
    };
  }
  content.news = articles.map((article: Article) => ({
    id: String(article.id),
    category: article.categories[0]?.name ?? "Berita",
    date: formatDate(article.published_at),
    title: article.title,
    excerpt: article.excerpt?.trim() || plainText(article.content_html).slice(0, 180),
    visualLabel: `Gambar berita: ${article.title}`,
    imageUrl: article.thumbnail_url ?? undefined,
    href: `/berita/${encodeURIComponent(article.slug)}`,
  }));
  content.announcements = announcements.map((announcement: Announcement) => ({
    id: String(announcement.id),
    title: announcement.title,
    content: plainText(announcement.content_html).slice(0, 220),
    image: announcement.image_url ? { src: announcement.image_url, alt: "" } : undefined,
    dateLabel: formatDate(announcement.start_at),
    href: `/pengumuman/${encodeURIComponent(announcement.slug)}`,
  }));
  content.gallery = galleries.map((category: GalleryCategory, index: number) => ({
    id: String(category.id),
    title: category.title,
    dateLabel: formatDate(category.event_date),
    division: category.division_name ?? undefined,
    visualLabel: `Galeri kegiatan: ${category.title}`,
    imageUrl: category.items?.[0]?.preview_image_url ?? undefined,
    href: `/galeri/${encodeURIComponent(category.slug)}`,
    tone: (["navy", "blue", "amber", "slate"] as const)[index % 4],
  }));
  return <PublicHome content={content} />;
}
