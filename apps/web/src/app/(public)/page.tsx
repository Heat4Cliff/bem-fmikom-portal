import type { Metadata } from "next";
import { getLandingPageContent, isDemoContentEnabled } from "@/features/public/home/data/home-content";
import { PublicHome } from "@/features/public/home/public-home";
import { getPublicSiteOrigin } from "@/lib/seo/site-url";

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

export default function PublicHomePage() {
  const content = getLandingPageContent(isDemoContentEnabled());
  return <PublicHome content={content} />;
}
