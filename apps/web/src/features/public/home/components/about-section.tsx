import { BookOpenText, MessageCircleHeart, UsersRound } from "lucide-react";
import type { AboutContent } from "../types";
import { SectionHeading } from "./section-primitives";

const highlightIcons = [BookOpenText, UsersRound, MessageCircleHeart];

export function AboutSection({ about }: { about: AboutContent }) {
  return (
    <section id="tentang" className="scroll-mt-24 bg-surface-muted px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto grid max-w-[var(--container-public)] gap-8 lg:grid-cols-[0.95fr_1.05fr] lg:gap-14">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-secondary">Tentang BEM</p>
          <h2 className="mt-2 font-heading text-2xl font-bold leading-tight tracking-tight text-primary sm:text-3xl">
            {about.title}
          </h2>
          <p className="mt-4 max-w-xl text-sm leading-7 text-body sm:text-base">{about.description}</p>
          <p className="mt-4 inline-flex rounded-full bg-white px-3 py-1.5 text-xs font-semibold text-muted ring-1 ring-border/70">
            Informasi umum portal · bukan data periode aktif
          </p>
        </div>
        {about.highlights.length > 0 ? (
          <div className="grid gap-3 sm:grid-cols-3">
            {about.highlights.map((item, index) => {
              const Icon = highlightIcons[index % highlightIcons.length];
              return (
                <article
                  key={item.title}
                  className="rounded-xl border border-border/70 bg-white p-4 shadow-card sm:p-5"
                >
                  <span className="grid size-10 place-items-center rounded-lg bg-surface-soft text-primary">
                    <Icon aria-hidden="true" className="size-5" />
                  </span>
                  <h3 className="mt-4 font-heading text-base font-bold leading-snug text-primary">
                    {item.title}
                  </h3>
                  <p className="mt-2 text-sm leading-6 text-muted">{item.description}</p>
                </article>
              );
            })}
          </div>
        ) : null}
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
          <p className="mb-4 text-xs font-semibold text-amber-900">Contoh visual · bukan visi-misi resmi</p>
        ) : null}
        {vision || mission.length > 0 ? (
          <div className="grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
            <article className="rounded-2xl bg-primary p-6 text-white shadow-raised sm:p-8">
              <p className="text-xs font-bold uppercase tracking-[0.18em] text-amber-200">Visi</p>
              <p className="mt-5 font-heading text-xl font-bold leading-relaxed sm:text-2xl">
                {vision ?? "Visi belum tersedia."}
              </p>
            </article>
            <div className="grid gap-3 sm:grid-cols-2">
              {mission.map((item, index) => (
                <article key={item} className="rounded-xl border border-border/70 bg-white p-5 shadow-card">
                  <span className="grid size-8 place-items-center rounded-md bg-surface-soft text-sm font-bold text-primary">
                    {index + 1}
                  </span>
                  <p className="mt-3 text-sm leading-6 text-body">{item}</p>
                </article>
              ))}
            </div>
          </div>
        ) : (
          <div className="rounded-xl border border-dashed border-primary/20 bg-white/75 p-6 text-sm leading-6 text-muted">
            Konten visi dan misi resmi belum tersedia. Bagian ini akan menampilkan isi landing settings
            setelah data dikonfirmasi.
          </div>
        )}
      </div>
    </section>
  );
}
