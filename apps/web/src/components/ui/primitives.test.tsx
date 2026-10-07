import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from "@/components/ui/overlays";
import { Input } from "@/components/ui/primitives";

describe("core UI primitives", () => {
  it("keeps the selected button variant and disabled state", () => {
    render(
      <Button variant="danger" disabled>
        Hapus
      </Button>,
    );
    const button = screen.getByRole("button", { name: "Hapus" });

    expect(button).toBeDisabled();
    expect(button).toHaveClass("bg-danger");
  });

  it("connects input validation text to the control", () => {
    render(<Input id="email" type="email" error="Email wajib diisi" />);
    const input = screen.getByRole("textbox", { name: "" });

    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input).toHaveAttribute("aria-describedby", "email-error");
    expect(screen.getByText("Email wajib diisi")).toHaveAttribute("id", "email-error");
  });

  it("opens a dialog with an accessible name", async () => {
    render(
      <Dialog>
        <DialogTrigger asChild>
          <Button variant="outline">Buka dialog</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogTitle>Konfirmasi perubahan</DialogTitle>
          <p>Periksa kembali sebelum melanjutkan.</p>
        </DialogContent>
      </Dialog>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Buka dialog" }));
    const dialog = await screen.findByRole("dialog", { name: "Konfirmasi perubahan" });
    expect(dialog).toBeVisible();
    expect(screen.getByText("Buka dialog").closest("button")).toHaveAttribute("aria-expanded", "true");

    fireEvent.keyDown(document, { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
