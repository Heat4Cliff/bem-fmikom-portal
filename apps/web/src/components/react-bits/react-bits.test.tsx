import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { BlurText } from "./blur-text";
import { CountUp } from "./count-up";
import { FadeContent } from "./fade-content";

vi.mock("motion/react", async (importOriginal) => {
  const motionModule = await importOriginal<typeof import("motion/react")>();
  return { ...motionModule, useReducedMotion: () => true };
});

describe("React Bits motion adaptations", () => {
  it("keeps hero text readable without an animated hidden state for reduced motion", () => {
    render(<BlurText text="Sistem Informasi FMIKOM" className="hero-title" />);

    expect(screen.getByRole("heading", { name: "Sistem Informasi FMIKOM" })).toBeVisible();
    expect(screen.getByRole("heading").textContent).toBe("Sistem Informasi FMIKOM");
    expect(document.querySelectorAll("[style]")).toHaveLength(0);
  });

  it("keeps fade content visible and usable when reduced motion is enabled", () => {
    render(
      <FadeContent>
        <a href="#section">Konten tetap tersedia</a>
      </FadeContent>,
    );

    expect(screen.getByRole("link", { name: "Konten tetap tersedia" })).toBeVisible();
    expect(document.querySelectorAll("[style]")).toHaveLength(0);
  });

  it("renders the actual dashboard value accessibly with reduced motion", () => {
    render(<CountUp to={142} />);

    expect(screen.getByText("142", { selector: ".sr-only" })).toBeInTheDocument();
    expect(screen.getByText("142", { selector: "span[aria-hidden='true']" })).toBeInTheDocument();
  });
});
