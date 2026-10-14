import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ImgHTMLAttributes } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { PublicOrganization } from "@/features/content/types";
import { PublicOrganizationPage } from "./public-organization-page";

vi.mock("next/image", () => ({
  default: (props: ImgHTMLAttributes<HTMLImageElement> & { fill?: boolean; sizes?: string }) => {
    const { fill, ...imageProps } = props;
    return (
      // biome-ignore lint/performance/noImgElement: This mock preserves native image error semantics for MediaImage tests.
      <img {...imageProps} alt={imageProps.alt ?? ""} data-fill={fill ? "true" : undefined} />
    );
  },
}));

afterEach(cleanup);

const organization: PublicOrganization = {
  period: { name: "Nawasena 2026", start_date: "2026-01-01", end_date: "2026-12-31" },
  leadership: [
    { name: "Ketua", position: "KETUA", photo_url: null },
    { name: "Sekretaris", position: "SEKRETARIS", photo_url: "/api/v1/public/media/12" },
  ],
  divisions: [
    {
      name: "Divisi Riset",
      head: { name: "Kepala Riset", position: "KEPALA_BIDANG", photo_url: null },
      members: [{ name: "Anggota Riset", position: "ANGGOTA", photo_url: null }],
    },
    { name: "Divisi Media", head: null, members: [] },
  ],
};

function member(name: string): PublicOrganization["leadership"][number] {
  return { name, position: "ANGGOTA", photo_url: null };
}

function organizationWithDivision(
  members: PublicOrganization["divisions"][number]["members"],
  head: PublicOrganization["divisions"][number]["head"] = {
    name: "Kepala Riset",
    position: "KEPALA_BIDANG",
    photo_url: null,
  },
): PublicOrganization {
  return {
    ...organization,
    divisions: [{ name: "Divisi Riset", head, members }],
  };
}

