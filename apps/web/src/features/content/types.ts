export type PageMeta = { page: number; limit: number; total: number; total_pages: number };
export type ApiList<T> = { items: T[]; meta: PageMeta };

export type ContentCategory = { id: number; name: string; slug: string; description?: string | null };
export type Article = {
  id: number;
  title: string;
  slug: string;
  excerpt?: string | null;
  content_html: string;
  thumbnail_url?: string | null;
  division_id?: number | null;
  division_name?: string | null;
  status: "DRAFT" | "PUBLISHED";
  published_at?: string | null;
  created_at: string;
  updated_at: string;
  categories: ContentCategory[];
};
export type GalleryItem = {
  id: number;
  category_id: number;
  title: string;
  description?: string | null;
  google_photos_url: string;
  preview_image_url?: string | null;
  sort_order: number;
};
export type GalleryCategory = {
  id: number;
  title: string;
  slug: string;
  description?: string | null;
  period_id?: number | null;
  division_id?: number | null;
  division_name?: string | null;
  event_date?: string | null;
  item_count?: number;
  items?: GalleryItem[];
};
export type Announcement = {
  id: number;
  title: string;
  slug: string;
  content_html: string;
  image_url?: string | null;
  start_at: string;
  end_at?: string | null;
  is_active: boolean;
  status: "INACTIVE" | "SCHEDULED" | "ACTIVE" | "EXPIRED";
};
