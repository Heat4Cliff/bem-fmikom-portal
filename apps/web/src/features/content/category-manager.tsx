"use client";

import * as React from "react";
import { AdminBreadcrumb, AdminPageHeader } from "@/components/admin/admin-page-header";
import { apiRequest } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import type { ContentCategory } from "./types";

const blank = { name: "", slug: "", description: "" };

export function CategoryManager() {
  const { user } = useAuth();
  const [items, setItems] = React.useState<ContentCategory[]>([]);
  const [form, setForm] = React.useState(blank);
  const [editing, setEditing] = React.useState<number | null>(null);
  const [message, setMessage] = React.useState("");
  const [error, setError] = React.useState("");
  const [busy, setBusy] = React.useState(false);
  const canManage = user?.role !== "KEPALA_BIDANG";
  const load = React.useCallback(async () => {
    setBusy(true);
    try {
      setItems(await apiRequest<ContentCategory[]>("/admin/article-categories"));
      setError("");
    } catch {
      setError("Kategori gagal dimuat.");
    } finally {
      setBusy(false);
    }
  }, []);
  React.useEffect(() => {
    void load();
  }, [load]);
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setMessage("");
    try {
      await apiRequest(editing ? `/admin/article-categories/${editing}` : "/admin/article-categories", {
        method: editing ? "PATCH" : "POST",
        body: { ...form, description: form.description || null },
      });
      setForm(blank);
      setEditing(null);
      setMessage(editing ? "Kategori diperbarui." : "Kategori dibuat.");
      await load();
    } catch {
      setError("Kategori gagal disimpan. Periksa slug dan pastikan kategori belum dipakai.");
    }
  }
  function edit(item: ContentCategory) {
    setEditing(item.id);
    setForm({ name: item.name, slug: item.slug, description: item.description ?? "" });
    setMessage("");
  }
  async function remove(item: ContentCategory) {
    if (!window.confirm(`Hapus kategori “${item.name}”?`)) return;
    try {
      await apiRequest(`/admin/article-categories/${item.id}`, { method: "DELETE" });
      setMessage("Kategori dihapus.");
      await load();
    } catch {
      setError("Kategori tidak dapat dihapus jika masih digunakan artikel.");
    }
  }
  return (
    <div className="mx-auto max-w-5xl px-4 py-6 md:px-6 lg:px-8">
      <AdminBreadcrumb items={[{ label: "Berita", href: "/admin/berita" }, { label: "Kategori" }]} />
      <AdminPageHeader
        title="Kategori Berita"
        description="Kategori merupakan taksonomi bersama. Kepala bidang dapat melihat, sedangkan SUPER_ADMIN dan BPH mengelola."
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
      {canManage ? (
        <form
          onSubmit={submit}
          className="mt-6 grid gap-3 rounded-xl border border-border/70 bg-surface p-5 shadow-card md:grid-cols-2"
        >
          <h2 className="md:col-span-2 font-heading text-lg font-bold text-primary">
            {editing ? "Edit kategori" : "Tambah kategori"}
          </h2>
          <label className="text-sm font-semibold">
            Nama
            <input
              required
              maxLength={120}
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
            />
          </label>
          <label className="text-sm font-semibold">
            Slug
            <input
              value={form.slug}
              onChange={(e) => setForm({ ...form, slug: e.target.value })}
              placeholder="otomatis-dari-nama"
              className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
            />
          </label>
          <label className="text-sm font-semibold md:col-span-2">
            Deskripsi
            <textarea
              maxLength={2048}
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              rows={2}
              className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 font-normal"
            />
          </label>
          <div className="flex gap-2 md:col-span-2">
            <button
              type="submit"
              disabled={busy}
              className="min-h-10 rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground"
            >
              {editing ? "Simpan perubahan" : "Simpan kategori"}
            </button>
            {editing ? (
              <button
                type="button"
                onClick={() => {
                  setEditing(null);
                  setForm(blank);
                }}
                className="min-h-10 rounded-md border border-border px-4 text-sm"
              >
                Batal
              </button>
            ) : null}
          </div>
        </form>
      ) : null}
      <section className="mt-5 rounded-xl border border-border/70 bg-surface shadow-card">
        <h2 className="border-b border-border/70 p-4 font-semibold text-primary">Daftar kategori</h2>
        {busy ? (
          <p className="p-5 text-sm text-muted">Memuat kategori...</p>
        ) : items.length ? (
          <ul className="divide-y divide-border/70">
            {items.map((item) => (
              <li key={item.id} className="flex flex-wrap items-center justify-between gap-3 p-4">
                <div>
                  <p className="font-semibold text-primary">{item.name}</p>
                  <p className="text-xs text-muted">
                    /{item.slug}
                    {item.description ? ` · ${item.description}` : ""}
                  </p>
                </div>
                {canManage ? (
                  <div className="flex gap-4 text-sm font-semibold">
                    <button type="button" onClick={() => edit(item)} className="text-primary">
                      Edit
                    </button>
                    <button type="button" onClick={() => void remove(item)} className="text-danger">
                      Hapus
                    </button>
                  </div>
                ) : null}
              </li>
            ))}
          </ul>
        ) : (
          <p className="p-5 text-sm text-muted">Belum ada kategori.</p>
        )}
      </section>
    </div>
  );
}