describe("public organization detail", () => {
  it("renders active period, leadership, divisions, and members in API order without private data", () => {
    render(<PublicOrganizationPage organization={organization} />);

    expect(screen.getByRole("heading", { name: "Struktur Organisasi" })).toBeVisible();
    expect(screen.getByText("Periode Nawasena 2026")).toBeVisible();
    expect(screen.getAllByText("Ketua")).toHaveLength(2);
    expect(screen.getByText("Anggota Riset")).toBeVisible();
    expect(screen.getByText("Kepala Bidang belum ditetapkan")).toBeVisible();
    expect(screen.getByText("Belum ada anggota")).toBeVisible();
    const divisionNames = screen.getAllByRole("heading", { level: 3 }).map((heading) => heading.textContent);
    expect(divisionNames).toEqual(["Divisi Riset", "Divisi Media"]);
    expect(document.body.textContent).not.toContain("NIM");
  });

  it("keeps member images inside fixed square frames and provides fallback when absent or failed", () => {
    render(<PublicOrganizationPage organization={organization} />);

    expect(
      screen.getAllByTestId("media-frame").every((frame) => frame.className.includes("aspect-square")),
    ).toBe(true);
    expect(screen.getByRole("img", { name: "Foto Ketua tidak tersedia" })).toBeVisible();
    const image = screen.getByRole("img", { name: "Foto Sekretaris" });
    fireEvent.error(image);
    expect(screen.getByRole("img", { name: "Foto Sekretaris tidak tersedia" })).toBeVisible();
  });

  it("renders neutral empty and safe API error states", () => {
    const { rerender } = render(<PublicOrganizationPage organization={{ ...organization, period: null }} />);
    expect(screen.getByText(/Periode aktif belum tersedia/)).toBeVisible();
    rerender(<PublicOrganizationPage organization={null} loadError />);
    expect(screen.getByText(/belum dapat dimuat/)).toBeVisible();
    expect(screen.queryByText("Ketua")).not.toBeInTheDocument();
  });

  it("uses a single-column mobile layout that expands at tablet and desktop widths", () => {
    render(<PublicOrganizationPage organization={organization} />);
    const leadership = screen.getByRole("region", { name: "Badan Pengurus Harian" });
    const leadershipList = within(leadership).getByRole("list");
    expect(leadershipList.className).toContain("grid-cols-1");
    expect(leadershipList.className).toContain("sm:grid-cols-2");
    expect(leadershipList.className).toContain("xl:grid-cols-3");
    expect(screen.getByRole("heading", { name: "Struktur Organisasi" })).toBeVisible();
  });

  it("renders an odd number of divisions as sequential full-width charts", () => {
    const threeDivisions: PublicOrganization = {
      ...organization,
      divisions: [
        organization.divisions[0],
        organization.divisions[1],
        { name: "Divisi Ketiga", head: null, members: [] },
      ],
    };
    render(<PublicOrganizationPage organization={threeDivisions} />);

    const chartList = screen.getByTestId("division-chart-list");
    expect(chartList.className).toContain("space-y-8");
    expect(chartList.className).not.toMatch(/grid|xl:grid-cols/);
    expect(within(chartList).getAllByRole("article")).toHaveLength(3);
    expect(
      within(chartList)
        .getAllByRole("heading", { level: 3 })
        .map((heading) => heading.textContent),
    ).toEqual(["Divisi Riset", "Divisi Media", "Divisi Ketiga"]);
  });

  it("uses a direct vertical link for a head and one member without a horizontal branch", () => {
    render(<PublicOrganizationPage organization={organizationWithDivision([member("Anggota Tunggal")])} />);

    const chart = screen.getByRole("article", { name: "Divisi Riset" });
    expect(
      within(chart).getByRole("button", { name: "Lihat profil Kepala Riset, Kepala Bidang" }),
    ).toBeVisible();
    expect(
      within(chart).getByRole("button", { name: "Lihat profil Anggota Tunggal, Anggota" }),
    ).toBeVisible();
    expect(within(chart).queryByTestId("member-branch-line")).not.toBeInTheDocument();
    expect(
      within(chart).getByTestId("organization-children").querySelector('[aria-hidden="true"]'),
    ).toBeInTheDocument();
  });

  it.each([2, 3, 5])("branches a centered member group for %i members", (count) => {
    const members = Array.from({ length: count }, (_, index) => member(`Anggota ${index + 1}`));
    render(<PublicOrganizationPage organization={organizationWithDivision(members)} />);

    const chart = screen.getByRole("article", { name: "Divisi Riset" });
    const rows = within(chart).getAllByTestId("member-row");
    expect(rows).toHaveLength(count > 4 ? 2 : 1);
    expect(rows.flatMap((row) => within(row).getAllByRole("button"))).toHaveLength(count);
    expect(within(chart).getAllByTestId("member-branch-line")).toHaveLength(count > 1 ? 1 : 0);
    expect(within(chart).getAllByRole("button")).toHaveLength(count + 1);
  });

  it("keeps the member sorting from the API response", () => {
    const orderedMembers = [member("Anggota C"), member("Anggota A"), member("Anggota B")];
    render(<PublicOrganizationPage organization={organizationWithDivision(orderedMembers)} />);

    const chart = screen.getByRole("article", { name: "Divisi Riset" });
    expect(
      within(chart)
        .getAllByRole("button")
        .map((button) => button.getAttribute("aria-label")),
    ).toEqual([
      "Lihat profil Kepala Riset, Kepala Bidang",
      "Lihat profil Anggota C, Anggota",
      "Lihat profil Anggota A, Anggota",
      "Lihat profil Anggota B, Anggota",
    ]);
  });

  it("does not draw empty connectors when a division has no members", () => {
    render(<PublicOrganizationPage organization={organizationWithDivision([])} />);

    const chart = screen.getByRole("article", { name: "Divisi Riset" });
    expect(within(chart).getByText("Belum ada anggota")).toBeVisible();
    expect(within(chart).queryByTestId("organization-children")).not.toBeInTheDocument();
    expect(within(chart).queryByTestId("member-branch-line")).not.toBeInTheDocument();
  });

  it("keeps missing-head empty states compact and does not invent connectors", () => {
    render(
      <PublicOrganizationPage
        organization={organizationWithDivision([member("Anggota Tanpa Kepala")], null)}
      />,
    );

    const chart = screen.getByRole("article", { name: "Divisi Riset" });
    expect(within(chart).getByText("Kepala Bidang belum ditetapkan").className).toContain("max-w-[17.5rem]");
    expect(
      within(chart).getByRole("button", { name: "Lihat profil Anggota Tanpa Kepala, Anggota" }),
    ).toBeVisible();
    expect(within(chart).queryByTestId("member-branch-line")).not.toBeInTheDocument();
    expect(within(chart).queryByTestId("head-member-stem")).not.toBeInTheDocument();
  });

  it("renders only API-supplied public work divisions, excluding Faculty and BPH layers", () => {
    render(<PublicOrganizationPage organization={organization} />);

    const chartNames = screen
      .getAllByTestId("division-chart")
      .map((chart) => within(chart).getByRole("heading", { level: 3 }).textContent);
    expect(chartNames).toEqual(["Divisi Riset", "Divisi Media"]);
    expect(chartNames).not.toContain("Fakultas");
    expect(chartNames).not.toContain("BPH");
  });

  it("opens a public-safe member detail from an accessible card and supports keyboard dismissal", async () => {
    const user = userEvent.setup();
    render(<PublicOrganizationPage organization={organization} />);

    const card = screen.getByRole("button", { name: "Lihat profil Ketua, Ketua" });
    card.focus();
    await user.keyboard(" ");

    const dialog = await screen.findByRole("dialog", { name: "Profil Pengurus" });
    expect(within(dialog).getByRole("heading", { name: "Ketua" })).toBeVisible();
    expect(within(dialog).getByText("Nawasena 2026")).toBeVisible();
    expect(dialog.textContent).not.toMatch(/NIM|email|telepon|user.?id|database.?id/i);

    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await waitFor(() => expect(card).toHaveFocus());
  });
});
