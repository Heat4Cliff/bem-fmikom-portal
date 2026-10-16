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
    expect(screen.getByRole("link", { name: /Lihat selengkapnya/ })).toHaveAttribute("href", "/tentang");
    expect(screen.queryByRole("heading", { name: "Visi & Misi", level: 2 })).not.toBeInTheDocument();
    expect(screen.getByText("Arah gerak")).toBeInTheDocument();
    expect(screen.queryByRole("img", { name: /logo kabinet/i })).not.toBeInTheDocument();
  });

  it("uses working section anchors for hero calls to action", () => {
    render(<PublicHome content={getLandingPageContent(false)} />);

    expect(screen.getByRole("link", { name: "Jelajahi informasi" })).toHaveAttribute("href", "#berita");
    expect(screen.getByRole("link", { name: "Kirim aspirasi" })).toHaveAttribute("href", "#aspirasi");
    expect(document.getElementById("berita")).toBeInTheDocument();
    expect(document.getElementById("aspirasi")).toBeInTheDocument();
  });

  it("keeps non-CTA hero and about content outside every link while isolating news links", () => {
    const source = getLandingPageContent(true);
    const content = {
      ...source,
      news: source.news.map((item, index) => (index === 0 ? { ...item, href: "/berita/regression" } : item)),
    };
    render(<PublicHome content={content} />);

    const heroTitle = screen.getByRole("heading", { level: 1 });
    const heroArt = screen.getByTestId("public-hero-art");
    const heroSection = heroTitle.closest("section");
    expect(heroTitle.closest("a")).toBeNull();
    expect(heroArt.closest("a")).toBeNull();
    expect(heroSection).not.toBeNull();
    fireEvent.click(heroTitle);
    fireEvent.click(heroArt);
    if (heroSection) fireEvent.click(heroSection);
    expect(heroTitle.closest("a")).toBeNull();
    expect(screen.getByRole("link", { name: "Jelajahi informasi" })).toHaveAttribute("href", "#berita");

    const aboutSection = document.querySelector(".landing-about-section");
    const aboutTitle = screen.getByRole("heading", { name: "Ruang informasi dan kolaborasi mahasiswa" });
    const aboutDescription = aboutSection?.querySelector("p.line-clamp-4");
    const visionCard = aboutSection?.querySelector("article");
    expect(aboutSection).not.toBeNull();
    expect(aboutTitle.closest("a")).toBeNull();
    expect(aboutDescription?.closest("a")).toBeNull();
    expect(visionCard?.closest("a")).toBeNull();
    fireEvent.click(aboutTitle);
    if (aboutDescription) fireEvent.click(aboutDescription);
    if (visionCard) fireEvent.click(visionCard);
    if (aboutSection) fireEvent.click(aboutSection);
    expect(screen.getByRole("link", { name: /Lihat selengkapnya/ })).toHaveAttribute("href", "/tentang");

    const story = screen.getByRole("link", { name: "Contoh judul berita untuk tinjauan tampilan portal" });
    expect(story).toHaveAttribute("href", "/berita/regression");
    expect(story.closest("article")?.className).toContain("relative");
    expect(story.closest("article")?.className).toContain("isolate");
    expect(
      screen.getByRole("link", { name: "Baca Contoh judul berita untuk tinjauan tampilan portal" }),
    ).toHaveAttribute("href", "/berita/regression");
  });

  it("keeps the decorative hero frame fluid and independent from news content", () => {
    render(<PublicHome content={getLandingPageContent(true)} />);

    const artwork = screen.getByTestId("public-hero-art");
    expect(artwork.className).toContain("w-full");
    expect(artwork.className).toContain("min-w-0");
    expect(artwork.className).toContain("aspect-[4/3]");
    expect(
      screen.getByText("Dokumentasi kegiatan akan ditampilkan setelah aset resmi tersedia."),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("hero-news-carousel")).not.toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Berita terbaru" })).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1 }).className).toContain("break-words");
  });

  it("renders landing hero media as a documentation visual, not a news link", () => {
    const content = getLandingPageContent(false);
    content.hero.imageUrl = "/api/v1/public/media/42";
    content.hero.visualCaption = "Dokumentasi kegiatan dan aktivitas BEM FMIKOM UNUGHA.";
    render(<PublicHome content={content} />);

    const frame = screen.getByTestId("public-hero-art");
    const image = screen.getByRole("img", { name: content.hero.visualLabel });
    expect(frame.tagName).toBe("FIGURE");
    expect(frame).toHaveClass("aspect-[4/3]", "overflow-hidden", "rounded-2xl", "sm:aspect-[16/10]");
    expect(image).toHaveAttribute("src", expect.stringContaining("/api/v1/public/media/42"));
    expect(image).toHaveClass("object-cover", "object-center");
    expect(screen.getByText(content.hero.visualCaption)).toBeInTheDocument();
    expect(frame.querySelector("a")).toBeNull();
  });

  it("shows a themed documentation fallback when no hero image is configured", () => {
    const content = getLandingPageContent(false);
    content.hero.imageUrl = null;
    render(<PublicHome content={content} />);

    expect(screen.getAllByText("BEM FMIKOM · UNUGHA").length).toBeGreaterThan(0);
    expect(
      screen.getAllByText("Dokumentasi kegiatan akan ditampilkan setelah aset resmi tersedia.").length,
    ).toBeGreaterThan(0);
    expect(screen.getByTestId("public-hero-art").querySelector("a")).toBeNull();
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
      screen.queryByText("Misi tambahan untuk memastikan susunan list tetap bertambah dengan aman."),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Baca pengumuman/ }));
    expect(await screen.findByRole("dialog", { name: longHeadline })).toBeVisible();
    expect(screen.queryByRole("img", { name: /contoh visual/ })).not.toBeInTheDocument();
  });
});
