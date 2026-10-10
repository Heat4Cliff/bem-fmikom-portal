"use client";

import { Download, FileText, LockKeyhole, Pencil, Plus, Search, Trash2, UploadCloud, X } from "lucide-react";
import * as React from "react";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { Button } from "@/components/ui/button";
import { Input, Label, Select } from "@/components/ui/primitives";
import { getDivisions, getPeriods } from "@/features/organization/api";
import type { Division, Period } from "@/features/organization/types";
import { useAuth } from "@/lib/auth/auth-provider";
import {
  deleteDocument,
  downloadDocument,
  getDocument,
  getDocuments,
  updateDocument,
  uploadDocument,
  type ArchiveDocument,
} from "./api";

const categories = [
  { value: "SK", label: "SK" },
  { value: "PROPOSAL", label: "Proposal" },
  { value: "LPJ", label: "LPJ" },
  { value: "OTHER", label: "Lainnya" },
];
const categoryLabel = (value: string) => categories.find((item) => item.value === value)?.label ?? value;
const formatBytes = (bytes: number) =>
  bytes < 1024 * 1024
    ? `${Math.max(1, Math.round(bytes / 1024))} KB`
    : `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
const dateLabel = (value: string) =>
  new Intl.DateTimeFormat("id-ID", { dateStyle: "medium" }).format(new Date(value));

export function DocumentArchive() {
  const { user } = useAuth();
  const [items, setItems] = React.useState<ArchiveDocument[]>([]);
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [page, setPage] = React.useState(1);
  const [totalPages, setTotalPages] = React.useState(1);
  const [search, setSearch] = React.useState("");
  const [category, setCategory] = React.useState("");
  const [periodId, setPeriodId] = React.useState("");
  const [busy, setBusy] = React.useState(false);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  const [uploadOpen, setUploadOpen] = React.useState(false);
  const [edit, setEdit] = React.useState<ArchiveDocument | null>(null);
  const [remove, setRemove] = React.useState<ArchiveDocument | null>(null);
  const [file, setFile] = React.useState<File | null>(null);
  const [dragging, setDragging] = React.useState(false);
  const [uploadPeriod, setUploadPeriod] = React.useState("");
  const canChooseDivision = user?.role !== "KEPALA_BIDANG";

  const load = React.useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await getDocuments({ page, limit: 20, search, category, period_id: periodId });
      setItems(result.items);
      setTotalPages(Math.max(1, result.meta.total_pages));
    } catch {
      setError("Arsip dokumen gagal dimuat. Periksa koneksi dan akses akun.");
    } finally {
      setLoading(false);
    }
  }, [page, search, category, periodId]);
  React.useEffect(() => {
    void load();
  }, [load]);
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
        if (active) setError("Pilihan periode atau divisi gagal dimuat.");
      });
    return () => {
      active = false;
    };
  }, []);

  async function submitUpload(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!file) {
      setError("Pilih dokumen terlebih dahulu.");
      return;
    }
    const form = new FormData(event.currentTarget);
    form.set("file", file);
    if (user?.role === "KEPALA_BIDANG" && user.division?.id)
      form.set("division_id", String(user.division.id));
    setBusy(true);
    setError("");
    try {
      await uploadDocument(form);
      setUploadOpen(false);
      setFile(null);
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Dokumen gagal diunggah.");
    } finally {
      setBusy(false);
    }
  }
  async function doDownload(item: ArchiveDocument) {
    setBusy(true);
    setError("");
    try {
      const result = await downloadDocument(item.id);
      const url = URL.createObjectURL(result.blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = result.filename;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      setError("Unduhan gagal atau akses dokumen tidak tersedia.");
    } finally {
      setBusy(false);
    }
  }
  async function saveEdit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!edit) return;
    const data = new FormData(event.currentTarget);
    setBusy(true);
    setError("");
    try {
      await updateDocument(edit.id, {
        name: String(data.get("name")),
        category: String(data.get("category")),
      });
      setEdit(null);
      await load();
    } catch {
      setError("Metadata dokumen gagal diperbarui.");
    } finally {
      setBusy(false);
    }
  }
  async function openEdit(id: number) {
    setBusy(true);
    setError("");
    try {
      setEdit(await getDocument(id));
    } catch {
      setError("Metadata dokumen gagal dimuat.");
    } finally {
      setBusy(false);
    }
  }
  async function confirmDelete() {
    if (!remove) return;
    setBusy(true);
    setError("");
    try {
      await deleteDocument(remove.id);
      setRemove(null);
      await load();
    } catch {
      setError("Dokumen gagal dihapus. Arsip dapat dicoba kembali.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto max-w-7xl space-y-6">
      <AdminPageHeader
        title="Arsip Dokumen Internal"
        description="Simpan dokumen organisasi dengan akses sesuai periode dan divisi."
        actions={
          <Button
            onClick={() => {
              setError("");
              setFile(null);
              setUploadPeriod(String(divisions.find((d) => d.id === user?.division?.id)?.period_id ?? ""));
              setUploadOpen(true);
            }}
          >
            <Plus className="mr-2 size-4" />
            Unggah Dokumen
          </Button>
        }
      />
      <div className="flex items-start gap-3 rounded-xl border border-primary/15 bg-primary/5 p-4 text-sm text-primary">
        <LockKeyhole className="mt-0.5 size-5 shrink-0" />
        <div>
          <strong>Akses internal BEM FMIKOM</strong>
          <p className="mt-1 text-body">
            Dokumen tersimpan privat dan hanya dapat diunduh oleh akun yang memiliki akses pada periodenya.
            File tidak memiliki tautan publik.
          </p>
        </div>
      </div>
      {error && (
        <div role="alert" className="rounded-lg border border-red-300 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}
      <section
        className="space-y-4 rounded-2xl border border-border/70 bg-surface p-4 shadow-sm md:p-6"
        aria-label="Daftar arsip dokumen"
      >
        <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_200px_220px]">
          <div className="relative">
            <label htmlFor="archive-search" className="sr-only">
              Cari dokumen
            </label>
            <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted" />
            <Input
              id="archive-search"
              className="pl-9"
              placeholder="Cari nama atau file..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
            />
          </div>
          <div>
            <label htmlFor="archive-category" className="sr-only">
              Filter kategori
            </label>
            <Select
              id="archive-category"
              value={category}
              onChange={(e) => {
                setCategory(e.target.value);
                setPage(1);
              }}
            >
              <option value="">Semua Kategori</option>
              {categories.map((item) => (
                <option key={item.value} value={item.value}>
                  {item.label}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <label htmlFor="archive-period" className="sr-only">
              Filter periode
            </label>
            <Select
              id="archive-period"
              value={periodId}
              onChange={(e) => {
                setPeriodId(e.target.value);
                setPage(1);
              }}
            >
              <option value="">Semua Periode</option>
              {periods.map((period) => (
                <option key={period.id} value={period.id}>
                  {period.name}
                </option>
              ))}
            </Select>
          </div>
        </div>
        {loading ? (
          <div className="py-16 text-center text-sm text-muted">Memuat arsip...</div>
        ) : items.length === 0 ? (
          <div className="rounded-xl border border-dashed border-border p-10 text-center">
            <FileText className="mx-auto size-9 text-muted" />
            <h2 className="mt-3 font-heading font-bold text-primary">Belum ada dokumen</h2>
            <p className="mt-1 text-sm text-muted">Unggah dokumen resmi untuk mulai mengisi arsip.</p>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full min-w-[900px] text-left text-sm">
                <thead>
                  <tr className="border-b border-border text-xs uppercase tracking-wide text-muted">
                    <th className="px-3 py-3">Dokumen</th>
                    <th className="px-3 py-3">Kategori</th>
                    <th className="px-3 py-3">Periode / Divisi</th>
                    <th className="px-3 py-3">Diunggah</th>
                    <th className="px-3 py-3">Ukuran</th>
                    <th className="px-3 py-3 text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((item) => (
                    <tr key={item.id} className="border-b border-border/60 last:border-0">
                      <td className="px-3 py-4">
                        <div className="font-semibold text-primary">{item.name}</div>
                        <div className="mt-1 text-xs text-muted">{item.original_filename}</div>
                      </td>
                      <td className="px-3 py-4">{categoryLabel(item.category)}</td>
                      <td className="px-3 py-4">
                        {item.period_name}
                        <div className="mt-1 text-xs text-muted">{item.division_name}</div>
                      </td>
                      <td className="px-3 py-4">
                        {item.uploader_name}
                        <div className="mt-1 text-xs text-muted">{dateLabel(item.created_at)}</div>
                      </td>
                      <td className="px-3 py-4">{formatBytes(item.size_bytes)}</td>
                      <td className="px-3 py-4">
                        <div className="flex justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            aria-label={`Unduh ${item.name}`}
                            disabled={busy}
                            onClick={() => void doDownload(item)}
                          >
                            <Download className="size-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            aria-label={`Edit ${item.name}`}
                            onClick={() => void openEdit(item.id)}
                          >
                            <Pencil className="size-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            aria-label={`Hapus ${item.name}`}
                            onClick={() => setRemove(item)}
                          >
                            <Trash2 className="size-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <ul className="space-y-3 md:hidden">
              {items.map((item) => (
                <li key={item.id} className="rounded-xl border border-border/70 p-4">
                  <div className="flex items-start gap-3">
                    <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/5 text-primary">
                      <FileText className="size-5" />
                    </span>
                    <div className="min-w-0 flex-1">
                      <div className="font-semibold text-primary">{item.name}</div>
                      <p className="mt-1 truncate text-xs text-muted">{item.original_filename}</p>
                      <div className="mt-2 flex flex-wrap gap-2 text-[11px]">
                        <span className="rounded-full bg-primary/5 px-2 py-1 text-primary">
                          {categoryLabel(item.category)}
                        </span>
                        <span className="rounded-full bg-surface-muted px-2 py-1">
                          {item.file_extension.slice(1).toUpperCase()} · {formatBytes(item.size_bytes)}
                        </span>
                      </div>
                      <p className="mt-3 text-xs text-body">
                        {item.period_name} · {item.division_name}
                      </p>
                      <p className="mt-1 text-xs text-muted">
                        {item.uploader_name} · {dateLabel(item.created_at)}
                      </p>
                    </div>
                  </div>
                  <div className="mt-4 flex gap-2">
                    <Button
                      variant="outline"
                      className="flex-1"
                      disabled={busy}
                      onClick={() => void doDownload(item)}
                    >
                      <Download className="mr-2 size-4" />
                      Unduh
                    </Button>
                    <Button
                      variant="outline"
                      size="icon"
                      aria-label={`Edit ${item.name}`}
                      onClick={() => void openEdit(item.id)}
                    >
                      <Pencil className="size-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="icon"
                      aria-label={`Hapus ${item.name}`}
                      onClick={() => setRemove(item)}
                    >
                      <Trash2 className="size-4" />
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
            <div className="flex items-center justify-between border-t border-border/60 pt-4 text-sm">
              <span className="text-muted">
                Halaman {page} dari {totalPages}
              </span>
              <div className="flex gap-2">
                <Button variant="outline" disabled={page <= 1} onClick={() => setPage((v) => v - 1)}>
                  Sebelumnya
                </Button>
                <Button variant="outline" disabled={page >= totalPages} onClick={() => setPage((v) => v + 1)}>
                  Berikutnya
                </Button>
              </div>
            </div>
          </>
        )}
      </section>
      {uploadOpen && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/45 p-3">
          <section
            role="dialog"
            aria-modal="true"
            aria-labelledby="document-upload-title"
            className="max-h-[95vh] w-full max-w-xl overflow-y-auto rounded-2xl bg-surface p-5 shadow-xl md:p-7"
          >
            <div className="mb-5 flex items-start justify-between">
              <div>
                <h2 id="document-upload-title" className="font-heading text-xl font-bold text-primary">
                  Unggah Dokumen
                </h2>
                <p className="mt-1 text-sm text-muted">Pilih lingkup periode dan divisi arsip.</p>
              </div>
              <button
                type="button"
                aria-label="Tutup dialog"
                className="rounded-md p-2 hover:bg-surface-muted"
                onClick={() => setUploadOpen(false)}
              >
                <X className="size-5" />
              </button>
            </div>
            <form className="space-y-4" onSubmit={(e) => void submitUpload(e)}>
              <div>
                <Label htmlFor="document-name">Nama Dokumen</Label>
                <Input
                  id="document-name"
                  name="name"
                  required
                  maxLength={240}
                  placeholder="Contoh: SK Pengurus 2026"
                />
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <Label htmlFor="document-category">Kategori</Label>
                  <Select id="document-category" name="category" required defaultValue="">
                    <option value="" disabled>
                      Pilih kategori
                    </option>
                    {categories.map((item) => (
                      <option key={item.value} value={item.value}>
                        {item.label}
                      </option>
                    ))}
                  </Select>
                </div>
                <div>
                  <Label htmlFor="document-period">Periode</Label>
                  <Select
                    id="document-period"
                    name="period_id"
                    required
                    value={uploadPeriod}
                    onChange={(e) => setUploadPeriod(e.target.value)}
                  >
                    <option value="" disabled>
                      Pilih periode
                    </option>
                    {periods
                      .filter(
                        (period) =>
                          canChooseDivision ||
                          period.id === divisions.find((d) => d.id === user?.division?.id)?.period_id,
                      )
                      .map((period) => (
                        <option key={period.id} value={period.id}>
                          {period.name}
                        </option>
                      ))}
                  </Select>
                </div>
              </div>
              <div>
                <Label htmlFor="document-division">Divisi</Label>
                <Select
                  id="document-division"
                  name="division_id"
                  required
                  defaultValue={user?.role === "KEPALA_BIDANG" ? String(user.division?.id ?? "") : ""}
                  disabled={!canChooseDivision}
                >
                  <option value="" disabled>
                    Pilih divisi
                  </option>
                  {divisions
                    .filter((d) => !uploadPeriod || d.period_id === Number(uploadPeriod))
                    .map((division) => (
                      <option key={division.id} value={division.id}>
                        {division.name}
                      </option>
                    ))}
                </Select>
                {!canChooseDivision && (
                  <>
                    <input type="hidden" name="division_id" value={user.division?.id ?? ""} />
                    <p className="mt-1 text-xs text-muted">Divisi Anda terisi otomatis sesuai hak akses.</p>
                  </>
                )}
              </div>
              <div>
                <Label htmlFor="document-file">File Dokumen</Label>
                <label
                  htmlFor="document-file"
                  onDragOver={(e) => {
                    e.preventDefault();
                    setDragging(true);
                  }}
                  onDragLeave={() => setDragging(false)}
                  onDrop={(e) => {
                    e.preventDefault();
                    setDragging(false);
                    const next = e.dataTransfer.files[0];
                    if (next) setFile(next);
                  }}
                  className={`mt-1 flex min-h-32 cursor-pointer flex-col items-center justify-center rounded-xl border border-dashed p-4 text-center ${dragging ? "border-primary bg-primary/5" : "border-border bg-surface-muted/40"}`}
                >
                  <UploadCloud className="size-7 text-primary" />
                  <span className="mt-2 text-sm font-semibold text-primary">
                    {file?.name ?? "Pilih file atau tarik ke sini"}
                  </span>
                  <span className="mt-1 text-xs text-muted">PDF, DOC, DOCX, XLS, XLSX · maksimal 10 MiB</span>
                  <input
                    id="document-file"
                    type="file"
                    className="sr-only"
                    accept=".pdf,.doc,.docx,.xls,.xlsx"
                    required
                    onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                  />
                </label>
              </div>
              <div className="flex flex-col-reverse gap-2 border-t border-border/60 pt-4 sm:flex-row sm:justify-end">
                <Button type="button" variant="outline" onClick={() => setUploadOpen(false)}>
                  Batal
                </Button>
                <Button type="submit" disabled={busy}>
                  {busy ? "Mengunggah..." : "Simpan ke Arsip"}
                </Button>
              </div>
            </form>
          </section>
        </div>
      )}
      {edit && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/45 p-3">
          <section
            role="dialog"
            aria-modal="true"
            aria-labelledby="document-edit-title"
            className="w-full max-w-md rounded-2xl bg-surface p-6"
          >
            <h2 id="document-edit-title" className="font-heading text-lg font-bold text-primary">
              Edit Metadata
            </h2>
            <form className="mt-4 space-y-4" onSubmit={(e) => void saveEdit(e)}>
              <div>
                <Label htmlFor="edit-name">Nama dokumen</Label>
                <Input id="edit-name" name="name" required maxLength={240} defaultValue={edit.name} />
              </div>
              <div>
                <Label htmlFor="edit-category">Kategori</Label>
                <Select id="edit-category" name="category" defaultValue={edit.category}>
                  {categories.map((item) => (
                    <option key={item.value} value={item.value}>
                      {item.label}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="flex justify-end gap-2">
                <Button type="button" variant="outline" onClick={() => setEdit(null)}>
                  Batal
                </Button>
                <Button disabled={busy}>Simpan</Button>
              </div>
            </form>
          </section>
        </div>
      )}
      {remove && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/45 p-3">
          <section
            role="alertdialog"
            aria-modal="true"
            aria-labelledby="document-delete-title"
            className="w-full max-w-md rounded-2xl bg-surface p-6"
          >
            <h2 id="document-delete-title" className="font-heading text-lg font-bold text-primary">
              Hapus dokumen?
            </h2>
            <p className="mt-2 text-sm text-body">
              File <strong>{remove.name}</strong> akan dihapus dari arsip internal.
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setRemove(null)}>
                Batal
              </Button>
              <Button variant="danger" disabled={busy} onClick={() => void confirmDelete()}>
                Hapus
              </Button>
            </div>
          </section>
        </div>
      )}
    </div>
  );
}
