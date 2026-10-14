"use client";

import type * as React from "react";
import { useOnceInView } from "./use-once-in-view";

/**
 * Lightweight, reduced-motion-safe adaptation of React Bits Fade Content:
 * https://reactbits.dev/animations/fade-content
 * Static children remain visible in server HTML and when IntersectionObserver
 * is unavailable. Animation uses stylesheet classes, respecting the nonce-only
 * style CSP and the user's prefers-reduced-motion setting.
 */
export function FadeContent({ children, className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const { ref, inView } = useOnceInView<HTMLDivElement>();

  return (
    <div
      ref={ref}
      className={`${className ?? ""} ${inView ? "react-bits-fade-visible" : ""}`.trim()}
      {...props}
    >
      {children}
    </div>
  );
}
