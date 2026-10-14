"use client";

import * as React from "react";
import { UserRound } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/overlays";
import { MediaFrame } from "@/components/ui/media-frame";
import { MediaImage } from "@/components/ui/media-image";
import type { PublicOrganization, PublicOrganizationMember } from "@/features/content/types";

const positionLabels: Record<PublicOrganizationMember["position"], string> = {
  KETUA: "Ketua",
  WAKIL_KETUA: "Wakil Ketua",
  SEKRETARIS: "Sekretaris",
  BENDAHARA: "Bendahara",
  KEPALA_BIDANG: "Kepala Bidang",
  ANGGOTA: "Anggota",
};

type PublicProfile = {
  member: PublicOrganizationMember;
  divisionName?: string | null;
  periodName: string;
};

function initials(name: string) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toLocaleUpperCase("id-ID") ?? "")
    .join("");
}

function PersonCard({
  member,
  divisionName,
  onSelect,
}: {
  member: PublicOrganizationMember;
  divisionName?: string | null;
  onSelect: (trigger: HTMLButtonElement) => void;
}) {
  const label = positionLabels[member.position];
  return (
    <button
      type="button"
      aria-label={`Lihat profil ${member.name}, ${label}`}
      onClick={(event) => onSelect(event.currentTarget)}
      className="group flex w-full min-w-0 items-center gap-3 rounded-xl border border-border bg-white p-3 text-left shadow-card transition hover:border-secondary/50 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary focus-visible:ring-offset-2 sm:flex-col sm:items-stretch sm:p-4 md:w-40 lg:w-48"
    >
      <MediaFrame
        className="size-14 shrink-0 self-center rounded-full sm:size-24 sm:rounded-xl"
        ratio="aspect-square"
      >
        <MediaImage
          value={member.photo_url}
          alt={`Foto ${member.name}`}
          fill
          sizes="(min-width: 640px) 96px, 56px"
          className="object-cover object-center"
          fallbackClassName="bg-primary-fixed text-primary"
          fallbackText={initials(member.name)}
          fallback={<UserRound aria-hidden="true" className="size-6" />}
        />
      </MediaFrame>
      <span className="min-w-0 flex-1 sm:text-center">
        <span className="block break-words font-heading text-sm font-bold leading-5 text-primary group-hover:text-secondary">
          {member.name}
        </span>
        <span className="mt-1 block break-words text-xs leading-5 text-muted">{label}</span>
        {divisionName ? (
          <span className="mt-1 block break-words text-xs leading-5 text-muted">{divisionName}</span>
        ) : null}
      </span>
    </button>
  );
}

function EmptyMessage({ children }: { children: string }) {
  return (
    <p className="rounded-lg border border-dashed border-border bg-surface-soft px-4 py-3 text-center text-sm leading-6 text-muted">
      {children}
    </p>
  );
}

type OrganizationDivision = PublicOrganization["divisions"][number];

function groupMembers(members: PublicOrganizationMember[], groupSize = 4) {
  const groups: PublicOrganizationMember[][] = [];
  for (let index = 0; index < members.length; index += groupSize) {
    groups.push(members.slice(index, index + groupSize));
  }
  return groups;
}

