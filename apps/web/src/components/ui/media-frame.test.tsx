import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MediaFrame } from "./media-frame";

describe("MediaFrame", () => {
  it("bounds content in a fixed aspect viewport independent of image intrinsic size", () => {
    render(
      <MediaFrame ratio="aspect-[4/3]" className="max-h-72">
        <span
          role="img"
          aria-label="preview"
          className="absolute inset-0 size-full object-contain object-center"
        />
      </MediaFrame>,
    );
    const frame = screen.getByTestId("media-frame");
    expect(frame.className).toContain("aspect-[4/3]");
    expect(frame.className).toContain("max-h-72");
    expect(frame.className).toContain("overflow-hidden");
    expect(screen.getByRole("img", { name: "preview" }).className).toContain("object-contain");
  });
});
