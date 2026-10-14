"use client";

import { ArrowRight, Image as ImageIcon } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { MediaImage } from "@/components/ui/media-image";
import { MediaFrame } from "@/components/ui/media-frame";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/overlays";

export type AnnouncementPopupProps = {
  title: string;
  content: string;
  image?: { src: string; alt: string };
  dateLabel?: string;
  href?: string;
  open: boolean;
  onClose: () => void;
};

export function AnnouncementPopup({
  title,
  content,
  image,
  dateLabel,
  href,
  open,
  onClose,
}: AnnouncementPopupProps) {
  return (
    <Dialog open={open} onOpenChange={(nextOpen) => !nextOpen && onClose()}>
      <DialogContent
        aria-describedby="announcement-popup-description"
        className="max-h-[calc(100dvh-2rem)] w-[min(92vw,42rem)] overflow-y-auto p-0"
      >
        {image ? (
          <MediaFrame ratio="h-[min(38dvh,18rem)]" className="rounded-t-xl">
            <MediaImage
              value={image.src}
              alt={image.alt}
              fill
              sizes="(max-width: 672px) 92vw, 672px"
              className="object-contain object-center"
              fallbackClassName="absolute inset-0 grid place-items-center bg-surface-muted text-primary"
              fallback={<ImageIcon aria-hidden="true" className="size-8" />}
            />
          </MediaFrame>
        ) : null}
        <div className="p-6 sm:p-8">
          {dateLabel ? (
            <p className="text-xs font-semibold uppercase tracking-wide text-secondary">{dateLabel}</p>
          ) : null}
          <DialogTitle className="mt-2 pr-8 font-heading text-xl font-bold leading-snug text-primary sm:text-2xl">
            {title}
          </DialogTitle>
          <DialogDescription id="announcement-popup-description" className="mt-3 text-sm leading-7 text-body">
            {content}
          </DialogDescription>
          {href ? (
            <Button asChild className="mt-6">
              <Link href={href} onClick={onClose}>
                Lihat informasi <ArrowRight aria-hidden="true" className="size-4" />
              </Link>
            </Button>
          ) : null}
        </div>
      </DialogContent>
    </Dialog>
  );
}
