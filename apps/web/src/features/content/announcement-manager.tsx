"use client";

import * as React from "react";
import { AdminBreadcrumb, AdminPageHeader } from "@/components/admin/admin-page-header";
import { MediaUpload } from "@/components/admin/media-upload";
import { apiRequest, apiRequestEnvelope } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import { announcementPayload } from "./content-payloads";
import { contentFormErrorMessage } from "./form-errors";
import type { Announcement } from "./types";

type Form = {
  title: string;
  slug: string;
  content_html: string;
  image_url: string;
  start_at: string;
  end_at: string;
};
const blank: Form = { title: "", slug: "", content_html: "", image_url: "", start_at: "", end_at: "" };
function localDate(value: string) {
  if (!value) return "";
  const date = new Date(value);
  const offset = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

export function AnnouncementManager() {
  const { user } = useAuth();
  const [items, setItems] = React.useState<Announcement[]>([]);
  const [form, setForm] = React.useState(blank);
  const [editing, setEditing] = React.useState<number | null>(null);
  const [open, setOpen] = React.useState(false);
  const [busy, setBusy] = React.useState(false);
  const [uploadPending, setUploadPending] = React.useState(false);
  const [error, setError] = React.useState("");
  const [message, setMessage] = React.useState("");
  const [search, setSearch] = React.useState("");
  const [canManage, setCanManage] = React.useState(false);
  React.useEffect(() => setCanManage(user?.role === "SUPER_ADMIN" || user?.role === "BPH"), [user]);
  const load = React.useCallback(async () => {
    setBusy(true);
    try {
      const q = new URLSearchParams({ page: "1", limit: "100" });
      if (search) q.set("search", search);
      const result = await apiRequestEnvelope<Announcement[]>(`/admin/announcements?${q}`);
      setItems(result.data);
      setError("");
    } catch {
      setError("Pengumuman gagal dimuat.");
    } finally {
      setBusy(false);
    }
  }, [search]);
  React.useEffect(() => {
    void load();
  }, [load]);
  function beginEdit(item: Announcement) {
    setEditing(item.id);
    setForm({
      title: item.title,
      slug: item.slug,
      content_html: item.content_html,
      image_url: item.image_url ?? "",
      start_at: localDate(item.start_at),
      end_at: localDate(item.end_at ?? ""),
    });
    setOpen(true);
  }
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (busy || uploadPending) return;
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await apiRequest(editing ? `/admin/announcements/${editing}` : "/admin/announcements", {
        method: editing ? "PATCH" : "POST",
        body: announcementPayload(form),
      });
      setMessage(editing ? "Pengumuman diperbarui." : "Pengumuman dibuat sebagai nonaktif.");
      setOpen(false);
      setEditing(null);
      setForm(blank);
      await load();
    } catch (error: unknown) {
      setError(contentFormErrorMessage(error, "Pengumuman gagal disimpan. Periksa data lalu coba lagi."));
    } finally {
      setBusy(false);
    }
  }
  async function toggle(item: Announcement) {
    setError("");
    try {
      await apiRequest(`/admin/announcements/${item.id}/${item.is_active ? "deactivate" : "activate"}`, {
        method: "POST",
        body: {},
      });
      setMessage("Status pengumuman diperbarui.");
      await load();
    } catch {
      setError("Status pengumuman gagal diubah.");
    }
  }
  async function remove(item: Announcement) {
    if (!window.confirm(`Hapus pengumuman “${item.title}”?`)) return;
    try {
      await apiRequest(`/admin/announcements/${item.id}`, { method: "DELETE" });
      setMessage("Pengumuman dihapus.");
      await load();
    } catch {
      setError("Pengumuman gagal dihapus.");
    }
  }
  const statusLabel = (v: Announcement["status"]) =>
    ({ ACTIVE: "Aktif", INACTIVE: "Nonaktif", SCHEDULED: "Terjadwal", EXPIRED: "Kedaluwarsa" })[v];
  return (
    <div className="mx-auto max-w-[1440px] px-4 py-6 md:px-6 lg:px-8">
      <AdminBreadcrumb items={[{ label: "Dashboard", href: "/admin" }, { label: "Pengumuman" }]} />
      <AdminPageHeader
        title="Manajemen Pengumuman Resmi"
        description="Atur isi dan periode tampil pengumuman publik."
        actions={
          canManage ? (
            <button
              type="button"
              onClick={() => {
                setEditing(null);
                setForm(blank);
                setOpen(true);
              }}
              className="min-h-11 rounded-md bg-primary px-5 text-sm font-semibold text-white"
            >
              + Tambah pengumuman
            </button>
          ) : undefined
        }
      />
      {error ? (
        <p role="alert" className="mt-4 rounded-md bg-danger/5 p-3 text-sm text-danger">
          {error}
        </p>
      ) : null}
      {message ? (
        <p role="status" className="mt-4 rounded-md bg-success/10 p-3 text-sm text-success">
          {message}
        </p>
      ) : null}
      <section className="mt-5 rounded-xl border border-border/70 bg-white p-4 shadow-card">
        <label className="block max-w-xl text-sm font-semibold">
          Cari pengumuman
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Judul pengumuman..."
            className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
          />
        </label>
      </section>
      <section
        className="mt-5 overflow-hidden rounded-xl border border-border/70 bg-white shadow-card"
        aria-busy={busy}
      >
        <div className="hidden overflow-x-auto md:block">
          <table className="w-full min-w-[760px] text-left text-sm">
            <thead className="bg-surface-muted text-xs uppercase text-muted">
              <tr>
                <th className="p-4">Judul</th>
                <th className="p-4">Status</th>
                <th className="p-4">Mulai</th>
                <th className="p-4">Berakhir</th>
                <th className="p-4">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/70">
              {items.map((item) => (
                <tr key={item.id}>
                  <td className="max-w-md p-4">
                    <span className="block truncate font-semibold text-primary">{item.title}</span>
                    <span className="text-xs text-muted">/{item.slug}</span>
                  </td>
                  <td className="p-4">
                    <span className="rounded-full bg-surface-muted px-2.5 py-1 text-xs font-semibold">
                      {statusLabel(item.status)}
                    </span>
                  </td>
                  <td className="p-4">{new Date(item.start_at).toLocaleString("id-ID")}</td>
                  <td className="p-4">
                    {item.end_at ? new Date(item.end_at).toLocaleString("id-ID") : "Tanpa batas"}
                  </td>
                  <td className="p-4">
                    {canManage ? (
                      <div className="flex gap-3">
                        <button
                          type="button"
                          onClick={() => beginEdit(item)}
                          className="font-semibold text-primary"
                        >
                          Edit
                        </button>
                        <button
                          type="button"
                          onClick={() => void toggle(item)}
                          className="font-semibold text-primary"
                        >
                          {item.is_active ? "Nonaktifkan" : "Aktifkan"}
                        </button>
                        <button
                          type="button"
                          onClick={() => void remove(item)}
                          className="font-semibold text-danger"
                        >
                          Hapus
                        </button>
                      </div>
                    ) : (
                      <span className="text-xs text-muted">Baca saja</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {items.length === 0 ? (
          <p className="p-8 text-sm text-muted">{busy ? "Memuat pengumuman..." : "Belum ada pengumuman."}</p>
        ) : null}
        <ul className="divide-y divide-border/70 md:hidden">
          {items.map((item) => (
            <li key={item.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <p className="font-semibold text-primary">{item.title}</p>
                <span className="shrink-0 rounded-full bg-surface-muted px-2 py-1 text-[10px] font-semibold">
                  {statusLabel(item.status)}
                </span>
              </div>
              <p className="mt-2 text-xs text-muted">
                Mulai {new Date(item.start_at).toLocaleString("id-ID")}
              </p>
              {canManage ? (
                <div className="mt-3 flex flex-wrap gap-4 text-sm font-semibold">
                  <button type="button" onClick={() => beginEdit(item)} className="text-primary">
                    Edit
                  </button>
                  <button type="button" onClick={() => void toggle(item)} className="text-primary">
                    {item.is_active ? "Nonaktifkan" : "Aktifkan"}
                  </button>
                  <button type="button" onClick={() => void remove(item)} className="text-danger">
                    Hapus
                  </button>
                </div>
              ) : null}
            </li>
          ))}
        </ul>
      </section>
      {open && canManage ? (
        <div className="fixed inset-0 z-50 grid place-items-center bg-midnight/50 p-3" role="presentation">
          <section
            role="dialog"
            aria-modal="true"
            aria-labelledby="announcement-form-title"
            className="max-h-[95dvh] w-full max-w-2xl overflow-y-auto rounded-xl bg-white shadow-float"
          >
            <header className="sticky top-0 flex items-center justify-between border-b border-border bg-surface-muted p-4">
              <h2 id="announcement-form-title" className="font-heading text-lg font-bold text-primary">
                {editing ? "Edit pengumuman" : "Tambah pengumuman"}
              </h2>
              <button
                type="button"
                onClick={() => setOpen(false)}
                aria-label="Tutup form"
                className="size-10 rounded-md hover:bg-white"
              >
                ×
              </button>
            </header>
            <form onSubmit={submit} className="grid gap-4 p-5">
              <label className="text-sm font-semibold">
                Judul
                <input
                  required
                  maxLength={240}
                  value={form.title}
                  onChange={(e) => setForm({ ...form, title: e.target.value })}
                  className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                />
              </label>
              <label className="text-sm font-semibold">
                Slug
                <input
                  value={form.slug}
                  onChange={(e) => setForm({ ...form, slug: e.target.value })}
                  placeholder="otomatis dari judul"
                  className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                />
              </label>
              <div className="grid gap-3 sm:grid-cols-2">
                <label className="text-sm font-semibold">
                  Mulai tampil
                  <input
                    type="datetime-local"
                    required
                    value={form.start_at}
                    onChange={(e) => setForm({ ...form, start_at: e.target.value })}
                    className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                  />
                </label>
                <label className="text-sm font-semibold">
                  Akhir tampil
                  <input
                    type="datetime-local"
                    value={form.end_at}
                    onChange={(e) => setForm({ ...form, end_at: e.target.value })}
                    className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                  />
                </label>
              </div>
              <div className="text-sm font-semibold">
                <MediaUpload
                  id="announcement-image"
                  label="Gambar pengumuman"
                  purpose="ANNOUNCEMENT_IMAGE"
                  value={form.image_url}
                  onChange={(image_url) => setForm((current) => ({ ...current, image_url }))}
                  onUploadStateChange={setUploadPending}
                  disabled={busy}
                  optional
                  compact
                />
              </div>
              <label className="text-sm font-semibold">
                Isi pengumuman
                <textarea
                  required
                  rows={10}
                  value={form.content_html}
                  onChange={(e) => setForm({ ...form, content_html: e.target.value })}
                  className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 font-mono text-xs font-normal leading-6"
                />
              </label>
              <p className="text-xs text-muted">
                Konten rich text disanitasi oleh API. Status awal nonaktif; aktifkan setelah menyimpan.
              </p>
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setOpen(false)}
                  className="min-h-10 rounded-md border border-border px-4 text-sm"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  disabled={busy || uploadPending}
                  className="min-h-10 rounded-md bg-primary px-4 text-sm font-semibold text-white"
                >
                  Simpan
                </button>
              </div>
            </form>
          </section>
        </div>
      ) : null}
    </div>
  );
}
