import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import AdminLoginPage from "./page";
import { BrandingProvider } from "@/features/content/branding-context";

vi.mock("next/navigation", () => ({ useRouter: () => ({ replace: vi.fn() }) }));
vi.mock("@/lib/auth/auth-provider", () => ({ useAuth: () => ({ login: vi.fn() }) }));

afterEach(cleanup);

describe("admin login branding", () => {
  it("uses the inherited global logo and selected presentation with an accessible name", () => {
    render(
      <BrandingProvider
        value={{ logoUrl: "/api/v1/public/media/17", logoShape: "rounded", logoFit: "contain" }}
      >
        <AdminLoginPage />
      </BrandingProvider>,
    );
    const logo = screen.getByRole("img", { name: "BEM FMIKOM UNUGHA" });
    expect(logo).toHaveAttribute("src", expect.stringContaining("/api/v1/public/media/17"));
    expect(logo).toHaveClass("object-contain");
    expect(document.querySelector('[data-logo-shape="rounded"]')).toBeInTheDocument();
  });
});
