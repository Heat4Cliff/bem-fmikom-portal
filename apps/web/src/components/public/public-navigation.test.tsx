import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { PublicHeader, PublicNavigation } from "./public-navigation";

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

    expect(screen.getByRole("link", { name: "Berita" })).toHaveAttribute("href", "/#berita");
    expect(screen.getByRole("link", { name: "Pengumuman" })).toHaveAttribute("href", "/#pengumuman");
    expect(screen.getByRole("link", { name: "Pendaftaran" })).toHaveAttribute("href", "/#pendaftaran");
  });

  it("opens a keyboard accessible mobile drawer and closes it with Escape", async () => {
    render(<PublicHeader />);

    fireEvent.click(screen.getByRole("button", { name: "Buka menu navigasi" }));
    expect(await screen.findByRole("dialog", { name: "Navigasi BEM FMIKOM" })).toBeVisible();
    expect(screen.getByRole("link", { name: "Kirim Aspirasi" })).toHaveAttribute("href", "/#aspirasi");

    fireEvent.keyDown(document, { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
