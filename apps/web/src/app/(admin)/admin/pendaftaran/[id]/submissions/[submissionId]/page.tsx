import type { Metadata } from "next";
import { RegistrationSubmissionDetail } from "@/features/registrations/registration-admin";

export const metadata: Metadata = { title: "Detail Pendaftar" };

export default async function RegistrationSubmissionDetailPage({
  params,
}: {
  params: Promise<{ id: string; submissionId: string }>;
}) {
  const { id, submissionId } = await params;
  return <RegistrationSubmissionDetail id={Number(id)} submissionId={Number(submissionId)} />;
}
