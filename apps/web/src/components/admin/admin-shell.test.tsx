import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AdminSidebar, AdminTopbar } from "./admin-shell";
import { BrandingProvider } from "@/features/content/branding-context";

vi.mock("next/navigation", () => ({ usePathname: () => "/admin" }));

afterEach(cleanup);

describe("admin global branding", () => {
  it("uses one configured global logo in the sidebar and mobile header", () => {
    render(
      <BrandingProvider value={{ logoUrl: "/api/v1/public/media/42", logoShape: "circle", logoFit: "cover" }}>
        <AdminSidebar adminRole="SUPER_ADMIN" />
        <AdminTopbar
          onMenuClick={vi.fn()}
          user={{
            id: 1,
            name: "Admin Uji",
            email: "admin@example.test",
            role: "SUPER_ADMIN",
            division: null,
          }}
          onLogout={vi.fn()}
        />
      </BrandingProvider>,
    );
    expect(screen.getByRole("link", { name: "BEM FMIKOM Admin Portal" })).toBeInTheDocument();
    const logos = document.querySelectorAll('[data-logo-shape="circle"][data-logo-fit="cover"]');
    expect(logos).toHaveLength(2);
    for (const logo of logos)
      expect(logo.querySelector("img")).toHaveAttribute(
        "src",
        expect.stringContaining("/api/v1/public/media/42"),
      );
  });
});
