import { describe, expect, it } from "vitest";
import { ApiError } from "@/lib/api/client";
import { contentFormErrorMessage } from "./form-errors";

describe("content form API errors", () => {
  it("renders field-level validation details with user-facing field labels", () => {
    const error = new ApiError("Validation failed", 422, "VALIDATION_ERROR", {
      end_at: ["Tanggal selesai harus setelah tanggal mulai."],
      image_url: ["Format gambar tidak valid."],
    });

    expect(contentFormErrorMessage(error, "Gagal menyimpan.")).toBe(
      "Tanggal selesai: Tanggal selesai harus setelah tanggal mulai. Format gambar: Format gambar tidak valid.",
    );
  });

  it("uses a safe fallback for server errors and non-API failures", () => {
    expect(contentFormErrorMessage(new Error("SQL at /var/private"), "Pengumuman gagal disimpan.")).toBe(
      "Pengumuman gagal disimpan.",
    );
    expect(
      contentFormErrorMessage(new ApiError("Validation failed", 422, "VALIDATION_ERROR"), "fallback"),
    ).toBe("Periksa data yang dimasukkan, lalu coba lagi.");
  });
});
