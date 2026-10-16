"use client";

import * as React from "react";
import type { LogoFit, LogoShape } from "./api";

export type BrandingPresentation = {
  logoUrl: string | null;
  logoShape: LogoShape;
  logoFit: LogoFit;
};

const defaultBranding: BrandingPresentation = {
  logoUrl: null,
  logoShape: "original",
  logoFit: "contain",
};

const BrandingContext = React.createContext<BrandingPresentation>(defaultBranding);

export function BrandingProvider({
  value,
  children,
}: {
  value: BrandingPresentation;
  children: React.ReactNode;
}) {
  return <BrandingContext.Provider value={value}>{children}</BrandingContext.Provider>;
}

export function useBranding() {
  return React.useContext(BrandingContext);
}
