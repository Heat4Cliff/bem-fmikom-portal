"use client";

import { ArrowRight } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { Button } from "@/components/ui/button";
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
      <DialogContent aria-describedby="announcement-popup-description" className="w-[min(92vw,42rem)] p-0">
        {image ? (
          <div className="relative aspect-[16/7] overflow-hidden rounded-t-xl bg-surface-muted">
            <Image
              src={image.src}
              alt={image.alt}
              fill
              sizes="(max-width: 672px) 92vw, 672px"
              className="object-cover"
              unoptimized
            />
          </div>
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
