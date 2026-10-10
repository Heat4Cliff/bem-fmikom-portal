import type { Metadata } from "next";
import { AspirationAdminList } from "@/features/aspirations/aspiration-admin";

export const metadata: Metadata = { title: "Manajemen Aspirasi" };

export default function AdminAspirationsPage() {
  return <AspirationAdminList />;
}
