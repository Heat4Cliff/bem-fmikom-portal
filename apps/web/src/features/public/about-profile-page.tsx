import { Building2, Compass, Eye } from "lucide-react";
import { BrandLogo } from "@/components/ui/brand-logo";
import { PublicContainer } from "@/components/public/public-shell";
import type { PublicLandingSettings } from "@/features/content/api";

export function AboutProfilePage({ settings }: { settings: PublicLandingSettings | null }) {
  const missions = settings?.missions ?? [];
  const hasCabinetIdentity = Boolean(
    settings?.cabinet_name ||
      settings?.cabinet_period ||
      settings?.cabinet_tagline ||
      settings?.cabinet_logo_url,
  );

  return (
    <div className="min-h-[60dvh] bg-background py-10 sm:py-14 lg:py-20">
      <PublicContainer className="grid grid-cols-[minmax(0,1fr)] min-w-0 gap-10 sm:gap-12 lg:gap-14">
        <header className="min-w-0 max-w-3xl">
          <p className="text-xs font-bold uppercase tracking-[0.18em] text-secondary">Profil organisasi</p>
          <h1 className="mt-2 break-words font-heading text-3xl font-bold tracking-tight text-primary sm:text-4xl">
            Tentang BEM FMIKOM UNUGHA
          </h1>
          <p className="mt-4 break-words text-sm leading-7 text-muted sm:text-base">
            Informasi organisasi dan arah gerak yang dikelola melalui pengaturan portal.
          </p>
        </header>

        <section
          aria-labelledby="cabinet-heading"
          className="flex min-w-0 flex-col items-center gap-4 text-center sm:gap-5"
        >
          {settings?.cabinet_logo_url ? (
            <BrandLogo
              src={settings.cabinet_logo_url}
              alt={`Logo ${settings.cabinet_name || "kabinet BEM FMIKOM"}`}
              shape={settings.cabinet_logo_shape}
              fit="contain"
              className="size-28 max-w-full border-0 bg-transparent shadow-none sm:size-36 lg:size-44"
            />
          ) : (
            <span
              className="grid size-28 max-w-full place-items-center bg-transparent text-muted sm:size-36 lg:size-44"
              role="img"
              aria-label="Logo kabinet belum tersedia"
            >
              <Building2 aria-hidden="true" className="size-10" />
            </span>
          )}
          <div className="flex max-w-2xl flex-col items-center gap-1.5">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-secondary">
              {settings?.cabinet_period ? `Periode ${settings.cabinet_period}` : "Identitas kabinet"}
            </p>
            <h2
              id="cabinet-heading"
              className="break-words font-heading text-2xl font-bold text-primary sm:text-3xl"
            >
              {settings?.cabinet_name ||
                (hasCabinetIdentity ? "Nama kabinet belum diatur" : "Identitas kabinet belum tersedia")}
            </h2>
            {settings?.cabinet_tagline ? (
              <p className="break-words text-sm leading-6 text-muted sm:text-base">
                {settings.cabinet_tagline}
              </p>
            ) : null}
          </div>
        </section>

        <section
          aria-labelledby="about-heading"
          className="min-w-0 rounded-2xl border border-border bg-surface p-5 shadow-card sm:p-8"
        >
          <div className="flex items-start gap-3">
            <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-surface-muted text-primary">
              <Compass aria-hidden="true" className="size-5" />
            </span>
            <div className="min-w-0">
              <h2 id="about-heading" className="break-words font-heading text-xl font-bold text-primary">
                {settings?.about_title || "Tentang BEM FMIKOM"}
              </h2>
              <p className="mt-3 whitespace-pre-line text-sm leading-7 text-body sm:text-base">
                {settings?.about_description?.trim() || "Deskripsi organisasi belum tersedia."}
              </p>
            </div>
          </div>
        </section>

        <div className="grid gap-5 lg:grid-cols-2">
          <section
            aria-labelledby="vision-heading"
            className="rounded-2xl border border-primary/20 bg-brand-panel p-6 text-brand-panel-foreground shadow-card sm:p-8"
          >
            <span className="grid size-10 place-items-center rounded-lg bg-brand-panel-foreground/10">
              <Eye aria-hidden="true" className="size-5" />
            </span>
            <h2 id="vision-heading" className="mt-4 font-heading text-xl font-bold">
              Visi
            </h2>
            <p className="mt-3 whitespace-pre-line text-sm leading-7 sm:text-base">
              {settings?.vision?.trim() || "Visi organisasi belum tersedia."}
            </p>
          </section>
          <section
            aria-labelledby="mission-heading"
            className="rounded-2xl border border-border bg-surface p-6 shadow-card sm:p-8"
          >
            <h2 id="mission-heading" className="font-heading text-xl font-bold text-primary">
              Misi
            </h2>
            {missions.length ? (
              <ol className="mt-4 grid gap-3">
                {missions.map((mission, index) => (
                  <li
                    key={mission}
                    className="flex min-w-0 gap-3 rounded-lg border border-border/70 bg-surface-muted p-3 text-sm leading-6 text-body"
                  >
                    <span className="font-bold text-secondary">{index + 1}.</span>
                    <span className="min-w-0">{mission}</span>
                  </li>
                ))}
              </ol>
            ) : (
              <p className="mt-3 text-sm leading-6 text-muted">Misi organisasi belum tersedia.</p>
            )}
          </section>
        </div>
      </PublicContainer>
    </div>
  );
}
