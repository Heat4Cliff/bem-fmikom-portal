import { describe, expect, it } from "vitest";
import { announcementPayload, galleryItemPayload } from "./content-payloads";

const start = "2026-10-12T10:30";

describe("content create payloads", () => {
  it("sends the completed announcement media reference and keeps empty optional image null", () => {
    const base = {
      title: "Info kegiatan",
      slug: "info-kegiatan",
      content_html: "<p>Isi</p>",
      start_at: start,
      end_at: "",
    };
    expect(
      announcementPayload({ ...base, image_url: "http://localhost:8080/api/v1/public/media/42" }),
    ).toMatchObject({
      image_url: "http://localhost:8080/api/v1/public/media/42",
      start_at: new Date(start).toISOString(),
      end_at: null,
    });
    expect(announcementPayload({ ...base, image_url: "" }).image_url).toBeNull();
    expect(() => announcementPayload({ ...base, image_url: "blob:http://localhost:3290/preview" })).toThrow(
      "belum selesai diunggah",
    );
  });

  it("keeps the external album URL separate from uploaded preview media and uses the selected category id", () => {
    const payload = galleryItemPayload(
      {
        title: "Dokumentasi",
        description: "",
        google_photos_url: "https://photos.app.goo.gl/album-id",
        preview_image_url: "/api/v1/public/media/54",
        sort_order: "2",
      },
      17,
    );
    expect(payload).toMatchObject({
      category_id: 17,
      google_photos_url: "https://photos.app.goo.gl/album-id",
      preview_image_url: "/api/v1/public/media/54",
      description: null,
      sort_order: 2,
    });
    expect(() =>
      galleryItemPayload(
        {
          title: "Dokumentasi",
          description: "",
          google_photos_url: "https://photos.app.goo.gl/album-id",
          preview_image_url: "blob:http://localhost:3290/preview",
          sort_order: "0",
        },
        17,
      ),
    ).toThrow("belum selesai diunggah");
    expect(() =>
      galleryItemPayload(
        {
          title: "Dokumentasi",
          description: "",
          google_photos_url: "https://photos.app.goo.gl/album-id",
          preview_image_url: "",
          sort_order: "0",
        },
        0,
      ),
    ).toThrow("Pilih kategori");
  });
});
