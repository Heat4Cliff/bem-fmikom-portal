import type { ReactNode } from "react";
import { PublicFooter, PublicHeader, MobileNavigation } from "@/components/public/public-navigation";
import { getPublicLandingSettings } from "@/features/content/api";

export async function PublicShell({ children }: { children: ReactNode }) {
  const settings = await getPublicLandingSettings().catch(() => null);
  return (
    <div className="min-h-dvh pb-[calc(3rem+0.375rem+max(env(safe-area-inset-bottom),0.35rem)+1px)] md:pb-0">
      <a
        href="#main-content"
        className="sr-only z-[60] rounded-md bg-primary px-4 py-3 font-semibold text-primary-foreground focus:not-sr-only focus:fixed focus:left-4 focus:top-4"
      >
        Lewati ke konten utama
      </a>
      <PublicHeader
        logoUrl={settings?.navbar_logo_url ?? null}
        logoShape={settings?.brand_logo_shape ?? "original"}
        logoFit={settings?.brand_logo_fit ?? "contain"}
      />
      <main id="main-content" tabIndex={-1}>
        {children}
      </main>
      <PublicFooter
        logoUrl={
          settings?.footer_use_brand_logo === false ? settings.footer_logo_url : settings?.navbar_logo_url
        }
        logoShape={
          settings?.footer_use_brand_logo === false ? settings?.footer_logo_shape : settings?.brand_logo_shape
        }
        logoFit={
          settings?.footer_use_brand_logo === false ? settings?.footer_logo_fit : settings?.brand_logo_fit
        }
      />
      <MobileNavigation />
    </div>
  );
}

export function PublicContainer({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div
      className={`mx-auto w-full max-w-[var(--container-public)] px-4 md:px-8 lg:px-12 ${className ?? ""}`}
    >
      {children}
    </div>
  );
}

export function SectionContainer({
  children,
  className,
  id,
}: {
  children: ReactNode;
  className?: string;
  id?: string;
}) {
  return (
    <section id={id} className={`py-10 md:py-16 ${className ?? ""}`}>
      <PublicContainer>{children}</PublicContainer>
    </section>
  );
}
