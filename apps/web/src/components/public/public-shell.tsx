import type { ReactNode } from "react";
import { PublicFooter, PublicHeader, MobileNavigation } from "@/components/public/public-navigation";

export function PublicShell({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-dvh pb-16 md:pb-0">
      <a
        href="#main-content"
        className="sr-only z-[60] rounded-md bg-primary px-4 py-3 font-semibold text-white focus:not-sr-only focus:fixed focus:left-4 focus:top-4"
      >
        Lewati ke konten utama
      </a>
      <PublicHeader />
      <main id="main-content" tabIndex={-1}>
        {children}
      </main>
      <PublicFooter />
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
