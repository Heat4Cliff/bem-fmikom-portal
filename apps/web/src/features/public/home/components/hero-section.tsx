import { ArrowRight, CalendarDays, MessageSquareText } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { MediaImage } from "@/components/ui/media-image";
import { BlurText } from "@/components/react-bits/blur-text";
import { FadeContent } from "@/components/react-bits/fade-content";
import type { LandingHero } from "../types";

export function HeroSection({ hero, demoMode }: { hero: LandingHero; demoMode: boolean }) {
  return (
    <>
      {demoMode ? (
        <div className="border-b border-warning/30 bg-warning/10 px-4 py-2 text-center text-xs font-semibold text-warning-foreground">
          Mode demo: konten dan angka contoh hanya digunakan untuk tinjauan visual.
        </div>
      ) : null}
      <section className="overflow-hidden bg-background">
        <div className="mx-auto grid w-full max-w-[var(--container-public)] items-center gap-7 px-4 py-8 sm:gap-9 sm:py-14 md:px-8 lg:grid-cols-[0.9fr_1.1fr] lg:gap-12 lg:px-12 lg:py-16">
          <div className="flex min-w-0 flex-col items-start">
            <p className="mb-3 text-[10px] font-bold uppercase tracking-[0.18em] text-secondary sm:text-xs">
              {hero.eyebrow}
            </p>
            <BlurText
              text={hero.title}
              className="max-w-2xl break-words font-heading text-[clamp(1.875rem,7vw,2.15rem)] font-extrabold leading-[1.12] tracking-[-0.03em] text-primary sm:text-5xl lg:text-[3.45rem]"
            />
            <FadeContent className="mt-4 max-w-xl">
              <p className="text-base leading-7 text-body sm:text-lg sm:leading-8">{hero.description}</p>
            </FadeContent>
            <FadeContent className="mt-7 w-full sm:w-auto">
              <div className="flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
                <Button asChild size="lg" className="w-full sm:w-auto">
                  <Link href={hero.primaryAction.href}>
                    {hero.primaryAction.label}
                    <ArrowRight aria-hidden="true" className="size-4" />
                  </Link>
                </Button>
                {hero.secondaryAction ? (
                  <Button asChild size="lg" variant="outline" className="w-full sm:w-auto">
                    <Link href={hero.secondaryAction.href}>
                      <MessageSquareText aria-hidden="true" className="size-4 text-secondary" />
                      {hero.secondaryAction.label}
                    </Link>
                  </Button>
                ) : null}
              </div>
            </FadeContent>
          </div>

          <FadeContent>
            <figure
              aria-label={hero.visualLabel}
              data-testid="public-hero-art"
              className="public-hero-art relative isolate aspect-[4/3] w-full min-w-0 overflow-hidden rounded-2xl border border-primary/20 shadow-card sm:aspect-[16/10] sm:min-h-64"
            >
              {hero.imageUrl ? (
                <>
                  <MediaImage
                    value={hero.imageUrl}
                    alt={hero.visualLabel}
                    fill
                    sizes="(max-width: 1024px) 100vw, 50vw"
                    className="absolute inset-0 -z-10 size-full object-cover object-center"
                    fallbackClassName="absolute inset-0 -z-10"
                    fallback={
                      <span className="flex size-full flex-col items-center justify-center gap-2 bg-surface-soft px-6 text-center text-primary">
                        <span className="font-heading text-sm font-bold tracking-wide sm:text-base">
                          BEM FMIKOM · UNUGHA
                        </span>
                        <span className="max-w-sm text-xs leading-5 text-muted sm:text-sm">
                          Dokumentasi kegiatan akan ditampilkan setelah aset resmi tersedia.
                        </span>
                      </span>
                    }
                  />
                  <div
                    aria-hidden="true"
                    className="absolute inset-0 -z-10 bg-gradient-to-t from-foreground/40 to-transparent"
                  />
                </>
              ) : (
                <div className="absolute inset-0 -z-10 flex flex-col items-center justify-center gap-2 px-6 text-center text-primary">
                  <p className="font-heading text-sm font-bold tracking-wide sm:text-base">
                    BEM FMIKOM · UNUGHA
                  </p>
                  <p className="max-w-sm text-xs leading-5 text-muted sm:text-sm">
                    Dokumentasi kegiatan akan ditampilkan setelah aset resmi tersedia.
                  </p>
                </div>
              )}
              {hero.imageUrl ? (
                <figcaption className="absolute inset-x-3 bottom-3 rounded-xl border border-border/80 bg-surface/95 p-4 shadow-raised backdrop-blur-sm sm:inset-x-6 sm:bottom-6 sm:p-5">
                  <div className="flex items-start gap-3">
                    <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                      <CalendarDays aria-hidden="true" className="size-5" />
                    </span>
                    <div className="min-w-0">
                      <p className="text-[10px] font-bold uppercase tracking-[0.15em] text-primary sm:text-xs">
                        BEM FMIKOM · UNUGHA
                      </p>
                      <p className="mt-1 text-xs leading-5 text-muted sm:text-sm sm:leading-6">
                        {hero.visualCaption}
                      </p>
                    </div>
                  </div>
                </figcaption>
              ) : null}
            </figure>
          </FadeContent>
        </div>
      </section>
    </>
  );
}

export function DemoStatistics({ items }: { items: { label: string; value: string }[] }) {
  if (items.length === 0) return null;

  return (
    <section
      aria-label="Statistik contoh untuk tinjauan visual"
      className="bg-surface-muted px-4 py-4 md:px-8 lg:px-12"
    >
      <div className="mx-auto grid max-w-[var(--container-public)] grid-cols-2 gap-3 lg:grid-cols-4">
        {items.map((item) => (
          <div
            key={item.label}
            className="rounded-xl border border-border/60 bg-surface p-4 shadow-card sm:p-5"
          >
            <p className="font-heading text-2xl font-extrabold tracking-tight text-primary sm:text-3xl">
              {item.value}
            </p>
            <p className="mt-1 text-xs font-medium leading-5 text-muted sm:text-sm">{item.label}</p>
            <span className="mt-3 inline-flex rounded-full bg-warning/15 px-2 py-0.5 text-[10px] font-bold text-warning-foreground">
              Contoh
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}
