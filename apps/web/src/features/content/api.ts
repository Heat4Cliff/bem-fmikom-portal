import { apiRequest, apiRequestEnvelope, type RequestOptions } from "@/lib/api/client";
import type {
  Announcement,
  Article,
  ContentCategory,
  GalleryCategory,
  GalleryItem,
  PageMeta,
  PublicOrganization,
} from "./types";

type Envelope<T> = { success: boolean; message: string; data: T; code?: string; meta?: PageMeta };
const apiBase = (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1").replace(/\/$/, "");

async function publicGet<T>(path: string): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    cache: "no-store",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) throw new Error(`Public API returned ${response.status}`);
  const envelope = (await response.json()) as Envelope<T>;
  if (!envelope.success) throw new Error(envelope.message || "Public API request failed");
  return envelope.data;
}

async function publicList<T>(path: string): Promise<T[]> {
  const response = await fetch(`${apiBase}${path}`, {
    cache: "no-store",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) throw new Error(`Public API returned ${response.status}`);
  const envelope = (await response.json()) as Envelope<T[]>;
  if (!envelope.success) throw new Error(envelope.message || "Public API request failed");
  return envelope.data ?? [];
}

async function publicPage<T>(path: string): Promise<{ items: T[]; meta: PageMeta }> {
  const response = await fetch(`${apiBase}${path}`, {
    cache: "no-store",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) throw new Error(`Public API returned ${response.status}`);
  const envelope = (await response.json()) as Envelope<T[]>;
  if (!envelope.success) throw new Error(envelope.message || "Public API request failed");
  return {
    items: envelope.data ?? [],
    meta: envelope.meta ?? { page: 1, limit: 20, total: 0, total_pages: 0 },
  };
}

export const getPublicArticles = (page = 1, limit = 20, search = "") =>
  publicList<Article>(`/public/articles?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`);
export const getPublicArticlePage = (page = 1, limit = 20, search = "") =>
  publicPage<Article>(`/public/articles?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`);
export const getPublicArticle = (slug: string) =>
  publicGet<Article>(`/public/articles/${encodeURIComponent(slug)}`);
export const getPublicAnnouncements = (page = 1, limit = 20, search = "") =>
  publicList<Announcement>(
    `/public/announcements?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`,
  );
export const getPublicAnnouncementPage = (page = 1, limit = 20, search = "") =>
  publicPage<Announcement>(
    `/public/announcements?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`,
  );
export const getActiveAnnouncements = (page = 1, limit = 20) =>
  publicList<Announcement>(`/public/announcements/active?page=${page}&limit=${limit}`);
export const getPublicAnnouncement = (slug: string) =>
  publicGet<Announcement>(`/public/announcements/${encodeURIComponent(slug)}`);
export const getPublicGallery = (page = 1, limit = 20) =>
  publicList<GalleryCategory>(`/public/gallery-categories?page=${page}&limit=${limit}`);
export const getPublicGalleryPage = (page = 1, limit = 20) =>
  publicPage<GalleryCategory>(`/public/gallery-categories?page=${page}&limit=${limit}`);
export const getPublicGalleryCategory = (slug: string) =>
  publicGet<GalleryCategory>(`/public/gallery-categories/${encodeURIComponent(slug)}`);
export const getPublicOrganization = () => publicGet<PublicOrganization>("/public/organization");

export type ActivePublicRegistration = {
  title: string;
  slug: string;
  description: string | null;
  period_name: string;
  division_name: string | null;
  open_at: string;
  close_at: string;
  public_status: "OPEN";
};

// Opening/closing windows are time-dependent; always ask the API for the current UTC state.
export const getActivePublicRegistrations = (limit = 6) =>
  publicList<ActivePublicRegistration>(`/public/registrations?status=OPEN&page=1&limit=${limit}`);

export type PublicLandingSettings = {
  hero_headline: string;
  hero_subheadline: string;
  hero_primary_cta_label: string;
  hero_primary_cta_url: string;
  hero_secondary_cta_label: string | null;
  hero_secondary_cta_url: string | null;
  hero_image_url: string | null;
  about_title: string;
  about_description: string;
  vision: string;
  missions: string[];
  contact_email: string | null;
  contact_phone: string | null;
  contact_address: string | null;
  location_url: string | null;
  instagram_url: string | null;
  youtube_url: string | null;
  tiktok_url: string | null;
  other_social_url: string | null;
};

export const getPublicLandingSettings = () => publicGet<PublicLandingSettings>("/public/settings");

export async function adminList<T>(
  path: string,
  options: RequestOptions = {},
): Promise<{ items: T[]; meta: PageMeta }> {
  const response = await apiRequestEnvelope<T[]>(path, options);
  return {
    items: response.data,
    meta: response.meta ?? {
      page: 1,
      limit: response.data.length,
      total: response.data.length,
      total_pages: 1,
    },
  };
}

export const getAdminCategories = () => apiRequest<ContentCategory[]>("/admin/article-categories");
export const getAdminGalleryItems = (categoryID?: number) =>
  adminList<GalleryItem>(`/admin/gallery-items?limit=100${categoryID ? `&category_id=${categoryID}` : ""}`);
