import Link from "next/link";
import type { AboutContent } from "../types";
import { SectionHeading } from "./section-primitives";

export function AboutSection({
  about,
  vision,
  mission,
}: {
  about: AboutContent;
  vision: string | null;
  mission: string[];
}) {
  return (
    <section
      id="tentang"
      className="landing-about-section scroll-mt-24 bg-surface-muted px-4 py-14 md:px-8 lg:px-12 lg:py-20"
    >
      <div className="mx-auto grid max-w-[var(--container-public)] gap-6 lg:grid-cols-[1.1fr_.9fr] lg:gap-10">
        <div className="min-w-0">
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-secondary">Tentang BEM FMIKOM</p>
          <h2 className="mt-2 font-heading text-2xl font-bold leading-tight tracking-tight text-primary sm:text-3xl">
            {about.title}
          </h2>
          <p className="mt-4 line-clamp-4 max-w-2xl text-sm leading-7 text-body sm:text-base">
            {about.description}
          </p>
          <Link
            href="/tentang"
            className="mt-5 inline-flex min-h-11 items-center gap-2 rounded-md font-semibold text-primary underline decoration-primary/40 underline-offset-4 transition-colors hover:text-secondary focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-secondary"
          >
            Lihat selengkapnya <span aria-hidden="true">→</span>
          </Link>
        </div>
        <article className="min-w-0 rounded-2xl border border-border/70 bg-surface p-5 shadow-card sm:p-7">
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-secondary">Arah gerak</p>
          <h3 className="mt-2 font-heading text-xl font-bold text-primary">Visi &amp; Misi</h3>
          {vision ? (
            <p className="mt-3 line-clamp-3 text-sm leading-6 text-body">{vision}</p>
          ) : (
            <p className="mt-3 text-sm text-muted">Visi belum tersedia.</p>
          )}
          {mission.length ? (
            <ul className="mt-4 grid gap-2 text-sm leading-6 text-body">
              {mission.slice(0, 2).map((item, index) => (
                <li key={item} className="flex gap-2">
                  <span className="font-bold text-secondary">{index + 1}.</span>
                  <span className="line-clamp-2">{item}</span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="mt-3 text-sm text-muted">Misi belum tersedia.</p>
          )}
        </article>
      </div>
    </section>
  );
}

export function VisionMissionSection({
  vision,
  mission,
  demoMode,
}: {
  vision: string | null;
  mission: string[];
  demoMode: boolean;
}) {
  return (
    <section id="visi-misi" className="scroll-mt-24 bg-surface-soft px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow="Arah gerak"
          title="Visi & Misi"
          description="Ruang untuk menyampaikan arah dan prioritas organisasi dari konten resmi yang telah dikonfirmasi."
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-warning-foreground">
            Contoh visual · bukan visi-misi resmi
          </p>
        ) : null}
        {vision || mission.length > 0 ? (
          <div className="grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
            <article className="rounded-2xl border border-primary/20 bg-brand-panel p-6 text-brand-panel-foreground shadow-raised sm:p-8">
              <p className="text-xs font-bold uppercase tracking-[0.18em] text-brand-panel-foreground/75">
                Visi
              </p>
              <p className="mt-5 font-heading text-xl font-bold leading-relaxed sm:text-2xl">
                {vision ?? "Visi belum tersedia."}
              </p>
            </article>
            <div className="grid gap-3 sm:grid-cols-2">
              {mission.map((item, index) => (
                <article key={item} className="rounded-xl border border-border/70 bg-surface p-5 shadow-card">
                  <span className="grid size-8 place-items-center rounded-md bg-surface-soft text-sm font-bold text-primary">
                    {index + 1}
                  </span>
                  <p className="mt-3 text-sm leading-6 text-body">{item}</p>
                </article>
              ))}
            </div>
          </div>
        ) : (
          <div className="rounded-xl border border-dashed border-primary/20 bg-surface/75 p-6 text-sm leading-6 text-muted">
            Konten visi dan misi resmi belum tersedia. Bagian ini akan menampilkan isi landing settings
            setelah data dikonfirmasi.
          </div>
        )}
      </div>
    </section>
  );
}
