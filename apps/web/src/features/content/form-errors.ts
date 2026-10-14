import { ApiError } from "@/lib/api/client";

const fieldLabels: Record<string, string> = {
  title: "Judul",
  slug: "Slug",
  content_html: "Isi",
  image_url: "Format gambar",
  start_at: "Tanggal mulai",
  end_at: "Tanggal selesai",
  description: "Deskripsi",
  category_id: "Kategori",
  google_photos_url: "Tautan album Google Photos",
  preview_image_url: "Gambar pratinjau",
  event_date: "Tanggal kegiatan",
};

export function contentFormErrorMessage(error: unknown, fallback: string): string {
  if (!(error instanceof ApiError)) return fallback;

  const fieldMessages = Object.entries(error.errors ?? {}).flatMap(([field, messages]) =>
    messages.map((message) => `${fieldLabels[field] ?? "Data"}: ${message}`),
  );
  if (fieldMessages.length) return fieldMessages.join(" ");
  if (error.code === "VALIDATION_ERROR") return "Periksa data yang dimasukkan, lalu coba lagi.";
  if (error.code === "CONFLICT") return error.message;
  return fallback;
}
