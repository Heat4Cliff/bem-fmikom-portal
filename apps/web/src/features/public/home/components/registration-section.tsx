import { ArrowRight, CalendarClock } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import type { RegistrationPreview } from "../types";
import { EmptyPreview, SectionHeading } from "./section-primitives";

export function RegistrationSection({
  registrations,
  demoMode,
}: {
  registrations: RegistrationPreview[];
  demoMode: boolean;
}) {
  return (
    <section id="pendaftaran" className="scroll-mt-24 bg-surface-muted px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow="Peluang berpartisipasi"
          title="Pendaftaran"
          description="Informasi pendaftaran yang sedang dibuka akan ditampilkan di sini."
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-amber-900">
            Contoh visual · status dan jadwal bukan informasi resmi
          </p>
        ) : null}
        {registrations.length ? (
          <div className="grid grid-cols-1 items-stretch gap-4 lg:grid-cols-2">
            {registrations.map((registration) => (
              <article
                key={registration.id}
                className="grid h-full w-full min-w-0 gap-5 rounded-2xl border border-primary/10 bg-white p-5 shadow-card sm:p-7 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
              >
                <div>
                  <span className="inline-flex rounded-full bg-amber-100 px-3 py-1 text-[10px] font-bold uppercase tracking-wide text-amber-950">
                    {registration.status}
                  </span>
                  <h3 className="mt-3 font-heading text-xl font-bold leading-snug text-primary sm:text-2xl">
                    {registration.title}
                  </h3>
                  {registration.description ? (
                    <p className="mt-2 line-clamp-3 text-sm leading-6 text-muted">
                      {registration.description}
                    </p>
                  ) : null}
                  <p className="mt-2 text-sm text-muted">
                    {registration.division ?? "Pendaftaran BEM FMIKOM"}
                  </p>
                  <div className="mt-4 flex flex-wrap gap-2 text-xs text-body">
                    {registration.openDate ? (
                      <span className="rounded-md bg-surface-muted px-3 py-2">
                        Buka: {registration.openDate}
                      </span>
                    ) : null}
                    {registration.closeDate ? (
                      <span className="rounded-md bg-surface-muted px-3 py-2">
                        Tutup: {registration.closeDate}
                      </span>
                    ) : null}
                  </div>
                </div>
                <div className="flex flex-col items-start gap-3 md:items-end">
                  <CalendarClock aria-hidden="true" className="size-8 text-secondary" />
                  <Button asChild variant="outline">
                    <Link
                      href={
                        registration.slug
                          ? `/pendaftaran/${encodeURIComponent(registration.slug)}/form`
                          : "/pendaftaran"
                      }
                    >
                      Daftar sekarang <ArrowRight aria-hidden="true" className="size-4" />
                    </Link>
                  </Button>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <EmptyPreview title="Belum ada pendaftaran aktif">
            Informasi pendaftaran muncul setelah formulir resmi dibuka. Bagian ini belum menerima pendaftaran.
          </EmptyPreview>
        )}
      </div>
    </section>
  );
}
