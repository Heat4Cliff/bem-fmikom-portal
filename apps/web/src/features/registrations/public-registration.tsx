"use client";

import { CalendarDays, CheckCircle2, ChevronRight, Clock3, FileText } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { Button } from "@/components/ui/button";
import { MediaFrame } from "@/components/ui/media-frame";
import { MediaImage } from "@/components/ui/media-image";
import {
  Checkbox,
  Input,
  Label,
  RadioOption,
  RadioGroup,
  Select,
  Textarea,
} from "@/components/ui/primitives";
import { ApiError } from "@/lib/api/client";
import { getPublicRegistration, getPublicRegistrations, submitRegistration } from "./api";
import type {
  PublicRegistration,
  PublicRegistrationPreview,
  PublicRegistrationStatus,
  RegistrationField,
} from "./types";

const statuses: { value: PublicRegistrationStatus; label: string }[] = [
  { value: "OPEN", label: "Dibuka" },
  { value: "UPCOMING", label: "Akan Datang" },
  { value: "CLOSED", label: "Ditutup" },
];

export function PublicRegistrationList() {
  const [status, setStatus] = React.useState<PublicRegistrationStatus>("OPEN");
  const [items, setItems] = React.useState<PublicRegistrationPreview[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  React.useEffect(() => {
    let active = true;
    setLoading(true);
    getPublicRegistrations({ status, page: 1, limit: 30 })
      .then((result) => {
        if (active) setItems(result.items);
      })
      .catch(() => {
        if (active) setError("Informasi pendaftaran belum dapat dimuat.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [status]);
  return (
    <main className="mx-auto w-full max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl text-center">
        <p className="text-xs font-bold uppercase tracking-[0.18em] text-secondary">BEM FMIKOM UNUGHA</p>
        <h1 className="mt-3 font-heading text-3xl font-bold text-primary sm:text-4xl">Pendaftaran</h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          Temukan kesempatan untuk berkembang dan berkontribusi bersama.
        </p>
      </div>
      <fieldset className="mx-auto mt-8 flex max-w-2xl flex-wrap justify-center gap-2">
        <legend className="sr-only">Status pendaftaran</legend>
        {statuses.map((item) => (
          <button
            key={item.value}
            type="button"
            aria-pressed={status === item.value}
            onClick={() => setStatus(item.value)}
            className={`min-h-10 rounded-full px-5 text-sm font-semibold transition ${status === item.value ? "bg-primary text-white" : "border border-border bg-white text-muted hover:text-primary"}`}
          >
            {item.label}
          </button>
        ))}
      </fieldset>
      {error ? (
        <p role="alert" className="mx-auto mt-8 max-w-3xl rounded-lg bg-danger/5 p-4 text-sm text-danger">
          {error}
        </p>
      ) : null}
      {loading ? (
        <p role="status" className="mt-12 text-center text-sm text-muted">
          Memuat pendaftaran…
        </p>
      ) : items.length ? (
        <div className="mt-8 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          {items.map((item) => (
            <article
              key={item.slug}
              className="flex min-h-64 flex-col rounded-2xl border border-border bg-white p-6 shadow-card"
            >
              <div className="flex items-center justify-between gap-3">
                <span
                  className={`rounded-full px-3 py-1 text-xs font-bold ${item.public_status === "OPEN" ? "bg-success/10 text-success" : "bg-surface-muted text-muted"}`}
                >
                  {statuses.find((entry) => entry.value === item.public_status)?.label}
                </span>
                <FileText aria-hidden="true" className="size-5 text-secondary" />
              </div>
              <h2 className="mt-5 font-heading text-xl font-bold text-primary">{item.title}</h2>
              <p className="mt-2 text-sm text-muted">{item.division_name ?? "BEM FMIKOM"}</p>
              <div className="mt-4 grid gap-2 text-xs text-body sm:grid-cols-2">
                <span className="inline-flex items-center gap-2">
                  <CalendarDays className="size-4 text-secondary" />
                  {formatDate(item.open_at)}
                </span>
                <span className="inline-flex items-center gap-2">
                  <Clock3 className="size-4 text-secondary" />
                  s.d. {formatDate(item.close_at)}
                </span>
              </div>
              <Link
                href={`/pendaftaran/${encodeURIComponent(item.slug)}`}
                className="mt-auto inline-flex min-h-11 items-center gap-1 pt-5 text-sm font-bold text-primary hover:text-secondary"
              >
                Lihat detail <ChevronRight aria-hidden="true" className="size-4" />
              </Link>
            </article>
          ))}
        </div>
      ) : !loading ? (
        <div className="mx-auto mt-10 max-w-xl rounded-2xl border border-dashed border-border bg-white p-8 text-center">
          <CalendarDays aria-hidden="true" className="mx-auto size-8 text-secondary" />
          <h2 className="mt-3 font-heading text-lg font-bold text-primary">
            Belum ada pendaftaran{" "}
            {status === "OPEN" ? "yang dibuka" : status === "UPCOMING" ? "yang akan datang" : "yang ditutup"}
          </h2>
          <p className="mt-2 text-sm text-muted">
            Silakan periksa kembali informasi pendaftaran secara berkala.
          </p>
        </div>
      ) : null}
    </main>
  );
}

export function PublicRegistrationDetail({ slug }: { slug: string }) {
  const [registration, setRegistration] = React.useState<PublicRegistration | null>(null);
  const [error, setError] = React.useState("");
  React.useEffect(() => {
    let active = true;
    getPublicRegistration(slug)
      .then((data) => {
        if (active) setRegistration(data);
      })
      .catch(() => {
        if (active) setError("Pendaftaran tidak ditemukan atau belum dipublikasikan.");
      });
    return () => {
      active = false;
    };
  }, [slug]);
  if (error)
    return (
      <main className="mx-auto max-w-3xl px-4 py-16 text-center">
        <h1 className="font-heading text-2xl font-bold text-primary">Pendaftaran tidak tersedia</h1>
        <p role="alert" className="mt-3 text-sm text-muted">
          {error}
        </p>
        <Link className="mt-6 inline-block text-sm font-bold text-primary" href="/pendaftaran">
          Kembali ke daftar
        </Link>
      </main>
    );
  if (!registration)
    return (
      <main className="mx-auto max-w-3xl px-4 py-16 text-center text-sm text-muted" role="status">
        Memuat informasi pendaftaran…
      </main>
    );
  const isOpen = registration.status === "OPEN";
  return (
    <main className="mx-auto w-full max-w-4xl px-4 py-10 sm:px-6 lg:py-16">
      <Link className="text-sm font-semibold text-primary hover:text-secondary" href="/pendaftaran">
        ← Semua pendaftaran
      </Link>
      <article className="mt-5 rounded-2xl border border-border bg-white p-5 shadow-card sm:p-9">
        <div className="flex flex-wrap items-center gap-3">
          <span className="rounded-full bg-primary/5 px-3 py-1 text-xs font-bold uppercase text-primary">
            {registration.status === "OPEN"
              ? "Dibuka"
              : registration.status === "UPCOMING"
                ? "Akan Datang"
                : "Ditutup"}
          </span>
          <span className="text-sm text-muted">{registration.division_name ?? "BEM FMIKOM"}</span>
        </div>
        <h1 className="mt-4 font-heading text-3xl font-bold text-primary sm:text-4xl">
          {registration.title}
        </h1>
        <div className="mt-5 grid gap-3 rounded-xl bg-surface-muted p-4 text-sm sm:grid-cols-2">
          <p>
            <span className="block text-xs text-muted">Periode organisasi</span>
            <span className="mt-1 block font-semibold text-primary">{registration.period_name}</span>
          </p>
          <p>
            <span className="block text-xs text-muted">Waktu pendaftaran</span>
            <span className="mt-1 block font-semibold text-primary">
              {formatDate(registration.open_at)} — {formatDate(registration.close_at)}
            </span>
          </p>
        </div>
        {registration.description ? (
          <p className="mt-6 whitespace-pre-wrap text-sm leading-7 text-body">{registration.description}</p>
        ) : null}
        {isOpen ? (
          <Link
            href={`/pendaftaran/${encodeURIComponent(registration.slug)}/form`}
            className="mt-8 inline-flex min-h-12 items-center justify-center rounded-lg bg-primary px-6 text-sm font-bold text-white hover:bg-primary/90"
          >
            Isi Formulir Pendaftaran
          </Link>
        ) : (
          <p className="mt-8 rounded-lg border border-border bg-surface-muted p-4 text-sm text-muted">
            {registration.status === "UPCOMING"
              ? "Formulir dapat diisi saat masa pendaftaran dimulai."
              : "Masa pendaftaran telah berakhir."}
          </p>
        )}
      </article>
    </main>
  );
}

export function PublicRegistrationForm({ slug }: { slug: string }) {
  const [registration, setRegistration] = React.useState<PublicRegistration | null>(null);
  const [values, setValues] = React.useState<Record<string, string | string[]>>({});
  const [images, setImages] = React.useState<Record<string, File>>({});
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [message, setMessage] = React.useState("");
  const [success, setSuccess] = React.useState(false);
  const [receiptID, setReceiptID] = React.useState<number | null>(null);
  const [submitting, setSubmitting] = React.useState(false);
  React.useEffect(() => {
    let active = true;
    getPublicRegistration(slug)
      .then((item) => {
        if (active) setRegistration(item);
      })
      .catch(() => {
        if (active) setMessage("Formulir pendaftaran tidak ditemukan.");
      });
    return () => {
      active = false;
    };
  }, [slug]);
  function setValue(key: string, value: string | string[]) {
    setValues((current) => ({ ...current, [key]: value }));
    setErrors((current) => ({ ...current, [key]: "" }));
  }
  function validate(fields: RegistrationField[]) {
    const next: Record<string, string> = {};
    fields.forEach((field) => {
      const value = values[field.field_key];
      if (field.field_type === "IMAGE") {
        const file = images[field.field_key];
        if (field.required && !file) next[field.field_key] = "Pilih gambar untuk melanjutkan.";
        if (file && file.size > 5 * 1024 * 1024) next[field.field_key] = "Ukuran gambar maksimal 5 MiB.";
        return;
      }
      const empty = value === undefined || value === "" || (Array.isArray(value) && !value.length);
      if (field.required && empty) next[field.field_key] = "Bagian ini wajib diisi.";
      if (
        typeof value === "string" &&
        field.field_type === "EMAIL" &&
        value &&
        !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
      )
        next[field.field_key] = "Masukkan alamat email yang valid.";
      if (typeof value === "string" && field.field_type === "TEXTAREA" && value.length > 5000)
        next[field.field_key] = "Maksimal 5.000 karakter.";
      if (typeof value === "string" && field.field_type === "TEXT" && value.length > 500)
        next[field.field_key] = "Maksimal 500 karakter.";
    });
    setErrors(next);
    return next;
  }
  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage("");
    if (!registration) return;
    const nextErrors = validate(registration.fields);
    const first = Object.keys(nextErrors)[0];
    if (first) {
      focusRegistrationField(first);
      return;
    }
    setSubmitting(true);
    try {
      const receipt = await submitRegistration(registration.id, values, images);
      setReceiptID(receipt.submission_id);
      setSuccess(true);
    } catch (error) {
      if (error instanceof ApiError && error.code === "DUPLICATE_SUBMISSION") {
        setErrors((current) => ({ ...current, nim: "NIM ini sudah terdaftar pada pendaftaran tersebut." }));
        setMessage("NIM ini sudah terdaftar pada pendaftaran tersebut.");
        focusRegistrationField("nim");
      } else if (error instanceof ApiError && error.status === 429)
        setMessage("Pengiriman terlalu sering. Silakan coba lagi nanti.");
      else if (error instanceof ApiError && error.code === "REGISTRATION_CLOSED")
        setMessage("Masa pendaftaran telah berakhir.");
      else if (error instanceof ApiError && error.errors && Object.keys(error.errors).length) {
        const nextErrors = Object.fromEntries(
          Object.entries(error.errors ?? {}).flatMap(([key, messages]) =>
            messages.length ? [[key, messages[0]]] : [],
          ),
        );
        if (Object.keys(nextErrors).length) {
          setErrors((current) => ({ ...current, ...nextErrors }));
          setMessage("Periksa kolom yang ditandai, lalu kirim kembali.");
          focusRegistrationField(Object.keys(nextErrors)[0]);
        } else {
          setMessage("Periksa kolom yang ditandai, lalu kirim kembali.");
        }
      } else if (error instanceof ApiError && error.status === 422) {
        const safeValidationMessage = error.message.trim();
        setMessage(
          safeValidationMessage && safeValidationMessage !== "Validation failed"
            ? safeValidationMessage
            : "Periksa kembali isian formulir Anda.",
        );
      } else setMessage("Pendaftaran gagal dikirim. Coba kembali beberapa saat lagi.");
    } finally {
      setSubmitting(false);
    }
  }
  if (success)
    return (
      <main className="mx-auto max-w-2xl px-4 py-16 text-center">
        <div className="rounded-2xl border border-success/20 bg-white p-8 shadow-card">
          <CheckCircle2 aria-hidden="true" className="mx-auto size-12 text-success" />
          <h1 className="mt-4 font-heading text-2xl font-bold text-primary">Pendaftaran berhasil dikirim</h1>
          <p className="mt-2 text-sm text-muted">
            Terima kasih telah mengisi formulir {registration?.title}.
          </p>
          {receiptID ? (
            <p className="mt-3 text-sm text-body">
              Nomor referensi: <strong className="font-semibold text-primary">#{receiptID}</strong>
            </p>
          ) : null}
          <Link className="mt-6 inline-block text-sm font-bold text-primary" href="/pendaftaran">
            Kembali ke pendaftaran
          </Link>
        </div>
      </main>
    );
  if (!registration)
    return (
      <main className="mx-auto max-w-2xl px-4 py-16 text-center text-sm text-muted" role="status">
        Memuat formulir…
      </main>
    );
  if (registration.status !== "OPEN")
    return (
      <main className="mx-auto max-w-2xl px-4 py-16">
        <div className="rounded-xl border border-border bg-white p-6 text-center">
          <h1 className="font-heading text-xl font-bold text-primary">Formulir tidak sedang dibuka</h1>
          <p className="mt-2 text-sm text-muted">
            {registration.status === "UPCOMING"
              ? "Pendaftaran akan dibuka sesuai jadwal."
              : "Masa pendaftaran telah berakhir."}
          </p>
          <Link
            href={`/pendaftaran/${encodeURIComponent(slug)}`}
            className="mt-5 inline-block text-sm font-semibold text-primary"
          >
            Kembali ke detail
          </Link>
        </div>
      </main>
    );
  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-8 sm:px-6 lg:py-12">
      <Link className="text-sm font-semibold text-primary" href={`/pendaftaran/${encodeURIComponent(slug)}`}>
        ← Kembali ke informasi
      </Link>
      <header className="mt-5 rounded-2xl bg-primary p-5 text-white sm:p-8">
        <p className="text-xs font-bold uppercase tracking-[0.15em] text-white/70">Formulir Pendaftaran</p>
        <h1 className="mt-2 font-heading text-2xl font-bold sm:text-3xl">{registration.title}</h1>
        <p className="mt-2 text-sm text-white/75">
          {registration.division_name ?? "BEM FMIKOM"} · ditutup {formatDate(registration.close_at)}
        </p>
      </header>
      <form
        onSubmit={onSubmit}
        noValidate
        className="mt-5 space-y-5 rounded-2xl border border-border bg-white p-5 shadow-card sm:p-8"
      >
        {message ? (
          <p role="alert" className="rounded-lg bg-danger/5 p-3 text-sm text-danger">
            {message}
          </p>
        ) : null}
        {registration.fields.map((field) => (
          <RegistrationInput
            key={field.field_key}
            field={field}
            value={values[field.field_key]}
            image={images[field.field_key]}
            error={errors[field.field_key]}
            onChange={(value) => setValue(field.field_key, value)}
            onImageChange={(file) => {
              setImages((current) => {
                const next = { ...current };
                if (file) next[field.field_key] = file;
                else delete next[field.field_key];
                return next;
              });
              setErrors((current) => ({ ...current, [field.field_key]: "" }));
            }}
          />
        ))}
        <p className="text-xs text-muted">
          Informasi yang Anda kirim digunakan untuk keperluan proses pendaftaran ini.
        </p>
        <Button type="submit" disabled={submitting} className="w-full sm:w-auto">
          {submitting ? "Mengirim…" : "Kirim Pendaftaran"}
        </Button>
      </form>
    </main>
  );
}

function focusRegistrationField(fieldKey: string) {
  const control = document.getElementById(`registration-${fieldKey}`);
  control?.scrollIntoView?.({ block: "center", behavior: "smooth" });
  control?.focus({ preventScroll: true });
}

function RegistrationInput({
  field,
  value,
  image,
  error,
  onChange,
  onImageChange,
}: {
  field: RegistrationField;
  value?: string | string[];
  image?: File;
  error?: string;
  onChange: (value: string | string[]) => void;
  onImageChange: (file: File | undefined) => void;
}) {
  const id = `registration-${field.field_key}`;
  const required = field.required;
  const label = (
    <Label htmlFor={id}>
      {field.label}
      {required ? (
        <span aria-hidden="true" className="ml-1 text-danger">
          *
        </span>
      ) : null}
    </Label>
  );
  const common = {
    id,
    required,
    "aria-invalid": Boolean(error),
    "aria-describedby": error ? `${id}-error` : undefined,
    placeholder: field.placeholder ?? "",
  };
  return (
    <div className="space-y-2">
      {field.field_type === "IMAGE" ? (
        <RegistrationImageInput
          id={id}
          label={field.label}
          required={required}
          file={image}
          error={error}
          onChange={onImageChange}
        />
      ) : field.field_type === "RADIO" ? (
        <fieldset aria-describedby={error ? `${id}-error` : undefined}>
          <legend className="text-sm font-semibold text-primary">
            {field.label}
            {required ? (
              <span aria-hidden="true" className="ml-1 text-danger">
                *
              </span>
            ) : null}
          </legend>
          <RadioGroup className="mt-2 flex-col gap-2">
            {field.options.map((option) => (
              <RadioOption
                key={option.value}
                id={`${id}-${option.value}`}
                name={id}
                value={option.value}
                label={option.label}
                checked={value === option.value}
                onChange={() => onChange(option.value)}
              />
            ))}
          </RadioGroup>
        </fieldset>
      ) : field.field_type === "CHECKBOX" ? (
        <fieldset>
          <legend className="text-sm font-semibold text-primary">
            {field.label}
            {required ? (
              <span aria-hidden="true" className="ml-1 text-danger">
                *
              </span>
            ) : null}
          </legend>
          <div className="mt-2 grid gap-2">
            {field.options.map((option, index) => {
              const selected = Array.isArray(value) ? value : [];
              const optionID = `${id}-option-${index}`;
              return (
                <div key={option.value} className="flex min-h-10 items-center gap-3 text-sm text-body">
                  <Checkbox
                    id={optionID}
                    checked={selected.includes(option.value)}
                    onChange={(event) =>
                      onChange(
                        event.target.checked
                          ? [...selected, option.value]
                          : selected.filter((item) => item !== option.value),
                      )
                    }
                  />
                  <Label htmlFor={optionID}>{option.label}</Label>
                </div>
              );
            })}
          </div>
        </fieldset>
      ) : (
        <>
          {label}
          {field.field_type === "TEXTAREA" ? (
            <Textarea
              {...common}
              maxLength={5000}
              value={typeof value === "string" ? value : ""}
              onChange={(event) => onChange(event.target.value)}
            />
          ) : field.field_type === "SELECT" ? (
            <Select
              {...common}
              value={typeof value === "string" ? value : ""}
              onChange={(event) => onChange(event.target.value)}
            >
              <option value="">Pilih {field.label.toLowerCase()}</option>
              {field.options.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
          ) : (
            <Input
              {...common}
              type={
                field.field_type === "EMAIL"
                  ? "email"
                  : field.field_type === "NUMBER"
                    ? "number"
                    : field.field_type === "DATE"
                      ? "date"
                      : "text"
              }
              maxLength={field.field_type === "TEXT" ? 500 : undefined}
              value={typeof value === "string" ? value : ""}
              onChange={(event) => onChange(event.target.value)}
            />
          )}
        </>
      )}
      {error ? (
        <p id={`${id}-error`} className="text-sm text-danger">
          {error}
        </p>
      ) : null}
    </div>
  );
}

const allowedRegistrationImage = /\.(jpe?g|jfif|png|webp|gif|tiff?|bmp)$/i;

function RegistrationImageInput({
  id,
  label,
  required,
  file,
  error,
  onChange,
}: {
  id: string;
  label: string;
  required: boolean;
  file?: File;
  error?: string;
  onChange: (file: File | undefined) => void;
}) {
  const input = React.useRef<HTMLInputElement>(null);
  const [preview, setPreview] = React.useState("");
  const [localError, setLocalError] = React.useState("");
  React.useEffect(() => {
    if (!file) {
      setPreview("");
      return;
    }
    const url = URL.createObjectURL(file);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);
  function choose(candidate?: File) {
    if (!candidate) return;
    if (!allowedRegistrationImage.test(candidate.name) || candidate.size > 5 * 1024 * 1024) {
      setLocalError(
        !allowedRegistrationImage.test(candidate.name)
          ? "Pilih gambar JPEG, PNG, WebP, GIF, TIFF, atau BMP."
          : "Ukuran gambar maksimal 5 MiB.",
      );
      onChange(undefined);
      return;
    }
    setLocalError("");
    onChange(candidate);
  }
  const displayedError = error || localError;
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>
        {label}
        {required ? (
          <span aria-hidden="true" className="ml-1 text-danger">
            *
          </span>
        ) : null}
      </Label>
      <div className="space-y-3 rounded-xl border border-dashed border-border bg-surface-muted p-4">
        {file && preview ? (
          <MediaFrame ratio="aspect-[4/3]" className="mx-auto max-w-sm rounded-lg bg-white">
            {/* Blob URLs are local previews only; submission bytes use multipart and never a public URL. */}
            <MediaImage
              value={preview}
              alt={`Pratinjau ${label}`}
              fill
              allowBlobPreview
              className="object-contain object-center"
            />
          </MediaFrame>
        ) : (
          <div className="grid min-h-24 place-items-center text-center text-sm text-muted">
            Seret dan lepaskan gambar di sini atau pilih dari perangkat.
          </div>
        )}
        {file ? (
          <p className="break-all text-xs text-body">
            {file.name} · {(file.size / (1024 * 1024)).toFixed(2)} MiB
          </p>
        ) : null}
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" onClick={() => input.current?.click()}>
            {file ? "Ganti gambar" : "Pilih gambar"}
          </Button>
          {file ? (
            <Button type="button" variant="ghost" onClick={() => onChange(undefined)}>
              Hapus
            </Button>
          ) : null}
        </div>
        <input
          ref={input}
          id={id}
          type="file"
          accept=".jpg,.jpeg,.jfif,.png,.webp,.gif,.tif,.tiff,.bmp,image/jpeg,image/png,image/webp,image/gif,image/tiff,image/bmp"
          className="sr-only"
          required={required}
          aria-invalid={Boolean(displayedError)}
          aria-describedby={displayedError ? `${id}-error` : undefined}
          onChange={(event) => {
            choose(event.currentTarget.files?.[0]);
            event.currentTarget.value = "";
          }}
        />
        <p className="text-xs text-muted">
          JPEG, PNG, WebP, GIF statis, TIFF, BMP · maksimal 5 MiB. GIF animasi tidak diterima.
        </p>
      </div>
      {displayedError ? (
        <p id={`${id}-error`} className="text-sm text-danger">
          {displayedError}
        </p>
      ) : null}
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeZone: "Asia/Jakarta" }).format(
    new Date(value),
  );
}
