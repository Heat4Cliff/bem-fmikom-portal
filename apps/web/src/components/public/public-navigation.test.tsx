import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MobileNavigation, PublicHeader, PublicNavigation } from "./public-navigation";
import { PublicFooter } from "./public-navigation";

const navigation = vi.hoisted(() => ({ pathname: "/" }));
vi.mock("next/navigation", () => ({ usePathname: () => navigation.pathname }));

describe("public navigation", () => {
  beforeEach(() => {
    navigation.pathname = "/";
  });

  afterEach(() => {
    cleanup();
    document.body.style.pointerEvents = "";
  });

  it("uses stable links for the landing page sections", () => {
    render(<PublicNavigation />);

    const labels = [
      "Beranda",
      "Tentang",
      "Struktur",
      "Berita",
      "Galeri",
      "Pengumuman",
      "Pendaftaran",
      "Aspirasi",
    ];
    expect(screen.getAllByRole("link").map((link) => link.textContent)).toEqual(labels);
    expect(screen.queryByRole("link", { name: "Kirim Aspirasi" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Portal Admin" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Struktur" })).toHaveAttribute("href", "/struktur");
    expect(screen.getByRole("link", { name: "Tentang" })).toHaveAttribute("href", "/tentang");
    expect(screen.getByRole("link", { name: "Berita" })).toHaveAttribute("href", "/#berita");
    expect(screen.getByRole("link", { name: "Pengumuman" })).toHaveAttribute("href", "/#pengumuman");
    expect(screen.getByRole("link", { name: "Pendaftaran" })).toHaveAttribute("href", "/#pendaftaran");
    expect(screen.getByRole("link", { name: "Aspirasi" })).toHaveAttribute("href", "/#aspirasi");
  });

  it("uses the same public-only links in the mobile drawer list", () => {
    render(<PublicNavigation mobile />);

    expect(screen.getAllByRole("link").map((link) => link.textContent)).toEqual([
      "Beranda",
      "Tentang",
      "Struktur",
      "Berita",
      "Galeri",
      "Pengumuman",
      "Pendaftaran",
      "Aspirasi",
    ]);
    expect(screen.queryByText("Kirim Aspirasi")).not.toBeInTheDocument();
    expect(screen.queryByText("Portal Admin")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Struktur" })).toHaveAttribute("href", "/struktur");
  });

  it("opens a keyboard accessible mobile drawer and closes it with Escape", async () => {
    render(<PublicHeader />);

    fireEvent.click(screen.getByRole("button", { name: "Buka menu navigasi" }));
    expect(await screen.findByRole("dialog", { name: "Navigasi BEM FMIKOM" })).toBeVisible();
    expect(screen.getByRole("link", { name: "Aspirasi" })).toHaveAttribute("href", "/#aspirasi");
    expect(screen.queryByRole("link", { name: "Portal Admin" })).not.toBeInTheDocument();

    fireEvent.keyDown(document, { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("uses configured header and footer logos in fixed contain boxes", () => {
    const { rerender } = render(<PublicHeader logoUrl="/api/v1/public/media/11" logoShape="circle" />);
    const header = screen.getByRole("link", { name: "BEM FMIKOM UNUGHA, beranda" });
    const headerLogo = header.querySelector("img");
    if (!headerLogo) throw new Error("Global logo image was not rendered in the header");
    expect(headerLogo).toHaveAttribute("src", expect.stringContaining("/api/v1/public/media/11"));
    expect(headerLogo.className).toContain("object-contain");
    expect(header.querySelector('[data-logo-shape="circle"]')).toBeInTheDocument();

    rerender(<PublicFooter logoUrl="/api/v1/public/media/12" logoShape="rounded" logoFit="cover" />);
    const footer = screen.getByRole("contentinfo");
    const footerLogo = footer.querySelector("img");
    if (!footerLogo) throw new Error("Global logo image was not rendered in the footer");
    expect(footerLogo).toHaveAttribute("src", expect.stringContaining("/api/v1/public/media/12"));
    expect(footerLogo.className).toContain("object-cover");
    expect(footer.querySelector('[data-logo-shape="rounded"]')).toBeInTheDocument();
  });

  it("keeps the fixed mobile navigation touch targets above the safe area", () => {
    render(<MobileNavigation />);

    const nav = screen.getByRole("navigation", { name: "Navigasi bawah" });
    expect(nav.className).toContain("pb-[max(env(safe-area-inset-bottom),0.35rem)]");
    expect(screen.getAllByRole("link").every((link) => link.className.includes("min-h-12"))).toBe(true);
    expect(screen.queryByRole("link", { name: "Portal Admin" })).not.toBeInTheDocument();
  });
});
