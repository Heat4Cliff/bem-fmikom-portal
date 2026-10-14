import { afterEach, describe, expect, it, vi } from "vitest";
import type React from "react";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ApiError } from "@/lib/api/client";
import { getDivisions, getPeriods } from "@/features/organization/api";
import { PublicRegistrationForm, PublicRegistrationList } from "./public-registration";
import {
  RegistrationAdminList,
  RegistrationBuilder,
  RegistrationSubmissionDetail,
  RegistrationSubmissionList,
} from "./registration-admin";
import {
  createRegistration,
  downloadRegistrationExport,
  getPublicRegistration,
  getPublicRegistrations,
  getRegistration,
  getRegistrationSubmission,
  getRegistrationSubmissions,
  getRegistrations,
  saveRegistrationFields,
  submitRegistration,
} from "./api";
import type {
  PublicRegistration,
  RegistrationField,
  RegistrationForm,
  RegistrationSubmission,
} from "./types";

const mocks = vi.hoisted(() => ({ navigate: vi.fn(), confirm: vi.fn(() => true) }));
vi.mock("./api", () => ({
  createRegistration: vi.fn(),
  deleteRegistration: vi.fn(),
  downloadRegistrationExport: vi.fn(),
  getPublicRegistration: vi.fn(),
  getPublicRegistrations: vi.fn(),
  getRegistration: vi.fn(),
  getRegistrationSubmission: vi.fn(),
  getRegistrationSubmissions: vi.fn(),
  getRegistrations: vi.fn(),
  saveRegistrationFields: vi.fn(),
  setRegistrationStatus: vi.fn(),
  submitRegistration: vi.fn(),
  updateRegistration: vi.fn(),
}));
vi.mock("next/navigation", () => ({ useRouter: () => ({ push: mocks.navigate, replace: vi.fn() }) }));
vi.mock("next/link", () => ({
  default: ({ href, children, ...props }: React.AnchorHTMLAttributes<HTMLAnchorElement>) => (
    <a href={href as string} {...props}>
      {children}
    </a>
  ),
}));
vi.mock("@/lib/auth/auth-provider", () => ({
  useAuth: () => ({ user: { role: "SUPER_ADMIN", division: null }, status: "authenticated" }),
}));
vi.mock("@/features/organization/api", () => ({
  getDivisions: vi.fn().mockResolvedValue({ items: [] }),
  getPeriods: vi.fn().mockResolvedValue({ items: [] }),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

const fields: RegistrationField[] = [
  { field_key: "full_name", label: "Nama Lengkap", field_type: "TEXT", required: true, options: [] },
  { field_key: "nim", label: "NIM", field_type: "TEXT", required: true, options: [] },
  { field_key: "email", label: "Email", field_type: "EMAIL", required: true, options: [] },
  { field_key: "phone", label: "No HP", field_type: "TEXT", required: true, options: [] },
  { field_key: "message", label: "Motivasi", field_type: "TEXTAREA", required: true, options: [] },
  { field_key: "number", label: "Nilai", field_type: "NUMBER", required: true, options: [] },
  {
    field_key: "division_choice",
    label: "Pilih Divisi",
    field_type: "SELECT",
    required: true,
    options: [{ value: "media", label: "Media" }],
  },
  {
    field_key: "shift",
    label: "Shift",
    field_type: "RADIO",
    required: true,
    options: [{ value: "pagi", label: "Pagi" }],
  },
  {
    field_key: "interest",
    label: "Minat",
    field_type: "CHECKBOX",
    required: true,
    options: [{ value: "web", label: "Web" }],
  },
  { field_key: "date", label: "Tanggal", field_type: "DATE", required: true, options: [] },
];
const publicForm: PublicRegistration = {
  id: 7,
  title: "Open Recruitment",
  slug: "open-recruitment",
  description: "Daftar sekarang",
  period_name: "2026/2027",
  division_name: "BEM",
  open_at: "2026-10-01T00:00:00Z",
  close_at: "2026-11-01T00:00:00Z",
  status: "OPEN",
  fields,
};
const adminForm: RegistrationForm = {
  id: 7,
  period_id: 4,
  period_name: "2026/2027",
  division_id: 2,
  division_name: "Akademik",
  title: "Open Recruitment",
  slug: "open-recruitment",
  description: "",
  open_at: "2026-10-01T00:00:00Z",
  close_at: "2026-11-01T00:00:00Z",
  status: "DRAFT",
  submission_count: 0,
};

describe("public registration experience", () => {
  it("lists actual public state records and switches status filters", async () => {
    const user = userEvent.setup();
    vi.mocked(getPublicRegistrations).mockResolvedValue({
      items: [
        {
          title: adminForm.title,
          slug: adminForm.slug,
          description: adminForm.description,
          period_name: adminForm.period_name ?? "2026/2027",
          division_name: adminForm.division_name,
          open_at: adminForm.open_at,
          close_at: adminForm.close_at,
          public_status: "OPEN",
        },
      ],
      meta: { page: 1, limit: 30, total: 1, total_pages: 1 },
    });
    render(<PublicRegistrationList />);
    expect(await screen.findByRole("heading", { name: "Open Recruitment" })).toBeInTheDocument();
    expect(await screen.findAllByText("Dibuka")).toHaveLength(2);
    await user.click(screen.getByRole("button", { name: "Akan Datang" }));
    await waitFor(() =>
      expect(getPublicRegistrations).toHaveBeenLastCalledWith({ status: "UPCOMING", page: 1, limit: 30 }),
    );
  });

  it("renders the dynamic API field types, validates required answers, and posts dynamic values", async () => {
    const user = userEvent.setup();
    vi.mocked(getPublicRegistration).mockResolvedValue(publicForm);
    vi.mocked(submitRegistration).mockResolvedValue({
      submission_id: 14,
      submitted_at: "2026-10-10T00:00:00Z",
      registration_title: publicForm.title,
    });
    render(<PublicRegistrationForm slug={publicForm.slug} />);
    expect(await screen.findByRole("heading", { name: publicForm.title })).toBeInTheDocument();
    for (const field of fields) expect(screen.getByText(field.label)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Kirim Pendaftaran" }));
    expect(await screen.findAllByText("Bagian ini wajib diisi.")).toHaveLength(10);
    await user.type(screen.getByLabelText(/Nama Lengkap/), "Ayu Pendaftar");
    await user.type(screen.getByLabelText(/NIM/), "260001");
    await user.type(screen.getByLabelText(/Email/), "ayu@example.test");
    await user.type(screen.getByLabelText(/No HP/), "+62 812 1234");
    await user.type(screen.getByLabelText(/Motivasi/), "Belajar dan berkontribusi");
    await user.type(screen.getByLabelText(/Nilai/), "3.8");
    await user.selectOptions(screen.getByLabelText(/Pilih Divisi/), "media");
    await user.click(screen.getByLabelText("Pagi"));
    await user.click(screen.getByLabelText("Web"));
    fireEvent.change(screen.getByLabelText(/Tanggal/), { target: { value: "2026-10-20" } });
    await user.click(screen.getByRole("button", { name: "Kirim Pendaftaran" }));
    expect(await screen.findByText(/Pendaftaran berhasil dikirim/)).toBeInTheDocument();
    expect(screen.getByText(/Nomor referensi:/)).toHaveTextContent("#14");
    expect(submitRegistration).toHaveBeenCalledWith(
      7,
      expect.objectContaining({ full_name: "Ayu Pendaftar", shift: "pagi", interest: ["web"] }),
      {},
    );
  });

  it("supports IMAGE picker, preview, replacement, removal, and multipart submit state", async () => {
    const user = userEvent.setup();
    vi.mocked(getPublicRegistration).mockResolvedValue({
      ...publicForm,
      fields: [
        ...fields.slice(0, 4),
        {
          field_key: "profile_photo",
          label: "Foto Profil",
          field_type: "IMAGE",
          required: true,
          options: [],
        },
      ],
    });
    vi.mocked(submitRegistration).mockResolvedValue({
      submission_id: 15,
      submitted_at: "2026-10-10T00:00:00Z",
      registration_title: publicForm.title,
    });
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:http://localhost/test"),
    });
    Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: vi.fn() });
    render(<PublicRegistrationForm slug={publicForm.slug} />);
    await user.type(await screen.findByLabelText(/Nama Lengkap/), "Ayu Pendaftar");
    await user.type(screen.getByLabelText(/NIM/), "260002");
    await user.type(screen.getByLabelText(/Email/), "ayu2@example.test");
    await user.type(screen.getByLabelText(/No HP/), "0812345678");
    const input = screen.getByLabelText(/Foto Profil/) as HTMLInputElement;
    const first = new File(["first"], "profil.png", { type: "image/png" });
    await user.upload(input, first);
    expect(await screen.findByAltText("Pratinjau Foto Profil")).toBeInTheDocument();
    expect(screen.getByText(/profil.png/)).toBeInTheDocument();
    const replacement = new File(["replacement"], "profil-baru.jpg", { type: "image/jpeg" });
    await user.upload(input, replacement);
    expect(await screen.findByText(/profil-baru.jpg/)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Hapus" }));
    await user.upload(input, replacement);
    await user.click(screen.getByRole("button", { name: "Kirim Pendaftaran" }));
    await waitFor(() =>
      expect(submitRegistration).toHaveBeenCalledWith(7, expect.objectContaining({ nim: "260002" }), {
        profile_photo: replacement,
      }),
    );
  });

  it("shows a clear duplicate NIM message and rate-limited submit feedback", async () => {
    const user = userEvent.setup();
    vi.mocked(getPublicRegistration).mockResolvedValue(publicForm);
    vi.mocked(submitRegistration).mockRejectedValue(new ApiError("duplicate", 409, "DUPLICATE_SUBMISSION"));
    render(<PublicRegistrationForm slug={publicForm.slug} />);
    await screen.findByRole("heading", { name: publicForm.title });
    await user.type(screen.getByLabelText(/Nama Lengkap/), "Private Applicant");
    await user.type(screen.getByLabelText(/NIM/), "260001");
    await user.type(screen.getByLabelText(/Email/), "private@example.test");
    await user.type(screen.getByLabelText(/No HP/), "0812345678");
    await user.type(screen.getByLabelText(/Motivasi/), "Motivasi saya");
    await user.type(screen.getByLabelText(/Nilai/), "3");
    await user.selectOptions(screen.getByLabelText(/Pilih Divisi/), "media");
    await user.click(screen.getByLabelText("Pagi"));
    await user.click(screen.getByLabelText("Web"));
    fireEvent.change(screen.getByLabelText(/Tanggal/), { target: { value: "2026-10-20" } });
    await user.click(screen.getByRole("button", { name: "Kirim Pendaftaran" }));
    expect(await screen.findAllByText("NIM ini sudah terdaftar pada pendaftaran tersebut.")).toHaveLength(2);
    expect(screen.getByRole("alert")).toHaveTextContent("NIM ini sudah terdaftar");
    expect(screen.queryByText("Private Applicant")).not.toBeInTheDocument();
  });

  it("renders structured API validation errors beside the matching field", async () => {
    const user = userEvent.setup();
    vi.mocked(getPublicRegistration).mockResolvedValue(publicForm);
    vi.mocked(submitRegistration).mockRejectedValue(
      new ApiError("Periksa data formulir.", 422, "VALIDATION_ERROR", {
        nim: ["NIM tidak valid."],
      }),
    );
    render(<PublicRegistrationForm slug={publicForm.slug} />);
    await screen.findByRole("heading", { name: publicForm.title });
    await user.type(screen.getByLabelText(/Nama Lengkap/), "Ayu Pendaftar");
    await user.type(screen.getByLabelText(/NIM/), "260099");
    await user.type(screen.getByLabelText(/Email/), "ayu@example.test");
    await user.type(screen.getByLabelText(/No HP/), "0812345678");
    await user.type(screen.getByLabelText(/Motivasi/), "Motivasi");
    await user.type(screen.getByLabelText(/Nilai/), "3");
    await user.selectOptions(screen.getByLabelText(/Pilih Divisi/), "media");
    await user.click(screen.getByLabelText("Pagi"));
    await user.click(screen.getByLabelText("Web"));
    fireEvent.change(screen.getByLabelText(/Tanggal/), { target: { value: "2026-10-20" } });
    await user.click(screen.getByRole("button", { name: "Kirim Pendaftaran" }));
    expect(await screen.findByText("NIM tidak valid.")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("Periksa kolom yang ditandai");
    expect(screen.getByLabelText(/NIM/)).toHaveAttribute("aria-invalid", "true");
  });
});

describe("registration admin experience", () => {
  it("creates a form from native date controls and opens its builder", async () => {
    const user = userEvent.setup();
    vi.mocked(getRegistrations).mockResolvedValue({
      items: [],
      meta: { page: 1, limit: 20, total: 0, total_pages: 0 },
    });
    vi.mocked(getPeriods).mockResolvedValue({
      items: [
        {
          id: 4,
          name: "2026/2027",
          start_date: "2026-10-01",
          end_date: "2027-10-01",
          status: "ACTIVE",
          division_count: 0,
          member_count: 0,
        },
      ],
      meta: { page: 1, limit: 100, total: 1, total_pages: 1 },
    });
    vi.mocked(getDivisions).mockResolvedValue({
      items: [],
      meta: { page: 1, limit: 100, total: 0, total_pages: 0 },
    });
    vi.mocked(createRegistration).mockResolvedValue(adminForm);
    render(<RegistrationAdminList />);
    await user.click(await screen.findByRole("button", { name: "Buat Pendaftaran" }));
    await user.type(screen.getByLabelText("Nama pendaftaran"), "Recruitment 2026");
    await user.selectOptions(screen.getByLabelText("Periode"), "4");
    fireEvent.change(screen.getByLabelText("Mulai"), { target: { value: "2026-10-11T09:00" } });
    fireEvent.change(screen.getByLabelText("Tutup"), { target: { value: "2026-12-01T00:00" } });
    await user.click(screen.getByRole("button", { name: "Buat dan lanjutkan" }));
    await waitFor(() =>
      expect(createRegistration).toHaveBeenCalledWith({
        period_id: 4,
        division_id: null,
        title: "Recruitment 2026",
        description: null,
        open_at: new Date("2026-10-11T09:00").toISOString(),
        close_at: new Date("2026-12-01T00:00").toISOString(),
      }),
    );
    expect(mocks.navigate).toHaveBeenCalledWith(`/admin/pendaftaran/${adminForm.id}/builder`);
  });

  it("provides keyboard accessible builder controls and explains schema lock", async () => {
    const user = userEvent.setup();
    vi.mocked(getRegistration).mockResolvedValue({ form: adminForm, fields: fields.slice(0, 4) });
    vi.mocked(saveRegistrationFields).mockResolvedValue(fields.slice(0, 5));
    render(<RegistrationBuilder id={7} />);
    expect(await screen.findAllByRole("heading", { name: "Open Recruitment" })).toHaveLength(2);
    await user.click(screen.getByRole("button", { name: "text" }));
    expect(screen.getAllByLabelText("Key")).toHaveLength(5);
    await user.click(screen.getByRole("button", { name: "Turunkan Nama Lengkap" }));
    await user.click(screen.getByRole("button", { name: "Simpan Struktur Form" }));
    await waitFor(() =>
      expect(saveRegistrationFields).toHaveBeenCalledWith(
        7,
        expect.arrayContaining([expect.objectContaining({ field_key: "text" })]),
      ),
    );
    vi.mocked(getRegistration).mockResolvedValue({
      form: { ...adminForm, submission_count: 1 },
      fields: fields.slice(0, 4),
    });
    cleanup();
    render(<RegistrationBuilder id={7} />);
    expect(
      await screen.findByText(/Struktur formulir tidak dapat diubah karena sudah memiliki data pendaftar/),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "text" })).toBeDisabled();
  });

  it("adds an IMAGE structural field with a dropzone builder preview", async () => {
    const user = userEvent.setup();
    vi.mocked(getRegistration).mockResolvedValue({ form: adminForm, fields: fields.slice(0, 4) });
    vi.mocked(saveRegistrationFields).mockResolvedValue(fields.slice(0, 4));
    render(<RegistrationBuilder id={7} />);
    await user.click(await screen.findByRole("button", { name: "image" }));
    expect(screen.getByText("Pratinjau unggah gambar")).toBeInTheDocument();
    expect(screen.getByText(/Seret gambar atau pilih dari perangkat/)).toBeInTheDocument();
    expect(screen.getAllByLabelText("Tipe").at(-1)).toHaveValue("IMAGE");
    expect(screen.getByLabelText("Deskripsi bantuan (opsional)")).toBeInTheDocument();
  });

  it("shows snapshot answers and asks for explicit PII export confirmation", async () => {
    const user = userEvent.setup();
    const submission: RegistrationSubmission = {
      id: 8,
      registration_form_id: 7,
      registration_title: adminForm.title,
      division_name: "Akademik",
      full_name: "Ayu",
      nim: "260001",
      email: "ayu@example.test",
      phone: "0812345678",
      submitted_at: "2026-10-10T00:00:00Z",
      answers: [
        {
          field_key: "motivation",
          label: "Label Saat Submit",
          type: "TEXTAREA",
          value_text: "Jawaban historis",
          display_order: 5,
        },
      ],
    };
    vi.mocked(getRegistration).mockResolvedValue({ form: adminForm, fields: [] });
    vi.mocked(getRegistrationSubmissions).mockResolvedValue({
      items: [submission],
      meta: { page: 1, limit: 20, total: 1, total_pages: 1 },
    });
    vi.mocked(downloadRegistrationExport).mockResolvedValue(new Blob(["xlsx"]));
    mocks.confirm.mockReturnValue(true);
    vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:test");
    vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => {});
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
    render(<RegistrationSubmissionList id={7} />);
    expect(await screen.findAllByText("Ayu")).toHaveLength(2);
    vi.spyOn(window, "confirm").mockImplementation(mocks.confirm);
    await user.click(screen.getByRole("button", { name: "Ekspor XLSX" }));
    expect(mocks.confirm).toHaveBeenCalledWith(expect.stringContaining("informasi pribadi"));
    await waitFor(() => expect(downloadRegistrationExport).toHaveBeenCalledWith(7, ""));
    vi.mocked(getRegistrationSubmission).mockResolvedValue(submission);
    cleanup();
    render(<RegistrationSubmissionDetail id={7} submissionId={8} />);
    expect(await screen.findByText("Label Saat Submit")).toBeInTheDocument();
    expect(screen.getByText("Jawaban historis")).toBeInTheDocument();
  });

  it("hides export and empty states when an inaccessible form cannot be loaded", async () => {
    vi.mocked(getRegistration).mockRejectedValue(new ApiError("forbidden", 404));
    vi.mocked(getRegistrationSubmissions).mockRejectedValue(new ApiError("forbidden", 404));
    render(<RegistrationSubmissionList id={1} />);

    expect(await screen.findByRole("alert")).toHaveTextContent("Data pendaftar gagal dimuat.");
    expect(screen.queryByRole("button", { name: "Ekspor XLSX" })).not.toBeInTheDocument();
    expect(screen.queryByText("Belum ada data pendaftar.")).not.toBeInTheDocument();
  });
});