function DivisionOrganizationChart({
  division,
  onSelect,
}: {
  division: OrganizationDivision;
  onSelect: (member: PublicOrganizationMember, divisionName: string) => (trigger: HTMLButtonElement) => void;
}) {
  const memberGroups = groupMembers(division.members);

  return (
    <article
      aria-labelledby={`division-${division.name.replace(/[^a-z0-9]+/gi, "-")}`}
      className="min-w-0 rounded-2xl border border-border bg-white p-4 shadow-card sm:p-6 lg:p-8"
      data-testid="division-chart"
    >
      <h3
        id={`division-${division.name.replace(/[^a-z0-9]+/gi, "-")}`}
        className="break-words text-center font-heading text-lg font-bold text-primary"
      >
        {division.name}
      </h3>

      <div className="mt-5 flex min-w-0 flex-col items-center sm:mt-6">
        {division.head ? (
          <PersonCard
            member={division.head}
            divisionName={division.name}
            onSelect={onSelect(division.head, division.name)}
          />
        ) : (
          <p className="w-full max-w-[17.5rem] rounded-lg border border-dashed border-border bg-surface-soft px-4 py-3 text-center text-sm leading-5 text-muted">
            Kepala Bidang belum ditetapkan
          </p>
        )}

        {division.members.length === 0 ? (
          <div className="mt-4 w-full max-w-[17.5rem]">
            <EmptyMessage>Belum ada anggota</EmptyMessage>
          </div>
        ) : (
          <div className="mt-5 w-full md:mt-0" data-testid="organization-children">
            {division.head ? (
              <div
                aria-hidden="true"
                className="mx-auto h-5 w-px bg-secondary/45 md:h-7"
                data-testid="head-member-stem"
              />
            ) : null}
            {memberGroups.map((members, groupIndex) => (
              <React.Fragment key={members.map((member) => `${member.position}-${member.name}`).join("|")}>
                {division.head && groupIndex > 0 ? (
                  <div aria-hidden="true" className="mx-auto h-4 w-px bg-secondary/45 md:h-6" />
                ) : null}
                <div className="relative mx-auto w-full md:w-fit" data-testid="member-branch-group">
                  {division.head && members.length > 1 ? (
                    <span
                      aria-hidden="true"
                      className="absolute left-1/2 top-0 hidden h-px w-[calc(100%_-_10rem)] -translate-x-1/2 bg-secondary/45 md:block lg:w-[calc(100%_-_12rem)]"
                      data-testid="member-branch-line"
                    />
                  ) : null}
                  <ul
                    className={[
                      "flex w-full max-w-[17.5rem] flex-col items-stretch gap-2 md:grid md:w-fit md:max-w-none md:grid-flow-col md:justify-center md:gap-x-2 md:gap-y-2",
                      division.head ? "md:pt-7" : "",
                      "md:grid-cols-[repeat(var(--member-count),10rem)] lg:grid-cols-[repeat(var(--member-count),12rem)]",
                      "lg:gap-x-3 xl:gap-x-6",
                    ].join(" ")}
                    style={{ "--member-count": members.length } as React.CSSProperties}
                    data-testid="member-row"
                  >
                    {members.map((member, memberIndex) => (
                      <li
                        key={`${member.position}-${member.name}-${member.photo_url ?? "no-photo"}`}
                        className={[
                          "relative min-w-0",
                          division.head && memberIndex > 0
                            ? "before:absolute before:-top-2 before:left-1/2 before:h-2 before:w-px before:-translate-x-1/2 before:bg-secondary/45 md:before:hidden"
                            : "",
                          division.head
                            ? "md:before:absolute md:before:left-1/2 md:before:top-0 md:before:h-7 md:before:w-px md:before:-translate-x-1/2 md:before:bg-secondary/45"
                            : "",
                        ].join(" ")}
                      >
                        <PersonCard
                          member={member}
                          divisionName={division.name}
                          onSelect={onSelect(member, division.name)}
                        />
                      </li>
                    ))}
                  </ul>
                </div>
              </React.Fragment>
            ))}
          </div>
        )}
      </div>
    </article>
  );
}

