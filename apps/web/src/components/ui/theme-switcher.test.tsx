import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { ThemeSwitcher } from "./theme-switcher";

describe("multi-theme switcher", () => {
  beforeEach(() => {
    document.documentElement.dataset.theme = "nawasena";
    window.localStorage.clear();
  });

  afterEach(() => {
    cleanup();
    delete document.documentElement.dataset.theme;
    document.body.style.pointerEvents = "";
  });

  it("defaults to Nawasena and offers a compact accessible theme button", () => {
    render(<ThemeSwitcher />);
    expect(screen.getByRole("button", { name: "Ubah tema" })).toHaveTextContent("Tema");
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it.each([
    ["nawasena", "Nawasena"],
    ["classic", "Classic"],
    ["salt", "Salt"],
    ["matcha", "Matcha"],
    ["mint", "Mint"],
  ] as const)("selects %s immediately and persists the preference", async (theme, name) => {
    const user = userEvent.setup();
    render(<ThemeSwitcher />);
    await user.click(screen.getByRole("button", { name: "Ubah tema" }));
    const option = await screen.findByRole("button", { name: new RegExp(name) });
    await user.click(option);

    expect(document.documentElement).toHaveAttribute("data-theme", theme);
    expect(window.localStorage.getItem("bem-fmikom-theme")).toBe(theme);
    expect(option).toHaveAttribute("aria-pressed", "true");
  });

  it("shows five theme previews in the requested order and identifies the current choice", async () => {
    render(<ThemeSwitcher />);
    fireEvent.click(screen.getByRole("button", { name: "Ubah tema" }));
    const dialog = await screen.findByRole("dialog", { name: "Pilih Tema" });
    expect(dialog).toBeVisible();
    const choices = within(dialog)
      .getAllByRole("button")
      .filter((choice) => choice.hasAttribute("aria-pressed"));
    expect(choices.map((choice) => choice.textContent?.replace(/\s+/g, " ").trim())).toEqual([
      expect.stringContaining("Nawasena"),
      expect.stringContaining("Classic"),
      expect.stringContaining("Salt"),
      expect.stringContaining("Matcha"),
      expect.stringContaining("Mint"),
    ]);
    expect(screen.getByRole("button", { name: /Nawasena/ })).toHaveAttribute("aria-pressed", "true");
  });

  it("closes on Escape and restores keyboard focus to the theme button", async () => {
    const user = userEvent.setup();
    render(<ThemeSwitcher />);
    const trigger = screen.getByRole("button", { name: "Ubah tema" });
    await user.click(trigger);
    expect(await screen.findByRole("dialog", { name: "Pilih Tema" })).toBeVisible();
    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(trigger).toHaveFocus();
  });

  it("keeps the preference across public/admin remounts and safely falls back for invalid values", async () => {
    const user = userEvent.setup();
    const firstRender = render(<ThemeSwitcher />);
    await user.click(screen.getByRole("button", { name: "Ubah tema" }));
    await user.click(await screen.findByRole("button", { name: /Salt/ }));
    firstRender.unmount();

    const secondRender = render(<ThemeSwitcher />);
    await user.click(screen.getByRole("button", { name: "Ubah tema" }));
    expect(await screen.findByRole("button", { name: /Salt/ })).toHaveAttribute("aria-pressed", "true");

    secondRender.unmount();
    document.documentElement.dataset.theme = "unsupported";
    window.localStorage.setItem("bem-fmikom-theme", "unsupported");
    render(<ThemeSwitcher />);
    fireEvent.click(screen.getByRole("button", { name: "Ubah tema" }));
    expect(await screen.findByRole("button", { name: /Nawasena/ })).toHaveAttribute("aria-pressed", "true");
  });

  it("opens the picker from the keyboard", async () => {
    const user = userEvent.setup();
    render(<ThemeSwitcher />);
    const trigger = screen.getByRole("button", { name: "Ubah tema" });
    trigger.focus();
    await user.keyboard("{Enter}");
    expect(await screen.findByRole("dialog", { name: "Pilih Tema" })).toBeVisible();
  });
});
