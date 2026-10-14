import { headers } from "next/headers";
import type { ReactNode } from "react";

export const dynamic = "force-dynamic";

export default async function AdminLoginLayout({ children }: Readonly<{ children: ReactNode }>) {
  // The per-request CSP nonce is propagated by the proxy during dynamic rendering.
  await headers();
  return children;
}
