import type { Metadata } from "next";
import { AboutProfilePage } from "@/features/public/about-profile-page";
import { getPublicLandingSettings } from "@/features/content/api";
import { getPublicSiteOrigin } from "@/lib/seo/site-url";

export const dynamic = "force-dynamic";

const publicSiteOrigin = getPublicSiteOrigin();

export const metadata: Metadata = {
  title: "Tentang BEM FMIKOM UNUGHA",
  description: "Profil organisasi, identitas kabinet, visi, dan misi BEM FMIKOM UNUGHA.",
  alternates: publicSiteOrigin ? { canonical: new URL("/tentang", publicSiteOrigin).toString() } : undefined,
};

export default async function AboutPage() {
  const settings = await getPublicLandingSettings().catch(() => null);
  return <AboutProfilePage settings={settings} />;
}
