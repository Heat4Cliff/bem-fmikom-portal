"use client";

import * as React from "react";
import { ArrowUpRight, CalendarDays, Ellipsis, ImagePlus, Layers3, Plus, Rows3 } from "lucide-react";
import { AdminBreadcrumb, AdminPageHeader } from "@/components/admin/admin-page-header";
import { MediaImage } from "@/components/ui/media-image";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DropdownMenu,
} from "@/components/ui/overlays";
import { ActionGroup } from "@/components/ui/primitives";
import { MediaUpload } from "@/components/admin/media-upload";
import { getDivisions, getPeriods } from "@/features/organization/api";
import type { Division, Period } from "@/features/organization/types";
import { apiRequest, apiRequestEnvelope } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import type { GalleryCategory, GalleryItem } from "./types";
import { galleryItemPayload } from "./content-payloads";
import { contentFormErrorMessage } from "./form-errors";

const blankCategory = {
  title: "",
  slug: "",
  description: "",
  period_id: "",
  division_id: "",
  event_date: "",
};
const blankItem = {
  category_id: "",
  title: "",
  description: "",
  google_photos_url: "",
  preview_image_url: "",
  sort_order: "0",
};
const fieldClass =
  "mt-1 min-h-11 w-full rounded-md border border-border bg-surface-muted px-3 text-sm font-normal text-body focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary";

function albumHost(value: string) {
  try {
    return new URL(value).hostname.replace(/^www\./, "");
  } catch {
    return "Google Photos";
  }
}

