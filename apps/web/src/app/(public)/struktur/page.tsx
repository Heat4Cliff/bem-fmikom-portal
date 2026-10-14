import type { Metadata } from "next";
import { PublicOrganizationPage } from "@/features/organization/public-organization-page";
import { getPublicOrganization } from "@/features/content/api";
import { getPublicSiteOrigin } from "@/lib/seo/site-url";

export const dynamic = "force-dynamic";

const publicSiteOrigin = getPublicSiteOrigin();

export const metadata: Metadata = {
  title: "Struktur Organisasi",
  description: "Susunan pengurus BEM FMIKOM UNUGHA pada periode aktif.",
  alternates: publicSiteOrigin ? { canonical: new URL("/struktur", publicSiteOrigin).toString() } : undefined,
};

export default async function OrganizationPage() {
  try {
    const organization = await getPublicOrganization();
    return <PublicOrganizationPage organization={organization} />;
  } catch {
    return <PublicOrganizationPage organization={null} loadError />;
  }
}
