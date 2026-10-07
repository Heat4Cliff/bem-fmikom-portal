import type { ReactNode } from "react";
import { AdminRouteGuard, AuthProvider } from "@/lib/auth/auth-provider";

export default function AdminLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <AuthProvider>
      <AdminRouteGuard>{children}</AdminRouteGuard>
    </AuthProvider>
  );
}
