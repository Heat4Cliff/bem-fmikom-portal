import { apiRequest, apiRequestEnvelope } from "@/lib/api/client";
import type { PageMeta } from "@/features/content/types";
import type { PublicLandingSettings } from "@/features/content/api";

export type LandingSettings = PublicLandingSettings & { updated_at?: string };
export type ManagedUser = {
  id: number;
  name: string;
  email: string;
  role: "SUPER_ADMIN" | "BPH" | "KEPALA_BIDANG";
  division_id: number | null;
  division: { id: number; name: string } | null;
  status: "ACTIVE" | "INACTIVE";
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
};
export type Activity = {
  id: number;
  timestamp: string;
  actor: { id: number; name: string; email: string; role: string } | null;
  action: string;
  entity_type: string;
  entity_id: string | null;
  details: unknown;
};
export type DashboardSummary = {
  generated_at: string;
  scope: { type: "organization" | "division"; division: { id: number; name: string } | null };
  active_period: { id: number; name: string; start_date: string; end_date: string } | null;
  articles: { total: number; published: number; draft: number };
  registrations: { active_forms: number; active_form_submissions: number };
  aspirations: { total: number; pending: number; read: number; resolved: number };
  documents: { total: number };
  announcements: { total: number };
};
export type Filters = Record<string, string | number | undefined>;
function query(filters: Filters) {
  const p = new URLSearchParams();
  for (const [key, value] of Object.entries(filters))
    if (value !== undefined && value !== "") p.set(key, String(value));
  return p.toString();
}
export const getSettings = () => apiRequest<LandingSettings>("/admin/settings");
export const getDashboard = () => apiRequest<DashboardSummary>("/admin/dashboard");
export const saveSettings = (value: LandingSettings) =>
  apiRequest<LandingSettings>("/admin/settings", { method: "PATCH", body: value });
export async function getUsers(filters: Filters) {
  const r = await apiRequestEnvelope<ManagedUser[]>(`/admin/users?${query(filters)}`);
  return {
    items: r.data ?? [],
    meta: r.meta ?? ({ page: 1, limit: 20, total: 0, total_pages: 0 } satisfies PageMeta),
  };
}
export const getUser = (id: number) => apiRequest<ManagedUser>(`/admin/users/${id}`);
export const createUser = (value: Record<string, unknown>) =>
  apiRequest<ManagedUser>("/admin/users", { method: "POST", body: value });
export const updateUser = (id: number, value: Record<string, unknown>) =>
  apiRequest<ManagedUser>(`/admin/users/${id}`, { method: "PATCH", body: value });
export const deactivateUser = (id: number) =>
  apiRequest(`/admin/users/${id}/deactivate`, { method: "POST", body: {} });
export const reactivateUser = (id: number) =>
  apiRequest(`/admin/users/${id}/reactivate`, { method: "POST", body: {} });
export const resetUserPassword = (id: number, password: string, confirm_password: string) =>
  apiRequest(`/admin/users/${id}/reset-password`, { method: "POST", body: { password, confirm_password } });
export async function getActivities(filters: Filters) {
  const r = await apiRequestEnvelope<Activity[]>(`/admin/activity-logs?${query(filters)}`);
  return {
    items: r.data ?? [],
    meta: r.meta ?? ({ page: 1, limit: 20, total: 0, total_pages: 0 } satisfies PageMeta),
  };
}
export const getActivity = (id: number) => apiRequest<Activity>(`/admin/activity-logs/${id}`);
