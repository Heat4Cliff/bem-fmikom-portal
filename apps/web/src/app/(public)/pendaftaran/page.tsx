import type { Metadata } from "next";
import { PublicRegistrationList } from "@/features/registrations/public-registration";

export const metadata: Metadata = { title: "Pendaftaran" };

export default function RegistrationPage() {
  return <PublicRegistrationList />;
}
