import { cleanup, render } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { SocialBrandIcon } from "./social-brand-icon";

afterEach(cleanup);

describe("social brand icons", () => {
  it.each(["instagram", "tiktok", "youtube", "facebook", "x", "whatsapp"])(
    "renders a recognizable local %s icon",
    (platform) => {
      const { container } = render(<SocialBrandIcon platform={platform} />);
      expect(container.querySelector("svg")).toBeInTheDocument();
      expect(container.querySelector("svg")).toHaveAttribute("aria-hidden", "true");
    },
  );

  it("uses the recognizable LinkedIn mark and a generic icon for unknown platforms", () => {
    const { container, rerender } = render(<SocialBrandIcon platform="linkedin" />);
    expect(container).toHaveTextContent("in");
    rerender(<SocialBrandIcon platform="future-platform" />);
    expect(container.querySelector("svg")).toBeInTheDocument();
  });
});
