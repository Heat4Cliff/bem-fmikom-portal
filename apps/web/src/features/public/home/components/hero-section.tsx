import { ArrowRight, CalendarDays, MessageSquareText, Sparkles } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import type { LandingHero } from "../types";

export function HeroSection({ hero, demoMode }: { hero: LandingHero; demoMode: boolean }) {
  return (
    <>
      {demoMode ? (
        <div className="border-b border-amber-200 bg-amber-50 px-4 py-2 text-center text-xs font-semibold text-amber-950">
          Mode demo: konten dan angka contoh hanya digunakan untuk tinjauan visual.
        </div>
      ) : null}
      <section className="overflow-hidden bg-background">
        <div className="mx-auto grid w-full max-w-[var(--container-public)] items-center gap-9 px-4 py-10 sm:py-14 md:px-8 lg:grid-cols-[0.9fr_1.1fr] lg:gap-12 lg:px-12 lg:py-16">
          <div className="flex flex-col items-start">
            <span className="mb-4 inline-flex items-center gap-2 rounded-full border border-border bg-white px-3 py-1.5 text-[11px] font-bold uppercase tracking-wide text-primary">
              <Sparkles aria-hidden="true" className="size-3.5 text-secondary" />
              {hero.eyebrow}
            </span>
            <h1 className="max-w-2xl font-heading text-[2.15rem] font-extrabold leading-[1.12] tracking-[-0.03em] text-primary sm:text-5xl lg:text-[3.45rem]">
              {hero.title}
            </h1>
            <p className="mt-4 max-w-xl text-base leading-7 text-body sm:text-lg sm:leading-8">
              {hero.description}
            </p>
            <div className="mt-7 flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
              <Button asChild size="lg" className="w-full sm:w-auto">
                <Link href={hero.primaryAction.href}>
                  {hero.primaryAction.label}
                  <ArrowRight aria-hidden="true" className="size-4" />
                </Link>
              </Button>
              <Button asChild size="lg" variant="outline" className="w-full sm:w-auto">
                <Link href={hero.secondaryAction.href}>
                  <MessageSquareText aria-hidden="true" className="size-4 text-secondary" />
                  {hero.secondaryAction.label}
                </Link>
              </Button>
            </div>
          </div>

          <div
            role="img"
            aria-label={hero.visualLabel}
            className="public-hero-art relative isolate flex aspect-[16/10] min-h-64 items-end overflow-hidden rounded-2xl border border-primary/10 p-4 shadow-card sm:p-6"
          >
            <div aria-hidden="true" className="absolute inset-0 -z-10 opacity-60">
              <div className="absolute -right-10 -top-16 size-64 rounded-full border-[28px] border-white/40" />
              <div className="absolute left-[16%] top-[14%] h-[60%] w-[46%] rounded-t-[5rem] border border-white/70 bg-white/25 shadow-inner" />
              <div className="absolute bottom-[22%] left-[8%] h-2 w-[70%] rounded-full bg-primary/10" />
              <div className="absolute bottom-[14%] left-[14%] h-2 w-[56%] rounded-full bg-primary/10" />
            </div>
            <div className="relative z-10 flex w-full items-end justify-between gap-3 rounded-xl border border-white/60 bg-white/90 p-4 shadow-raised backdrop-blur">
              <div className="min-w-0">
                <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-secondary">
                  BEM FMIKOM · UNUGHA
                </p>
                <p className="mt-1 font-heading text-base font-bold leading-snug text-primary sm:text-lg">
                  {hero.visualCaption}
                </p>
              </div>
              <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary text-white">
                <CalendarDays aria-hidden="true" className="size-5" />
              </span>
            </div>
          </div>
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
            className="rounded-xl border border-border/60 bg-white p-4 shadow-card sm:p-5"
          >
            <p className="font-heading text-2xl font-extrabold tracking-tight text-primary sm:text-3xl">
              {item.value}
            </p>
            <p className="mt-1 text-xs font-medium leading-5 text-muted sm:text-sm">{item.label}</p>
            <span className="mt-3 inline-flex rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-bold text-amber-950">
              Contoh
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}
