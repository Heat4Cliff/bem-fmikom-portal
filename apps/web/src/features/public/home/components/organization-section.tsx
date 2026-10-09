import Image from "next/image";
import { ArrowRight, UserRound } from "lucide-react";
import Link from "next/link";
import type { OrganizationMemberPreview, OrganizationPreview } from "../types";
import { EmptyPreview, SectionHeading } from "./section-primitives";

const titles: Record<string, string> = {
  KETUA: "Ketua",
  WAKIL_KETUA: "Wakil Ketua",
  SEKRETARIS: "Sekretaris",
  BENDAHARA: "Bendahara",
  KEPALA_BIDANG: "Kepala Bidang",
  ANGGOTA: "Anggota",
};

function MemberCard({ member }: { member: OrganizationMemberPreview }) {
  return (
    <li className="flex min-w-0 items-center gap-3 rounded-lg border border-border/70 bg-white p-3 shadow-card">
      {member.photoUrl ? (
        <Image
          src={member.photoUrl}
          alt={`Foto ${member.name}`}
          width={48}
          height={48}
          unoptimized
          className="size-12 shrink-0 rounded-full object-cover"
        />
      ) : (
        <span
          aria-hidden="true"
          className="grid size-12 shrink-0 place-items-center rounded-full bg-primary-fixed text-primary"
        >
          <UserRound className="size-5" />
        </span>
      )}
      <span className="min-w-0">
        <span className="block text-xs font-semibold text-secondary">
          {titles[member.role] ?? member.role}
        </span>
        <span className="block break-words font-heading text-sm font-bold text-primary">{member.name}</span>
        {member.division ? <span className="block text-xs text-muted">{member.division}</span> : null}
      </span>
    </li>
  );
}

export function OrganizationSection({
  organization,
  demoMode,
}: {
  organization: OrganizationPreview;
  demoMode: boolean;
}) {
  const leadership = organization.leadership ?? [];
  const divisions = organization.divisions ?? [];
  const hasOrganization =
    leadership.length > 0 || divisions.some((division) => division.head || division.members.length > 0);

  return (
    <section id="struktur" className="scroll-mt-24 bg-background px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow={organization.period?.name ?? (organization.periodLabel ? "Periode aktif" : "Organisasi")}
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
        {organization.period ? (
          <p className="mb-5 text-sm font-semibold text-muted">Periode {organization.period.name}</p>
        ) : organization.periodLabel ? (
          <p className="mb-5 text-sm font-semibold text-muted">{organization.periodLabel}</p>
        ) : null}
        {hasOrganization ? (
          <div className="space-y-6">
            {leadership.length ? (
              <section
                aria-labelledby="public-leadership-heading"
                className="rounded-xl bg-primary-container/5 p-4 md:p-6"
              >
                <h3
                  id="public-leadership-heading"
                  className="mb-3 font-heading text-base font-bold text-primary"
                >
                  Pengurus Harian
                </h3>
                <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                  {leadership.map((member) => (
                    <MemberCard key={String(member.id)} member={member} />
                  ))}
                </ul>
              </section>
            ) : null}
            {divisions.filter((division) => division.head || division.members.length > 0).length ? (
              <section aria-labelledby="public-divisions-heading">
                <h3
                  id="public-divisions-heading"
                  className="mb-3 font-heading text-base font-bold text-primary"
                >
                  Divisi
                </h3>
                <ul className="grid gap-4 md:grid-cols-2">
                  {divisions
                    .filter((division) => division.head || division.members.length > 0)
                    .map((division) => (
                      <li
                        key={division.id}
                        className="min-w-0 rounded-xl border border-border/70 bg-white p-4 shadow-card md:p-5"
                      >
                        <h4 className="break-words font-heading text-lg font-bold text-primary">
                          {division.name}
                        </h4>
                        <ul className="mt-3 space-y-2">
                          {division.head ? (
                            <MemberCard member={{ ...division.head, division: division.name }} />
                          ) : null}
                          {division.members.map((member) => (
                            <MemberCard
                              key={String(member.id)}
                              member={{ ...member, division: division.name }}
                            />
                          ))}
                        </ul>
                      </li>
                    ))}
                </ul>
              </section>
            ) : null}
          </div>
        ) : (
          <EmptyPreview title="Data struktur belum tersedia">
            Susunan pengurus periode aktif akan ditampilkan setelah pengurus organisasi mengisi data resminya.
          </EmptyPreview>
        )}
      </div>
    </section>
  );
}
