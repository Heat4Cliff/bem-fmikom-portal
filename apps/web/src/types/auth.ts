export type AdminRole = "SUPER_ADMIN" | "BPH" | "KEPALA_BIDANG";

/** Typed shell context only. Phase 1 does not enforce permissions or scope. */
export interface AdminNavigationContext {
  role: AdminRole | null;
  divisionId?: number | null;
  permissions?: readonly string[];
}
