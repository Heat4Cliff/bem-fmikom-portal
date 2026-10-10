import { apiRequest, apiRequestEnvelope } from "@/lib/api/client";
import type { Aspiration, AspirationDivision, AspirationPage, AspirationReceipt } from "./types";

function query(values: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== "") params.set(key, String(value));
  }
  return params.toString();
}

export function getPublicAspirationDivisions() {
  return apiRequest<AspirationDivision[]>("/public/divisions", { retryOnUnauthorized: false });
}

export function submitAspiration(input: {
  division_id: number;
  name?: string;
  nim?: string;
  email?: string;
  subject: string;
  message: string;
  is_anonymous: boolean;
}) {
  return apiRequest<AspirationReceipt>("/public/aspirations", {
    method: "POST",
    body: input,
    retryOnUnauthorized: false,
  });
}

export async function getAspirations(filters: Record<string, string | number | undefined>) {
  const response = await apiRequestEnvelope<Aspiration[]>(`/admin/aspirations?${query(filters)}`);
  return {
    items: response.data ?? [],
    meta: response.meta ?? { page: 1, limit: 20, total: 0, total_pages: 0 },
  } satisfies AspirationPage;
}

export function getAspiration(id: number) {
  return apiRequest<Aspiration>(`/admin/aspirations/${id}`);
}

export function updateAspirationStatus(id: number, status: "READ" | "RESOLVED") {
  return apiRequest<Aspiration>(`/admin/aspirations/${id}/status`, { method: "PATCH", body: { status } });
}
