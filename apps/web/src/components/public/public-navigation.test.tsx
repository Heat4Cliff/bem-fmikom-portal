import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MobileNavigation, PublicHeader, PublicNavigation } from "./public-navigation";

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

  it("keeps the fixed mobile navigation touch targets above the safe area", () => {
    render(<MobileNavigation />);

    const nav = screen.getByRole("navigation", { name: "Navigasi bawah" });
    expect(nav.className).toContain("pb-[max(env(safe-area-inset-bottom),0.35rem)]");
    expect(screen.getAllByRole("link").every((link) => link.className.includes("min-h-12"))).toBe(true);
    expect(screen.queryByRole("link", { name: "Portal Admin" })).not.toBeInTheDocument();
  });
});
