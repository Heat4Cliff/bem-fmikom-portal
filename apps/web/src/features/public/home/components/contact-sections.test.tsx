import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { ContactSection } from "./contact-sections";

afterEach(cleanup);

describe("public contact social links", () => {
  it("shows local brand marks, readable labels, and safe keyboardable external links", () => {
    render(
      <ContactSection
        demoMode={false}
        contact={{
          socialLinks: [
            { platform: "instagram", label: "Instagram", href: "https://instagram.com/bem" },
            { platform: "tiktok", label: "TikTok", href: "https://tiktok.com/@bem" },
            { platform: "future-platform", label: "Kanal baru", href: "https://example.test/social" },
          ],
        }}
      />,
    );

    const instagram = screen.getByRole("link", { name: "Instagram (tautan eksternal)" });
    expect(instagram).toHaveAttribute("href", "https://instagram.com/bem");
    expect(instagram).toHaveAttribute("target", "_blank");
    expect(instagram).toHaveAttribute("rel", "noopener noreferrer");
    expect(instagram.querySelector("svg")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "TikTok (tautan eksternal)" })).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Kanal baru (tautan eksternal)" }).querySelector("svg"),
    ).toBeInTheDocument();
  });
});
