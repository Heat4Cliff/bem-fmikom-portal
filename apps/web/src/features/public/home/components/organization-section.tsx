import { ArrowRight, UserRound } from "lucide-react";
import { MediaImage } from "@/components/ui/media-image";
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
    <li className="flex min-w-0 items-center gap-3 rounded-lg border border-border/70 bg-surface p-3 shadow-card">
      <MediaImage
        value={member.photoUrl}
        alt={`Foto ${member.name}`}
        width={48}
        height={48}
        className="size-12 rounded-full object-cover object-center"
        fallbackClassName="bg-primary-fixed text-primary"
        fallbackText={member.name
          .trim()
          .split(/\s+/)
          .slice(0, 2)
          .map((part) => part[0]?.toUpperCase() ?? "")
          .join("")}
        fallback={<UserRound aria-hidden="true" className="size-5" />}
      />
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
  const leadership = (organization.leadership ?? []).slice(0, 4);

  return (
    <section id="struktur" className="scroll-mt-24 bg-background px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow={organization.period?.name ?? (organization.periodLabel ? "Periode aktif" : "Organisasi")}
          title="Struktur Organisasi"
          description="Kenali susunan pengurus BEM FMIKOM UNUGHA pada periode yang sedang berjalan."
          action={
            <Link
              href="/struktur"
              className="inline-flex min-h-11 items-center gap-2 rounded-md px-3 text-sm font-semibold text-primary hover:bg-surface-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
            >
              Lihat struktur <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          }
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-warning-foreground">
            Contoh visual · nama dan periode bukan data resmi
          </p>
        ) : null}
        {organization.period ? (
          <p className="mb-5 text-sm font-semibold text-muted">Periode {organization.period.name}</p>
        ) : organization.periodLabel ? (
          <p className="mb-5 text-sm font-semibold text-muted">{organization.periodLabel}</p>
        ) : null}
        {organization.status === "error" ? (
          <EmptyPreview title="Struktur belum dapat dimuat">
            Informasi struktur organisasi belum bisa dimuat. Silakan coba lagi nanti.
          </EmptyPreview>
        ) : organization.status === "empty" || leadership.length === 0 ? (
          <EmptyPreview title="Data struktur belum tersedia">
            {organization.period
              ? "Ringkasan pengurus periode aktif akan ditampilkan setelah data resmi tersedia."
              : "Periode aktif belum tersedia. Informasi struktur akan tampil setelah periode organisasi ditetapkan."}
          </EmptyPreview>
        ) : null}
        {leadership.length > 0 && organization.status !== "error" ? (
          <ul className="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {leadership.map((member) => (
              <MemberCard key={String(member.id)} member={member} />
            ))}
          </ul>
        ) : null}
      </div>
    </section>
  );
}
