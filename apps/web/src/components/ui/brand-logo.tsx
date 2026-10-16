import { cn } from "@/lib/utils/cn";
import { MediaImage } from "./media-image";
import type { LogoFit, LogoShape } from "@/features/content/api";

const shapeClasses: Record<LogoShape, string> = {
  original: "rounded-none",
  rounded: "rounded-xl",
  circle: "rounded-full",
};

export function BrandLogo({
  src,
  shape = "original",
  fit = "contain",
  alt = "BEM FMIKOM UNUGHA",
  className,
  fallbackText = "BEM",
}: {
  src?: string | null;
  shape?: LogoShape;
  fit?: LogoFit;
  alt?: string;
  className?: string;
  fallbackText?: string;
}) {
  const safeShape: LogoShape = shapeClasses[shape] ? shape : "original";
  const safeFit = fit === "cover" ? "cover" : "contain";
  return (
    <span
      className={cn(
        "relative inline-grid size-10 shrink-0 place-items-center overflow-hidden bg-surface-muted text-primary",
        shapeClasses[safeShape],
        className,
      )}
      data-logo-shape={safeShape}
      data-logo-fit={safeFit}
    >
      {src ? (
        <MediaImage
          value={src}
          alt={alt}
          width={160}
          height={160}
          className={cn("size-full", safeFit === "cover" ? "object-cover" : "object-contain")}
          fallbackClassName="bg-primary font-heading text-xs font-extrabold text-primary-foreground"
          fallback={<span aria-hidden="true">{fallbackText}</span>}
        />
      ) : (
        <span
          role="img"
          aria-label={alt || undefined}
          aria-hidden={alt ? undefined : true}
          className="grid size-full place-items-center bg-primary font-heading text-xs font-extrabold text-primary-foreground"
        >
          {fallbackText}
        </span>
      )}
    </span>
  );
}
