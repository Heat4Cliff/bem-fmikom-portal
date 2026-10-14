"use client";

import { useMotionValue, useReducedMotion, useSpring } from "motion/react";
import * as React from "react";
import { useOnceInView } from "./use-once-in-view";

/** Adapted from the free React Bits Count Up source: https://reactbits.dev/text-animations/count-up */
export function CountUp({ to, className }: { to: number; className?: string }) {
  const shouldReduceMotion = useReducedMotion();
  const { ref, inView } = useOnceInView<HTMLSpanElement>();
  const value = useMotionValue(0);
  const spring = useSpring(value, { damping: 28, stiffness: 120 });
  const [display, setDisplay] = React.useState(() => to.toLocaleString("id-ID"));

  React.useEffect(() => {
    if (shouldReduceMotion || !inView) return;
    setDisplay("0");
    const timeout = window.setTimeout(() => value.set(to), 60);
    return () => window.clearTimeout(timeout);
  }, [inView, shouldReduceMotion, to, value]);

  React.useEffect(() => {
    return spring.on("change", (latest) => setDisplay(Math.round(latest).toLocaleString("id-ID")));
  }, [spring]);

  return (
    <>
      <span ref={ref} className={className} aria-hidden="true">
        {display}
      </span>
      <span className="sr-only">{to.toLocaleString("id-ID")}</span>
    </>
  );
}
