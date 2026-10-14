import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { MediaImage } from "./media-image";

afterEach(cleanup);

describe("MediaImage", () => {
  it("renders an intentional initials placeholder when there is no image", () => {
    render(
      <MediaImage
        value={null}
        alt="Foto Lukluk"
        width={40}
        height={40}
        className="size-10 rounded-full"
        fallbackText="LU"
      />,
    );

    expect(screen.getByRole("img", { name: "Foto Lukluk tidak tersedia" })).toHaveTextContent("LU");
    expect(screen.queryByRole("img", { name: "Foto Lukluk" })).not.toBeInTheDocument();
  });

  it("resolves an internal media reference and falls back after one failed image request", () => {
    const failure = vi.spyOn(console, "error").mockImplementation(() => {});
    render(
      <MediaImage
        value="media:1"
        alt="Foto Lukluk"
        width={40}
        height={40}
        className="size-10 rounded-full"
        fallbackText="LU"
      />,
    );

    const image = screen.getByRole("img", { name: "Foto Lukluk" });
    expect(image).toHaveAttribute("src", "http://localhost:8080/api/v1/public/media/1");
    fireEvent.error(image);

    expect(screen.getByRole("img", { name: "Foto Lukluk tidak tersedia" })).toHaveTextContent("LU");
    expect(failure).toHaveBeenCalledOnce();
    failure.mockRestore();
  });
});
