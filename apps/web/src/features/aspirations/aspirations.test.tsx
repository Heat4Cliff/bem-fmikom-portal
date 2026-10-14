import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ApiError } from "@/lib/api/client";
import { AspirationAdminList, AspirationDetail } from "./aspiration-admin";
import { PublicAspirationForm } from "./public-aspiration-form";
import {
  getAspiration,
  getAspirations,
  getPublicAspirationDivisions,
  submitAspiration,
  updateAspirationStatus,
} from "./api";
import type { Aspiration } from "./types";

vi.mock("./api", () => ({
  getAspiration: vi.fn(),
  getAspirations: vi.fn(),
  getPublicAspirationDivisions: vi.fn(),
  submitAspiration: vi.fn(),
  updateAspirationStatus: vi.fn(),
}));
vi.mock("@/features/organization/api", () => ({ getDivisions: vi.fn().mockResolvedValue({ items: [] }) }));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

const identified: Aspiration = {
  id: 41,
  division_id: 2,
  division_name: "Akademik",
  name: "Ayu",
  nim: "12345",
  email: "ayu@example.test",
  subject: "Perbaikan ruang belajar",
  message: "Mohon lampu ruang belajar diperbaiki.",
  is_anonymous: false,
  status: "PENDING",
  submitted_at: "2026-10-10T01:00:00Z",
  read_at: null,
  resolved_at: null,
  created_at: "2026-10-10T01:00:00Z",
  updated_at: "2026-10-10T01:00:00Z",
};

describe("aspiration public form", () => {
  it("loads active destination options, strips visible identity in anonymous mode, and confirms submission", async () => {
    const user = userEvent.setup();
    vi.mocked(getPublicAspirationDivisions).mockResolvedValue([{ id: 2, name: "Akademik" }]);
    vi.mocked(submitAspiration).mockResolvedValue({
      id: 41,
      status: "PENDING",
      submitted_at: identified.submitted_at,
    });
    render(<PublicAspirationForm />);

    await user.selectOptions(await screen.findByLabelText(/Tujuan Aspirasi/), "2");
    await user.type(screen.getByLabelText("Nama"), "Nama pribadi");
    await user.type(screen.getByLabelText("NIM"), "9988");
    await user.click(screen.getByRole("checkbox", { name: /Kirim sebagai anonim/ }));
    expect(screen.queryByLabelText("Nama")).not.toBeInTheDocument();
    expect(
      screen.getByText("Aspirasi akan dikirim tanpa identitas; nama, NIM, dan email tidak akan disimpan."),
    ).toBeInTheDocument();
    await user.type(screen.getByLabelText(/Judul Aspirasi/), "Lampu ruang belajar");
    await user.type(screen.getByLabelText(/Isi Aspirasi/), "Mohon diperbaiki.");
    await user.click(screen.getByRole("button", { name: "Kirim Aspirasi" }));

    expect(await screen.findByRole("heading", { name: "Aspirasi berhasil dikirim." })).toBeInTheDocument();
    expect(screen.getByText("ASP-41")).toBeInTheDocument();
    const successActions = screen.getByRole("group", { name: "Aksi setelah aspirasi terkirim" });
    const reference = successActions.querySelector('[data-slot="aspiration-reference"]');
    const retryAction = screen.getByRole("button", { name: "Kirim aspirasi lain" });
    expect(reference).toHaveTextContent("Nomor referensi: ASP-41");
    expect(successActions).toHaveClass("mt-5", "flex-col", "gap-3", "sm:flex-row");
    expect(retryAction).toHaveClass("min-h-10", "whitespace-nowrap", "w-full", "sm:w-auto");
    await user.tab();
    expect(retryAction).toHaveFocus();
    expect(submitAspiration).toHaveBeenCalledWith(
      expect.objectContaining({
        division_id: 2,
        is_anonymous: true,
        name: undefined,
        nim: undefined,
        email: undefined,
      }),
    );
  });

  it("shows required-field errors and readable rate-limit feedback", async () => {
    const user = userEvent.setup();
    vi.mocked(getPublicAspirationDivisions).mockResolvedValue([{ id: 2, name: "Akademik" }]);
    vi.mocked(submitAspiration).mockRejectedValue(new ApiError("rate limited", 429, "RATE_LIMITED"));
    render(<PublicAspirationForm />);
    await screen.findByRole("option", { name: "Akademik" });
    await user.click(screen.getByRole("button", { name: "Kirim Aspirasi" }));
    expect(await screen.findByText("Pilih tujuan aspirasi.")).toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText(/Tujuan Aspirasi/), "2");
    await user.type(screen.getByLabelText(/Judul Aspirasi/), "Masukan");
    await user.type(screen.getByLabelText(/Isi Aspirasi/), "Isi masukan");
    await user.click(screen.getByRole("button", { name: "Kirim Aspirasi" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Pengiriman terlalu sering");
  });
});

describe("aspiration admin surfaces", () => {
  it("renders anonymous sender and responsive record link in the list", async () => {
    vi.mocked(getAspirations).mockResolvedValue({
      items: [{ ...identified, is_anonymous: true, name: null, nim: null, email: null }],
      meta: { page: 1, limit: 20, total: 1, total_pages: 1 },
    });
    render(<AspirationAdminList />);
    expect(await screen.findByRole("heading", { name: "Manajemen Aspirasi" })).toBeInTheDocument();
    expect(await screen.findAllByText("Anonim")).not.toHaveLength(0);
    expect(screen.getAllByRole("link", { name: /detail/i })[0]).toHaveAttribute("href", "/admin/aspirasi/41");
  });

  it("shows sender data only when identified and exposes only the next valid status action", async () => {
    const user = userEvent.setup();
    vi.mocked(getAspiration).mockResolvedValue(identified);
    vi.mocked(updateAspirationStatus).mockResolvedValue({
      ...identified,
      status: "READ",
      read_at: "2026-10-10T02:00:00Z",
    });
    render(<AspirationDetail id={41} />);
    expect(await screen.findByText("ayu@example.test")).toBeInTheDocument();
    expect(screen.getByText(identified.message)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tandai Dibaca" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Tandai Selesai" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Tandai Dibaca" }));
    await waitFor(() => expect(updateAspirationStatus).toHaveBeenCalledWith(41, "READ"));
    expect(await screen.findByRole("button", { name: "Tandai Selesai" })).toBeInTheDocument();
  });
});
