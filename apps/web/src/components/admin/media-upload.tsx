"use client";

import { ImagePlus, LoaderCircle, Trash2, Upload } from "lucide-react";
import * as React from "react";
import { MediaImage } from "@/components/ui/media-image";
import { apiRequest } from "@/lib/api/client";
import { resolveMediaUrl } from "@/lib/media/resolve-media-url";
import { FadeContent } from "@/components/react-bits/fade-content";
import { ActionGroup } from "@/components/ui/primitives";

export type MediaPurpose =
  | "ARTICLE_COVER"
  | "ANNOUNCEMENT_IMAGE"
  | "GALLERY_PREVIEW"
  | "MEMBER_PHOTO"
  | "LANDING_HERO"
  | "LANDING_BRANDING";

type UploadedMedia = { public_url: string };

export const MEDIA_UPLOAD_ACCEPT =
  "image/jpeg,image/png,image/webp,image/gif,image/tiff,image/bmp,.jpg,.jpeg,.jfif,.png,.webp,.gif,.tif,.tiff,.bmp";

const supportedExtensions = new Set([
  ".jpg",
  ".jpeg",
  ".jfif",
  ".png",
  ".webp",
  ".gif",
  ".tif",
  ".tiff",
  ".bmp",
]);

export function isSupportedMediaFilename(filename: string) {
  const extension = filename.slice(filename.lastIndexOf(".")).toLowerCase();
  return supportedExtensions.has(extension);
}

