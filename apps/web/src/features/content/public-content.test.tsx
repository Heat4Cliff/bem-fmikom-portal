import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { AnnouncementCard, ArticleCard, ContentPagination, GalleryCard } from "./public-content";
import type { Announcement, Article, GalleryCategory } from "./types";

afterEach(cleanup);

const article: Article = {
  id: 4,
  title: "Berita Resmi",
  slug: "berita-resmi",
  excerpt: "Ringkasan berita",
  content_html: "<p>Konten berita</p>",
  status: "PUBLISHED",
  published_at: "2026-10-09T08:00:00Z",
  created_at: "2026-10-08T08:00:00Z",
  updated_at: "2026-10-09T08:00:00Z",
  categories: [{ id: 1, name: "Kegiatan", slug: "kegiatan" }],
};

describe("public content cards", () => {
  it("links a published article card to the canonical detail route", () => {
    render(<ArticleCard article={article} />);
    expect(screen.getByRole("heading", { name: "Berita Resmi" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Baca selengkapnya" })).toHaveAttribute(
      "href",
      "/berita/berita-resmi",
    );
    expect(screen.getByText("Kegiatan")).toBeInTheDocument();
  });

  it("links announcements to public detail and uses their publication date", () => {
    const announcement: Announcement = {
      id: 5,
      title: "Jadwal Kegiatan",
      slug: "jadwal-kegiatan",
      content_html: "<p>Informasi jadwal.</p>",
      start_at: "2026-10-10T01:00:00Z",
      is_active: true,
      status: "ACTIVE",
    };
    render(<AnnouncementCard announcement={announcement} />);
    expect(screen.getByRole("link", { name: /Baca pengumuman/ })).toHaveAttribute(
      "href",
      "/pengumuman/jadwal-kegiatan",
    );
  });

  it("opens album links in a safe new context and paginates with search", () => {
    const category: GalleryCategory = {
      id: 7,
      title: "Seminar FMIKOM",
      slug: "seminar-fmikom",
      item_count: 1,
      items: [
        {
          id: 1,
          category_id: 7,
          title: "Album Seminar",
          google_photos_url: "https://photos.app.goo.gl/example",
          sort_order: 1,
        },
      ],
    };
    render(
      <>
        <GalleryCard category={category} />
        <ContentPagination pathname="/berita" page={2} totalPages={4} search="ui ux" />
      </>,
    );
    expect(screen.getByRole("link", { name: "Lihat galeri Seminar FMIKOM" })).toHaveAttribute(
      "href",
      "/galeri/seminar-fmikom",
    );
    expect(screen.getByRole("link", { name: "Selanjutnya" })).toHaveAttribute(
      "href",
      "/berita?page=3&search=ui%20ux",
    );
  });
});