function MemberDetail({
  profile,
  onClose,
  returnFocusRef,
}: {
  profile: PublicProfile | null;
  onClose: () => void;
  returnFocusRef: React.RefObject<HTMLButtonElement | null>;
}) {
  return (
    <Dialog open={profile !== null} onOpenChange={(open) => !open && onClose()}>
      {profile ? (
        <DialogContent
          className="w-[min(92vw,28rem)] p-5 sm:p-6"
          onCloseAutoFocus={(event) => {
            event.preventDefault();
            requestAnimationFrame(() => returnFocusRef.current?.focus());
          }}
        >
          <DialogTitle className="pr-10 font-heading text-xl font-bold text-primary">
            Profil Pengurus
          </DialogTitle>
          <DialogDescription className="mt-1 text-sm text-muted">
            Informasi publik pengurus pada periode aktif.
          </DialogDescription>
          <div className="mt-5 flex flex-col items-center text-center">
            <MediaFrame
              className="size-32 overflow-hidden rounded-2xl border border-border sm:size-40"
              ratio="aspect-square"
            >
              <MediaImage
                value={profile.member.photo_url}
                alt={`Foto ${profile.member.name}`}
                fill
                sizes="160px"
                className="object-cover object-center"
                fallbackClassName="bg-primary-fixed text-primary"
                fallbackText={initials(profile.member.name)}
                fallback={<UserRound aria-hidden="true" className="size-8" />}
              />
            </MediaFrame>
            <h3 className="mt-4 break-words font-heading text-lg font-bold text-primary">
              {profile.member.name}
            </h3>
            <p className="mt-1 text-sm font-semibold text-secondary">
              {positionLabels[profile.member.position]}
            </p>
          </div>
          <dl className="mt-5 grid gap-3 rounded-xl bg-surface-soft p-4 text-sm sm:grid-cols-2">
            {profile.divisionName ? (
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-muted">Bidang / Divisi</dt>
                <dd className="mt-1 break-words font-medium text-primary">{profile.divisionName}</dd>
              </div>
            ) : null}
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wide text-muted">Periode</dt>
              <dd className="mt-1 break-words font-medium text-primary">{profile.periodName}</dd>
            </div>
          </dl>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}

export function PublicOrganizationPage({
  organization,
  loadError = false,
}: {
  organization: PublicOrganization | null;
  loadError?: boolean;
}) {
  const [profile, setProfile] = React.useState<PublicProfile | null>(null);
  const returnFocusRef = React.useRef<HTMLButtonElement | null>(null);
  const activePeriod = organization?.period ?? null;
  const showProfile =
    (member: PublicOrganizationMember, divisionName?: string | null) => (trigger: HTMLButtonElement) => {
      returnFocusRef.current = trigger;
      setProfile({ member, divisionName, periodName: activePeriod?.name ?? "" });
    };
  const leadership = organization?.leadership ?? [];
  const chairIndex = leadership.findIndex((member) => member.position === "KETUA");
  const chair = chairIndex >= 0 ? leadership[chairIndex] : null;
  const otherLeaders = leadership.filter((_, index) => index !== chairIndex);

  return (
    <div className="min-w-0 overflow-x-clip bg-background px-4 pb-16 pt-10 md:px-8 md:pb-20 md:pt-14 lg:px-12">
      <div className="mx-auto w-full max-w-[var(--container-public)]">
        <header className="mb-10 max-w-3xl md:mb-14">
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-secondary">BEM FMIKOM UNUGHA</p>
          <h1 className="mt-2 font-heading text-3xl font-bold leading-tight tracking-tight text-primary md:text-4xl">
            Struktur Organisasi
          </h1>
          <p className="mt-4 max-w-2xl text-sm leading-6 text-muted md:text-base">
            Susunan pengurus Badan Eksekutif Mahasiswa Fakultas Matematika dan Ilmu Komputer pada periode yang
            sedang berjalan.
          </p>
          {activePeriod ? (
            <p className="mt-5 inline-flex min-h-11 items-center rounded-full bg-primary-fixed px-4 py-2 text-sm font-semibold text-primary">
              Periode {activePeriod.name}
            </p>
          ) : null}
        </header>

        {loadError ? (
          <EmptyMessage>Struktur organisasi belum dapat dimuat. Silakan coba kembali nanti.</EmptyMessage>
        ) : !activePeriod ? (
          <EmptyMessage>
            Periode aktif belum tersedia. Struktur organisasi akan ditampilkan setelah periode ditetapkan.
          </EmptyMessage>
        ) : (
          <div className="space-y-12 md:space-y-16">
            <section aria-labelledby="leadership-heading">
              <div className="mb-5 text-center">
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-secondary">Pimpinan</p>
                <h2
                  id="leadership-heading"
                  className="mt-1 font-heading text-xl font-bold text-primary md:text-2xl"
                >
                  Badan Pengurus Harian
                </h2>
              </div>
              {leadership.length ? (
                <div className="flex min-w-0 flex-col items-center">
                  {chair ? (
                    <PersonCard
                      member={chair}
                      divisionName={chair.division_name}
                      onSelect={showProfile(chair, chair.division_name)}
                    />
                  ) : null}
                  {otherLeaders.length ? (
                    <>
                      {chair ? <div aria-hidden="true" className="h-8 w-px bg-secondary/55" /> : null}
                      <ul
                        className={[
                          "relative grid min-w-0 w-full grid-cols-1 justify-items-center gap-3",
                          chair ? "sm:border-t sm:border-secondary/45 sm:pt-5" : "",
                          "sm:grid-cols-2 xl:grid-cols-3",
                        ].join(" ")}
                      >
                        {otherLeaders.map((member) => (
                          <li
                            key={`${member.position}-${member.name}-${member.photo_url ?? "no-photo"}`}
                            className="relative min-w-0 max-w-full sm:before:absolute sm:before:left-1/2 sm:before:-top-5 sm:before:h-5 sm:before:w-px sm:before:bg-secondary/45"
                          >
                            <PersonCard
                              member={member}
                              divisionName={member.division_name}
                              onSelect={showProfile(member, member.division_name)}
                            />
                          </li>
                        ))}
                      </ul>
                    </>
                  ) : null}
                  {!chair && otherLeaders.length === 0 ? (
                    <EmptyMessage>Pengurus harian periode ini belum tersedia.</EmptyMessage>
                  ) : null}
                </div>
              ) : (
                <EmptyMessage>Pengurus harian periode ini belum tersedia.</EmptyMessage>
              )}
            </section>

            <section aria-labelledby="divisions-heading">
              <div className="mb-8 text-center">
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-secondary">Bidang kerja</p>
                <h2
                  id="divisions-heading"
                  className="mt-1 font-heading text-xl font-bold text-primary md:text-2xl"
                >
                  Divisi dan Bidang
                </h2>
              </div>
              {organization?.divisions.length ? (
                <div className="space-y-8" data-testid="division-chart-list">
                  {organization.divisions.map((division) => (
                    <DivisionOrganizationChart
                      key={division.name}
                      division={division}
                      onSelect={showProfile}
                    />
                  ))}
                </div>
              ) : (
                <EmptyMessage>Bidang kerja untuk periode ini belum tersedia.</EmptyMessage>
              )}
            </section>
          </div>
        )}
      </div>
      <MemberDetail profile={profile} onClose={() => setProfile(null)} returnFocusRef={returnFocusRef} />
    </div>
  );
}
