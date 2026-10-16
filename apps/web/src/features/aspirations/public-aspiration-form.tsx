"use client";

import { CheckCircle2, LockKeyhole, Send } from "lucide-react";
import * as React from "react";
import { Button } from "@/components/ui/button";
import { ActionGroup, Checkbox, Input, Label, Select, Textarea } from "@/components/ui/primitives";
import { ApiError } from "@/lib/api/client";
import { getPublicAspirationDivisions, submitAspiration } from "./api";
import type { AspirationDivision, AspirationReceipt } from "./types";

type FormValues = {
  division_id: string;
  name: string;
  nim: string;
  email: string;
  subject: string;
  message: string;
  is_anonymous: boolean;
};
type FieldErrors = Partial<Record<keyof FormValues, string>>;

const initialValues: FormValues = {
  division_id: "",
  name: "",
  nim: "",
  email: "",
  subject: "",
  message: "",
  is_anonymous: false,
};

export function PublicAspirationForm() {
  const [divisions, setDivisions] = React.useState<AspirationDivision[]>([]);
  const [loadingDivisions, setLoadingDivisions] = React.useState(true);
  const [divisionError, setDivisionError] = React.useState("");
  const [receipt, setReceipt] = React.useState<AspirationReceipt | null>(null);
  const [submitError, setSubmitError] = React.useState("");
  const successRef = React.useRef<HTMLDivElement>(null);
  const [values, setValues] = React.useState<FormValues>(initialValues);
  const [errors, setErrors] = React.useState<FieldErrors>({});
  const [submitting, setSubmitting] = React.useState(false);
  const anonymous = values.is_anonymous;

  function change<K extends keyof FormValues>(key: K, value: FormValues[K]) {
    setValues((current) => ({ ...current, [key]: value }));
    setErrors((current) => ({ ...current, [key]: undefined }));
  }

  React.useEffect(() => {
    let active = true;
    getPublicAspirationDivisions()
      .then((items) => {
        if (active) setDivisions(items);
      })
      .catch(() => {
        if (active) setDivisionError("Tujuan aspirasi belum dapat dimuat. Coba muat ulang halaman.");
      })
      .finally(() => {
        if (active) setLoadingDivisions(false);
      });
    return () => {
      active = false;
    };
  }, []);

  React.useEffect(() => {
    if (receipt) successRef.current?.focus();
  }, [receipt]);

  function toggleAnonymous(checked: boolean) {
    setValues((current) => ({
      ...current,
      is_anonymous: checked,
      ...(checked ? { name: "", nim: "", email: "" } : {}),
    }));
  }

  function validate(): FieldErrors {
    const next: FieldErrors = {};
    if (!values.division_id) next.division_id = "Pilih tujuan aspirasi.";
    if (values.name.length > 160) next.name = "Nama maksimal 160 karakter.";
    if (values.nim.length > 32) next.nim = "NIM maksimal 32 karakter.";
    if (values.email.length > 254) next.email = "Email maksimal 254 karakter.";
    else if (values.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(values.email))
      next.email = "Masukkan alamat email yang valid.";
    if (!values.subject.trim()) next.subject = "Judul wajib diisi.";
    else if (values.subject.trim().length > 200) next.subject = "Judul maksimal 200 karakter.";
    if (!values.message.trim()) next.message = "Isi aspirasi wajib diisi.";
    else if (values.message.trim().length > 5000) next.message = "Isi aspirasi maksimal 5000 karakter.";
    return next;
  }

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitError("");
    const nextErrors = validate();
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) {
      const first = Object.keys(nextErrors)[0] as keyof FormValues;
      document.getElementById(`aspiration-${first}`)?.focus();
      return;
    }
    setSubmitting(true);
    try {
      const response = await submitAspiration({
        division_id: Number(values.division_id),
        name: values.is_anonymous ? undefined : values.name,
        nim: values.is_anonymous ? undefined : values.nim,
        email: values.is_anonymous ? undefined : values.email,
        subject: values.subject,
        message: values.message,
        is_anonymous: values.is_anonymous,
      });
      setReceipt(response);
      setValues(initialValues);
    } catch (error) {
      if (error instanceof ApiError && error.status === 429) {
        setSubmitError("Pengiriman terlalu sering. Silakan coba beberapa saat lagi.");
      } else if (error instanceof ApiError && error.status === 422) {
        setSubmitError("Periksa kembali tujuan dan isian aspirasi Anda.");
      } else {
        setSubmitError("Aspirasi belum dapat dikirim. Coba lagi beberapa saat lagi.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (receipt) {
    return (
      <div
        ref={successRef}
        tabIndex={-1}
        role="status"
        aria-live="polite"
        className="rounded-2xl border border-emerald-200 bg-surface p-6 shadow-card outline-none sm:p-9"
      >
        <span className="grid size-12 place-items-center rounded-xl bg-success/10 text-success-foreground">
          <CheckCircle2 aria-hidden="true" className="size-6" />
        </span>
        <h2 className="mt-5 font-heading text-2xl font-bold text-primary">Aspirasi berhasil dikirim.</h2>
        <p className="mt-2 text-sm leading-6 text-body">
          Terima kasih telah menyampaikan masukan. Pengurus terkait akan meninjaunya.
        </p>
        <ActionGroup
          role="group"
          aria-label="Aksi setelah aspirasi terkirim"
          data-slot="aspiration-success-actions"
          className="mt-5 w-full sm:justify-between"
        >
          <p
            data-slot="aspiration-reference"
            className="inline-flex min-h-10 w-fit max-w-full items-center rounded-md bg-surface-muted px-3 py-2 text-sm text-body"
          >
            Nomor referensi: <strong className="ml-1 whitespace-nowrap text-primary">ASP-{receipt.id}</strong>
          </p>
          <Button className="w-full sm:w-auto" onClick={() => setReceipt(null)}>
            Kirim aspirasi lain
          </Button>
        </ActionGroup>
      </div>
    );
  }

  return (
    <form
      onSubmit={onSubmit}
      noValidate
      className="space-y-5 rounded-2xl border border-border bg-surface p-5 shadow-card sm:p-8"
    >
      <div>
        <Label htmlFor="aspiration-division">
          Tujuan Aspirasi <span className="text-danger">*</span>
        </Label>
        <Select
          id="aspiration-division"
          value={values.division_id}
          onChange={(event) => change("division_id", event.target.value)}
          aria-invalid={!!errors.division_id}
          aria-describedby={errors.division_id ? "aspiration-division-error" : undefined}
          disabled={loadingDivisions || divisions.length === 0}
        >
          <option value="">{loadingDivisions ? "Memuat tujuan…" : "Pilih divisi tujuan"}</option>
          {divisions.map((division) => (
            <option key={division.id} value={division.id}>
              {division.name}
            </option>
          ))}
        </Select>
        {errors.division_id ? (
          <p id="aspiration-division-error" className="mt-1.5 text-sm text-danger" role="alert">
            {errors.division_id}
          </p>
        ) : null}
        {divisionError ? (
          <p className="mt-1.5 text-sm text-danger" role="alert">
            {divisionError}
          </p>
        ) : null}
        {!loadingDivisions && !divisionError && divisions.length === 0 ? (
          <p className="mt-1.5 text-sm text-muted">Belum ada divisi aktif pada periode berjalan.</p>
        ) : null}
      </div>

      <label
        htmlFor="aspiration-anonymous"
        className="flex min-h-16 cursor-pointer items-start gap-3 rounded-xl bg-surface-muted p-4"
      >
        <Checkbox
          id="aspiration-anonymous"
          checked={anonymous}
          onChange={(event) => toggleAnonymous(event.target.checked)}
          aria-describedby="anonymous-help"
          className="mt-0.5"
        />
        <span>
          <span className="block text-sm font-semibold text-primary">Kirim sebagai anonim</span>
          <span id="anonymous-help" className="mt-1 block text-xs leading-5 text-muted">
            Aspirasi akan dikirim tanpa identitas; nama, NIM, dan email tidak akan disimpan.
          </span>
        </span>
      </label>

      {!anonymous ? (
        <fieldset className="space-y-4">
          <legend className="text-xs font-bold uppercase tracking-[0.12em] text-primary">
            Informasi pengirim{" "}
            <span className="font-normal normal-case tracking-normal text-muted">(opsional)</span>
          </legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="aspiration-name">Nama</Label>
              <Input
                id="aspiration-name"
                maxLength={160}
                value={values.name}
                onChange={(event) => change("name", event.target.value)}
                error={errors.name}
              />
            </div>
            <div>
              <Label htmlFor="aspiration-nim">NIM</Label>
              <Input
                id="aspiration-nim"
                maxLength={32}
                value={values.nim}
                onChange={(event) => change("nim", event.target.value)}
                error={errors.nim}
              />
            </div>
          </div>
          <div>
            <Label htmlFor="aspiration-email">
              Email <span className="font-normal text-muted">(opsional)</span>
            </Label>
            <Input
              id="aspiration-email"
              type="email"
              autoComplete="email"
              maxLength={254}
              value={values.email}
              onChange={(event) => change("email", event.target.value)}
              error={errors.email}
            />
          </div>
        </fieldset>
      ) : null}

      <div>
        <Label htmlFor="aspiration-subject">
          Judul Aspirasi <span className="text-danger">*</span>
        </Label>
        <Input
          id="aspiration-subject"
          maxLength={200}
          placeholder="Ringkas pokok masukan Anda"
          value={values.subject}
          onChange={(event) => change("subject", event.target.value)}
          error={errors.subject}
        />
      </div>
      <div>
        <div className="flex items-center justify-between gap-2">
          <Label htmlFor="aspiration-message">
            Isi Aspirasi <span className="text-danger">*</span>
          </Label>
          <span className="text-xs text-muted">{values.message.length}/5000</span>
        </div>
        <Textarea
          id="aspiration-message"
          maxLength={5000}
          rows={7}
          placeholder="Ceritakan masukan atau keluhan dengan jelas dan santun…"
          value={values.message}
          onChange={(event) => change("message", event.target.value)}
          error={errors.message}
        />
      </div>

      <div className="flex items-start gap-3 rounded-lg border border-border bg-surface p-4 text-sm leading-6 text-body">
        <LockKeyhole aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-primary" />
        <p>
          Data yang Anda isi digunakan pengurus untuk meninjau aspirasi. Jika memilih anonim, identitas tidak
          disimpan bersama aspirasi. Hindari mengirim data pribadi yang tidak diperlukan.
        </p>
      </div>
      {submitError ? (
        <p
          role="alert"
          className="rounded-md border border-danger/25 bg-danger/5 p-3 text-sm text-danger-foreground"
        >
          {submitError}
        </p>
      ) : null}
      <Button
        type="submit"
        disabled={submitting || loadingDivisions || divisions.length === 0}
        className="min-h-12 w-full sm:w-auto"
      >
        <Send aria-hidden="true" className="mr-2 size-4" />
        {submitting ? "Mengirim…" : "Kirim Aspirasi"}
      </Button>
    </form>
  );
}
