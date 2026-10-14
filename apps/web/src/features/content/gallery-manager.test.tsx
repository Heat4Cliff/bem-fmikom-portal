import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GalleryManager } from "./gallery-manager";

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  apiRequestEnvelope: vi.fn(),
  getDivisions: vi.fn(),
  getPeriods: vi.fn(),
}));

vi.mock("@/lib/api/client", () => ({
  apiRequest: mocks.apiRequest,
  apiRequestEnvelope: mocks.apiRequestEnvelope,
}));
vi.mock("@/lib/auth/auth-provider", () => ({ useAuth: () => ({ user: { role: "SUPER_ADMIN" } }) }));
vi.mock("@/features/organization/api", () => ({
  getDivisions: mocks.getDivisions,
  getPeriods: mocks.getPeriods,
}));
vi.mock("@/components/admin/media-upload", () => ({
  MediaUpload: ({
    label,
    value,
    onChange,
  }: {
    label: string;
    value: string;
    onChange: (value: string) => void;
  }) => (
    <div>
      <span>{label}</span>
      {value ? <span>Pratinjau tersimpan</span> : null}
      <button type="button" onClick={() => onChange("/api/v1/public/media/41")}>
        Simulasikan unggah pratinjau
      </button>
    </div>
  ),
}));

const categorySeed = [
  {
    id: 11,
    title: "Workshop UI/UX",
    slug: "workshop-ui-ux",
    period_id: 4,
    period_name: "Periode 2024/2025",
    division_id: 8,
    division_name: "Biro Kominfo",
    item_count: 1,
  },
  { id: 12, title: "Seminar", slug: "seminar", item_count: 0 },
];
const itemSeed = [
  {
    id: 51,
    category_id: 11,
    title: "Dokumentasi Sesi Pertama",
    description: "Foto pelaksanaan kegiatan",
    google_photos_url: "https://photos.app.goo.gl/workshop-album-2026",
    preview_image_url: "/api/v1/public/media/41",
    sort_order: 1,
  },
];

let categories: typeof categorySeed;
let items: typeof itemSeed;

