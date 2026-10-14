"use client";

import { Eye, Plus, Save, Trash2 } from "lucide-react";
import * as React from "react";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { MediaUpload } from "@/components/admin/media-upload";
import { Button } from "@/components/ui/button";
import { Alert } from "@/components/ui/primitives";
import { Input, Label, Textarea } from "@/components/ui/primitives";
import { AboutSection, VisionMissionSection } from "@/features/public/home/components/about-section";
import { ContactSection } from "@/features/public/home/components/contact-sections";
import { HeroSection } from "@/features/public/home/components/hero-section";
import type { LandingPageContent } from "@/features/public/home/types";
import { ApiError } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import { getSettings, saveSettings, type LandingSettings } from "./api";

const empty: LandingSettings = {
  hero_headline: "",
  hero_subheadline: "",
  hero_primary_cta_label: "",
  hero_primary_cta_url: "",
  hero_secondary_cta_label: null,
  hero_secondary_cta_url: null,
  hero_image_url: null,
  about_title: "",
  about_description: "",
  vision: "",
  missions: [],
  contact_email: null,
  contact_phone: null,
  contact_address: null,
  location_url: null,
  instagram_url: null,
  youtube_url: null,
  tiktok_url: null,
  other_social_url: null,
};
const sections = ["Hero", "Tentang BEM", "Visi & Misi", "Kontak", "Media Sosial"] as const;
function toNullable(value: string) {
  return value.trim() ? value.trim() : null;
}
function socialLinks(settings: LandingSettings) {
  const candidates: { label: string; href: string | null }[] = [
    { label: "Instagram", href: settings.instagram_url },
    { label: "YouTube", href: settings.youtube_url },
    { label: "TikTok", href: settings.tiktok_url },
    { label: "Situs lainnya", href: settings.other_social_url },
  ];
  return candidates.flatMap(({ label, href }) => (href ? [{ label, href }] : []));
}

