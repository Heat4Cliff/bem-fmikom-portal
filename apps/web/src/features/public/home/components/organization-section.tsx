import { ArrowRight, UserRound } from "lucide-react";
import Link from "next/link";
import type { OrganizationPreview } from "../types";
import { EmptyPreview, SectionHeading } from "./section-primitives";

export function OrganizationSection({
  organization,
  demoMode,
}: {
  organization: OrganizationPreview;
  demoMode: boolean;
}) {
  return (
    <section id="struktur" className="scroll-mt-24 bg-background px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow="Periode aktif"
          title="Struktur Organisasi"
          description="Kenali susunan pengurus BEM FMIKOM UNUGHA pada periode yang sedang berjalan."
          action={
            <Link
              href="#struktur"
              className="inline-flex min-h-10 items-center gap-2 rounded-md px-3 text-sm font-semibold text-primary hover:bg-surface-muted focus-visible:outline-2 focus-visible:outline-secondary"
            >
              Lihat struktur <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          }
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-amber-900">
            Contoh visual · nama dan periode bukan data resmi
          </p>
        ) : null}
        {organization.members.length ? (
          <>
            {organization.periodLabel ? (
              <p className="mb-4 text-sm font-semibold text-muted">{organization.periodLabel}</p>
            ) : null}
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {organization.members.map((member, index) => (
                <article
                  key={member.id}
                  className="overflow-hidden rounded-xl border border-border/70 bg-white shadow-card"
                >
                  <div
                    className={`grid aspect-[4/3] place-items-center ${index % 2 ? "bg-surface-soft" : "bg-surface-muted"}`}
                  >
                    <span className="grid size-16 place-items-center rounded-full border-4 border-white bg-primary text-white shadow-raised">
                      <UserRound aria-hidden="true" className="size-7" />
                    </span>
                  </div>
                  <div className="p-4">
                    {member.division ? (
                      <span className="inline-flex rounded-full bg-surface-muted px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-primary">
                        {member.division}
                      </span>
                    ) : null}
                    <p className="mt-3 text-xs font-semibold text-secondary">{member.role}</p>
                    <h3 className="mt-1 font-heading text-base font-bold text-primary">{member.name}</h3>
                  </div>
                </article>
              ))}
            </div>
          </>
        ) : (
          <EmptyPreview title="Data struktur belum tersedia">
            Susunan pengurus periode aktif akan ditampilkan setelah layanan organisasi menyediakan data
            resminya.
          </EmptyPreview>
        )}
      </div>
    </section>
  );
}