export function GalleryManager() {
  const { user } = useAuth();
  const [categories, setCategories] = React.useState<GalleryCategory[]>([]);
  const [items, setItems] = React.useState<GalleryItem[]>([]);
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [selected, setSelected] = React.useState(0);
  const [catForm, setCatForm] = React.useState(blankCategory);
  const [itemForm, setItemForm] = React.useState(blankItem);
  const [editCategory, setEditCategory] = React.useState<number | null>(null);
  const [editItem, setEditItem] = React.useState<number | null>(null);
  const [categoryDialogOpen, setCategoryDialogOpen] = React.useState(false);
  const [itemDialogOpen, setItemDialogOpen] = React.useState(false);
  const [slugAdvanced, setSlugAdvanced] = React.useState(false);
  const [categoryToDelete, setCategoryToDelete] = React.useState<GalleryCategory | null>(null);
  const [itemToDelete, setItemToDelete] = React.useState<GalleryItem | null>(null);
  const [busy, setBusy] = React.useState(false);
  const [uploadPending, setUploadPending] = React.useState(false);
  const [error, setError] = React.useState("");
  const [message, setMessage] = React.useState("");
  const canManageCategories = user?.role !== "KEPALA_BIDANG";
  const activeCategory = categories.find((category) => category.id === selected) ?? null;

  const loadCategories = React.useCallback(async () => {
    try {
      const result = await apiRequestEnvelope<GalleryCategory[]>(
        "/admin/gallery-categories?page=1&limit=100",
      );
      setCategories(result.data);
      setSelected((current) =>
        result.data.some((category) => category.id === current) ? current : (result.data[0]?.id ?? 0),
      );
    } catch {
      setError("Kategori galeri gagal dimuat.");
    }
  }, []);

  const loadItems = React.useCallback(async (categoryID: number) => {
    if (!categoryID) {
      setItems([]);
      return;
    }
    try {
      const result = await apiRequestEnvelope<GalleryItem[]>(
        `/admin/gallery-items?category_id=${categoryID}&page=1&limit=100`,
      );
      setItems(result.data);
    } catch {
      setError("Dokumentasi galeri gagal dimuat.");
    }
  }, []);

  React.useEffect(() => {
    void loadCategories();
    if (canManageCategories) {
      void Promise.all([getPeriods({ page: 1, limit: 100 }), getDivisions({ page: 1, limit: 100 })])
        .then(([periodResult, divisionResult]) => {
          setPeriods(periodResult.items);
          setDivisions(divisionResult.items);
        })
        .catch(() => setError("Pilihan periode atau divisi gagal dimuat."));
    }
  }, [canManageCategories, loadCategories]);

  React.useEffect(() => {
    void loadItems(selected);
  }, [loadItems, selected]);

  function openNewCategory() {
    setEditCategory(null);
    setSlugAdvanced(false);
    setCatForm(blankCategory);
    setError("");
    setCategoryDialogOpen(true);
  }

  function beginEditCategory(category: GalleryCategory) {
    setEditCategory(category.id);
    setSlugAdvanced(true);
    setCatForm({
      title: category.title,
      slug: category.slug,
      description: category.description ?? "",
      period_id: category.period_id ? String(category.period_id) : "",
      division_id: category.division_id ? String(category.division_id) : "",
      event_date: category.event_date ?? "",
    });
    setError("");
    setCategoryDialogOpen(true);
  }

  function openNewItem() {
    setEditItem(null);
    setItemForm({ ...blankItem, category_id: selected ? String(selected) : "" });
    setError("");
    setItemDialogOpen(true);
  }

  function beginEditItem(item: GalleryItem) {
    setEditItem(item.id);
    setItemForm({
      category_id: String(item.category_id),
      title: item.title,
      description: item.description ?? "",
      google_photos_url: item.google_photos_url,
      preview_image_url: item.preview_image_url ?? "",
      sort_order: String(item.sort_order),
    });
    setError("");
    setItemDialogOpen(true);
  }

  async function saveCategory(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const payload = {
        ...catForm,
        slug: catForm.slug.trim(),
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
      setCategoryDialogOpen(false);
      setCatForm(blankCategory);
      setEditCategory(null);
      setMessage(editCategory ? "Kategori diperbarui." : "Kategori kegiatan dibuat.");
      await loadCategories();
    } catch (saveError: unknown) {
      setError(contentFormErrorMessage(saveError, "Kategori gagal disimpan. Periksa data lalu coba lagi."));
    } finally {
      setBusy(false);
    }
  }

  async function confirmRemoveCategory() {
    if (!categoryToDelete) return;
    try {
      await apiRequest(`/admin/gallery-categories/${categoryToDelete.id}`, { method: "DELETE" });
      setMessage("Kategori dihapus.");
      setCategoryToDelete(null);
      await loadCategories();
    } catch {
      setError("Kategori yang masih berisi dokumentasi tidak dapat dihapus.");
      setCategoryToDelete(null);
    }
  }

  async function saveItem(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy || uploadPending) return;
    const categoryID = Number(itemForm.category_id);
    if (!categoryID) {
      setError("Pilih kategori dokumentasi terlebih dahulu.");
      return;
    }
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await apiRequest(editItem ? `/admin/gallery-items/${editItem}` : "/admin/gallery-items", {
        method: editItem ? "PATCH" : "POST",
        body: galleryItemPayload(itemForm, categoryID),
      });
      setSelected(categoryID);
      setItemDialogOpen(false);
      setItemForm(blankItem);
      setEditItem(null);
      setMessage(editItem ? "Dokumentasi diperbarui." : "Dokumentasi ditambahkan.");
      await Promise.all([loadCategories(), loadItems(categoryID)]);
    } catch (saveError: unknown) {
      setError(
        contentFormErrorMessage(saveError, "Dokumentasi gagal disimpan. Periksa data lalu coba lagi."),
      );
    } finally {
      setBusy(false);
    }
  }

  async function confirmRemoveItem() {
    if (!itemToDelete) return;
    try {
      await apiRequest(`/admin/gallery-items/${itemToDelete.id}`, { method: "DELETE" });
      setMessage("Dokumentasi dihapus.");
      await Promise.all([loadCategories(), loadItems(itemToDelete.category_id)]);
      setItemToDelete(null);
    } catch {
      setError("Dokumentasi gagal dihapus.");
      setItemToDelete(null);
    }
  }

  const selectedDivisionOptions = catForm.period_id
    ? divisions.filter((division) => division.period_id === Number(catForm.period_id))
    : divisions;

  return (
    <div className="mx-auto max-w-[1440px] px-4 py-6 md:px-6 lg:px-8">
      <AdminBreadcrumb items={[{ label: "Dashboard", href: "/admin" }, { label: "Galeri" }]} />
      <AdminPageHeader
        title="Manajemen Galeri Kegiatan"
        description="Kelola dokumentasi kegiatan dan album eksternal."
        actions={
          <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
            {canManageCategories ? (
              <Button variant="outline" onClick={openNewCategory}>
                <Plus aria-hidden="true" className="size-4" />
                Tambah Kategori
              </Button>
            ) : null}
            <Button onClick={openNewItem} disabled={!activeCategory}>
              <Plus aria-hidden="true" className="size-4" />
              Tambah Item
            </Button>
          </div>
        }
      />

      {error ? (
        <p
          role="alert"
          className="mt-4 rounded-md border border-danger/20 bg-danger/5 p-3 text-sm text-danger"
        >
          {error}
        </p>
      ) : null}
      {message ? (
        <p
          role="status"
          className="mt-4 rounded-md border border-success/20 bg-success/5 p-3 text-sm text-success"
        >
          {message}
        </p>
      ) : null}

      <section aria-label="Ringkasan galeri" className="mt-5 grid grid-cols-2 gap-3 sm:max-w-xl">
        <div className="flex items-center gap-3 rounded-xl border border-border/70 bg-white p-4 shadow-card">
          <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/5 text-primary">
            <Layers3 aria-hidden="true" className="size-5" />
          </span>
          <div>
            <p className="text-xs text-muted">Total kategori</p>
            <p className="font-heading text-xl font-bold text-primary">{categories.length}</p>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border border-border/70 bg-white p-4 shadow-card">
          <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-secondary/15 text-primary">
            <Rows3 aria-hidden="true" className="size-5" />
          </span>
          <div>
            <p className="text-xs text-muted">Total dokumentasi</p>
            <p className="font-heading text-xl font-bold text-primary">
              {categories.reduce((total, category) => total + (category.item_count ?? 0), 0)}
            </p>
          </div>
        </div>
      </section>

      <section className="mt-6 rounded-xl border border-border/70 bg-white shadow-card">
        <div className="border-b border-border/70 p-4 sm:p-5">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-muted">Kategori kegiatan</p>
              <h2 className="mt-1 font-heading text-lg font-bold text-primary">
                Pilih dokumentasi yang ingin dikelola
              </h2>
            </div>
            {activeCategory ? (
              <div className="rounded-lg bg-primary/5 px-3 py-2 text-right">
                <p className="text-xs text-muted">Kategori aktif</p>
                <p className="max-w-56 truncate text-sm font-semibold text-primary">{activeCategory.title}</p>
              </div>
            ) : null}
          </div>
          {categories.length ? (
            <nav aria-label="Pilih kategori" className="mt-4 -mx-1 flex gap-2 overflow-x-auto px-1 pb-1">
              {categories.map((category) => (
                <div
                  key={category.id}
                  className={`flex shrink-0 items-center gap-1 rounded-lg border ${
                    selected === category.id
                      ? "border-primary bg-primary text-white"
                      : "border-border bg-surface text-body"
                  }`}
                >
                  <button
                    type="button"
                    aria-pressed={selected === category.id}
                    onClick={() => setSelected(category.id)}
                    className="min-h-10 max-w-56 rounded-lg px-3 text-left text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
                  >
                    <span className="block truncate font-semibold">{category.title}</span>
                    <span
                      className={`block text-xs ${selected === category.id ? "text-white/75" : "text-muted"}`}
                    >
                      {category.item_count ?? 0} dokumentasi
                    </span>
                  </button>
                  {canManageCategories ? (
                    <DropdownMenu
                      trigger={
                        <button
                          type="button"
                          aria-label={`Aksi kategori ${category.title}`}
                          className="mr-1 grid size-9 shrink-0 place-items-center rounded-md focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
                        >
                          <Ellipsis aria-hidden="true" className="size-4" />
                        </button>
                      }
                      items={[
                        { label: "Edit kategori", onSelect: () => beginEditCategory(category) },
                        {
                          label: "Hapus kategori",
                          destructive: true,
                          onSelect: () => setCategoryToDelete(category),
                        },
                      ]}
                    />
                  ) : null}
                </div>
              ))}
            </nav>
          ) : (
            <div className="mt-4 rounded-lg border border-dashed border-border bg-surface-muted/50 p-5 text-center">
              <p className="text-sm text-muted">Belum ada kategori galeri.</p>
              {canManageCategories ? (
                <Button variant="outline" className="mt-3" onClick={openNewCategory}>
                  <Plus aria-hidden="true" className="size-4" /> Tambah kategori pertama
                </Button>
              ) : null}
            </div>
          )}
        </div>

        <div className="p-4 sm:p-5">
          {activeCategory ? (
            <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
              <div className="min-w-0">
                <p className="text-xs text-muted">Kategori aktif</p>
                <h3 className="mt-1 truncate font-heading text-xl font-bold text-primary">
                  {activeCategory.title}
                </h3>
                <p className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted">
                  <span>{items.length} dokumentasi</span>
                  {activeCategory.event_date ? (
                    <span className="inline-flex items-center gap-1">
                      <CalendarDays aria-hidden="true" className="size-3.5" /> {activeCategory.event_date}
                    </span>
                  ) : null}
                  {activeCategory.division_name ? <span>{activeCategory.division_name}</span> : null}
                </p>
              </div>
              <Button onClick={openNewItem}>
                <Plus aria-hidden="true" className="size-4" /> Tambah Dokumentasi
              </Button>
            </div>
          ) : null}

          {!activeCategory ? (
            <div className="grid justify-items-center gap-2 rounded-xl border border-dashed border-border bg-surface-muted/40 px-5 py-12 text-center">
              <ImagePlus aria-hidden="true" className="size-8 text-primary/70" />
              <h3 className="font-semibold text-primary">Pilih atau buat kategori</h3>
              <p className="max-w-md text-sm text-muted">
                Dokumentasi akan ditampilkan setelah kategori kegiatan tersedia.
              </p>
            </div>
          ) : items.length ? (
            <ul aria-label={`Dokumentasi ${activeCategory.title}`} className="grid gap-3 md:grid-cols-2">
              {items.map((item) => (
                <li key={item.id}>
                  <article className="grid min-w-0 grid-cols-[5rem_minmax(0,1fr)_2.5rem] items-center gap-3 rounded-xl border border-border/70 bg-surface p-3 sm:grid-cols-[6rem_minmax(0,1fr)_2.5rem] sm:gap-4 sm:p-4">
                    <MediaImage
                      value={item.preview_image_url ?? ""}
                      alt={`Pratinjau ${item.title}`}
                      width={160}
                      height={112}
                      className="h-20 w-20 rounded-lg bg-surface-muted object-contain sm:h-24 sm:w-24"
                      fallbackClassName="grid h-20 w-20 place-items-center rounded-lg bg-surface-muted text-primary/70 sm:h-24 sm:w-24"
                      fallback={<ImagePlus aria-hidden="true" className="size-6" />}
                    />
                    <div className="min-w-0">
                      <h4 className="truncate font-semibold text-primary" title={item.title}>
                        {item.title}
                      </h4>
                      <p className="mt-1 line-clamp-2 text-sm text-muted">
                        {item.description || "Dokumentasi kegiatan"}
                      </p>
                      <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-muted">
                        <a
                          href={item.google_photos_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          aria-label={`Buka album ${item.title} di ${albumHost(item.google_photos_url)}`}
                          className="inline-flex min-h-8 items-center gap-1.5 rounded-md bg-white px-2 font-semibold text-primary underline decoration-primary/30 underline-offset-2 focus-visible:outline-2 focus-visible:outline-secondary"
                        >
                          Google Photos <ArrowUpRight aria-hidden="true" className="size-3.5" />
                        </a>
                        <span>Urutan {item.sort_order}</span>
                      </div>
                    </div>
                    <DropdownMenu
                      trigger={
                        <button
                          type="button"
                          aria-label={`Aksi dokumentasi ${item.title}`}
                          className="grid size-10 place-items-center rounded-md text-muted hover:bg-white hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
                        >
                          <Ellipsis aria-hidden="true" className="size-5" />
                        </button>
                      }
                      items={[
                        { label: "Edit", onSelect: () => beginEditItem(item) },
                        {
                          label: "Buka album",
                          onSelect: () =>
                            window.open(item.google_photos_url, "_blank", "noopener,noreferrer"),
                        },
                        {
                          label: "Hapus",
                          destructive: true,
                          onSelect: () => setItemToDelete(item),
                        },
                      ]}
                    />
                  </article>
                </li>
              ))}
            </ul>
          ) : (
            <div className="grid justify-items-center gap-2 rounded-xl border border-dashed border-border bg-surface-muted/40 px-5 py-12 text-center">
              <ImagePlus aria-hidden="true" className="size-8 text-primary/70" />
              <h3 className="font-semibold text-primary">Belum ada dokumentasi pada kategori ini.</h3>
              <p className="max-w-md text-sm text-muted">
                Tambahkan item untuk menghubungkan foto pratinjau dan album kegiatan.
              </p>
              <Button className="mt-2" onClick={openNewItem}>
                <Plus aria-hidden="true" className="size-4" /> Tambah Dokumentasi
              </Button>
            </div>
          )}
        </div>
      </section>

      <Dialog open={categoryDialogOpen} onOpenChange={setCategoryDialogOpen}>
        <DialogContent className="w-[min(96vw,44rem)] p-0 sm:w-[min(92vw,44rem)]">
          <div className="border-b border-border bg-primary/5 px-5 py-4 pr-14 sm:px-6">
            <DialogTitle className="font-heading text-lg font-bold text-primary">
              {editCategory ? "Edit kategori kegiatan" : "Tambah kategori kegiatan"}
            </DialogTitle>
            <DialogDescription className="mt-1 text-sm text-muted">
              Atur nama, periode, dan divisi yang terkait dengan dokumentasi.
            </DialogDescription>
          </div>
          <form
            onSubmit={saveCategory}
            className="grid max-h-[calc(94dvh-5rem)] gap-4 overflow-y-auto p-5 sm:p-6"
          >
            <label className="text-sm font-semibold text-primary">
              Judul kategori
              <input
                required
                maxLength={120}
                value={catForm.title}
                onChange={(event) => setCatForm((current) => ({ ...current, title: event.target.value }))}
                className={fieldClass}
                autoFocus
              />
            </label>
            <details open={slugAdvanced} onToggle={(event) => setSlugAdvanced(event.currentTarget.open)}>
              <summary className="cursor-pointer text-sm font-semibold text-primary focus-visible:outline-2 focus-visible:outline-secondary">
                Pengaturan URL (opsional)
              </summary>
              <label className="mt-3 block text-sm font-semibold text-primary">
                Slug
                <input
                  value={catForm.slug}
                  onChange={(event) => setCatForm((current) => ({ ...current, slug: event.target.value }))}
                  placeholder="Dibuat otomatis dari judul"
                  className={fieldClass}
                />
              </label>
            </details>
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="text-sm font-semibold text-primary">
                Tanggal kegiatan
                <input
                  type="date"
                  value={catForm.event_date}
                  onChange={(event) =>
                    setCatForm((current) => ({ ...current, event_date: event.target.value }))
                  }
                  className={fieldClass}
                />
              </label>
              <label className="text-sm font-semibold text-primary">
                Periode organisasi
                <select
                  value={catForm.period_id}
                  onChange={(event) =>
                    setCatForm((current) => ({ ...current, period_id: event.target.value, division_id: "" }))
                  }
                  className={fieldClass}
                >
                  <option value="">Pilih periode (opsional)</option>
                  {periods.map((period) => (
                    <option key={period.id} value={period.id}>
                      {period.name}
                      {period.status === "ACTIVE" ? " · Aktif" : ""}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <label className="text-sm font-semibold text-primary">
              Divisi penyelenggara
              <select
                value={catForm.division_id}
                onChange={(event) =>
                  setCatForm((current) => ({ ...current, division_id: event.target.value }))
                }
                className={fieldClass}
              >
                <option value="">Pilih divisi (opsional)</option>
                {selectedDivisionOptions.map((division) => (
                  <option key={division.id} value={division.id}>
                    {division.name}
                    {division.period_name ? ` · ${division.period_name}` : ""}
                  </option>
                ))}
              </select>
            </label>
            <label className="text-sm font-semibold text-primary">
              Deskripsi
              <textarea
                rows={3}
                value={catForm.description}
                onChange={(event) =>
                  setCatForm((current) => ({ ...current, description: event.target.value }))
                }
                className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 text-sm font-normal text-body focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
              />
            </label>
            {error ? (
              <p role="alert" className="text-sm text-danger">
                {error}
              </p>
            ) : null}
            <ActionGroup className="sticky bottom-0 -mx-5 -mb-5 mt-1 w-auto border-t border-border bg-surface px-5 py-4 sm:-mx-6 sm:-mb-6 sm:px-6">
              <Button
                type="button"
                variant="outline"
                disabled={busy}
                onClick={() => setCategoryDialogOpen(false)}
              >
                Batal
              </Button>
              <Button type="submit" loading={busy}>
                Simpan Kategori
              </Button>
            </ActionGroup>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={itemDialogOpen} onOpenChange={setItemDialogOpen}>
        <DialogContent className="w-[min(96vw,44rem)] p-0 sm:w-[min(92vw,44rem)]">
          <div className="border-b border-border bg-primary/5 px-5 py-4 pr-14 sm:px-6">
            <DialogTitle className="font-heading text-lg font-bold text-primary">
              {editItem ? "Edit dokumentasi kegiatan" : "Tambah dokumentasi kegiatan"}
            </DialogTitle>
            <DialogDescription className="mt-1 text-sm text-muted">
              Hubungkan satu item dokumentasi dengan album dan gambar pratinjau.
            </DialogDescription>
          </div>
          <form
            onSubmit={saveItem}
            className="grid max-h-[calc(94dvh-5rem)] gap-4 overflow-y-auto p-5 sm:p-6"
          >
            <label className="text-sm font-semibold text-primary">
              Kategori
              <select
                required
                value={itemForm.category_id}
                onChange={(event) =>
                  setItemForm((current) => ({ ...current, category_id: event.target.value }))
                }
                className={fieldClass}
              >
                <option value="">Pilih kategori</option>
                {categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.title}
                  </option>
                ))}
              </select>
            </label>
            <label className="text-sm font-semibold text-primary">
              Judul item
              <input
                required
                maxLength={160}
                value={itemForm.title}
                onChange={(event) => setItemForm((current) => ({ ...current, title: event.target.value }))}
                className={fieldClass}
                autoFocus
              />
            </label>
            <label className="text-sm font-semibold text-primary">
              Tautan album Google Photos
              <input
                required
                type="url"
                inputMode="url"
                placeholder="https://photos.app.goo.gl/..."
                value={itemForm.google_photos_url}
                onChange={(event) =>
                  setItemForm((current) => ({ ...current, google_photos_url: event.target.value }))
                }
                className={fieldClass}
              />
              <span className="mt-1 block text-xs font-normal text-muted">
                Tautan album/dokumentasi eksternal.
              </span>
            </label>
            <MediaUpload
              id="gallery-preview"
              label="Gambar pratinjau"
              purpose="GALLERY_PREVIEW"
              value={itemForm.preview_image_url}
              onChange={(preview_image_url) => setItemForm((current) => ({ ...current, preview_image_url }))}
              onUploadStateChange={setUploadPending}
              disabled={busy}
              optional
              compact
            />
            <label className="text-sm font-semibold text-primary">
              Deskripsi
              <textarea
                rows={3}
                value={itemForm.description}
                onChange={(event) =>
                  setItemForm((current) => ({ ...current, description: event.target.value }))
                }
                className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 text-sm font-normal text-body focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
              />
            </label>
            <label className="text-sm font-semibold text-primary sm:max-w-48">
              Urutan tampil
              <input
                type="number"
                min={0}
                value={itemForm.sort_order}
                onChange={(event) =>
                  setItemForm((current) => ({ ...current, sort_order: event.target.value }))
                }
                className={fieldClass}
              />
            </label>
            {error ? (
              <p role="alert" className="text-sm text-danger">
                {error}
              </p>
            ) : null}
            <ActionGroup className="sticky bottom-0 -mx-5 -mb-5 mt-1 w-auto border-t border-border bg-surface px-5 py-4 sm:-mx-6 sm:-mb-6 sm:px-6">
              <Button
                type="button"
                variant="outline"
                disabled={busy}
                onClick={() => setItemDialogOpen(false)}
              >
                Batal
              </Button>
              <Button
                type="submit"
                loading={busy || uploadPending}
                disabled={uploadPending || !itemForm.category_id}
              >
                Simpan
              </Button>
            </ActionGroup>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(categoryToDelete)} onOpenChange={(open) => !open && setCategoryToDelete(null)}>
        <DialogContent className="w-[min(92vw,28rem)]">
          <DialogTitle className="font-heading text-lg font-bold text-primary">Hapus kategori?</DialogTitle>
          <DialogDescription className="mt-2 text-sm leading-6 text-muted">
            Kategori “{categoryToDelete?.title}” akan dihapus. Kategori yang masih memiliki item tidak dapat
            dihapus.
          </DialogDescription>
          <ActionGroup className="mt-6 justify-end">
            <Button variant="outline" onClick={() => setCategoryToDelete(null)}>
              Batal
            </Button>
            <Button variant="danger" onClick={() => void confirmRemoveCategory()}>
              Hapus Kategori
            </Button>
          </ActionGroup>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(itemToDelete)} onOpenChange={(open) => !open && setItemToDelete(null)}>
        <DialogContent className="w-[min(92vw,28rem)]">
          <DialogTitle className="font-heading text-lg font-bold text-primary">
            Hapus dokumentasi?
          </DialogTitle>
          <DialogDescription className="mt-2 text-sm leading-6 text-muted">
            “{itemToDelete?.title}” akan dihapus dari daftar kategori ini.
          </DialogDescription>
          <ActionGroup className="mt-6 justify-end">
            <Button variant="outline" onClick={() => setItemToDelete(null)}>
              Batal
            </Button>
            <Button variant="danger" onClick={() => void confirmRemoveItem()}>
              Hapus Dokumentasi
            </Button>
          </ActionGroup>
        </DialogContent>
      </Dialog>
    </div>
  );
}
