import type * as React from "react";
import { cn } from "@/lib/utils/cn";

/** A fixed-ratio viewport for content images; image intrinsic dimensions never size the card. */
export function MediaFrame({
  children,
  className,
  ratio = "aspect-[16/9]",
}: {
  children: React.ReactNode;
  className?: string;
  ratio?: string;
}) {
  return (
    <div
      data-testid="media-frame"
      className={cn("relative isolate w-full min-w-0 overflow-hidden bg-surface-muted", ratio, className)}
    >
      {children}
    </div>
  );
}
