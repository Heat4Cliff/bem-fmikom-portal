"use client";

import { ArrowDown, ArrowUp, Download, Plus, Search, Trash2 } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import * as React from "react";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { Button } from "@/components/ui/button";
import { MediaImage } from "@/components/ui/media-image";
import { Checkbox, Input, Label, Select, Textarea } from "@/components/ui/primitives";
import { useAuth } from "@/lib/auth/auth-provider";
import { ApiError } from "@/lib/api/client";
import { getDivisions, getPeriods } from "@/features/organization/api";
import type { Division, Period } from "@/features/organization/types";
import {
  createRegistration,
  deleteRegistration,
  downloadRegistrationExport,
  getRegistration,
  getRegistrationSubmission,
  getRegistrationAttachmentBlob,
  getRegistrationSubmissions,
  getRegistrations,
  saveRegistrationFields,
  setRegistrationStatus,
  updateRegistration,
} from "./api";
import type {
  RegistrationField,
  RegistrationFieldType,
  RegistrationForm,
  RegistrationSubmission,
} from "./types";

const fieldTypes: RegistrationFieldType[] = [
  "TEXT",
  "TEXTAREA",
  "EMAIL",
  "NUMBER",
  "SELECT",
  "RADIO",
  "CHECKBOX",
  "DATE",
  "IMAGE",
];
const coreKeys = new Set(["full_name", "nim", "email", "phone"]);

