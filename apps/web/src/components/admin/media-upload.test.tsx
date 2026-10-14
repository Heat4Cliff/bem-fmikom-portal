import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { isSupportedMediaFilename, MEDIA_UPLOAD_ACCEPT, MediaUpload } from "./media-upload";

const { apiRequest } = vi.hoisted(() => ({ apiRequest: vi.fn() }));
vi.mock("@/lib/api/client", () => ({ apiRequest }));
vi.mock("@/components/ui/media-image", () => ({
  MediaImage: ({ value, alt }: { value: string; alt: string }) => (
    <div role="img" aria-label={alt} data-src={value} />
  ),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("MediaUpload", () => {
  it("advertises the exact supported formats and web optimization policy", () => {
    render(
      <MediaUpload id="cover" label="Gambar sampul" purpose="ARTICLE_COVER" value="" onChange={vi.fn()} />,
    );
    expect(screen.getByLabelText("Pilih gambar sampul")).toHaveAttribute("accept", MEDIA_UPLOAD_ACCEPT);
    expect(screen.getByText("JPG, PNG, WebP, GIF, TIFF, BMP · Maks. 5 MiB")).toBeInTheDocument();
    expect(screen.getByText("TIFF dan BMP akan dioptimalkan otomatis untuk web.")).toBeInTheDocument();
    for (const filename of [
      "photo.jpg",
      "photo.jpeg",
      "photo.jfif",
      "photo.png",
      "photo.webp",
      "photo.gif",
      "photo.tif",
      "photo.tiff",
      "photo.bmp",
    ]) {
      expect(isSupportedMediaFilename(filename)).toBe(true);
    }
    expect(isSupportedMediaFilename("vector.svg")).toBe(false);
    expect(isSupportedMediaFilename("photo.heic")).toBe(false);
    expect(isSupportedMediaFilename("photo.avif")).toBe(false);
  });

  it("uploads an accessible file-picker selection and supports removing the preview", async () => {
    const onChange = vi.fn();
    apiRequest.mockResolvedValue({ public_url: "/api/v1/public/media/21" });
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: vi.fn(() => "blob:preview") });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: vi.fn() });
    const { rerender } = render(
      <MediaUpload
        id="cover"
        label="Gambar sampul"
        purpose="ARTICLE_COVER"
        value=""
        onChange={onChange}
        optional
      />,
    );

    const input = screen.getByLabelText("Pilih gambar sampul") as HTMLInputElement;
    const file = new File([new Uint8Array([1, 2])], "sampul.png", { type: "image/png" });
    fireEvent.change(input, { target: { files: [file] } });

    await waitFor(() =>
      expect(onChange).toHaveBeenCalledWith("http://localhost:8080/api/v1/public/media/21"),
    );
    rerender(
      <MediaUpload
        id="cover"
        label="Gambar sampul"
        purpose="ARTICLE_COVER"
        value="http://localhost:8080/api/v1/public/media/21"
        onChange={onChange}
        optional
      />,
    );
    expect(apiRequest).toHaveBeenCalledWith(
      "/admin/media",
      expect.objectContaining({ method: "POST", body: expect.any(FormData) }),
    );
    expect(screen.getByRole("img", { name: "Pratinjau gambar sampul" })).toBeInTheDocument();
    expect(screen.getByText("sampul.png")).toBeInTheDocument();
    expect(screen.getByText("0.00 MiB")).toBeInTheDocument();
    const selectedCard = screen
      .getByRole("img", { name: "Pratinjau gambar sampul" })
      .closest('[data-slot="media-upload-selected-card"]');
    expect(selectedCard).toHaveClass("grid", "min-w-0", "gap-4");
    const actions = screen.getByRole("group", { name: "Aksi gambar sampul" });
    expect(actions).toHaveClass("flex-col", "gap-3", "max-w-sm");
    expect(actions.querySelectorAll("button")).toHaveLength(3);
    expect(screen.getByRole("button", { name: "Ganti gambar" })).toHaveClass(
      "whitespace-nowrap",
      "min-h-10",
      "focus-visible:outline-2",
    );
    expect(screen.getByRole("button", { name: "Hapus" })).toHaveClass(
      "whitespace-nowrap",
      "min-h-10",
      "focus-visible:outline-2",
    );
    expect(screen.getByRole("button", { name: "Gunakan URL eksternal" })).toHaveClass(
      "whitespace-nowrap",
      "focus-visible:outline-2",
    );
    fireEvent.click(screen.getByRole("button", { name: "Gunakan URL eksternal" }));
    expect(screen.getByLabelText("URL gambar eksternal (opsional)")).toHaveAttribute("type", "url");
    expect(screen.getByRole("button", { name: "Kembali ke upload gambar" })).toHaveClass(
      "min-h-10",
      "whitespace-nowrap",
      "focus-visible:outline-2",
    );
    fireEvent.click(screen.getByRole("button", { name: "Kembali ke upload gambar" }));
    expect(screen.getByRole("group", { name: "Aksi gambar sampul" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Hapus" }));
    expect(onChange).toHaveBeenLastCalledWith("");
    rerender(
      <MediaUpload
        id="cover"
        label="Gambar sampul"
        purpose="ARTICLE_COVER"
        value=""
        onChange={onChange}
        optional
      />,
    );
    expect(screen.queryByRole("img", { name: "Pratinjau gambar sampul" })).not.toBeInTheDocument();
  });

  it("rejects unsupported formats before making an API request", () => {
    Object.defineProperty(URL, "createObjectURL", { configurable: true, value: vi.fn(() => "blob:preview") });
    render(
      <MediaUpload
        id="gallery"
        label="Gambar pratinjau"
        purpose="GALLERY_PREVIEW"
        value=""
        onChange={vi.fn()}
      />,
    );
    const file = new File(["<svg/>"], "bad.svg", { type: "image/svg+xml" });
    fireEvent.change(screen.getByLabelText("Pilih gambar pratinjau"), { target: { files: [file] } });
    expect(screen.getByRole("alert")).toHaveTextContent("Format gambar belum didukung.");
    expect(apiRequest).not.toHaveBeenCalled();
  });

  it("accepts TIFF by extension for server-side signature validation", async () => {
    apiRequest.mockResolvedValue({ public_url: "/api/v1/public/media/31" });
    render(
      <MediaUpload id="article" label="Gambar sampul" purpose="ARTICLE_COVER" value="" onChange={vi.fn()} />,
    );
    const file = new File([new Uint8Array([1, 2, 3])], "scan.tiff", { type: "" });
    fireEvent.change(screen.getByLabelText("Pilih gambar sampul"), { target: { files: [file] } });
    await waitFor(() => expect(apiRequest).toHaveBeenCalled());
  });

  it("reports the upload-in-progress boundary so the parent can block create submission", async () => {
    const onUploadStateChange = vi.fn();
    let finishUpload: ((media: { public_url: string }) => void) | undefined;
    apiRequest.mockImplementation(
      () =>
        new Promise((resolve) => {
          finishUpload = resolve;
        }),
    );
    render(
      <MediaUpload
        id="gallery"
        label="Gambar pratinjau"
        purpose="GALLERY_PREVIEW"
        value=""
        onChange={vi.fn()}
        onUploadStateChange={onUploadStateChange}
        optional
      />,
    );
    const file = new File([new Uint8Array([1, 2])], "preview.png", { type: "image/png" });
    fireEvent.change(screen.getByLabelText("Pilih gambar pratinjau"), { target: { files: [file] } });
    expect(onUploadStateChange).toHaveBeenLastCalledWith(true);
    finishUpload?.({ public_url: "/api/v1/public/media/9" });
    await waitFor(() => expect(onUploadStateChange).toHaveBeenLastCalledWith(false));
  });
});
