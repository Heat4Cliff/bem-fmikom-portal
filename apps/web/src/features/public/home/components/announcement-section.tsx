"use client";

import { ArrowRight, BellRing } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { Button } from "@/components/ui/button";
import type { AnnouncementPreview } from "../types";
import { AnnouncementPopup } from "./announcement-popup";
import { EmptyPreview, SectionHeading } from "./section-primitives";

export function AnnouncementSection({
  announcements,
  demoMode,
}: {
  announcements: AnnouncementPreview[];
  demoMode: boolean;
}) {
  const [selected, setSelected] = React.useState<AnnouncementPreview | null>(null);

  return (
    <section id="pengumuman" className="scroll-mt-24 bg-background px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow="Informasi penting"
          title="Pengumuman"
          description="Informasi resmi dan pembaruan organisasi akan ditampilkan pada bagian ini."
          action={
            <Link
              href="/pengumuman"
              className="inline-flex min-h-10 items-center gap-2 rounded-md px-3 text-sm font-semibold text-primary hover:bg-surface-muted focus-visible:outline-2 focus-visible:outline-secondary"
            >
              Arsip pengumuman <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          }
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-warning-foreground">
            Contoh visual · bukan pengumuman resmi
          </p>
        ) : null}
        {announcements.length ? (
          <div className="grid gap-3">
            {announcements.map((announcement) => (
              <article
                key={announcement.id}
                className="flex flex-col gap-4 rounded-xl border-l-4 border-secondary bg-surface-muted p-5 sm:flex-row sm:items-center sm:justify-between sm:p-6"
              >
                <div className="flex min-w-0 items-start gap-3">
                  <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-warning/15 text-secondary">
                    <BellRing aria-hidden="true" className="size-5" />
                  </span>
                  <div className="min-w-0">
                    {announcement.dateLabel ? (
                      <p className="text-xs font-semibold text-muted">{announcement.dateLabel}</p>
                    ) : null}
                    <h3 className="mt-1 font-heading text-lg font-bold leading-snug text-primary">
                      {announcement.title}
                    </h3>
                    <p className="mt-2 line-clamp-2 text-sm leading-6 text-body">{announcement.content}</p>
                  </div>
                </div>
                <Button variant="outline" className="shrink-0" onClick={() => setSelected(announcement)}>
                  Baca pengumuman <ArrowRight aria-hidden="true" className="size-4" />
                </Button>
              </article>
            ))}
          </div>
        ) : (
          <EmptyPreview title="Belum ada pengumuman aktif">
            Pengumuman akan ditampilkan setelah informasi resmi tersedia. Popup tidak dibuka tanpa data
            pengumuman.
          </EmptyPreview>
        )}
      </div>
      {selected ? (
        <AnnouncementPopup {...selected} open={Boolean(selected)} onClose={() => setSelected(null)} />
      ) : null}
    </section>
  );
}
