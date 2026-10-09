"use client";

import Link from "next/link";
import * as React from "react";
import { AdminBreadcrumb, AdminPageHeader } from "@/components/admin/admin-page-header";
import { apiRequest, apiRequestEnvelope } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import type { Article, ContentCategory, PageMeta } from "./types";

const emptyMeta: PageMeta = { page: 1, limit: 10, total: 0, total_pages: 0 };

export function ArticleManager() {
  const { user } = useAuth();
  const [items, setItems] = React.useState<Article[]>([]);
  const [categories, setCategories] = React.useState<ContentCategory[]>([]);
  const [meta, setMeta] = React.useState(emptyMeta);
  const [page, setPage] = React.useState(1);
  const [search, setSearch] = React.useState("");
  const [searchInput, setSearchInput] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [category, setCategory] = React.useState("");
  const [busy, setBusy] = React.useState(false);
  const [error, setError] = React.useState("");

  const load = React.useCallback(async () => {
    setBusy(true);
    setError("");
    try {
      const query = new URLSearchParams({ page: String(page), limit: "10" });
      if (search) query.set("search", search);
      if (status) query.set("status", status);
      if (category) query.set("category", category);
      const [result, categoryRows] = await Promise.all([
        apiRequestEnvelope<Article[]>(`/admin/articles?${query.toString()}`),
        apiRequest<ContentCategory[]>("/admin/article-categories"),
      ]);
      setItems(result.data);
      setMeta(result.meta ?? emptyMeta);
      setCategories(categoryRows);
    } catch {
      setError("Daftar berita gagal dimuat. Periksa koneksi API lalu coba lagi.");
    } finally {
      setBusy(false);
    }
  }, [category, page, search, status]);
  React.useEffect(() => {
    void load();
  }, [load]);

  async function changeStatus(article: Article, action: "publish" | "unpublish") {
    setError("");
    try {
      await apiRequest(`/admin/articles/${article.id}/${action}`, { method: "POST", body: {} });
      await load();
    } catch {
      setError(`Status berita “${article.title}” gagal diperbarui.`);
    }
  }
  async function remove(article: Article) {
    if (!window.confirm(`Hapus berita “${article.title}”?`)) return;
    setError("");
    try {
      await apiRequest(`/admin/articles/${article.id}`, { method: "DELETE" });
      await load();
    } catch {
      setError("Berita gagal dihapus.");
    }
  }

  return (
    <div className="mx-auto max-w-[1440px] px-4 py-6 md:px-6 lg:px-8">
      <AdminBreadcrumb items={[{ label: "Dashboard", href: "/admin" }, { label: "Berita" }]} />
      <AdminPageHeader
        title="Manajemen Berita & Artikel"
        description="Kelola artikel publikasi kegiatan dan informasi BEM FMIKOM UNUGHA."
        actions={
          <Link
            href="/admin/berita/new"
            className="inline-flex min-h-11 items-center rounded-md bg-primary px-5 text-sm font-semibold text-white"
          >
            + Tulis berita baru
          </Link>
        }
      />
      <section
        className="mt-6 rounded-xl border border-border/70 bg-white p-4 shadow-card"
        aria-label="Filter berita"
      >
        <form
          onSubmit={(event) => {
            event.preventDefault();
            setSearch(searchInput.trim());
            setPage(1);
          }}
          className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"
        >
          <label className="sm:col-span-2 xl:col-span-2">
            <span className="sr-only">Cari berita</span>
            <input
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
              placeholder="Cari judul berita..."
              className="min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 text-sm"
            />
          </label>
          <label className="sr-only" htmlFor="article-category-filter">
            Kategori
          </label>
          <select
            id="article-category-filter"
            value={category}
            onChange={(event) => {
              setCategory(event.target.value);
              setPage(1);
            }}
            className="min-h-10 rounded-md border border-border bg-surface-muted px-3 text-sm"
          >
            <option value="">Semua kategori</option>
            {categories.map((item) => (
              <option key={item.id} value={item.slug}>
                {item.name}
              </option>
            ))}
          </select>
          {user?.role !== "KEPALA_BIDANG" ? (
            <select
              aria-label="Status berita"
              value={status}
              onChange={(event) => {
                setStatus(event.target.value);
                setPage(1);
              }}
              className="min-h-10 rounded-md border border-border bg-surface-muted px-3 text-sm"
            >
              <option value="">Semua status</option>
              <option value="PUBLISHED">Terbit</option>
              <option value="DRAFT">Draft</option>
            </select>
          ) : null}
        </form>
      </section>
      {error ? (
        <p
          role="alert"
          className="mt-4 rounded-md border border-danger/30 bg-danger/5 p-3 text-sm text-danger"
        >
          {error}
        </p>
      ) : null}
      <section
        className="mt-5 overflow-hidden rounded-xl border border-border/70 bg-white shadow-card"
        aria-label="Daftar berita"
        aria-busy={busy}
      >
        {busy ? (
          <p className="p-8 text-sm text-muted" role="status">
            Memuat berita...
          </p>
        ) : items.length === 0 ? (
          <p className="p-8 text-sm text-muted">Belum ada berita yang sesuai.</p>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full min-w-[760px] text-left text-sm">
                <thead className="bg-surface-muted text-xs uppercase tracking-wide text-muted">
                  <tr>
                    <th className="p-4">Judul berita & slug</th>
                    <th className="p-4">Kategori</th>
                    <th className="p-4">Divisi</th>
                    <th className="p-4">Status</th>
                    <th className="p-4">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/70">
                  {items.map((item) => (
                    <tr key={item.id}>
                      <td className="p-4">
                        <span className="block max-w-sm truncate font-semibold text-primary">
                          {item.title}
                        </span>
                        <span className="block text-xs text-muted">/{item.slug}</span>
                      </td>
                      <td className="p-4">{item.categories.map((c) => c.name).join(", ") || "—"}</td>
                      <td className="p-4">{item.division_name ?? "BPH"}</td>
                      <td className="p-4">
                        <span
                          className={`rounded-full px-2.5 py-1 text-xs font-semibold ${item.status === "PUBLISHED" ? "bg-success/10 text-success" : "bg-surface-muted text-muted"}`}
                        >
                          {item.status === "PUBLISHED" ? "TERBIT" : "DRAFT"}
                        </span>
                      </td>
                      <td className="p-4">
                        <div className="flex flex-wrap gap-2">
                          <Link href={`/admin/berita/${item.id}/edit`} className="font-semibold text-primary">
                            Edit
                          </Link>
                          {item.status === "DRAFT" ? (
                            <button
                              type="button"
                              onClick={() => void changeStatus(item, "publish")}
                              className="font-semibold text-primary"
                            >
                              Terbitkan
                            </button>
                          ) : (
                            <button
                              type="button"
                              onClick={() => void changeStatus(item, "unpublish")}
                              className="font-semibold text-muted"
                            >
                              Jadikan draft
                            </button>
                          )}
                          <button
                            type="button"
                            onClick={() => void remove(item)}
                            className="font-semibold text-danger"
                          >
                            Hapus
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <ul className="divide-y divide-border/70 md:hidden">
              {items.map((item) => (
                <li key={item.id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="font-heading font-bold text-primary">{item.title}</p>
                      <p className="mt-1 truncate text-xs text-muted">
                        {item.categories.map((c) => c.name).join(", ") || "Tanpa kategori"} ·{" "}
                        {item.division_name ?? "BPH"}
                      </p>
                    </div>
                    <span className="shrink-0 rounded-full bg-surface-muted px-2 py-1 text-[10px] font-semibold">
                      {item.status}
                    </span>
                  </div>
                  <div className="mt-3 flex flex-wrap gap-4 text-sm font-semibold">
                    <Link href={`/admin/berita/${item.id}/edit`} className="text-primary">
                      Edit
                    </Link>
                    {item.status === "DRAFT" ? (
                      <button
                        type="button"
                        onClick={() => void changeStatus(item, "publish")}
                        className="text-primary"
                      >
                        Terbitkan
                      </button>
                    ) : (
                      <button
                        type="button"
                        onClick={() => void changeStatus(item, "unpublish")}
                        className="text-muted"
                      >
                        Jadikan draft
                      </button>
                    )}
                    <button type="button" onClick={() => void remove(item)} className="text-danger">
                      Hapus
                    </button>
                  </div>
                </li>
              ))}
            </ul>
            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-border/70 p-4 text-sm">
              <span className="text-muted">
                {meta.total} berita · halaman {meta.page} dari {Math.max(meta.total_pages, 1)}
              </span>
              <div className="flex gap-2">
                <button
                  type="button"
                  disabled={page <= 1}
                  onClick={() => setPage((v) => v - 1)}
                  className="min-h-9 rounded-md border border-border px-3 disabled:opacity-40"
                >
                  Sebelumnya
                </button>
                <button
                  type="button"
                  disabled={page >= meta.total_pages}
                  onClick={() => setPage((v) => v + 1)}
                  className="min-h-9 rounded-md border border-border px-3 disabled:opacity-40"
                >
                  Berikutnya
                </button>
              </div>
            </div>
          </>
        )}
      </section>
    </div>
  );
}