export function LandingSettingsAdmin() {
  const { user } = useAuth();
  const [value, setValue] = React.useState<LandingSettings>(empty);
  const [saved, setSaved] = React.useState<LandingSettings>(empty);
  const [loading, setLoading] = React.useState(true);
  const [saving, setSaving] = React.useState(false);
  const [error, setError] = React.useState("");
  const [notice, setNotice] = React.useState("");
  const [preview, setPreview] = React.useState(false);
  const [active, setActive] = React.useState<(typeof sections)[number]>("Hero");
  const dirty = JSON.stringify(value) !== JSON.stringify(saved);
  React.useEffect(() => {
    if (user?.role !== "SUPER_ADMIN" && user?.role !== "BPH") {
      setLoading(false);
      return;
    }
    void getSettings()
      .then((v) => {
        setValue(v);
        setSaved(v);
      })
      .catch(() => setError("Pengaturan landing page gagal dimuat."))
      .finally(() => setLoading(false));
  }, [user?.role]);
  React.useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);
  const update = <K extends keyof LandingSettings>(key: K, next: LandingSettings[K]) =>
    setValue((current) => ({ ...current, [key]: next }));
  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError("");
    setNotice("");
    setSaving(true);
    try {
      const result = await saveSettings(value);
      setValue(result);
      setSaved(result);
      setNotice("Perubahan Landing Page berhasil disimpan.");
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 422
          ? "Periksa kembali isian dan URL. Semua tautan eksternal harus menggunakan HTTPS."
          : "Pengaturan gagal disimpan. Periksa koneksi dan akses akun.",
      );
    } finally {
      setSaving(false);
    }
  };
  const textField = (key: keyof LandingSettings, label: string, max?: number) => (
    <div className="grid gap-1.5" key={key as string}>
      <Label htmlFor={String(key)}>{label}</Label>
      <Input
        id={String(key)}
        maxLength={max}
        value={String(value[key] ?? "")}
        onChange={(e) => update(key, e.target.value as never)}
      />
    </div>
  );
  if (loading)
    return (
      <div className="p-6 text-sm text-muted" role="status">
        Memuat pengaturan…
      </div>
    );
  if (user?.role !== "SUPER_ADMIN" && user?.role !== "BPH")
    return (
      <div className="p-6">
        <Alert variant="danger" title="Akses ditolak">
          Role akun Anda tidak dapat membuka pengaturan landing page.
        </Alert>
      </div>
    );
  const hero = {
    eyebrow: "Portal publik BEM FMIKOM UNUGHA",
    title: value.hero_headline,
    description: value.hero_subheadline,
    primaryAction: { label: value.hero_primary_cta_label, href: value.hero_primary_cta_url },
    secondaryAction:
      value.hero_secondary_cta_label && value.hero_secondary_cta_url
        ? { label: value.hero_secondary_cta_label, href: value.hero_secondary_cta_url }
        : null,
    visualLabel: "Gambar hero BEM FMIKOM UNUGHA",
    visualCaption: "Informasi resmi organisasi",
    imageUrl: value.hero_image_url,
  };
  const previewContent = {
    demoMode: false,
    hero,
    about: { title: value.about_title, description: value.about_description, highlights: [] },
    visionMission: { vision: value.vision || null, mission: value.missions },
    statistics: [],
    organization: { leadership: [], period: null, status: "empty" },
    news: [],
    announcements: [],
    gallery: [],
    registrations: [],
    contact: {
      address: value.contact_address ?? undefined,
      email: value.contact_email ?? undefined,
      phone: value.contact_phone ?? undefined,
      locationUrl: value.location_url ?? undefined,
      socialLinks: socialLinks(value),
    },
  } satisfies LandingPageContent;
  const sectionId = active.toLowerCase().replaceAll(" ", "-").replace("&", "");
  return (
    <div className="space-y-5 p-4 md:p-6 lg:p-8">
      <AdminPageHeader
        title="Pengaturan Landing Page"
        description="Kelola konten hero, profil, visi-misi, dan kontak yang tampil di beranda publik."
        actions={
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" onClick={() => setPreview(true)}>
              <Eye className="size-4" />
              Pratinjau
            </Button>
            <Button onClick={submit} disabled={saving || !dirty}>
              <Save className="size-4" />
              {saving ? "Menyimpan…" : "Simpan Perubahan"}
            </Button>
          </div>
        }
      />
      {dirty && (
        <Alert variant="warning" title="Perubahan belum disimpan">
          Perubahan draf belum diterapkan ke halaman publik.
        </Alert>
      )}
      {error && (
        <Alert variant="danger" title="Terjadi kesalahan">
          {error}
        </Alert>
      )}
      {notice && <Alert variant="success">{notice}</Alert>}
      <nav
        aria-label="Bagian pengaturan"
        className="flex flex-wrap gap-2 rounded-lg border border-border bg-surface p-2"
      >
        {sections.map((s) => (
          <button
            key={s}
            type="button"
            aria-pressed={active === s}
            onClick={() => setActive(s)}
            className={`min-h-10 rounded-md px-3 text-sm font-semibold ${active === s ? "bg-primary text-white" : "text-body hover:bg-surface-muted"}`}
          >
            {s}
          </button>
        ))}
      </nav>
      <form onSubmit={submit} className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(18rem,.72fr)]">
        <section
          className="min-w-0 rounded-xl border border-border bg-surface p-5 shadow-card"
          aria-labelledby="settings-section-title"
        >
          <h2 id="settings-section-title" className="font-heading text-xl font-bold text-primary">
            {sections.find((s) => s.toLowerCase().replaceAll(" ", "-").replace("&", "") === sectionId)}
          </h2>
          {active === "Hero" && (
            <div className="mt-5 grid gap-4">
              {textField("hero_headline", "Headline utama", 180)}
              <div className="grid gap-1.5">
                <Label htmlFor="hero_subheadline">Subheadline</Label>
                <Textarea
                  id="hero_subheadline"
                  maxLength={500}
                  value={value.hero_subheadline}
                  onChange={(e) => update("hero_subheadline", e.target.value)}
                />
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                {textField("hero_primary_cta_label", "Label CTA utama", 80)}
                {textField("hero_primary_cta_url", "URL CTA utama")}
                {textField("hero_secondary_cta_label", "Label CTA kedua", 80)}
                {textField("hero_secondary_cta_url", "URL CTA kedua")}
              </div>
              <MediaUpload
                id="landing-hero-image"
                label="Gambar hero"
                purpose="LANDING_HERO"
                value={value.hero_image_url ?? ""}
                onChange={(url) => update("hero_image_url", toNullable(url))}
                optional
              />
            </div>
          )}
          {active === "Tentang BEM" && (
            <div className="mt-5 grid gap-4">
              {textField("about_title", "Judul", 180)}
              <div className="grid gap-1.5">
                <Label htmlFor="about_description">Deskripsi</Label>
                <Textarea
                  id="about_description"
                  maxLength={5000}
                  value={value.about_description}
                  onChange={(e) => update("about_description", e.target.value)}
                />
              </div>
              <p className="text-sm text-muted">Konten statistik tidak dikelola dari bagian ini.</p>
            </div>
          )}
          {active === "Visi & Misi" && (
            <div className="mt-5 grid gap-4">
              <div className="grid gap-1.5">
                <Label htmlFor="vision">Visi</Label>
                <Textarea
                  id="vision"
                  maxLength={3000}
                  value={value.vision}
                  onChange={(e) => update("vision", e.target.value)}
                />
              </div>
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-primary">Poin Misi</h3>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => update("missions", [...value.missions, ""])}
                  disabled={value.missions.length >= 10}
                >
                  <Plus className="size-4" />
                  Tambah Misi
                </Button>
              </div>
              <ol className="grid gap-2">
                {value.missions.map((mission, index) => (
                  <li
                    key={mission}
                    className="flex items-start gap-2 rounded-lg border border-border bg-surface-muted p-2"
                  >
                    <span className="pt-3 text-sm font-bold text-primary">{index + 1}.</span>
                    <Textarea
                      aria-label={`Misi ${index + 1}`}
                      maxLength={300}
                      value={mission}
                      onChange={(e) =>
                        update(
                          "missions",
                          value.missions.map((item, i) => (i === index ? e.target.value : item)),
                        )
                      }
                      className="min-h-14 flex-1"
                    />
                    <div className="grid shrink-0">
                      <Button
                        aria-label={`Pindahkan misi ${index + 1} ke atas`}
                        type="button"
                        size="icon"
                        variant="ghost"
                        disabled={index === 0}
                        onClick={() => {
                          const a = [...value.missions];
                          [a[index - 1], a[index]] = [a[index], a[index - 1]];
                          update("missions", a);
                        }}
                      >
                        ↑
                      </Button>
                      <Button
                        aria-label={`Pindahkan misi ${index + 1} ke bawah`}
                        type="button"
                        size="icon"
                        variant="ghost"
                        disabled={index === value.missions.length - 1}
                        onClick={() => {
                          const a = [...value.missions];
                          [a[index + 1], a[index]] = [a[index], a[index + 1]];
                          update("missions", a);
                        }}
                      >
                        ↓
                      </Button>
                    </div>
                    <Button
                      aria-label={`Hapus misi ${index + 1}`}
                      type="button"
                      size="icon"
                      variant="ghost"
                      onClick={() =>
                        update(
                          "missions",
                          value.missions.filter((_, i) => i !== index),
                        )
                      }
                    >
                      <Trash2 className="size-4 text-danger" />
                    </Button>
                  </li>
                ))}
              </ol>
            </div>
          )}
          {active === "Kontak" && (
            <div className="mt-5 grid gap-4">
              {textField("contact_email", "Email kontak", 254)}
              {textField("contact_phone", "Telepon", 64)}
              <div className="grid gap-1.5">
                <Label htmlFor="contact_address">Alamat</Label>
                <Textarea
                  id="contact_address"
                  maxLength={500}
                  value={value.contact_address ?? ""}
                  onChange={(e) => update("contact_address", toNullable(e.target.value))}
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="location_url">URL lokasi (HTTPS)</Label>
                <Input
                  id="location_url"
                  type="url"
                  value={value.location_url ?? ""}
                  onChange={(e) => update("location_url", toNullable(e.target.value))}
                />
              </div>
            </div>
          )}
          {active === "Media Sosial" && (
            <div className="mt-5 grid gap-4">
              {(["instagram_url", "youtube_url", "tiktok_url", "other_social_url"] as const).map((key, i) => (
                <div key={key} className="grid gap-1.5">
                  <Label htmlFor={key}>
                    {["Instagram", "YouTube", "TikTok", "Situs lainnya"][i]} (HTTPS)
                  </Label>
                  <Input
                    id={key}
                    type="url"
                    value={value[key] ?? ""}
                    onChange={(e) => update(key, toNullable(e.target.value))}
                  />
                </div>
              ))}
            </div>
          )}
        </section>
        <aside className="hidden xl:block">
          <div className="sticky top-6 rounded-xl border border-border bg-surface p-4">
            <h2 className="font-semibold text-primary">Status sinkronisasi</h2>
            <p className="mt-2 text-sm leading-6 text-muted">
              Beranda publik membaca pengaturan terbaru tanpa cache halaman server.
            </p>
            <p className="mt-4 text-xs text-muted">Simpan untuk menerapkan perubahan.</p>
          </div>
        </aside>
      </form>
      {preview && (
        <div
          className="fixed inset-0 z-50 overflow-y-auto bg-slate-950/55 p-3 md:p-8"
          role="dialog"
          aria-modal="true"
          aria-labelledby="preview-title"
        >
          <div className="mx-auto max-w-5xl rounded-xl bg-background p-4 md:p-7">
            <div className="mb-4 flex items-center justify-between">
              <h2 id="preview-title" className="font-heading text-lg font-bold text-primary">
                Pratinjau konten landing
              </h2>
              <Button variant="outline" onClick={() => setPreview(false)}>
                Tutup
              </Button>
            </div>
            <HeroSection hero={previewContent.hero} demoMode={false} />
            <AboutSection about={previewContent.about} />
            <VisionMissionSection
              vision={previewContent.visionMission.vision}
              mission={previewContent.visionMission.mission}
              demoMode={false}
            />
            <ContactSection contact={previewContent.contact} demoMode={false} />
          </div>
        </div>
      )}
    </div>
  );
}
