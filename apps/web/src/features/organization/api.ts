import { apiRequest, apiRequestEnvelope } from "@/lib/api/client";
import type { ApiPage, Division, OrganizationMember, Period } from "./types";

function query(values: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(values))
    if (value !== undefined && value !== "") params.set(key, String(value));
  return params.toString();
}

export async function organizationList<T>(path: string): Promise<ApiPage<T>> {
  const response = await apiRequestEnvelope<T[]>(path);
  return {
    items: response.data ?? [],
    meta: response.meta ?? { page: 1, limit: 20, total: 0, total_pages: 0 },
  };
}

export const getPeriods = (filters: Record<string, string | number | undefined> = {}) =>
  organizationList<Period>(`/admin/periods?${query(filters)}`);
export const savePeriod = (body: Pick<Period, "name" | "start_date" | "end_date">, id?: number) =>
  apiRequest<Period>(id ? `/admin/periods/${id}` : "/admin/periods", { method: id ? "PATCH" : "POST", body });
export const activatePeriod = (id: number) =>
  apiRequest(`/admin/periods/${id}/activate`, { method: "POST", body: {} });
export const archivePeriod = (id: number) =>
  apiRequest(`/admin/periods/${id}/archive`, { method: "POST", body: {} });

export const getDivisions = (filters: Record<string, string | number | undefined> = {}) =>
  organizationList<Division>(`/admin/divisions?${query(filters)}`);
export const saveDivision = (
  body: Omit<Division, "id" | "period_name" | "head" | "member_count">,
  id?: number,
) =>
  apiRequest<Division>(id ? `/admin/divisions/${id}` : "/admin/divisions", {
    method: id ? "PATCH" : "POST",
    body,
  });
export const deleteDivision = (id: number) =>
  apiRequest<null>(`/admin/divisions/${id}`, { method: "DELETE" });

export const getMembers = (filters: Record<string, string | number | undefined> = {}) =>
  organizationList<OrganizationMember>(`/admin/members?${query(filters)}`);
export const getMember = (id: number) => apiRequest<OrganizationMember>(`/admin/members/${id}`);
export const saveMember = (
  body: Omit<OrganizationMember, "id" | "period_name" | "division_name">,
  id?: number,
) =>
  apiRequest<OrganizationMember>(id ? `/admin/members/${id}` : "/admin/members", {
    method: id ? "PATCH" : "POST",
    body,
  });
export const deleteMember = (id: number) => apiRequest<null>(`/admin/members/${id}`, { method: "DELETE" });
