"use client";

import * as React from "react";
import { useOnceInView } from "./use-once-in-view";

/**
 * Adapted from React Bits Blur Text (free TS + Tailwind source):
 * https://reactbits.dev/text-animations/blur-text
 * Keeps a plain, server-rendered heading until it is in view, then applies
 * stylesheet-based word animations. This avoids inline styles under the app's
 * nonce-only style CSP and preserves visible content when JS is unavailable.
 */
export function BlurText({ text, className }: { text: string; className: string }) {
  const { ref, inView } = useOnceInView<HTMLHeadingElement>();
  const segmentOccurrences = new Map<string, number>();
  let wordOrder = 0;

  function segmentKey(segment: string) {
    const occurrence = segmentOccurrences.get(segment) ?? 0;
    segmentOccurrences.set(segment, occurrence + 1);
    return `${segment}-${occurrence}`;
  }

  if (!inView) {
    return (
      <h1 ref={ref} className={className}>
        {text}
      </h1>
    );
  }

  return (
    <h1 ref={ref} aria-label={text} className={className}>
      {text.split(/(\s+)/).map((part) => {
        const key = segmentKey(part);
        return /^\s+$/.test(part) ? (
          <React.Fragment key={key}>{part}</React.Fragment>
        ) : (
          <span
            key={key}
            aria-hidden="true"
            className={`react-bits-blur-word react-bits-blur-delay-${Math.min(wordOrder++, 12)}`}
          >
            {part}
          </span>
        );
      })}
    </h1>
  );
}
