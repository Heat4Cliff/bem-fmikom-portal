import type { Metadata } from "next";
import { PublicRegistrationDetail } from "@/features/registrations/public-registration";

export const metadata: Metadata = { title: "Detail Pendaftaran" };

export default async function RegistrationDetailPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <PublicRegistrationDetail slug={slug} />;
}
