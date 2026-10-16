import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { BrandLogo } from "./brand-logo";

afterEach(cleanup);

describe("BrandLogo", () => {
  it.each([
    ["original", "contain", "rounded-none", "object-contain"],
    ["rounded", "cover", "rounded-xl", "object-cover"],
    ["circle", "contain", "rounded-full", "object-contain"],
  ] as const)(
    "renders %s shape with %s fit without changing the image",
    (shape, fit, shapeClass, fitClass) => {
      const { container } = render(
        <BrandLogo src="/api/v1/public/media/24" shape={shape} fit={fit} alt="BEM FMIKOM UNUGHA" />,
      );
      const frame = container.querySelector(`[data-logo-shape="${shape}"]`);
      const image = screen.getByRole("img", { name: "BEM FMIKOM UNUGHA" });
      expect(frame?.className).toContain(shapeClass);
      expect(frame).toHaveAttribute("data-logo-fit", fit);
      expect(image.className).toContain(fitClass);
      expect(image).toHaveAttribute("src", expect.stringContaining("/api/v1/public/media/24"));
    },
  );

  it("shows a safe BEM fallback when no global logo is configured", () => {
    render(<BrandLogo />);
    expect(screen.getByRole("img", { name: "BEM FMIKOM UNUGHA" })).toHaveTextContent("BEM");
  });
});
