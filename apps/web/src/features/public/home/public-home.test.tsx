import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { getLandingPageContent } from "./data/home-content";
import { PublicHome } from "./public-home";

describe("public home foundation", () => {
  afterEach(() => {
    cleanup();
    document.body.style.pointerEvents = "";
  });

  it("renders the requested landing sections with one meaningful page heading", () => {
    render(<PublicHome content={getLandingPageContent(false)} />);

    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(screen.getByRole("heading", { name: /Sistem Informasi & Wadah Sinergi/ })).toBeInTheDocument();
    const eyebrow = screen.getByText("Portal publik BEM FMIKOM UNUGHA");
    expect(eyebrow.tagName).toBe("P");
    expect(eyebrow.className).not.toMatch(/rounded-full|border|bg-/);
    for (const heading of [
      "Ruang informasi dan kolaborasi mahasiswa",
      "Visi & Misi",
      "Struktur Organisasi",
      "Berita & Aktivitas Mahasiswa",
      "Pengumuman",
      "Galeri Kegiatan Fakultas",
      "Pendaftaran",
      "Aspirasi Anda membangun fakultas lebih baik",
      "Kontak & Media Sosial",
    ]) {
      expect(screen.getByRole("heading", { name: heading })).toBeInTheDocument();
    }
  });

  it("uses working section anchors for hero calls to action", () => {
    render(<PublicHome content={getLandingPageContent(false)} />);

    expect(screen.getByRole("link", { name: "Jelajahi informasi" })).toHaveAttribute("href", "#berita");
    expect(screen.getByRole("link", { name: "Kirim aspirasi" })).toHaveAttribute("href", "#aspirasi");
    expect(document.getElementById("berita")).toBeInTheDocument();
    expect(document.getElementById("aspirasi")).toBeInTheDocument();
  });

  it("keeps the mobile hero artwork and caption fluid for narrow screens", () => {
    render(<PublicHome content={getLandingPageContent(true)} />);

    const artwork = screen.getByTestId("public-hero-art");
    const caption = screen.getByTestId("public-hero-caption");
    expect(artwork.className).toContain("w-full");
    expect(artwork.className).toContain("min-w-0");
    expect(artwork.className).toContain("aspect-[4/3]");
    expect(caption.className).toContain("min-w-0");
    expect(screen.getByRole("heading", { level: 1 }).className).toContain("break-words");
  });

  it("shows safe empty states and no demo statistics when preview data is unavailable", () => {
    render(<PublicHome content={getLandingPageContent(false)} />);

    expect(screen.getByText("Data struktur belum tersedia")).toBeInTheDocument();
    expect(screen.getByText("Belum ada berita untuk ditampilkan")).toBeInTheDocument();
    expect(screen.getByText("Belum ada pengumuman aktif")).toBeInTheDocument();
    expect(screen.getByText("Dokumentasi belum tersedia")).toBeInTheDocument();
    expect(screen.getByText("Belum ada pendaftaran aktif")).toBeInTheDocument();
    expect(screen.getByText("Kontak resmi belum tersedia")).toBeInTheDocument();
    expect(screen.queryByText("450+")).not.toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("links the organization preview to its canonical detail route without an accordion", () => {
    render(<PublicHome content={getLandingPageContent(true)} />);
    expect(screen.getByRole("link", { name: "Lihat struktur" })).toHaveAttribute("href", "/struktur");
    expect(screen.queryByRole("button", { name: /struktur/i })).not.toBeInTheDocument();
    expect(screen.queryByText("Sembunyikan struktur")).not.toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Divisi" })).not.toBeInTheDocument();
  });

  it("keeps the canonical structure link and a neutral empty preview without an active period", () => {
    render(<PublicHome content={getLandingPageContent(false)} />);
    expect(screen.getByText("Data struktur belum tersedia")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Lihat struktur" })).toHaveAttribute("href", "/struktur");
  });

  it("renders preview stories as semantic links to the landing section", () => {
    render(<PublicHome content={getLandingPageContent(true)} />);

    const storyLink = screen.getByRole("link", {
      name: "Baca Contoh judul berita untuk tinjauan tampilan portal",
    });
    expect(storyLink.tagName).toBe("A");
    expect(storyLink).toHaveAttribute("href", "#berita");
    expect(storyLink.closest("article")).toBeInTheDocument();
  });

  it("opens and closes the announcement dialog accessibly", async () => {
    render(<PublicHome content={getLandingPageContent(true)} />);

    fireEvent.click(screen.getByRole("button", { name: /Baca pengumuman/ }));
    expect(
      await screen.findByRole("dialog", { name: "Contoh pengumuman untuk tinjauan tampilan" }),
    ).toBeVisible();
    fireEvent.keyDown(document, { key: "Escape" });
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("fits portrait announcement imagery inside a bounded responsive modal", async () => {
    const content = getLandingPageContent(true);
    content.announcements = [
      { ...content.announcements[0], image: { src: "https://example.test/portrait.png", alt: "Portrait" } },
    ];
    render(<PublicHome content={content} />);
    fireEvent.click(screen.getByRole("button", { name: /Baca pengumuman/ }));
    const dialog = await screen.findByRole("dialog", { name: content.announcements[0].title });
    expect(dialog.className).toContain("max-h-[calc(100dvh-2rem)]");
    expect(screen.getByRole("img", { name: "Portrait" }).className).toContain("object-contain");
    expect(screen.getByRole("img", { name: "Portrait" }).className).not.toContain("object-cover");
  });

  it("keeps long content, multiple mission items, and an announcement without an image renderable", async () => {
    const content = getLandingPageContent(true);
    const longHeadline = "Headline contoh yang panjang untuk menguji pemenggalan dan lebar konten portal "
      .repeat(3)
      .trim();
    content.hero = { ...content.hero, title: longHeadline };
    content.news = content.news.map((item, index) => (index === 0 ? { ...item, title: longHeadline } : item));
    content.announcements = [{ ...content.announcements[0], title: longHeadline, image: undefined }];
    content.organization = {
      ...content.organization,
      leadership: content.organization.leadership.map((member, index) =>
        index === 0
          ? { ...member, division: "Divisi dengan nama yang sangat panjang untuk menguji layout" }
          : member,
      ),
    };
    content.visionMission = {
      ...content.visionMission,
      mission: [
        ...content.visionMission.mission,
        "Misi tambahan untuk memastikan susunan list tetap bertambah dengan aman.",
      ],
    };
    render(<PublicHome content={content} />);

    expect(screen.getByRole("heading", { level: 1, name: longHeadline })).toBeInTheDocument();
    expect(screen.getAllByRole("heading", { level: 3, name: longHeadline })).toHaveLength(2);
    expect(
      screen.getByText("Misi tambahan untuk memastikan susunan list tetap bertambah dengan aman."),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Baca pengumuman/ }));
    expect(await screen.findByRole("dialog", { name: longHeadline })).toBeVisible();
    expect(screen.queryByRole("img", { name: /contoh visual/ })).not.toBeInTheDocument();
  });
});
