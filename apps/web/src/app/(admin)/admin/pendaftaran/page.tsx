import type { Metadata } from "next";
import { RegistrationAdminList } from "@/features/registrations/registration-admin";

export const metadata: Metadata = { title: "Manajemen Pendaftaran" };

export default function AdminRegistrationPage() {
  return <RegistrationAdminList />;
}
