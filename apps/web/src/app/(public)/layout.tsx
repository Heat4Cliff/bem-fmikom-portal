import type { ReactNode } from "react";
import { PublicShell } from "@/components/public/public-shell";

export default function PublicLayout({ children }: Readonly<{ children: ReactNode }>) {
  return <PublicShell>{children}</PublicShell>;
}
