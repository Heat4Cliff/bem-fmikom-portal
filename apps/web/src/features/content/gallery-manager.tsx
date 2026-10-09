"use client";

import * as React from "react";
import { AdminBreadcrumb, AdminPageHeader } from "@/components/admin/admin-page-header";
import { apiRequest, apiRequestEnvelope } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import type { GalleryCategory, GalleryItem, PageMeta } from "./types";

const blankCategory = {
  title: "",
  slug: "",
  description: "",
  period_id: "",
  division_id: "",
  event_date: "",
};
const blankItem = {
  title: "",
  description: "",
  google_photos_url: "",
  preview_image_url: "",
  sort_order: "0",
};

export function GalleryManager() {
  const { user } = useAuth();
  const [categories, setCategories] = React.useState<GalleryCategory[]>([]);
  const [items, setItems] = React.useState<GalleryItem[]>([]);
  const [selected, setSelected] = React.useState<number>(0);
  const [catForm, setCatForm] = React.useState(blankCategory);
  const [itemForm, setItemForm] = React.useState(blankItem);
  const [editCategory, setEditCategory] = React.useState<number | null>(null);
  const [editItem, setEditItem] = React.useState<number | null>(null);
  const [meta, setMeta] = React.useState<PageMeta>({ page: 1, limit: 100, total: 0, total_pages: 0 });
  const [busy, setBusy] = React.useState(false);
  const [error, setError] = React.useState("");
  const [message, setMessage] = React.useState("");
  const canManageCategories = user?.role !== "KEPALA_BIDANG";
  const loadCategories = React.useCallback(async () => {
    try {
      const res = await apiRequestEnvelope<GalleryCategory[]>("/admin/gallery-categories?page=1&limit=100");
      setCategories(res.data);
      setMeta(res.meta ?? { page: 1, limit: 100, total: res.data.length, total_pages: 1 });
      if (!selected && res.data.length) setSelected(res.data[0].id);
    } catch {
      setError("Kategori galeri gagal dimuat.");
    }
  }, [selected]);
  const loadItems = React.useCallback(async () => {
    if (!selected) {
      setItems([]);
      return;
    }
    try {
      const res = await apiRequestEnvelope<GalleryItem[]>(
        `/admin/gallery-items?category_id=${selected}&page=1&limit=100`,
      );
      setItems(res.data);
    } catch {
      setError("Item galeri gagal dimuat.");
    }
  }, [selected]);
  React.useEffect(() => {
    void loadCategories();
  }, [loadCategories]);
  React.useEffect(() => {
    void loadItems();
  }, [loadItems]);
  async function saveCategory(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const payload = {
        ...catForm,
        description: catForm.description || null,
        period_id: catForm.period_id ? Number(catForm.period_id) : null,
        division_id: catForm.division_id ? Number(catForm.division_id) : null,
        event_date: catForm.event_date || null,
      };
      const result = await apiRequest<GalleryCategory>(
        editCategory ? `/admin/gallery-categories/${editCategory}` : "/admin/gallery-categories",
        { method: editCategory ? "PATCH" : "POST", body: payload },
      );
      setSelected(result.id);
      setCatForm(blankCategory);
      setEditCategory(null);
      setMessage(editCategory ? "Kategori diperbarui." : "Kategori kegiatan dibuat.");
      await loadCategories();
    } catch {
      setError("Kategori tidak dapat disimpan. Periksa slug, tanggal, periode, dan ID divisi.");
    } finally {
      setBusy(false);
    }
  }
  function beginEditCategory(item: GalleryCategory) {
    setEditCategory(item.id);
    setCatForm({
      title: item.title,
      slug: item.slug,
      description: item.description ?? "",
      period_id: item.period_id ? String(item.period_id) : "",
      division_id: item.division_id ? String(item.division_id) : "",
      event_date: item.event_date ?? "",
    });
  }
  async function removeCategory(item: GalleryCategory) {
    if (!window.confirm(`Hapus kategori “${item.title}”?`)) return;
    try {
      await apiRequest(`/admin/gallery-categories/${item.id}`, { method: "DELETE" });
      if (selected === item.id) setSelected(0);
      setMessage("Kategori dihapus.");
      await loadCategories();
    } catch {
      setError("Kategori yang berisi item dokumentasi tidak dapat dihapus.");
    }
  }
  async function saveItem(event: React.FormEvent) {
    event.preventDefault();
    if (!selected) {
      setError("Pilih kategori galeri terlebih dahulu.");
      return;
    }
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await apiRequest(editItem ? `/admin/gallery-items/${editItem}` : "/admin/gallery-items", {
        method: editItem ? "PATCH" : "POST",
        body: {
          ...itemForm,
          category_id: selected,
          description: itemForm.description || null,
          preview_image_url: itemForm.preview_image_url || null,
          sort_order: Number(itemForm.sort_order) || 0,
        },
      });
      setItemForm(blankItem);
      setEditItem(null);
      setMessage(editItem ? "Item diperbarui." : "Item dokumentasi ditambahkan.");
      await loadItems();
    } catch {
      setError("Item galeri gagal disimpan. Tautan album wajib menggunakan HTTPS.");
    } finally {
      setBusy(false);
    }
  }
  function beginEditItem(item: GalleryItem) {
    setEditItem(item.id);
    setItemForm({
      title: item.title,
      description: item.description ?? "",
      google_photos_url: item.google_photos_url,
      preview_image_url: item.preview_image_url ?? "",
      sort_order: String(item.sort_order),
    });
  }
  async function removeItem(item: GalleryItem) {
    if (!window.confirm(`Hapus item “${item.title}”?`)) return;
    try {
      await apiRequest(`/admin/gallery-items/${item.id}`, { method: "DELETE" });
      setMessage("Item galeri dihapus.");
      await loadItems();
    } catch {
      setError("Item galeri gagal dihapus.");
    }
  }
  return (
    <div className="mx-auto max-w-[1440px] px-4 py-6 md:px-6 lg:px-8">
      <AdminBreadcrumb items={[{ label: "Dashboard", href: "/admin" }, { label: "Galeri" }]} />
      <AdminPageHeader
        title="Manajemen Galeri Kegiatan"
        description="Kelola kategori dokumentasi dan tautan album Google Photos."
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
      <section className="mt-5 grid gap-5 xl:grid-cols-[minmax(18rem,0.8fr)_minmax(0,1.2fr)]">
        <div className="space-y-4">
          <section className="rounded-xl border border-border/70 bg-white shadow-card">
            <div className="flex items-center justify-between gap-3 border-b border-border/70 p-4">
              <h2 className="font-heading font-bold text-primary">Kategori kegiatan</h2>
              <span className="text-xs text-muted">{meta.total} album</span>
            </div>
            {categories.length ? (
              <ul className="divide-y divide-border/70">
                {categories.map((item) => (
                  <li key={item.id} className={`p-4 ${selected === item.id ? "bg-primary-container/5" : ""}`}>
                    <button type="button" onClick={() => setSelected(item.id)} className="w-full text-left">
                      <span className="block font-semibold text-primary">{item.title}</span>
                      <span className="text-xs text-muted">
                        {item.item_count ?? 0} item{item.division_name ? ` · ${item.division_name}` : ""}
                      </span>
                    </button>
                    {canManageCategories ? (
                      <div className="mt-2 flex gap-4 text-xs font-semibold">
                        <button
                          type="button"
                          onClick={() => beginEditCategory(item)}
                          className="text-primary"
                        >
                          Edit metadata
                        </button>
                        <button
                          type="button"
                          onClick={() => void removeCategory(item)}
                          className="text-danger"
                        >
                          Hapus
                        </button>
                      </div>
                    ) : null}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="p-4 text-sm text-muted">Belum ada kategori galeri.</p>
            )}
          </section>
          {canManageCategories ? (
            <form
              onSubmit={saveCategory}
              className="grid gap-3 rounded-xl border border-border/70 bg-white p-5 shadow-card"
            >
              <h2 className="font-heading font-bold text-primary">
                {editCategory ? "Edit kategori" : "Tambah kategori"}
              </h2>
              <label className="text-sm font-semibold">
                Judul
                <input
                  required
                  value={catForm.title}
                  onChange={(e) => setCatForm({ ...catForm, title: e.target.value })}
                  className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                />
              </label>
              <label className="text-sm font-semibold">
                Slug
                <input
                  value={catForm.slug}
                  onChange={(e) => setCatForm({ ...catForm, slug: e.target.value })}
                  placeholder="otomatis"
                  className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                />
              </label>
              <label className="text-sm font-semibold">
                Tanggal kegiatan
                <input
                  type="date"
                  value={catForm.event_date}
                  onChange={(e) => setCatForm({ ...catForm, event_date: e.target.value })}
                  className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                />
              </label>
              {user?.role !== "KEPALA_BIDANG" ? (
                <div className="grid grid-cols-2 gap-2">
                  <label className="text-sm font-semibold">
                    ID divisi
                    <input
                      inputMode="numeric"
                      value={catForm.division_id}
                      onChange={(e) => setCatForm({ ...catForm, division_id: e.target.value })}
                      className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                    />
                  </label>
                  <label className="text-sm font-semibold">
                    ID periode
                    <input
                      inputMode="numeric"
                      value={catForm.period_id}
                      onChange={(e) => setCatForm({ ...catForm, period_id: e.target.value })}
                      className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
                    />
                  </label>
                </div>
              ) : null}
              <label className="text-sm font-semibold">
                Deskripsi
                <textarea
                  rows={3}
                  value={catForm.description}
                  onChange={(e) => setCatForm({ ...catForm, description: e.target.value })}
                  className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 font-normal"
                />
              </label>
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={busy}
                  className="min-h-10 rounded-md bg-primary px-4 text-sm font-semibold text-white"
                >
                  Simpan kategori
                </button>
                {editCategory ? (
                  <button
                    type="button"
                    onClick={() => {
                      setEditCategory(null);
                      setCatForm(blankCategory);
                    }}
                    className="min-h-10 rounded-md border px-4 text-sm"
                  >
                    Batal
                  </button>
                ) : null}
              </div>
            </form>
          ) : null}
        </div>
        <div className="space-y-4">
          <section className="rounded-xl border border-border/70 bg-white shadow-card">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border/70 p-4">
              <div>
                <h2 className="font-heading font-bold text-primary">Item dokumentasi</h2>
                <p className="text-xs text-muted">
                  {categories.find((item) => item.id === selected)?.title ?? "Pilih kategori"}
                </p>
              </div>
            </div>
            {items.length ? (
              <ul className="divide-y divide-border/70">
                {items.map((item) => (
                  <li key={item.id} className="flex flex-wrap items-center justify-between gap-3 p-4">
                    <div className="min-w-0">
                      <p className="font-semibold text-primary">{item.title}</p>
                      <a
                        href={item.google_photos_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block max-w-lg truncate text-xs text-muted underline"
                      >
                        {item.google_photos_url}
                      </a>
                    </div>
                    <div className="flex gap-4 text-sm font-semibold">
                      <button type="button" onClick={() => beginEditItem(item)} className="text-primary">
                        Edit
                      </button>
                      <button type="button" onClick={() => void removeItem(item)} className="text-danger">
                        Hapus
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="p-5 text-sm text-muted">Belum ada item pada kategori ini.</p>
            )}
          </section>
          <form
            onSubmit={saveItem}
            className="grid gap-3 rounded-xl border border-border/70 bg-white p-5 shadow-card"
          >
            <h2 className="font-heading font-bold text-primary">
              {editItem ? "Edit item" : "Tambah tautan album"}
            </h2>
            <label className="text-sm font-semibold">
              Judul item
              <input
                required
                value={itemForm.title}
                onChange={(e) => setItemForm({ ...itemForm, title: e.target.value })}
                className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
              />
            </label>
            <label className="text-sm font-semibold">
              Google Photos URL (HTTPS)
              <input
                required
                type="url"
                value={itemForm.google_photos_url}
                onChange={(e) => setItemForm({ ...itemForm, google_photos_url: e.target.value })}
                className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
              />
            </label>
            <label className="text-sm font-semibold">
              URL pratinjau (opsional)
              <input
                type="url"
                value={itemForm.preview_image_url}
                onChange={(e) => setItemForm({ ...itemForm, preview_image_url: e.target.value })}
                className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
              />
            </label>
            <label className="text-sm font-semibold">
              Deskripsi
              <textarea
                rows={2}
                value={itemForm.description}
                onChange={(e) => setItemForm({ ...itemForm, description: e.target.value })}
                className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 font-normal"
              />
            </label>
            <label className="text-sm font-semibold">
              Urutan
              <input
                type="number"
                value={itemForm.sort_order}
                onChange={(e) => setItemForm({ ...itemForm, sort_order: e.target.value })}
                className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
              />
            </label>
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={busy || !selected}
                className="min-h-10 rounded-md bg-primary px-4 text-sm font-semibold text-white"
              >
                Simpan item
              </button>
              {editItem ? (
                <button
                  type="button"
                  onClick={() => {
                    setEditItem(null);
                    setItemForm(blankItem);
                  }}
                  className="min-h-10 rounded-md border px-4 text-sm"
                >
                  Batal
                </button>
              ) : null}
            </div>
          </form>
        </div>
      </section>
    </div>
  );
}
