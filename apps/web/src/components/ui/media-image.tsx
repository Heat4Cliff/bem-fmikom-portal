"use client";

import { UserRound } from "lucide-react";
import Image from "next/image";
import * as React from "react";
import { resolveMediaUrl } from "@/lib/media/resolve-media-url";
import { cn } from "@/lib/utils/cn";

type MediaImageProps = {
  value: string | null | undefined;
  alt: string;
  width?: number;
  height?: number;
  fill?: boolean;
  sizes?: string;
  className?: string;
  fallbackClassName?: string;
  fallbackText?: string;
  fallback?: React.ReactNode;
  allowBlobPreview?: boolean;
};

export function MediaImage({
  value,
  alt,
  width = 1200,
  height = 800,
  fill = false,
  sizes,
  className,
  fallbackClassName,
  fallbackText,
  fallback,
  allowBlobPreview = false,
}: MediaImageProps) {
  const src = resolveMediaUrl(value, allowBlobPreview);
  const [failedSource, setFailedSource] = React.useState<string | null>(null);

  const failed = !src || failedSource === src;
  if (failed) {
    return (
      <span
        role="img"
        aria-label={alt ? `${alt} tidak tersedia` : "Gambar tidak tersedia"}
        className={cn(
          "inline-grid shrink-0 place-items-center overflow-hidden",
          className,
          fallbackClassName,
        )}
      >
        {fallback ??
          (fallbackText ? (
            <span aria-hidden="true" className="font-semibold">
              {fallbackText}
            </span>
          ) : (
            <UserRound aria-hidden="true" className="size-5" />
          ))}
      </span>
    );
  }

  const onError = () => {
    setFailedSource(src);
    if (process.env.NODE_ENV !== "production") {
      try {
        const failedURL = new URL(src, window.location.origin);
        console.error("[media] Gambar gagal dimuat", `${failedURL.origin}${failedURL.pathname}`);
      } catch {
        console.error("[media] Gambar gagal dimuat", "URL tidak valid");
      }
    }
  };

  if (fill) {
    return (
      <Image
        src={src}
        alt={alt}
        fill
        sizes={sizes ?? "100vw"}
        unoptimized
        onError={onError}
        className={className}
      />
    );
  }

  return (
    <Image
      src={src}
      alt={alt}
      width={width}
      height={height}
      unoptimized
      onError={onError}
      className={className}
    />
  );
}
