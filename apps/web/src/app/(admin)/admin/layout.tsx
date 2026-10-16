import { headers } from "next/headers";
import type { ReactNode } from "react";
import { AdminRouteGuard, AuthProvider } from "@/lib/auth/auth-provider";
import { getPublicLandingSettings } from "@/features/content/api";
import { BrandingProvider } from "@/features/content/branding-context";

export const dynamic = "force-dynamic";

export default async function AdminLayout({ children }: Readonly<{ children: ReactNode }>) {
  // Dynamic rendering lets Next.js propagate the per-request CSP nonce to scripts.
  await headers();
  const settings = await getPublicLandingSettings().catch(() => null);

  return (
    <BrandingProvider
      value={{
        logoUrl: settings?.navbar_logo_url ?? null,
        logoShape: settings?.brand_logo_shape ?? "original",
        logoFit: settings?.brand_logo_fit ?? "contain",
      }}
    >
      <AuthProvider>
        <AdminRouteGuard>{children}</AdminRouteGuard>
      </AuthProvider>
    </BrandingProvider>
  );
}
