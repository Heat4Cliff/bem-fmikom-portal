import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import type { RegistrationPreview } from "../types";
import { RegistrationSection } from "./registration-section";

afterEach(cleanup);

const makeRegistration = (id: string): RegistrationPreview => ({
  id,
  slug: id,
  title: `Pendaftaran ${id}`,
  status: "DIBUKA",
});

describe("registration section grid", () => {
  it.each([2, 3])("uses equal two-column desktop tracks and normal card width for %i cards", (count) => {
    const registrations = Array.from({ length: count }, (_, index) =>
      makeRegistration(`agenda-${index + 1}`),
    );
    const { container } = render(<RegistrationSection registrations={registrations} demoMode={false} />);
    const grid = container.querySelector("section > div > div.grid");
    const cards = screen.getAllByRole("article");

    expect(grid?.className).toContain("grid-cols-1");
    expect(grid?.className).toContain("lg:grid-cols-2");
    expect(cards).toHaveLength(count);
    expect(
      cards.every((card) => card.className.includes("w-full") && card.className.includes("h-full")),
    ).toBe(true);
    expect(cards.every((card) => !card.className.includes("col-span"))).toBe(true);
  });
});
