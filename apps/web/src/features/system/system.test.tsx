import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { LandingSettingsAdmin } from "./settings-admin";
import { ActivityAdmin } from "./activity-admin";
import { DashboardAdmin } from "./dashboard-admin";

const { authState, getSettings, getActivities, getDashboard } = vi.hoisted(() => ({
  authState: { user: null as { id: number; role: string } | null },
  getSettings: vi.fn(),
  getActivities: vi.fn(),
  getDashboard: vi.fn(),
}));

vi.mock("@/lib/auth/auth-provider", () => ({ useAuth: () => ({ user: authState.user }) }));
vi.mock("./api", () => ({
  getSettings,
  saveSettings: vi.fn(),
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

  it("does not request the read-only audit API for a non-SUPER_ADMIN", async () => {
    authState.user = { id: 2, role: "BPH" };
    render(<ActivityAdmin />);

    expect(screen.getByText("Activity Log hanya tersedia untuk SUPER_ADMIN.")).toBeInTheDocument();
    await waitFor(() => expect(getActivities).not.toHaveBeenCalled());
  });
});