beforeEach(() => {
  categories = [...categorySeed];
  items = [...itemSeed];
  mocks.apiRequestEnvelope.mockImplementation(async (path: string) => {
    if (path.startsWith("/admin/gallery-categories")) {
      return {
        success: true,
        data: categories,
        meta: { page: 1, limit: 100, total: categories.length, total_pages: 1 },
      };
    }
    if (path.startsWith("/admin/gallery-items")) {
      const categoryID = Number(new URLSearchParams(path.split("?")[1]).get("category_id"));
      return { success: true, data: items.filter((item) => item.category_id === categoryID) };
    }
    return { success: true, data: [] };
  });
  mocks.getPeriods.mockResolvedValue({
    items: [{ id: 4, name: "Periode 2024/2025", status: "ACTIVE" }],
    meta: { page: 1, limit: 100, total: 1, total_pages: 1 },
  });
  mocks.getDivisions.mockResolvedValue({
    items: [
      { id: 8, period_id: 4, period_name: "Periode 2024/2025", name: "Biro Kominfo", status: "ACTIVE" },
    ],
    meta: { page: 1, limit: 100, total: 1, total_pages: 1 },
  });
  mocks.apiRequest.mockImplementation(
    async (path: string, options: { method?: string; body?: Record<string, unknown> } = {}) => {
      if (path === "/admin/gallery-categories" && options.method === "POST") {
        const created = {
          id: 13,
          title: String(options.body?.title),
          slug: "workshop-desain",
          period_id: options.body?.period_id as number,
          period_name: "Periode 2024/2025",
          division_id: options.body?.division_id as number,
          division_name: "Biro Kominfo",
          item_count: 0,
        };
        categories = [...categories, created];
        return created;
      }
      if (path === "/admin/gallery-items" && options.method === "POST") {
        const created = {
          id: 52,
          ...options.body,
        } as (typeof itemSeed)[number];
        items = [...items, created];
        const categoryIndex = categories.findIndex((category) => category.id === created.category_id);
        if (categoryIndex >= 0) categories[categoryIndex] = { ...categories[categoryIndex], item_count: 2 };
        return created;
      }
      if (path.includes("/admin/gallery-items/") && options.method === "PATCH") {
        const id = Number(path.split("/").at(-1));
        items = items.map((item) =>
          item.id === id ? ({ ...item, ...options.body } as (typeof itemSeed)[number]) : item,
        );
        return items.find((item) => item.id === id);
      }
      if (path.includes("/admin/gallery-items/") && options.method === "DELETE") {
        const id = Number(path.split("/").at(-1));
        items = items.filter((item) => item.id !== id);
        return null;
      }
      return null;
    },
  );
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("GalleryManager master-detail workflow", () => {
  it("shows compact category navigation, active context and scannable album cards", async () => {
    render(<GalleryManager />);
    expect(await screen.findByRole("heading", { name: "Workshop UI/UX" })).toBeVisible();
    expect(screen.getByRole("navigation", { name: "Pilih kategori" })).toHaveClass("overflow-x-auto");
    const activeCategory = within(screen.getByRole("navigation", { name: "Pilih kategori" }))
      .getAllByRole("button")
      .find(
        (button) => button.hasAttribute("aria-pressed") && button.getAttribute("aria-pressed") === "true",
      );
    expect(activeCategory).toHaveTextContent(/Workshop UI\/UX/);
    expect(activeCategory).toHaveTextContent("1 dokumentasi");
    expect(screen.getByRole("link", { name: /Buka album.*photos\.app\.goo\.gl/ })).toHaveTextContent(
      "Google Photos",
    );
    expect(screen.queryByText("https://photos.app.goo.gl/workshop-album-2026")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Simpan kategori" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Simpan item" })).not.toBeInTheDocument();
    const itemList = screen.getByRole("list", { name: "Dokumentasi Workshop UI/UX" });
    expect(within(itemList).getByText("Urutan 1")).toBeInTheDocument();
    expect(itemList).toHaveClass("md:grid-cols-2");
  });

  it("creates categories in a dialog using human-readable period and division selectors", async () => {
    render(<GalleryManager />);
    fireEvent.click(await screen.findByRole("button", { name: "Tambah Kategori" }));
    const dialog = await screen.findByRole("dialog", { name: "Tambah kategori kegiatan" });
    fireEvent.change(within(dialog).getByLabelText("Judul kategori"), {
      target: { value: "Workshop Desain" },
    });
    fireEvent.change(within(dialog).getByLabelText("Periode organisasi"), { target: { value: "4" } });
    fireEvent.change(within(dialog).getByLabelText("Divisi penyelenggara"), { target: { value: "8" } });
    fireEvent.click(within(dialog).getByRole("button", { name: "Simpan Kategori" }));
    await waitFor(() =>
      expect(mocks.apiRequest).toHaveBeenCalledWith(
        "/admin/gallery-categories",
        expect.objectContaining({
          body: expect.objectContaining({ title: "Workshop Desain", period_id: 4, division_id: 8, slug: "" }),
        }),
      ),
    );
    await waitFor(() => {
      const active = within(screen.getByRole("navigation", { name: "Pilih kategori" }))
        .getAllByRole("button")
        .find(
          (button) => button.hasAttribute("aria-pressed") && button.getAttribute("aria-pressed") === "true",
        );
      expect(active).toHaveTextContent("Workshop Desain");
    });
    expect(screen.queryByRole("dialog", { name: "Tambah kategori kegiatan" })).not.toBeInTheDocument();
  });

  it("preselects the active category and saves album URL and uploaded preview separately", async () => {
    render(<GalleryManager />);
    fireEvent.click((await screen.findAllByRole("button", { name: "Tambah Dokumentasi" }))[0]);
    const dialog = await screen.findByRole("dialog", { name: "Tambah dokumentasi kegiatan" });
    expect(within(dialog).getByLabelText("Kategori")).toHaveValue("11");
    fireEvent.change(within(dialog).getByLabelText("Judul item"), {
      target: { value: "Dokumentasi Workshop" },
    });
    fireEvent.change(within(dialog).getByLabelText(/Tautan album Google Photos/), {
      target: { value: "https://photos.app.goo.gl/new-workshop" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "Simulasikan unggah pratinjau" }));
    fireEvent.click(within(dialog).getByRole("button", { name: "Simpan" }));
    await waitFor(() =>
      expect(mocks.apiRequest).toHaveBeenCalledWith(
        "/admin/gallery-items",
        expect.objectContaining({
          method: "POST",
          body: expect.objectContaining({
            category_id: 11,
            title: "Dokumentasi Workshop",
            google_photos_url: "https://photos.app.goo.gl/new-workshop",
            preview_image_url: "/api/v1/public/media/41",
          }),
        }),
      ),
    );
    expect(await screen.findByRole("heading", { name: "Dokumentasi Workshop" })).toBeVisible();
    expect(screen.getByRole("link", { name: /Buka album Dokumentasi Workshop/ })).toHaveAttribute(
      "href",
      "https://photos.app.goo.gl/new-workshop",
    );
  });

  it("edits and deletes items only after a confirmation dialog", async () => {
    render(<GalleryManager />);
    await screen.findByRole("heading", { name: "Workshop UI/UX" });
    const user = userEvent.setup();
    await user.click(
      await screen.findByRole("button", { name: "Aksi dokumentasi Dokumentasi Sesi Pertama" }),
    );
    await user.click(await screen.findByRole("menuitem", { name: "Edit" }));
    const dialog = await screen.findByRole("dialog", { name: "Edit dokumentasi kegiatan" });
    expect(within(dialog).getByLabelText("Kategori")).toHaveValue("11");
    fireEvent.change(within(dialog).getByLabelText("Judul item"), {
      target: { value: "Dokumentasi Diperbarui" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "Simpan" }));
    await waitFor(() =>
      expect(mocks.apiRequest).toHaveBeenCalledWith(
        "/admin/gallery-items/51",
        expect.objectContaining({
          method: "PATCH",
          body: expect.objectContaining({ title: "Dokumentasi Diperbarui" }),
        }),
      ),
    );

    await user.click(await screen.findByRole("button", { name: "Aksi dokumentasi Dokumentasi Diperbarui" }));
    await user.click(await screen.findByRole("menuitem", { name: "Hapus" }));
    const confirmation = await screen.findByRole("dialog", { name: "Hapus dokumentasi?" });
    expect(confirmation).toHaveTextContent("Dokumentasi Diperbarui");
    fireEvent.click(within(confirmation).getByRole("button", { name: "Hapus Dokumentasi" }));
    await waitFor(() =>
      expect(mocks.apiRequest).toHaveBeenCalledWith(
        "/admin/gallery-items/51",
        expect.objectContaining({ method: "DELETE" }),
      ),
    );
  });

  it("shows a useful empty state without keeping an item form open", async () => {
    items = [];
    render(<GalleryManager />);
    expect(await screen.findByText("Belum ada dokumentasi pada kategori ini.")).toBeVisible();
    expect(screen.getAllByRole("button", { name: "Tambah Dokumentasi" }).length).toBeGreaterThan(0);
    expect(screen.queryByLabelText("Tautan album Google Photos")).not.toBeInTheDocument();
  });
});
