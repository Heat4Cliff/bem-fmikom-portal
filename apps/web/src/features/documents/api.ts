import { apiDownload, apiRequest, apiRequestEnvelope } from "@/lib/api/client";

export type ArchiveDocument = {
  id: number;
  period_id: number;
  period_name: string;
  division_id: number;
  division_name: string;
  uploaded_by: number;
  uploader_name: string;
  name: string;
  category: "SK" | "PROPOSAL" | "LPJ" | "OTHER";
  original_filename: string;
  mime_type: string;
  file_extension: string;
  size_bytes: number;
  sha256: string;
  created_at: string;
  updated_at: string;
};

export async function getDocuments(filters: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters))
    if (value !== undefined && value !== "") params.set(key, String(value));
  const result = await apiRequestEnvelope<ArchiveDocument[]>(`/admin/documents?${params}`);
  return { items: result.data ?? [], meta: result.meta ?? { page: 1, limit: 20, total: 0, total_pages: 0 } };
}

export function uploadDocument(body: FormData) {
  return apiRequest<ArchiveDocument>("/admin/documents", { method: "POST", body });
}

export function getDocument(id: number) {
  return apiRequest<ArchiveDocument>(`/admin/documents/${id}`);
}

export function updateDocument(id: number, body: { name: string; category: string }) {
  return apiRequest<ArchiveDocument>(`/admin/documents/${id}`, { method: "PATCH", body });
}

export function deleteDocument(id: number) {
  return apiRequest(`/admin/documents/${id}`, { method: "DELETE" });
}

export const downloadDocument = (id: number) => apiDownload(`/admin/documents/${id}/download`);
