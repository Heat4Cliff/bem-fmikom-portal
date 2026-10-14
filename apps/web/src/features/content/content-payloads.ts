type AnnouncementDraft = {
  title: string;
  slug: string;
  content_html: string;
  image_url: string;
  start_at: string;
  end_at: string;
};

type GalleryItemDraft = {
  title: string;
  description: string;
  google_photos_url: string;
  preview_image_url: string;
  sort_order: string;
};

function persistedImageReference(value: string, field: string): string | null {
  const reference = value.trim();
  if (!reference) return null;
  if (reference.startsWith("blob:")) throw new Error(`${field} belum selesai diunggah.`);
  return reference;
}

export function announcementPayload(form: AnnouncementDraft) {
  return {
    ...form,
    image_url: persistedImageReference(form.image_url, "Gambar pengumuman"),
    start_at: new Date(form.start_at).toISOString(),
    end_at: form.end_at ? new Date(form.end_at).toISOString() : null,
  };
}

export function galleryItemPayload(form: GalleryItemDraft, categoryID: number) {
  if (!Number.isInteger(categoryID) || categoryID < 1)
    throw new Error("Pilih kategori galeri terlebih dahulu.");
  return {
    ...form,
    category_id: categoryID,
    description: form.description || null,
    preview_image_url: persistedImageReference(form.preview_image_url, "Gambar pratinjau"),
    sort_order: Number(form.sort_order) || 0,
  };
}
