import type { Metadata } from "next";
import { UsersAdmin } from "@/features/system/users-admin";

export const metadata: Metadata = { title: "Manajemen Pengguna" };
export default function UsersPage() {
  return <UsersAdmin />;
}
