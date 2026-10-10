import type { Metadata } from "next";
import { RegistrationBuilder } from "@/features/registrations/registration-admin";

export const metadata: Metadata = { title: "Form Builder" };

export default async function RegistrationBuilderPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <RegistrationBuilder id={Number(id)} />;
}
