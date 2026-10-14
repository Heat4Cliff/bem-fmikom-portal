import { describe, expect, it } from "vitest";
import { resolveMediaUrl } from "./resolve-media-url";

describe("resolveMediaUrl", () => {
  it("resolves internal media IDs and relative API references to the configured API origin", () => {
    expect(resolveMediaUrl("media:7")).toBe("http://localhost:8080/api/v1/public/media/7");
    expect(resolveMediaUrl("/api/v1/public/media/7")).toBe("http://localhost:8080/api/v1/public/media/7");
  });

  it("canonicalizes stale internal hosts without changing the media ID", () => {
    expect(resolveMediaUrl("http://127.0.0.1:8080/api/v1/public/media/7")).toBe(
      "http://localhost:8080/api/v1/public/media/7",
    );
  });

  it("preserves valid legacy HTTPS URLs and rejects unsafe or malformed sources", () => {
    expect(resolveMediaUrl("https://images.example.test/member.jpg")).toBe(
      "https://images.example.test/member.jpg",
    );
    expect(resolveMediaUrl(null)).toBeNull();
    expect(resolveMediaUrl("  ")).toBeNull();
    expect(resolveMediaUrl("http://untrusted.example.test/photo.jpg")).toBeNull();
    expect(resolveMediaUrl("javascript:alert(1)")).toBeNull();
    expect(resolveMediaUrl("C:\\media\\photo.jpg")).toBeNull();
  });

  it("allows object URLs only for temporary upload previews", () => {
    expect(resolveMediaUrl("blob:http://localhost/preview-id")).toBeNull();
    expect(resolveMediaUrl("blob:http://localhost/preview-id", true)).toBe(
      "blob:http://localhost/preview-id",
    );
  });
});
