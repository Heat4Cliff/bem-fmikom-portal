import {
  apiRequest,
  apiRequestEnvelope,
  ApiError,
  getMemoryAccessToken,
  refreshAccessToken,
} from "@/lib/api/client";
import type {
  PageMeta,
  PublicRegistration,
  PublicRegistrationPreview,
  RegistrationField,
  RegistrationForm,
  RegistrationSubmission,
} from "./types";

function query(values: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(values))
    if (value !== undefined && value !== "") params.set(key, String(value));
  return params.toString();
}

function page<T>(items: T[], meta?: PageMeta) {
  return { items: items ?? [], meta: meta ?? { page: 1, limit: 20, total: 0, total_pages: 0 } };
}

export async function getPublicRegistrations(filters: Record<string, string | number | undefined> = {}) {
  const response = await apiRequestEnvelope<PublicRegistrationPreview[]>(
    `/public/registrations?${query(filters)}`,
    { retryOnUnauthorized: false },
  );
  return page(response.data, response.meta);
}
export function getPublicRegistration(slug: string) {
  return apiRequest<PublicRegistration>(`/public/registrations/${encodeURIComponent(slug)}`, {
    retryOnUnauthorized: false,
  });
}
export function submitRegistration(
  id: number,
  answers: Record<string, unknown>,
  images: Record<string, File> = {},
) {
  if (!Object.keys(images).length) {
    return apiRequest<{ submission_id: number; submitted_at: string; registration_title: string }>(
      `/public/registrations/${id}/submissions`,
      { method: "POST", body: { answers }, retryOnUnauthorized: false },
    );
  }
  const body = new FormData();
  body.set("answers", JSON.stringify(answers));
  for (const [fieldKey, file] of Object.entries(images)) body.set(`image:${fieldKey}`, file, file.name);
  return apiRequest<{ submission_id: number; submitted_at: string; registration_title: string }>(
    `/public/registrations/${id}/submissions`,
    { method: "POST", body, retryOnUnauthorized: false },
  );
}
export async function getRegistrations(filters: Record<string, string | number | undefined> = {}) {
  const response = await apiRequestEnvelope<RegistrationForm[]>(`/admin/registrations?${query(filters)}`);
  return page(response.data, response.meta);
}
export function getRegistration(id: number) {
  return apiRequest<{ form: RegistrationForm; fields: RegistrationField[] }>(`/admin/registrations/${id}`);
}
export function createRegistration(input: Partial<RegistrationForm>) {
  return apiRequest<RegistrationForm>("/admin/registrations", { method: "POST", body: input });
}
export function updateRegistration(id: number, input: Partial<RegistrationForm>) {
  return apiRequest<RegistrationForm>(`/admin/registrations/${id}`, { method: "PATCH", body: input });
}
export function saveRegistrationFields(id: number, fields: RegistrationField[]) {
  const input = fields.map((field) => ({
    field_key: field.field_key,
    label: field.label,
    field_type: field.field_type,
    placeholder: field.placeholder,
    required: field.required,
    options: field.options.map(({ value, label }) => ({ value, label })),
  }));
  return apiRequest<RegistrationField[]>(`/admin/registrations/${id}/fields`, { method: "PUT", body: input });
}
export function setRegistrationStatus(id: number, status: "open" | "close") {
  return apiRequest<RegistrationForm>(`/admin/registrations/${id}/${status}`, { method: "POST" });
}
export function deleteRegistration(id: number) {
  return apiRequest<null>(`/admin/registrations/${id}`, { method: "DELETE" });
}
export async function getRegistrationSubmissions(
  id: number,
  filters: Record<string, string | number | undefined> = {},
) {
  const response = await apiRequestEnvelope<RegistrationSubmission[]>(
    `/admin/registrations/${id}/submissions?${query(filters)}`,
  );
  return page(response.data, response.meta);
}
export function getRegistrationSubmission(id: number, submissionId: number) {
  return apiRequest<RegistrationSubmission>(`/admin/registrations/${id}/submissions/${submissionId}`);
}
export async function getRegistrationAttachmentBlob(downloadPath: string) {
  let token = getMemoryAccessToken();
  const url = `${(process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1").replace(/\/$/, "")}${downloadPath}`;
  let response = await fetch(url, {
    credentials: "include",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (response.status === 401 && token) {
    token = await refreshAccessToken();
    response = await fetch(url, { credentials: "include", headers: { Authorization: `Bearer ${token}` } });
  }
  if (!response.ok) throw new ApiError("Lampiran tidak dapat dimuat", response.status);
  return response.blob();
}
export async function downloadRegistrationExport(id: number, search = "") {
  let token = getMemoryAccessToken();
  const url = `${(process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1").replace(/\/$/, "")}/admin/registrations/${id}/export?${query({ search })}`;
  let response = await fetch(url, {
    credentials: "include",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (response.status === 401 && token) {
    token = await refreshAccessToken();
    response = await fetch(url, { credentials: "include", headers: { Authorization: `Bearer ${token}` } });
  }
  if (!response.ok) {
    let message = "Ekspor gagal";
    let code: string | undefined;
    try {
      const body = (await response.json()) as { message?: string; code?: string };
      message = body.message ?? message;
      code = body.code;
    } catch {
      /* non-JSON proxy response */
    }
    throw new ApiError(message, response.status, code);
  }
  return response.blob();
}