export function RegistrationAdminList() {
  const router = useRouter();
  const { user } = useAuth();
  const [items, setItems] = React.useState<RegistrationForm[]>([]);
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [search, setSearch] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [page, setPage] = React.useState(1);
  const [total, setTotal] = React.useState(0);
  const [totalPages, setTotalPages] = React.useState(0);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  const [creating, setCreating] = React.useState(false);
  const [busy, setBusy] = React.useState(false);
  const [draft, setDraft] = React.useState({
    period_id: "",
    division_id: "",
    title: "",
    open_at: "",
    close_at: "",
    description: "",
  });
  const canChooseDivision = user?.role !== "KEPALA_BIDANG";

  const load = React.useCallback(() => {
    let active = true;
    setLoading(true);
    setError("");
    getRegistrations({ page, limit: 20, search, status })
      .then((result) => {
        if (active) {
          setItems(result.items);
          setTotal(result.meta.total);
          setTotalPages(result.meta.total_pages);
        }
      })
      .catch(() => {
        if (active) setError("Data pendaftaran gagal dimuat.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [page, search, status]);
  React.useEffect(() => load(), [load]);
  React.useEffect(() => {
    let active = true;
    Promise.all([getPeriods({ page: 1, limit: 100 }), getDivisions({ page: 1, limit: 100 })])
      .then(([p, d]) => {
        if (active) {
          setPeriods(p.items);
          setDivisions(d.items);
        }
      })
      .catch(() => {
        if (active) setError("Daftar periode atau divisi gagal dimuat.");
      });
    return () => {
      active = false;
    };
  }, []);

  async function create(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      const formData = new FormData(event.currentTarget);
      const openAt = String(formData.get("open_at") ?? "");
      const closeAt = String(formData.get("close_at") ?? "");
      const item = await createRegistration({
        period_id: Number(formData.get("period_id")),
        division_id: formData.get("division_id") ? Number(formData.get("division_id")) : null,
        title: String(formData.get("title") ?? ""),
        description: String(formData.get("description") ?? "") || null,
        open_at: new Date(openAt).toISOString(),
        close_at: new Date(closeAt).toISOString(),
      });
      router.push(`/admin/pendaftaran/${item.id}/builder`);
    } catch {
      setError("Formulir gagal dibuat. Periksa periode, divisi, dan rentang waktunya.");
    } finally {
      setBusy(false);
    }
  }
  async function changeStatus(item: RegistrationForm, action: "open" | "close") {
    setBusy(true);
    setError("");
    try {
      await setRegistrationStatus(item.id, action);
      load();
    } catch {
      setError("Status formulir tidak dapat diperbarui.");
    } finally {
      setBusy(false);
    }
  }
  async function remove(item: RegistrationForm) {
    if (!window.confirm(`Hapus formulir “${item.title}”?`)) return;
    setBusy(true);
    try {
      await deleteRegistration(item.id);
      load();
    } catch {
      setError("Formulir tidak dapat dihapus. Formulir dengan data pendaftar tidak dapat dihapus.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1600px] p-4 md:p-6 lg:p-8">
      <AdminPageHeader
        title="Manajemen Pendaftaran"
        description="Atur periode pendaftaran, formulir dinamis, dan data pendaftar."
        actions={
          <Button onClick={() => setCreating((value) => !value)}>
            <Plus aria-hidden="true" className="size-4" /> Buat Pendaftaran
          </Button>
        }
      />
      {error ? (
        <p role="alert" className="mb-4 rounded-lg bg-danger/5 p-3 text-sm text-danger">
          {error}
        </p>
      ) : null}
      {creating ? (
        <form
          onSubmit={create}
          className="mb-6 grid gap-4 rounded-xl border border-border bg-surface p-4 shadow-card sm:grid-cols-2 lg:grid-cols-3"
        >
          <div>
            <Label htmlFor="registration-title">Nama pendaftaran</Label>
            <Input
              id="registration-title"
              name="title"
              required
              maxLength={200}
              value={draft.title}
              onChange={(e) => setDraft({ ...draft, title: e.target.value })}
            />
          </div>
          <div>
            <Label htmlFor="registration-period">Periode</Label>
            <Select
              id="registration-period"
              name="period_id"
              required
              value={draft.period_id}
              onChange={(e) => setDraft({ ...draft, period_id: e.target.value })}
            >
              <option value="">Pilih periode</option>
              {periods.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name}
                </option>
              ))}
            </Select>
          </div>
          {canChooseDivision ? (
            <div>
              <Label htmlFor="registration-division">
                Divisi <span className="font-normal text-muted">(opsional)</span>
              </Label>
              <Select
                id="registration-division"
                name="division_id"
                value={draft.division_id}
                onChange={(e) => setDraft({ ...draft, division_id: e.target.value })}
              >
                <option value="">BEM FMIKOM</option>
                {divisions.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
              </Select>
            </div>
          ) : (
            <p className="self-end pb-2 text-sm text-muted">Divisi akan mengikuti divisi akun Anda.</p>
          )}
          <div>
            <Label htmlFor="registration-open">Mulai</Label>
            <Input
              id="registration-open"
              name="open_at"
              type="datetime-local"
              required
              value={draft.open_at}
              onChange={(e) => setDraft({ ...draft, open_at: e.target.value })}
            />
          </div>
          <div>
            <Label htmlFor="registration-close">Tutup</Label>
            <Input
              id="registration-close"
              name="close_at"
              type="datetime-local"
              required
              value={draft.close_at}
              onChange={(e) => setDraft({ ...draft, close_at: e.target.value })}
            />
          </div>
          <div className="sm:col-span-2 lg:col-span-3">
            <Label htmlFor="registration-description">
              Deskripsi <span className="font-normal text-muted">(opsional)</span>
            </Label>
            <Textarea
              id="registration-description"
              name="description"
              maxLength={5000}
              value={draft.description}
              onChange={(e) => setDraft({ ...draft, description: e.target.value })}
            />
          </div>
          <div className="flex gap-2 sm:col-span-2 lg:col-span-3">
            <Button type="submit" disabled={busy}>
              Buat dan lanjutkan
            </Button>
            <Button type="button" variant="outline" onClick={() => setCreating(false)}>
              Batal
            </Button>
          </div>
        </form>
      ) : null}
      <div className="mb-4 grid gap-3 rounded-xl border border-border bg-surface p-4 shadow-card sm:grid-cols-[minmax(14rem,1fr)_12rem]">
        <div className="relative">
          <Label htmlFor="registration-search" className="sr-only">
            Cari pendaftaran
          </Label>
          <Search className="absolute left-3 top-3 size-4 text-muted" aria-hidden="true" />
          <Input
            id="registration-search"
            className="pl-9"
            placeholder="Cari pendaftaran…"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>
        <div>
          <Label htmlFor="registration-status" className="sr-only">
            Status
          </Label>
          <Select
            id="registration-status"
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
          >
            <option value="">Semua status</option>
            <option value="DRAFT">Draft</option>
            <option value="OPEN">Dibuka</option>
            <option value="CLOSED">Ditutup</option>
            <option value="UPCOMING">Akan datang</option>
          </Select>
        </div>
      </div>
      {loading ? (
        <p role="status" className="p-8 text-center text-sm text-muted">
          Memuat pendaftaran…
        </p>
      ) : (
        <>
          <div className="hidden overflow-x-auto rounded-xl border border-border bg-surface shadow-card md:block">
            <table className="w-full min-w-[850px] text-left text-sm">
              <thead className="bg-surface-muted text-xs uppercase tracking-wide text-muted">
                <tr>
                  <th className="p-4">Nama</th>
                  <th className="p-4">Divisi</th>
                  <th className="p-4">Periode</th>
                  <th className="p-4">Pendaftaran</th>
                  <th className="p-4">Status</th>
                  <th className="p-4">Pendaftar</th>
                  <th className="p-4">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {items.map((item) => (
                  <tr key={item.id}>
                    <td className="p-4 font-semibold text-primary">{item.title}</td>
                    <td className="p-4">{item.division_name ?? "BEM FMIKOM"}</td>
                    <td className="p-4">{item.period_name}</td>
                    <td className="p-4 text-xs">
                      {formatDate(item.open_at)} – {formatDate(item.close_at)}
                    </td>
                    <td className="p-4">
                      <StatusBadge status={item.status} />
                    </td>
                    <td className="p-4">{item.submission_count}</td>
                    <td className="p-4">
                      <div className="flex flex-wrap gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => router.push(`/admin/pendaftaran/${item.id}/builder`)}
                        >
                          Builder
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => router.push(`/admin/pendaftaran/${item.id}/submissions`)}
                        >
                          Data
                        </Button>
                        {item.status === "DRAFT" ? (
                          <Button size="sm" onClick={() => void changeStatus(item, "open")} disabled={busy}>
                            Buka
                          </Button>
                        ) : item.status === "OPEN" ? (
                          <Button
                            size="sm"
                            variant="secondary"
                            onClick={() => void changeStatus(item, "close")}
                            disabled={busy}
                          >
                            Tutup
                          </Button>
                        ) : null}
                        {item.submission_count === 0 && item.status === "DRAFT" ? (
                          <Button
                            aria-label={`Hapus ${item.title}`}
                            size="sm"
                            variant="danger"
                            onClick={() => void remove(item)}
                          >
                            <Trash2 className="size-4" />
                          </Button>
                        ) : null}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {items.length === 0 ? (
              <p className="p-8 text-center text-sm text-muted">Belum ada formulir pendaftaran.</p>
            ) : null}
          </div>
          <div className="grid gap-3 md:hidden">
            {items.map((item) => (
              <article key={item.id} className="rounded-xl border border-border bg-surface p-4 shadow-card">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h2 className="font-heading font-bold text-primary">{item.title}</h2>
                    <p className="mt-1 text-xs text-muted">
                      {item.division_name ?? "BEM FMIKOM"} · {item.period_name}
                    </p>
                  </div>
                  <StatusBadge status={item.status} />
                </div>
                <p className="mt-3 text-xs text-body">
                  {formatDate(item.open_at)} – {formatDate(item.close_at)} · {item.submission_count} pendaftar
                </p>
                <div className="mt-3 flex flex-wrap gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => router.push(`/admin/pendaftaran/${item.id}/builder`)}
                  >
                    Builder
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => router.push(`/admin/pendaftaran/${item.id}/submissions`)}
                  >
                    Data
                  </Button>
                  {item.status === "DRAFT" ? (
                    <Button size="sm" onClick={() => void changeStatus(item, "open")} disabled={busy}>
                      Buka
                    </Button>
                  ) : item.status === "OPEN" ? (
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => void changeStatus(item, "close")}
                      disabled={busy}
                    >
                      Tutup
                    </Button>
                  ) : null}
                </div>
              </article>
            ))}
            {items.length === 0 ? (
              <p className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted">
                Belum ada formulir pendaftaran.
              </p>
            ) : null}
          </div>
          <div className="mt-4 flex items-center justify-between text-sm text-muted">
            <span>{total} formulir</span>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                disabled={page <= 1}
                onClick={() => setPage((value) => value - 1)}
              >
                Sebelumnya
              </Button>
              <span className="px-2 py-2">
                {page} / {Math.max(1, totalPages)}
              </span>
              <Button
                size="sm"
                variant="outline"
                disabled={page >= totalPages}
                onClick={() => setPage((value) => value + 1)}
              >
                Berikutnya
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export function RegistrationBuilder({ id }: { id: number }) {
  const router = useRouter();
  const { user } = useAuth();
  const [form, setForm] = React.useState<RegistrationForm | null>(null);
  const [fields, setFields] = React.useState<RegistrationField[]>([]);
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [meta, setMeta] = React.useState({
    title: "",
    description: "",
    period_id: "",
    division_id: "",
    open_at: "",
    close_at: "",
  });
  const [mobileTab, setMobileTab] = React.useState<"form" | "settings">("form");
  const [loading, setLoading] = React.useState(true);
  const [busy, setBusy] = React.useState(false);
  const [message, setMessage] = React.useState("");
  const locked = Boolean(form && form.submission_count > 0);
  const canChooseDivision = user?.role !== "KEPALA_BIDANG";

  const load = React.useCallback(() => {
    let active = true;
    Promise.all([
      getRegistration(id),
      getPeriods({ page: 1, limit: 100 }),
      getDivisions({ page: 1, limit: 100 }),
    ])
      .then(([result, p, d]) => {
        if (!active) return;
        setForm(result.form);
        setFields(result.fields);
        setPeriods(p.items);
        setDivisions(d.items);
        setMeta({
          title: result.form.title,
          description: result.form.description ?? "",
          period_id: String(result.form.period_id),
          division_id: result.form.division_id ? String(result.form.division_id) : "",
          open_at: toDateTimeLocal(result.form.open_at),
          close_at: toDateTimeLocal(result.form.close_at),
        });
      })
      .catch(() => {
        if (active) setMessage("Data formulir gagal dimuat.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [id]);
  React.useEffect(() => load(), [load]);

  async function saveMetadata(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setMessage("");
    try {
      const updated = await updateRegistration(id, {
        title: meta.title,
        description: meta.description || null,
        period_id: Number(meta.period_id),
        division_id: meta.division_id ? Number(meta.division_id) : null,
        open_at: new Date(meta.open_at).toISOString(),
        close_at: new Date(meta.close_at).toISOString(),
      });
      setForm((current) => (current ? { ...current, ...updated } : updated));
      setMessage("Perubahan formulir tersimpan.");
    } catch {
      setMessage("Metadata tidak dapat disimpan. Periksa judul dan rentang waktu.");
    } finally {
      setBusy(false);
    }
  }
  function changeField(index: number, changes: Partial<RegistrationField>) {
    setFields((current) => current.map((field, at) => (at === index ? { ...field, ...changes } : field)));
  }
  function addField(type: RegistrationFieldType) {
    const base = type.toLowerCase();
    let key = base;
    let suffix = 2;
    while (fields.some((field) => field.field_key === key)) key = `${base}_${suffix++}`;
    setFields((current) => [
      ...current,
      {
        field_key: key,
        label: "Pertanyaan baru",
        field_type: type,
        placeholder: "",
        required: false,
        options: ["SELECT", "RADIO", "CHECKBOX"].includes(type) ? [{ value: "opsi_1", label: "Opsi 1" }] : [],
      },
    ]);
  }
  async function saveFields() {
    setBusy(true);
    setMessage("");
    try {
      const result = await saveRegistrationFields(id, fields);
      setFields(result);
      setMessage("Struktur formulir tersimpan.");
    } catch (error) {
      setMessage(
        error instanceof ApiError && error.status === 409
          ? "Struktur formulir tidak dapat diubah karena sudah memiliki data pendaftar."
          : "Struktur formulir gagal disimpan. Periksa key, core field, dan opsi.",
      );
    } finally {
      setBusy(false);
    }
  }
  async function openForm() {
    setBusy(true);
    try {
      const value = await setRegistrationStatus(id, "open");
      setForm(value);
      setMessage("Pendaftaran berhasil dibuka.");
    } catch {
      setMessage("Formulir belum siap dibuka. Lengkapi metadata dan field wajib terlebih dahulu.");
    } finally {
      setBusy(false);
    }
  }
  async function closeForm() {
    setBusy(true);
    try {
      const value = await setRegistrationStatus(id, "close");
      setForm(value);
      setMessage("Pendaftaran ditutup.");
    } catch {
      setMessage("Pendaftaran tidak dapat ditutup.");
    } finally {
      setBusy(false);
    }
  }

  if (loading)
    return (
      <div role="status" className="p-8 text-center text-sm text-muted">
        Memuat builder…
      </div>
    );
  if (!form)
    return (
      <div role="alert" className="p-8 text-center text-sm text-danger">
        Formulir tidak ditemukan.
      </div>
    );
  return (
    <div className="mx-auto w-full max-w-[1600px] p-4 md:p-6 lg:p-8">
      <nav aria-label="Breadcrumb" className="mb-4 text-xs text-muted">
        <Link href="/admin/pendaftaran" className="hover:text-primary">
          Pendaftaran
        </Link>
        <span className="mx-2">/</span>Form Builder
      </nav>
      <AdminPageHeader
        title={form.title}
        description="Atur identitas, susunan pertanyaan, dan tampilan formulir pendaftaran."
        actions={
          <div className="flex gap-2">
            {form.status === "DRAFT" ? (
              <Button onClick={() => void openForm()} disabled={busy}>
                Buka Pendaftaran
              </Button>
            ) : form.status === "OPEN" ? (
              <Button variant="secondary" onClick={() => void closeForm()} disabled={busy}>
                Tutup Pendaftaran
              </Button>
            ) : (
              <span className="self-center text-sm font-semibold text-muted">Ditutup</span>
            )}
            <Button variant="outline" onClick={() => router.push(`/admin/pendaftaran/${id}/submissions`)}>
              Data Pendaftar
            </Button>
          </div>
        }
      />
      {message ? (
        <p role="status" className="mb-4 rounded-lg bg-primary/5 p-3 text-sm text-primary">
          {message}
        </p>
      ) : null}
      {locked ? (
        <p className="mb-4 rounded-lg border border-warning/30 bg-warning/5 p-3 text-sm text-body">
          Struktur formulir tidak dapat diubah karena sudah memiliki data pendaftar. Anda masih dapat mengubah
          label dan metadata yang aman.
        </p>
      ) : null}
      <div className="mb-4 grid grid-cols-2 rounded-lg border border-border bg-surface p-1 md:hidden">
        <button
          type="button"
          aria-pressed={mobileTab === "form"}
          onClick={() => setMobileTab("form")}
          className={`min-h-10 rounded-md text-sm font-semibold ${mobileTab === "form" ? "bg-primary text-primary-foreground" : "text-muted"}`}
        >
          Form
        </button>
        <button
          type="button"
          aria-pressed={mobileTab === "settings"}
          onClick={() => setMobileTab("settings")}
          className={`min-h-10 rounded-md text-sm font-semibold ${mobileTab === "settings" ? "bg-primary text-primary-foreground" : "text-muted"}`}
        >
          Field Settings
        </button>
      </div>
      <div className="grid gap-5 lg:grid-cols-[minmax(0,1.3fr)_minmax(19rem,0.7fr)]">
        <section
          className={`${mobileTab === "settings" ? "hidden md:block" : ""} space-y-4`}
          aria-label="Preview formulir"
        >
          <div className="rounded-xl border border-border bg-surface p-5 shadow-card sm:p-7">
            <p className="text-xs font-bold uppercase tracking-wide text-secondary">Preview Formulir</p>
            <h2 className="mt-2 font-heading text-2xl font-bold text-primary">
              {meta.title || "Judul pendaftaran"}
            </h2>
            <p className="mt-2 whitespace-pre-wrap text-sm text-muted">
              {meta.description || "Deskripsi formulir akan tampil di sini."}
            </p>
            <div className="mt-6 space-y-4">
              {fields.map((field) => (
                <div key={field.field_key} className="rounded-lg border border-border p-3">
                  <p className="text-sm font-semibold text-primary">
                    {field.label}
                    {field.required ? <span className="ml-1 text-danger">*</span> : null}
                  </p>
                  <p className="mt-1 text-xs text-muted">
                    {field.field_type}
                    {field.options.length ? ` · ${field.options.length} opsi` : ""}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>
        <section
          className={`${mobileTab === "form" ? "hidden md:block" : ""} space-y-4`}
          aria-label="Pengaturan field formulir"
        >
          <form
            onSubmit={saveMetadata}
            className="space-y-3 rounded-xl border border-border bg-surface p-4 shadow-card"
          >
            <h2 className="font-heading text-lg font-bold text-primary">Informasi Pendaftaran</h2>
            <div>
              <Label htmlFor="builder-title">Judul</Label>
              <Input
                id="builder-title"
                required
                maxLength={200}
                value={meta.title}
                onChange={(e) => setMeta({ ...meta, title: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="builder-description">Deskripsi</Label>
              <Textarea
                id="builder-description"
                maxLength={5000}
                value={meta.description}
                onChange={(e) => setMeta({ ...meta, description: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="builder-period">Periode</Label>
              <Select
                id="builder-period"
                value={meta.period_id}
                disabled={locked}
                onChange={(e) => setMeta({ ...meta, period_id: e.target.value })}
              >
                {periods.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
              </Select>
            </div>
            {canChooseDivision ? (
              <div>
                <Label htmlFor="builder-division">Divisi</Label>
                <Select
                  id="builder-division"
                  value={meta.division_id}
                  disabled={locked}
                  onChange={(e) => setMeta({ ...meta, division_id: e.target.value })}
                >
                  <option value="">BEM FMIKOM</option>
                  {divisions.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
                </Select>
              </div>
            ) : null}
            <div>
              <Label htmlFor="builder-open">Mulai</Label>
              <Input
                id="builder-open"
                type="datetime-local"
                disabled={locked}
                value={meta.open_at}
                onChange={(e) => setMeta({ ...meta, open_at: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="builder-close">Tutup</Label>
              <Input
                id="builder-close"
                type="datetime-local"
                value={meta.close_at}
                onChange={(e) => setMeta({ ...meta, close_at: e.target.value })}
              />
            </div>
            <Button type="submit" disabled={busy}>
              Simpan Informasi
            </Button>
          </form>
          <div className="space-y-3 rounded-xl border border-border bg-surface p-4 shadow-card">
            <div className="flex items-center justify-between gap-2">
              <h2 className="font-heading text-lg font-bold text-primary">Field Settings</h2>
              <span className="text-xs text-muted">{fields.length} / 50</span>
            </div>
            <div className="flex flex-wrap gap-2">
              {fieldTypes.map((type) => (
                <Button
                  key={type}
                  type="button"
                  size="sm"
                  variant="outline"
                  disabled={locked || fields.length >= 50}
                  onClick={() => addField(type)}
                >
                  <Plus className="size-3" />
                  {type.toLowerCase()}
                </Button>
              ))}
            </div>
            {fields.map((field, index) => (
              <FieldEditor
                key={field.field_key}
                field={field}
                index={index}
                total={fields.length}
                locked={locked}
                onChange={(changes) => changeField(index, changes)}
                onRemove={() => setFields((current) => current.filter((_, at) => at !== index))}
                onMove={(direction) => setFields((current) => move(current, index, direction))}
              />
            ))}
            <Button type="button" onClick={() => void saveFields()} disabled={busy}>
              Simpan Struktur Form
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}

function FieldEditor({
  field,
  index,
  total,
  locked,
  onChange,
  onRemove,
  onMove,
}: {
  field: RegistrationField;
  index: number;
  total: number;
  locked: boolean;
  onChange: (changes: Partial<RegistrationField>) => void;
  onRemove: () => void;
  onMove: (direction: number) => void;
}) {
  const core = coreKeys.has(field.field_key);
  const choices = ["SELECT", "RADIO", "CHECKBOX"].includes(field.field_type);
  function updateOption(index: number, changes: Partial<RegistrationField["options"][number]>) {
    onChange({
      options: field.options.map((option, at) => (at === index ? { ...option, ...changes } : option)),
    });
  }
  return (
    <article className="space-y-3 rounded-lg border border-border p-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-bold uppercase text-muted">
          Field {index + 1}
          {core ? " · Identitas wajib" : ""}
        </span>
        <div className="flex gap-1">
          <Button
            type="button"
            aria-label={`Naikkan ${field.label}`}
            size="icon"
            variant="ghost"
            disabled={locked || index === 0}
            onClick={() => onMove(-1)}
          >
            <ArrowUp className="size-4" />
          </Button>
          <Button
            type="button"
            aria-label={`Turunkan ${field.label}`}
            size="icon"
            variant="ghost"
            disabled={locked || index === total - 1}
            onClick={() => onMove(1)}
          >
            <ArrowDown className="size-4" />
          </Button>
          {!core ? (
            <Button
              type="button"
              aria-label={`Hapus ${field.label}`}
              size="icon"
              variant="ghost"
              disabled={locked}
              onClick={onRemove}
            >
              <Trash2 className="size-4 text-danger" />
            </Button>
          ) : null}
        </div>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <Label htmlFor={`field-label-${index}`}>Label</Label>
          <Input
            id={`field-label-${index}`}
            maxLength={160}
            value={field.label}
            onChange={(e) => onChange({ label: e.target.value })}
          />
        </div>
        <div>
          <Label htmlFor={`field-key-${index}`}>Key</Label>
          <Input
            id={`field-key-${index}`}
            disabled={locked || core}
            pattern="[a-z][a-z0-9]*(?:_[a-z0-9]+)*"
            value={field.field_key}
            onChange={(e) => onChange({ field_key: e.target.value })}
          />
        </div>
        <div>
          <Label htmlFor={`field-type-${index}`}>Tipe</Label>
          <Select
            id={`field-type-${index}`}
            disabled={locked || core}
            value={field.field_type}
            onChange={(e) => {
              const field_type = e.target.value as RegistrationFieldType;
              onChange({
                field_type,
                options: ["SELECT", "RADIO", "CHECKBOX"].includes(field_type)
                  ? field.options.length
                    ? field.options
                    : [{ value: "opsi_1", label: "Opsi 1" }]
                  : [],
              });
            }}
          >
            {fieldTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </Select>
        </div>
        <div>
          <Label htmlFor={`field-placeholder-${index}`}>
            {field.field_type === "IMAGE" ? "Deskripsi bantuan (opsional)" : "Placeholder"}
          </Label>
          <Input
            id={`field-placeholder-${index}`}
            maxLength={255}
            value={field.placeholder ?? ""}
            onChange={(e) => onChange({ placeholder: e.target.value })}
            placeholder={field.field_type === "IMAGE" ? "Contoh: Unggah foto formal" : undefined}
          />
        </div>
        <div className="flex min-h-10 items-center gap-2 text-sm text-body">
          <Checkbox
            id={`field-required-${index}`}
            checked={field.required}
            disabled={locked || core}
            onChange={(e) => onChange({ required: e.target.checked })}
          />
          <Label htmlFor={`field-required-${index}`}>Wajib diisi</Label>
        </div>
      </div>
      {field.field_type === "IMAGE" ? (
        <div className="space-y-2 rounded-md border border-dashed border-border bg-surface-muted p-4">
          <p className="text-sm font-semibold text-primary">Pratinjau unggah gambar</p>
          <div className="grid min-h-28 place-items-center rounded-lg bg-surface text-sm text-muted">
            <span>Seret gambar atau pilih dari perangkat</span>
          </div>
          <p className="text-xs text-muted">JPEG, PNG, WebP, GIF statis, TIFF, BMP · maks. 5 MiB</p>
        </div>
      ) : choices ? (
        <div className="space-y-2 rounded-md bg-surface-muted p-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-primary">Pilihan</span>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={locked || field.options.length >= 100}
              onClick={() =>
                onChange({
                  options: [
                    ...field.options,
                    { value: `opsi_${field.options.length + 1}`, label: `Opsi ${field.options.length + 1}` },
                  ],
                })
              }
            >
              <Plus className="size-3" />
              Tambah opsi
            </Button>
          </div>
          {field.options.map((option, optionIndex) => (
            <div key={option.value} className="grid grid-cols-[1fr_1fr_auto] gap-2">
              <Input
                aria-label={`Nilai opsi ${optionIndex + 1}`}
                disabled={locked}
                value={option.value}
                onChange={(e) => updateOption(optionIndex, { value: e.target.value })}
              />
              <Input
                aria-label={`Label opsi ${optionIndex + 1}`}
                disabled={locked}
                value={option.label}
                onChange={(e) => updateOption(optionIndex, { label: e.target.value })}
              />
              <Button
                type="button"
                size="icon"
                variant="ghost"
                aria-label="Hapus opsi"
                disabled={locked || field.options.length <= 1}
                onClick={() => onChange({ options: field.options.filter((_, at) => at !== optionIndex) })}
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
          ))}
        </div>
      ) : null}
    </article>
  );
}

export function RegistrationSubmissionList({ id }: { id: number }) {
  const router = useRouter();
  const [form, setForm] = React.useState<RegistrationForm | null>(null);
  const [items, setItems] = React.useState<RegistrationSubmission[]>([]);
  const [search, setSearch] = React.useState("");
  const [page, setPage] = React.useState(1);
  const [total, setTotal] = React.useState(0);
  const [totalPages, setTotalPages] = React.useState(0);
  const [error, setError] = React.useState("");
  const [loading, setLoading] = React.useState(true);
  const [busy, setBusy] = React.useState(false);
  const load = React.useCallback(() => {
    let active = true;
    setLoading(true);
    Promise.all([getRegistration(id), getRegistrationSubmissions(id, { page, limit: 20, search })])
      .then(([detail, result]) => {
        if (active) {
          setForm(detail.form);
          setItems(result.items);
          setTotal(result.meta.total);
          setTotalPages(result.meta.total_pages);
        }
      })
      .catch(() => {
        if (active) setError("Data pendaftar gagal dimuat.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [id, page, search]);
  React.useEffect(() => load(), [load]);
  async function exportData() {
    if (
      !window.confirm(
        "Data pendaftar mengandung informasi pribadi. Pastikan file hasil ekspor digunakan dan disimpan sesuai kebutuhan organisasi.",
      )
    )
      return;
    setBusy(true);
    setError("");
    try {
      const blob = await downloadRegistrationExport(id, search);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = `pendaftar-${id}.xlsx`;
      anchor.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      setError(
        error instanceof ApiError && error.code === "EXPORT_LIMIT_EXCEEDED"
          ? "Ekspor dibatasi hingga 1.000 baris. Gunakan pencarian untuk mempersempit data."
          : "File ekspor gagal dibuat.",
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="mx-auto w-full max-w-[1600px] p-4 md:p-6 lg:p-8">
      <nav aria-label="Breadcrumb" className="mb-4 text-xs text-muted">
        <Link href="/admin/pendaftaran" className="hover:text-primary">
          Pendaftaran
        </Link>
        <span className="mx-2">/</span>Data Pendaftar
      </nav>
      <AdminPageHeader
        title="Data Pendaftar"
        description={form?.title ?? "Lihat data dan jawaban pendaftar."}
        actions={
          form ? (
            <Button onClick={() => void exportData()} loading={busy}>
              <Download className="size-4" />
              Ekspor XLSX
            </Button>
          ) : null
        }
      />
      {error ? (
        <p role="alert" className="mb-4 rounded-lg bg-danger/5 p-3 text-sm text-danger">
          {error}
        </p>
      ) : null}
      <div className="mb-4 max-w-lg">
        <Label htmlFor="submission-search" className="sr-only">
          Cari pendaftar
        </Label>
        <div className="relative">
          <Search className="absolute left-3 top-3 size-4 text-muted" />
          <Input
            id="submission-search"
            className="pl-9"
            placeholder="Cari nama, NIM, atau email…"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>
      </div>
      {loading ? (
        <p role="status" className="p-8 text-center text-sm text-muted">
          Memuat data pendaftar…
        </p>
      ) : error ? null : (
        <>
          <div className="hidden overflow-x-auto rounded-xl border border-border bg-surface shadow-card md:block">
            <table className="w-full min-w-[760px] text-left text-sm">
              <thead className="bg-surface-muted text-xs uppercase text-muted">
                <tr>
                  <th className="p-4">Nama</th>
                  <th className="p-4">NIM</th>
                  <th className="p-4">Email</th>
                  <th className="p-4">No. HP</th>
                  <th className="p-4">Waktu</th>
                  <th className="p-4">Detail</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {items.map((item) => (
                  <tr key={item.id}>
                    <td className="p-4 font-semibold text-primary">{item.full_name}</td>
                    <td className="p-4">{item.nim}</td>
                    <td className="p-4">{item.email}</td>
                    <td className="p-4">{item.phone}</td>
                    <td className="p-4 text-xs">{formatDate(item.submitted_at)}</td>
                    <td className="p-4">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => router.push(`/admin/pendaftaran/${id}/submissions/${item.id}`)}
                      >
                        Lihat jawaban
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {items.length === 0 ? (
              <p className="p-8 text-center text-sm text-muted">Belum ada data pendaftar.</p>
            ) : null}
          </div>
          <div className="grid gap-3 md:hidden">
            {items.map((item) => (
              <article key={item.id} className="rounded-xl border border-border bg-surface p-4">
                <h2 className="font-semibold text-primary">{item.full_name}</h2>
                <p className="mt-1 text-sm text-body">NIM {item.nim}</p>
                <p className="mt-1 break-all text-xs text-muted">
                  {item.email} · {item.phone}
                </p>
                <p className="mt-2 text-xs text-muted">{formatDate(item.submitted_at)}</p>
                <Button
                  className="mt-3 w-full"
                  size="sm"
                  variant="outline"
                  onClick={() => router.push(`/admin/pendaftaran/${id}/submissions/${item.id}`)}
                >
                  Lihat jawaban
                </Button>
              </article>
            ))}
            {items.length === 0 ? (
              <p className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted">
                Belum ada data pendaftar.
              </p>
            ) : null}
          </div>
          <div className="mt-4 flex items-center justify-between text-sm text-muted">
            <span>{total} pendaftar</span>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                disabled={page <= 1}
                onClick={() => setPage((value) => value - 1)}
              >
                Sebelumnya
              </Button>
              <span className="px-2 py-2">
                {page} / {Math.max(1, totalPages)}
              </span>
              <Button
                size="sm"
                variant="outline"
                disabled={page >= totalPages}
                onClick={() => setPage((value) => value + 1)}
              >
                Berikutnya
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export function RegistrationSubmissionDetail({ id, submissionId }: { id: number; submissionId: number }) {
  const [item, setItem] = React.useState<RegistrationSubmission | null>(null);
  const [error, setError] = React.useState("");
  React.useEffect(() => {
    let active = true;
    getRegistrationSubmission(id, submissionId)
      .then((data) => {
        if (active) setItem(data);
      })
      .catch(() => {
        if (active) setError("Data pendaftar tidak ditemukan atau Anda tidak memiliki akses.");
      });
    return () => {
      active = false;
    };
  }, [id, submissionId]);
  if (error)
    return (
      <div role="alert" className="p-8 text-center text-sm text-danger">
        {error}
      </div>
    );
  if (!item)
    return (
      <div role="status" className="p-8 text-center text-sm text-muted">
        Memuat jawaban…
      </div>
    );
  const value = (answer: NonNullable<RegistrationSubmission["answers"]>[number]) =>
    answer.value_text ??
    (Array.isArray(answer.value_json) ? answer.value_json.join("; ") : String(answer.value_json ?? "—"));
  return (
    <div className="mx-auto w-full max-w-4xl p-4 md:p-6 lg:p-8">
      <nav aria-label="Breadcrumb" className="mb-4 text-xs text-muted">
        <Link href="/admin/pendaftaran" className="hover:text-primary">
          Pendaftaran
        </Link>
        <span className="mx-2">/</span>
        <Link href={`/admin/pendaftaran/${id}/submissions`} className="hover:text-primary">
          Data Pendaftar
        </Link>
        <span className="mx-2">/</span>Detail
      </nav>
      <AdminPageHeader
        title={item.full_name}
        description={`${item.registration_title} · ${formatDate(item.submitted_at)}`}
      />
      <div className="grid gap-5 md:grid-cols-2">
        <section className="rounded-xl border border-border bg-surface p-5 shadow-card">
          <h2 className="font-heading text-lg font-bold text-primary">Identitas</h2>
          <dl className="mt-4 space-y-4">
            {[
              ["Nama Lengkap", item.full_name],
              ["NIM", item.nim],
              ["Email", item.email],
              ["Nomor HP", item.phone],
              ["Divisi", item.division_name ?? "BEM FMIKOM"],
            ].map(([label, val]) => (
              <div key={label}>
                <dt className="text-xs text-muted">{label}</dt>
                <dd className="mt-1 break-words text-sm font-semibold text-body">{val}</dd>
              </div>
            ))}
          </dl>
        </section>
        <section className="rounded-xl border border-border bg-surface p-5 shadow-card">
          <h2 className="font-heading text-lg font-bold text-primary">Jawaban Formulir</h2>
          <dl className="mt-4 space-y-4">
            {(item.answers ?? [])
              .filter((answer) => !coreKeys.has(answer.field_key))
              .map((answer) => (
                <div key={answer.field_key}>
                  <dt className="text-xs text-muted">{answer.label}</dt>
                  <dd className="mt-1 whitespace-pre-wrap break-words text-sm text-body">
                    {answer.type === "IMAGE" ? (
                      answer.attachment ? (
                        <PrivateRegistrationAttachment attachment={answer.attachment} />
                      ) : (
                        "Tidak ada gambar dilampirkan."
                      )
                    ) : (
                      value(answer)
                    )}
                  </dd>
                </div>
              ))}
          </dl>
          {!item.answers?.some((answer) => !coreKeys.has(answer.field_key)) ? (
            <p className="mt-3 text-sm text-muted">Tidak ada jawaban tambahan.</p>
          ) : null}
        </section>
      </div>
    </div>
  );
}

function PrivateRegistrationAttachment({
  attachment,
}: {
  attachment: NonNullable<NonNullable<RegistrationSubmission["answers"]>[number]["attachment"]>;
}) {
  const [url, setUrl] = React.useState("");
  React.useEffect(() => {
    let active = true;
    let objectURL = "";
    getRegistrationAttachmentBlob(attachment.download_url)
      .then((blob) => {
        objectURL = URL.createObjectURL(blob);
        if (active) setUrl(objectURL);
        else URL.revokeObjectURL(objectURL);
      })
      .catch(() => {
        if (active) setUrl("");
      });
    return () => {
      active = false;
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [attachment.download_url]);
  return (
    <div className="max-w-md space-y-2 rounded-lg border border-border bg-surface-muted p-3">
      {url ? (
        <div className="relative aspect-[4/3] w-full overflow-hidden rounded-md bg-surface">
          <MediaImage
            value={url}
            alt={`Lampiran ${attachment.original_filename}`}
            fill
            allowBlobPreview
            className="object-contain object-center"
          />
        </div>
      ) : (
        <p role="status" className="text-xs text-muted">
          Memuat lampiran privat…
        </p>
      )}
      <p className="break-all text-xs text-body">
        {attachment.original_filename} · {(attachment.size_bytes / (1024 * 1024)).toFixed(2)} MiB
      </p>
      {url ? (
        <a
          href={url}
          target="_blank"
          rel="noreferrer"
          className="inline-flex min-h-10 items-center font-semibold text-primary"
        >
          Buka gambar
        </a>
      ) : null}
    </div>
  );
}

function StatusBadge({ status }: { status: RegistrationForm["status"] }) {
  const label = status === "DRAFT" ? "Draft" : status === "OPEN" ? "Dibuka" : "Ditutup";
  return (
    <span
      className={`inline-flex rounded-full px-2.5 py-1 text-xs font-bold ${status === "OPEN" ? "bg-success/10 text-success" : status === "CLOSED" ? "bg-surface-muted text-muted" : "bg-warning/10 text-warning"}`}
    >
      {label}
    </span>
  );
}
function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeZone: "Asia/Jakarta" }).format(
    new Date(value),
  );
}
function toDateTimeLocal(value: string) {
  const date = new Date(value);
  const offset = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}
function move<T>(list: T[], index: number, direction: number) {
  const target = index + direction;
  if (target < 0 || target >= list.length) return list;
  const next = [...list];
  [next[index], next[target]] = [next[target], next[index]];
  return next;
}
