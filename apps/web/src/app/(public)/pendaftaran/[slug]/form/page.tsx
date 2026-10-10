import type { Metadata } from "next";
import { PublicRegistrationForm } from "@/features/registrations/public-registration";

export const metadata: Metadata = { title: "Formulir Pendaftaran" };

export default async function RegistrationFormPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <PublicRegistrationForm slug={slug} />;
}
