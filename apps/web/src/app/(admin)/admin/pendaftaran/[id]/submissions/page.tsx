import type { Metadata } from "next";
import { RegistrationSubmissionList } from "@/features/registrations/registration-admin";

export const metadata: Metadata = { title: "Data Pendaftar" };

export default async function RegistrationSubmissionsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <RegistrationSubmissionList id={Number(id)} />;
}
