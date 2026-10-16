import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { AboutProfilePage } from "./about-profile-page";

afterEach(cleanup);

const settings = {
  hero_headline: "Portal",
  hero_subheadline: "Sub",
  hero_primary_cta_label: "Lihat",
  hero_primary_cta_url: "#berita",
  hero_secondary_cta_label: null,
  hero_secondary_cta_url: null,
  hero_image_url: null,
  navbar_logo_url: null,
  brand_logo_shape: "original" as const,
  brand_logo_fit: "contain" as const,
  footer_logo_url: null,
  footer_use_brand_logo: true,
  footer_logo_shape: "original" as const,
  footer_logo_fit: "contain" as const,
  cabinet_logo_url: "/api/v1/public/media/8",
  cabinet_logo_shape: "circle" as const,
  cabinet_logo_fit: "cover" as const,
  cabinet_name: "Nawasena",
  cabinet_period: "2026",
  cabinet_tagline: "nawasena_abhiparya",
  about_title: "Tentang organisasi",
  about_description: "Deskripsi lengkap organisasi.",
  vision: "Visi lengkap organisasi.",
  missions: ["Misi satu lengkap", "Misi dua lengkap"],
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

describe("organization profile page", () => {
  it("renders CMS cabinet identity, full about, vision and mission without internal identifiers", () => {
    render(<AboutProfilePage settings={settings} />);

    const introDescription = screen.getByText(
      "Informasi organisasi dan arah gerak yang dikelola melalui pengaturan portal.",
    );
    const logo = screen.getByRole("img", { name: "Logo Nawasena" });
    const period = screen.getByText("Periode 2026");
    const cabinetName = screen.getByRole("heading", { name: "Nawasena" });
    const cabinetTagline = screen.getByText("nawasena_abhiparya");
    const introTitle = screen.getByRole("heading", { level: 1, name: "Tentang BEM FMIKOM UNUGHA" });
    const pageGrid = introTitle.parentElement?.parentElement;
    const identitySection = cabinetName.closest("section");

    expect(logo).toHaveAttribute("src", expect.stringContaining("/api/v1/public/media/8"));
    expect(logo).toHaveClass("object-contain");
    expect(identitySection).not.toBeNull();
    expect(identitySection).toHaveClass("flex", "flex-col", "items-center", "text-center");
    expect(identitySection).not.toHaveClass("border", "bg-surface", "shadow-card");
    expect(logo.parentElement).toHaveClass("bg-transparent", "border-0", "shadow-none");
    expect(period).toHaveClass("text-xs", "uppercase");
    expect(cabinetName).toHaveClass("text-2xl", "font-bold", "sm:text-3xl");
    expect(cabinetTagline).toHaveClass("text-sm", "text-muted", "sm:text-base");
    expect(introTitle).toHaveClass("break-words");
    expect(pageGrid).toHaveClass("grid-cols-[minmax(0,1fr)]", "min-w-0");
    expect(introDescription.compareDocumentPosition(logo) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(logo.compareDocumentPosition(period) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(period.compareDocumentPosition(cabinetName) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(
      cabinetName.compareDocumentPosition(cabinetTagline) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    expect(screen.getByText("Deskripsi lengkap organisasi.")).toBeInTheDocument();
    expect(screen.getByText("Visi lengkap organisasi.")).toBeInTheDocument();
    expect(screen.getByText("Misi satu lengkap")).toBeInTheDocument();
    expect(screen.queryByText(/admin\/|division_id|user_id/i)).not.toBeInTheDocument();
  });

  it("shows an intentional empty logo state without fabricated cabinet data", () => {
    render(
      <AboutProfilePage
        settings={{
          ...settings,
          cabinet_logo_url: null,
          cabinet_name: null,
          cabinet_period: null,
          cabinet_tagline: null,
        }}
      />,
    );
    expect(screen.getByRole("img", { name: "Logo kabinet belum tersedia" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Identitas kabinet belum tersedia" })).toBeInTheDocument();
  });

  it("keeps cabinet identity presentation centered and proportionate across breakpoints", () => {
    render(<AboutProfilePage settings={settings} />);

    const identity = screen.getByRole("heading", { name: "Nawasena" }).closest("section");
    const logo = screen.getByRole("img", { name: "Logo Nawasena" });

    expect(identity).toHaveClass("items-center", "text-center", "gap-4", "sm:gap-5");
    expect(logo.parentElement).toHaveClass("size-28", "max-w-full", "sm:size-36", "lg:size-44");
    expect(screen.getByText("Deskripsi lengkap organisasi.").closest("section")).not.toBe(identity);
  });
});