export function MediaUpload({
  id,
  label,
  purpose,
  value,
  onChange,
  onUploadStateChange,
  disabled = false,
  optional = false,
  compact = false,
}: {
  id: string;
  label: string;
  purpose: MediaPurpose;
  value: string;
  onChange: (url: string) => void;
  onUploadStateChange?: (busy: boolean) => void;
  disabled?: boolean;
  optional?: boolean;
  compact?: boolean;
}) {
  const inputRef = React.useRef<HTMLInputElement>(null);
  const [busy, setBusy] = React.useState(false);
  const [dragActive, setDragActive] = React.useState(false);
  const [error, setError] = React.useState("");
  const [externalURL, setExternalURL] = React.useState(false);
  const [localPreview, setLocalPreview] = React.useState("");
  const [filename, setFilename] = React.useState("");
  const [size, setSize] = React.useState<number | null>(null);

  React.useEffect(
    () => () => {
      if (localPreview) URL.revokeObjectURL(localPreview);
    },
    [localPreview],
  );

  function select(file: File | undefined) {
    setError("");
    if (!file) return;
    if (!isSupportedMediaFilename(file.name)) {
      setError("Format gambar belum didukung.");
      if (inputRef.current) inputRef.current.value = "";
      return;
    }
    if (file.size < 1 || file.size > 5 * 1024 * 1024) {
      setError("Ukuran gambar melebihi batas 5 MiB.");
      if (inputRef.current) inputRef.current.value = "";
      return;
    }
    const preview = URL.createObjectURL(file);
    setLocalPreview(preview);
    setFilename(file.name);
    setSize(file.size);
    setBusy(true);
    onUploadStateChange?.(true);
    const body = new FormData();
    body.set("purpose", purpose);
    body.set("file", file);
    void apiRequest<UploadedMedia>("/admin/media", { method: "POST", body })
      .then((media) => {
        onChange(resolveMediaUrl(media.public_url) ?? "");
        setExternalURL(false);
        setLocalPreview("");
      })
      .catch((uploadError: unknown) => {
        setError(uploadError instanceof Error ? uploadError.message : "Gambar gagal diunggah.");
        setLocalPreview("");
        setFilename("");
        setSize(null);
      })
      .finally(() => {
        setBusy(false);
        onUploadStateChange?.(false);
        if (inputRef.current) inputRef.current.value = "";
      });
  }

  function onDrop(event: React.DragEvent<HTMLFieldSetElement>) {
    event.preventDefault();
    setDragActive(false);
    const files = Array.from(event.dataTransfer.files);
    if (files.length !== 1) {
      setError("Pilih atau jatuhkan tepat satu gambar.");
      return;
    }
    select(files[0]);
  }

  const inputClass =
    "mt-1 min-h-10 w-full min-w-0 rounded-md border border-border bg-surface-muted px-3 text-sm";

  return (
    <div className="grid min-w-0 gap-2" aria-busy={busy}>
      <span className="text-sm font-semibold text-primary">
        {label}
        {optional ? " (opsional)" : ""}
      </span>
      {externalURL ? (
        <div className="grid gap-2">
          <label htmlFor={`${id}-url`} className="text-xs text-muted">
            URL gambar eksternal (opsional)
          </label>
          <input
            id={`${id}-url`}
            type="url"
            value={value}
            disabled={disabled || busy}
            onChange={(event) => onChange(event.target.value)}
            placeholder="https://…"
            className={inputClass}
          />
          <button
            type="button"
            disabled={disabled || busy}
            onClick={() => {
              setExternalURL(false);
              onChange("");
            }}
            className="inline-flex min-h-10 w-fit items-center whitespace-nowrap rounded-md px-3 text-sm font-semibold text-primary underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
          >
            Kembali ke upload gambar
          </button>
        </div>
      ) : value || localPreview ? (
        <FadeContent
          data-slot="media-upload-selected-card"
          className="grid min-w-0 gap-4 rounded-xl border border-border bg-surface p-4"
        >
          <MediaImage
            value={localPreview || value}
            alt={`Pratinjau ${label.toLowerCase()}`}
            width={1200}
            height={800}
            allowBlobPreview={Boolean(localPreview)}
            className={`w-full rounded-lg bg-surface-muted object-contain ${compact ? "h-32" : "h-48"}`}
            fallbackClassName="aspect-[3/2] place-items-center bg-surface-muted text-primary"
            fallback={<ImagePlus aria-hidden="true" className="size-7" />}
          />
          <div className="grid min-w-0 gap-2">
            <p className="truncate text-sm font-medium" title={filename}>
              {filename || "Gambar tersimpan"}
            </p>
            {size !== null ? (
              <p className="text-xs text-muted">{(size / 1024 / 1024).toFixed(2)} MiB</p>
            ) : null}
            {busy ? (
              <p role="status" className="flex items-center gap-2 text-sm text-muted">
                <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
                Mengunggah gambar…
              </p>
            ) : null}
            <ActionGroup
              role="group"
              aria-label={`Aksi ${label.toLowerCase()}`}
              data-slot="media-upload-actions"
              className="w-full max-w-sm flex-col sm:flex-col"
            >
              <button
                type="button"
                disabled={disabled || busy}
                onClick={() => inputRef.current?.click()}
                className="inline-flex min-h-10 w-full items-center justify-center gap-2 whitespace-nowrap rounded-md border border-border px-4 text-sm font-semibold text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
              >
                <Upload className="size-4" aria-hidden="true" />
                Ganti gambar
              </button>
              <button
                type="button"
                disabled={disabled || busy}
                onClick={() => {
                  onChange("");
                  setLocalPreview("");
                  setFilename("");
                  setSize(null);
                  setError("");
                }}
                className="inline-flex min-h-10 w-full items-center justify-center gap-2 whitespace-nowrap rounded-md border border-danger/30 px-4 text-sm font-semibold text-danger focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
              >
                <Trash2 className="size-4" aria-hidden="true" />
                Hapus
              </button>
              {optional ? (
                <button
                  type="button"
                  disabled={disabled || busy}
                  onClick={() => setExternalURL(true)}
                  className="min-h-10 w-full whitespace-nowrap rounded-md px-4 text-sm font-semibold text-primary underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
                >
                  Gunakan URL eksternal
                </button>
              ) : null}
            </ActionGroup>
          </div>
        </FadeContent>
      ) : (
        <fieldset
          aria-label={`Area unggah ${label.toLowerCase()}`}
          onDragEnter={(event) => {
            event.preventDefault();
            setDragActive(true);
          }}
          onDragOver={(event) => event.preventDefault()}
          onDragLeave={(event) => {
            if (!event.currentTarget.contains(event.relatedTarget as Node | null)) setDragActive(false);
          }}
          onDrop={onDrop}
          className={`grid min-w-0 justify-items-center gap-2 rounded-xl border border-dashed p-4 text-center sm:p-6 ${dragActive ? "border-primary bg-primary/5" : "border-border bg-surface-muted"}`}
        >
          <ImagePlus aria-hidden="true" className="size-7 text-primary" />
          <p className="text-sm font-medium">Jatuhkan gambar di sini</p>
          <span className="text-xs text-muted">atau</span>
          <button
            type="button"
            disabled={disabled || busy}
            onClick={() => inputRef.current?.click()}
            className="min-h-11 rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground"
          >
            Pilih gambar
          </button>
          <p className="text-xs text-muted">JPG, PNG, WebP, GIF, TIFF, BMP · Maks. 5 MiB</p>
          <p className="text-xs text-muted">TIFF dan BMP akan dioptimalkan otomatis untuk web.</p>
          {optional ? (
            <button
              type="button"
              disabled={disabled || busy}
              onClick={() => setExternalURL(true)}
              className="min-h-10 w-full whitespace-nowrap rounded-md px-4 text-sm font-semibold text-primary underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
            >
              Gunakan URL eksternal
            </button>
          ) : null}
        </fieldset>
      )}
      <input
        ref={inputRef}
        id={id}
        type="file"
        accept={MEDIA_UPLOAD_ACCEPT}
        className="sr-only"
        aria-label={`Pilih ${label.toLowerCase()}`}
        disabled={disabled || busy}
        onChange={(event) => select(event.currentTarget.files?.[0])}
      />
      {error ? (
        <p role="alert" className="text-sm text-danger">
          {error}
        </p>
      ) : null}
    </div>
  );
}
