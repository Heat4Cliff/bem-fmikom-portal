import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { LandingSettingsAdmin } from "./settings-admin";
import { ActivityAdmin } from "./activity-admin";
import { DashboardAdmin } from "./dashboard-admin";

const { authState, getSettings, saveSettings, getActivities, getDashboard } = vi.hoisted(() => ({
  authState: { user: null as { id: number; role: string } | null },
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
  getActivities: vi.fn(),
  getDashboard: vi.fn(),
}));

vi.mock("@/lib/auth/auth-provider", () => ({ useAuth: () => ({ user: authState.user }) }));
vi.mock("./api", () => ({
  getSettings,
  saveSettings,
  getActivities,
  getDashboard,
  getActivity: vi.fn(),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

const settings = {
  hero_headline: "Portal BEM",
  hero_subheadline: "Informasi organisasi",
  hero_primary_cta_label: "Lihat kegiatan",
  hero_primary_cta_url: "#kegiatan",
  hero_secondary_cta_label: null,
  hero_secondary_cta_url: null,
  hero_image_url: null,
  navbar_logo_url: null,
  brand_logo_shape: "original",
  brand_logo_fit: "contain",
  footer_logo_url: null,
  footer_use_brand_logo: true,
  footer_logo_shape: "original",
  footer_logo_fit: "contain",
  cabinet_logo_url: null,
  cabinet_logo_shape: "original",
  cabinet_logo_fit: "contain",
  cabinet_name: null,
  cabinet_period: null,
  cabinet_tagline: null,
  about_title: "Tentang BEM",
  about_description: "Profil organisasi",
  vision: "Visi bersama",
  missions: ["Misi satu", "Misi dua"],
  contact_email: null,
  contact_phone: null,
  contact_address: null,
  location_url: null,
  instagram_url: null,
  youtube_url: null,
  tiktok_url: null,
  other_social_url: null,
  social_links: [],
};

describe("system management screens", () => {
  it("renders operational metrics and the database-derived division scope", async () => {
    getDashboard.mockResolvedValue({
      generated_at: "2026-10-10T10:00:00Z",
      scope: { type: "division", division: { id: 4, name: "Kominfo" } },
      active_period: { id: 2, name: "2025/2026", start_date: "2025-08-01", end_date: "2026-07-31" },
      articles: { total: 5, published: 3, draft: 2 },
      registrations: { active_forms: 1, active_form_submissions: 12 },
      aspirations: { total: 8, pending: 2, read: 3, resolved: 3 },
      documents: { total: 6 },
      announcements: { total: 1 },
    });
    render(<DashboardAdmin />);

    expect(await screen.findByText("Ringkasan operasional · Divisi Kominfo")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Buka modul Pendaftaran aktif/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Buka modul Aspirasi masuk/ })).toBeInTheDocument();
    expect(screen.getByText("2025/2026")).toBeInTheDocument();
    expect(screen.getByText(/tidak mengumpulkan kunjungan/)).toBeInTheDocument();
  });
  it("lets an authorized editor reorder mission items accessibly", async () => {
    authState.user = { id: 1, role: "BPH" };
    getSettings.mockResolvedValue(settings);
    render(<LandingSettingsAdmin />);

    await screen.findByDisplayValue("Portal BEM");
    fireEvent.click(screen.getByRole("button", { name: "Visi & Misi" }));
    fireEvent.click(screen.getByRole("button", { name: "Pindahkan misi 2 ke atas" }));

    await waitFor(() => expect(screen.getByRole("textbox", { name: "Misi 1" })).toHaveValue("Misi dua"));
    expect(screen.getByRole("textbox", { name: "Misi 2" })).toHaveValue("Misi satu");
  });

  it("saves the nullable cabinet identity through the Branding settings section", async () => {
    authState.user = { id: 1, role: "BPH" };
    getSettings.mockResolvedValue(settings);
    saveSettings.mockImplementation(async (value) => ({ ...value, cabinet_name: "Kabinet Uji" }));
    render(<LandingSettingsAdmin />);

    await screen.findByDisplayValue("Portal BEM");
    fireEvent.click(screen.getByRole("button", { name: "Branding" }));
    expect(screen.getByText("Global Brand Logo")).toBeInTheDocument();
    expect(screen.getByLabelText("Gunakan Global Brand Logo untuk Footer")).toBeChecked();
    fireEvent.click(screen.getByLabelText("Gunakan Global Brand Logo untuk Footer"));
    expect(await screen.findByText("Pratinjau logo khusus Footer")).toBeInTheDocument();
    expect(screen.getByText(/Logo Kabinet/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Nama kabinet"), { target: { value: "Kabinet Uji" } });
    fireEvent.click(screen.getByRole("button", { name: "Simpan Perubahan" }));

    await waitFor(() =>
      expect(saveSettings).toHaveBeenCalledWith(expect.objectContaining({ cabinet_name: "Kabinet Uji" })),
    );
    expect(await screen.findByText("Perubahan Landing Page berhasil disimpan.")).toBeInTheDocument();
    expect(screen.queryByText("Perubahan draf belum diterapkan ke halaman publik.")).not.toBeInTheDocument();
  });

  it("manages ordered social channels with a platform preview and disabled switch", async () => {
    authState.user = { id: 1, role: "BPH" };
    getSettings.mockResolvedValue(settings);
    saveSettings.mockImplementation(async (value) => value);
    render(<LandingSettingsAdmin />);

    await screen.findByDisplayValue("Portal BEM");
    fireEvent.click(screen.getByRole("button", { name: "Media Sosial" }));
    fireEvent.click(screen.getByRole("button", { name: /Tambah media sosial/ }));
    fireEvent.change(screen.getByLabelText("URL (HTTPS)"), {
      target: { value: "https://instagram.com/bem" },
    });
    expect(screen.getByText("Pratinjau: Instagram")).toBeInTheDocument();
    fireEvent.click(screen.getByLabelText("Aktif"));
    fireEvent.click(screen.getByRole("button", { name: "Simpan Perubahan" }));

    await waitFor(() =>
      expect(saveSettings).toHaveBeenCalledWith(
        expect.objectContaining({
          social_links: [
            expect.objectContaining({
              platform: "instagram",
              url: "https://instagram.com/bem",
              active: false,
            }),
          ],
        }),
      ),
    );
  });

  it("keeps the draft and reports failed settings save without a contradictory dirty banner", async () => {
    authState.user = { id: 1, role: "BPH" };
    getSettings.mockResolvedValue(settings);
    saveSettings.mockRejectedValue(new Error("network"));
    render(<LandingSettingsAdmin />);

    await screen.findByDisplayValue("Portal BEM");
    fireEvent.click(screen.getByRole("button", { name: "Branding" }));
    fireEvent.change(screen.getByLabelText("Nama kabinet"), { target: { value: "Draf lokal" } });
    fireEvent.click(screen.getByRole("button", { name: "Simpan Perubahan" }));

    expect(await screen.findByText(/Draf tetap tersedia/)).toBeInTheDocument();
    expect(screen.getByLabelText("Nama kabinet")).toHaveValue("Draf lokal");
    expect(screen.queryByText("Perubahan draf belum diterapkan ke halaman publik.")).not.toBeInTheDocument();
  });

  it("does not request the read-only audit API for a non-SUPER_ADMIN", async () => {
    authState.user = { id: 2, role: "BPH" };
    render(<ActivityAdmin />);

    expect(screen.getByText("Activity Log hanya tersedia untuk SUPER_ADMIN.")).toBeInTheDocument();
    await waitFor(() => expect(getActivities).not.toHaveBeenCalled());
  });
});
