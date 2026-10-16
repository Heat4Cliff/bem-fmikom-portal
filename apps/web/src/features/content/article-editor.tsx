"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import * as React from "react";
import { AdminBreadcrumb, AdminPageHeader } from "@/components/admin/admin-page-header";
import { MediaUpload } from "@/components/admin/media-upload";
import { apiRequest } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import type { Article, ContentCategory } from "./types";

type FormState = {
  title: string;
  slug: string;
  excerpt: string;
  content_html: string;
  thumbnail_url: string;
  division_id: string;
  category_ids: number[];
};
const initial: FormState = {
  title: "",
  slug: "",
  excerpt: "",
  content_html: "",
  thumbnail_url: "",
  division_id: "",
  category_ids: [],
};

export function ArticleEditor({ id }: { id?: number }) {
  const router = useRouter();
  const { user } = useAuth();
  const [form, setForm] = React.useState(initial);
  const [categories, setCategories] = React.useState<ContentCategory[]>([]);
  const [busy, setBusy] = React.useState(false);
  const [error, setError] = React.useState("");
  const [message, setMessage] = React.useState("");
  const [articleID, setArticleID] = React.useState<number | undefined>(id);
  const [status, setStatus] = React.useState<Article["status"]>("DRAFT");
  React.useEffect(() => {
    let active = true;
    void (async () => {
      try {
        const cats = await apiRequest<ContentCategory[]>("/admin/article-categories");
        if (active) setCategories(cats);
        if (id) {
          const item = await apiRequest<Article>(`/admin/articles/${id}`);
          if (active) {
            setStatus(item.status);
            setForm({
              title: item.title,
              slug: item.slug,
              excerpt: item.excerpt ?? "",
              content_html: item.content_html,
              thumbnail_url: item.thumbnail_url ?? "",
              division_id: item.division_id ? String(item.division_id) : "",
              category_ids: item.categories.map((c) => c.id),
            });
          }
        }
      } catch {
        if (active) setError("Berita gagal dimuat. Anda mungkin tidak memiliki akses.");
      }
    })();
    return () => {
      active = false;
    };
  }, [id]);
  async function save(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const body = {
        ...form,
        excerpt: form.excerpt || null,
        thumbnail_url: form.thumbnail_url || null,
        division_id: form.division_id ? Number(form.division_id) : null,
      };
      const item = await apiRequest<Article>(articleID ? `/admin/articles/${articleID}` : "/admin/articles", {
        method: articleID ? "PATCH" : "POST",
        body,
      });
      setArticleID(item.id);
      setStatus(item.status);
      setForm({ ...form, slug: item.slug });
      setMessage(item.status === "PUBLISHED" ? "Perubahan berita terbit tersimpan." : "Draft tersimpan.");
      if (!articleID) router.replace(`/admin/berita/${item.id}/edit`);
    } catch {
      setError("Berita gagal disimpan. Periksa judul, gambar, kategori, dan format konten.");
    } finally {
      setBusy(false);
    }
  }
  async function publish() {
    if (!articleID) {
      setError("Simpan draft sebelum menerbitkan.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      await apiRequest(`/admin/articles/${articleID}/publish`, { method: "POST", body: {} });
      setStatus("PUBLISHED");
      setMessage("Berita berhasil dipublikasikan.");
      router.push("/admin/berita");
    } catch {
      setError("Berita gagal diterbitkan. Simpan sebagai draft dan periksa kembali isinya.");
    } finally {
      setBusy(false);
    }
  }
  function toggleCategory(id: number) {
    setForm((current) => ({
      ...current,
      category_ids: current.category_ids.includes(id)
        ? current.category_ids.filter((value) => value !== id)
        : [...current.category_ids, id],
    }));
  }
  const previewText = form.content_html
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
  return (
    <div className="mx-auto max-w-[1440px] px-4 py-6 md:px-6 lg:px-8">
      <AdminBreadcrumb
        items={[{ label: "Berita", href: "/admin/berita" }, { label: id ? "Edit berita" : "Tulis berita" }]}
      />
      <AdminPageHeader
        title={id ? "Edit Berita" : "Tulis Berita Baru"}
        description="Simpan sebagai draft atau publikasikan setelah konten siap."
        actions={
          <Link
            href="/admin/berita"
            className="min-h-10 rounded-md border border-border px-4 py-2 text-sm font-semibold text-primary"
          >
            Kembali ke daftar
          </Link>
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
      <form onSubmit={save} className="mt-5 grid items-start gap-5 xl:grid-cols-[minmax(0,1fr)_20rem]">
        <section className="space-y-4 rounded-xl border border-border/70 bg-surface p-5 shadow-card">
          <label className="block text-sm font-semibold text-primary">
            Judul berita
            <input
              required
              maxLength={240}
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="mt-1 min-h-11 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
            />
          </label>
          <label className="block text-sm font-semibold text-primary">
            Slug <span className="font-normal text-muted">(kosongkan untuk dibuat otomatis)</span>
            <input
              maxLength={240}
              value={form.slug}
              onChange={(e) => setForm({ ...form, slug: e.target.value })}
              className="mt-1 min-h-11 w-full rounded-md border border-border bg-surface-muted px-3 font-normal"
            />
          </label>
          <label className="block text-sm font-semibold text-primary">
            Ringkasan
            <textarea
              maxLength={500}
              value={form.excerpt}
              onChange={(e) => setForm({ ...form, excerpt: e.target.value })}
              rows={3}
              className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 font-normal"
            />
          </label>
          <label className="block text-sm font-semibold text-primary">
            Isi berita{" "}
            <span className="font-normal text-muted">
              (HTML aman: p, h2, h3, strong, em, ul, ol, li, blockquote, a, br)
            </span>
            <textarea
              required
              value={form.content_html}
              onChange={(e) => setForm({ ...form, content_html: e.target.value })}
              rows={18}
              className="mt-1 w-full rounded-md border border-border bg-surface-muted p-3 font-mono text-xs font-normal leading-6"
            />
          </label>
          <div className="rounded-lg border border-border bg-surface-muted p-4">
            <p className="text-xs font-semibold uppercase text-muted">Pratinjau teks</p>
            <p className="mt-2 whitespace-pre-wrap text-sm leading-6">
              {previewText || "Isi berita akan muncul di pratinjau."}
            </p>
          </div>
        </section>
        <aside className="space-y-4">
          <section className="rounded-xl border border-border/70 bg-surface p-5 shadow-card">
            <h2 className="font-heading font-bold text-primary">Kategori</h2>
            <div className="mt-3 space-y-2">
              {categories.map((category) => (
                <label key={category.id} className="flex min-h-9 items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={form.category_ids.includes(category.id)}
                    onChange={() => toggleCategory(category.id)}
                    className="size-4 accent-primary"
                  />
                  {category.name}
                </label>
              ))}
              {categories.length === 0 ? <p className="text-xs text-muted">Belum ada kategori.</p> : null}
            </div>
          </section>
          <section className="rounded-xl border border-border/70 bg-surface p-5 shadow-card">
            <MediaUpload
              id="article-cover"
              label="Gambar sampul"
              purpose="ARTICLE_COVER"
              value={form.thumbnail_url}
              onChange={(thumbnail_url) => setForm((current) => ({ ...current, thumbnail_url }))}
              optional
              compact
            />
          </section>
          {user?.role !== "KEPALA_BIDANG" ? (
            <section className="rounded-xl border border-border/70 bg-surface p-5 shadow-card">
              <h2 className="font-heading font-bold text-primary">Divisi pemilik</h2>
              <label className="mt-3 block text-sm">
                ID divisi
                <input
                  inputMode="numeric"
                  value={form.division_id}
                  onChange={(e) => setForm({ ...form, division_id: e.target.value })}
                  placeholder="Kosong untuk berita umum"
                  className="mt-1 min-h-10 w-full rounded-md border border-border bg-surface-muted px-3 text-sm"
                />
              </label>
              <p className="mt-2 text-xs text-muted">
                API akan memvalidasi divisi. Kepala bidang selalu menggunakan divisinya sendiri.
              </p>
            </section>
          ) : null}
          <section className="flex flex-col gap-2 rounded-xl border border-border/70 bg-surface p-4 shadow-card">
            <p className="text-xs font-semibold text-muted">
              Status: {status === "PUBLISHED" ? "Terbit" : "Draft"}
            </p>
            <button
              type="submit"
              disabled={busy}
              className="min-h-11 rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground"
            >
              {busy ? "Menyimpan..." : status === "PUBLISHED" ? "Simpan perubahan" : "Simpan draft"}
            </button>
            {articleID && status === "DRAFT" ? (
              <button
                type="button"
                disabled={busy}
                onClick={() => void publish()}
                className="min-h-11 rounded-md bg-secondary px-4 text-sm font-semibold text-secondary-foreground"
              >
                Terbitkan
              </button>
            ) : null}
          </section>
        </aside>
      </form>
    </div>
  );
}
