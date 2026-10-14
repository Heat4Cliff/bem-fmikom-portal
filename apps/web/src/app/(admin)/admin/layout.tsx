import { headers } from "next/headers";
import type { ReactNode } from "react";
import { AdminRouteGuard, AuthProvider } from "@/lib/auth/auth-provider";

export const dynamic = "force-dynamic";

export default async function AdminLayout({ children }: Readonly<{ children: ReactNode }>) {
  // Dynamic rendering lets Next.js propagate the per-request CSP nonce to scripts.
  await headers();

  return (
    <AuthProvider>
      <AdminRouteGuard>{children}</AdminRouteGuard>
    </AuthProvider>
  );
}
