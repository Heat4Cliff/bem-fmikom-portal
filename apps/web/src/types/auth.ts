export type AdminRole = "SUPER_ADMIN" | "BPH" | "KEPALA_BIDANG";

export interface AuthDivision {
  id: number;
  name: string;
}

export interface AuthUser {
  id: number;
  name: string;
  email: string;
  role: AdminRole;
  division: AuthDivision | null;
  permissions?: string[];
}

export interface AdminNavigationContext {
  role: AdminRole | null;
  divisionId?: number | null;
  permissions?: readonly string[];
}
