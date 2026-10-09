import type { LandingPageContent } from "../types";

const defaultContent: LandingPageContent = {
  demoMode: false,
  hero: {
    eyebrow: "Portal publik BEM FMIKOM UNUGHA",
    title: "Sistem Informasi & Wadah Sinergi Mahasiswa FMIKOM UNUGHA",
    description:
      "Temukan informasi organisasi, kegiatan, pengumuman, dan layanan BEM FMIKOM UNUGHA melalui satu portal.",
    primaryAction: { label: "Jelajahi informasi", href: "#berita" },
    secondaryAction: { label: "Kirim aspirasi", href: "#aspirasi" },
    visualLabel: "Visual kegiatan BEM FMIKOM UNUGHA",
    visualCaption: "Dokumentasi kegiatan akan ditampilkan setelah aset resmi tersedia.",
  },
  about: {
    title: "Ruang informasi dan kolaborasi mahasiswa",
    description:
      "Portal ini menjadi kanal publik untuk mengenal BEM FMIKOM UNUGHA dan menemukan informasi organisasi. Konten resmi dapat diperbarui ketika pengelolaan konten tersedia.",
    highlights: [
      {
        title: "Informasi terpusat",
        description: "Berita, pengumuman, dan dokumentasi organisasi tersedia dalam satu tempat.",
      },
      {
        title: "Terbuka untuk publik",
        description: "Informasi publik dapat dibaca tanpa membuat akun.",
      },
      {
        title: "Ruang partisipasi",
        description: "Portal disiapkan untuk menghubungkan mahasiswa dengan kegiatan organisasi.",
      },
    ],
  },
  visionMission: { vision: null, mission: [] },
  statistics: [],
  organization: { leadership: [], divisions: [], members: [] },
  news: [],
  announcements: [],
  gallery: [],
  registrations: [],
  contact: { socialLinks: [] },
};

const visualReviewContent: LandingPageContent = {
  ...defaultContent,
  demoMode: true,
  visionMission: {
    vision: "Menjadi ruang kolaborasi mahasiswa yang terbuka, bertanggung jawab, dan bertumbuh bersama.",
    mission: [
      "Menghubungkan gagasan mahasiswa dengan ruang kolaborasi yang inklusif.",
      "Menyampaikan informasi organisasi secara jelas dan mudah diakses.",
      "Mendukung kegiatan mahasiswa melalui komunikasi yang terbuka.",
      "Menjaga dokumentasi kegiatan agar mudah ditemukan kembali.",
    ],
  },
  statistics: [
    { label: "Divisi", value: "5" },
    { label: "Program kerja", value: "24" },
    { label: "Transparansi anggaran", value: "100%" },
    { label: "Mahasiswa terlayani", value: "450+" },
  ],
  organization: {
    periodLabel: "Contoh visual periode 2024/2025",
    leadership: [
      { id: "ketua-umum", role: "Ketua", name: "Nama contoh", division: "BPH" },
      { id: "wakil-ketua", role: "Wakil Ketua", name: "Nama contoh", division: "BPH" },
    ],
    divisions: [
      {
        id: 1,
        name: "Contoh divisi",
        head: { id: "kepala-bidang", role: "Kepala Bidang", name: "Nama contoh" },
        members: [{ id: "anggota", role: "Anggota", name: "Nama contoh" }],
      },
    ],
    members: [
      { id: "ketua-umum", role: "Ketua Umum", name: "Nama contoh", division: "BPH" },
      { id: "wakil-ketua", role: "Wakil Ketua", name: "Nama contoh", division: "BPH" },
      { id: "sekretaris", role: "Sekretaris", name: "Nama contoh", division: "Sekretariat" },
      {
        id: "koordinator-divisi",
        role: "Koordinator Divisi",
        name: "Nama contoh",
        division: "Contoh divisi",
      },
    ],
  },
  news: [
    {
      id: "visual-berita-akademik",
      category: "Akademik",
      date: "9 Oktober 2026",
      title: "Contoh judul berita untuk tinjauan tampilan portal",
      excerpt: "Konten contoh ini hanya dipakai untuk menilai hierarki judul, ringkasan, dan kartu berita.",
      visualLabel: "Ilustrasi pratinjau berita akademik",
      href: "#berita",
    },
    {
      id: "visual-berita-kegiatan",
      category: "Kegiatan",
      date: "7 Oktober 2026",
      title: "Contoh dokumentasi kegiatan mahasiswa FMIKOM",
      excerpt: "Teks contoh membantu memeriksa perilaku kartu ketika judul dan ringkasan berubah.",
      visualLabel: "Ilustrasi pratinjau kegiatan mahasiswa",
      href: "#berita",
    },
    {
      id: "visual-berita-organisasi",
      category: "Organisasi",
      date: "2 Oktober 2026",
      title: "Contoh informasi organisasi dan program kerja",
      excerpt: "Data ini bukan berita resmi dan tidak berasal dari layanan backend.",
      visualLabel: "Ilustrasi pratinjau informasi organisasi",
      href: "#berita",
    },
  ],
  announcements: [
    {
      id: "visual-pengumuman",
      title: "Contoh pengumuman untuk tinjauan tampilan",
      content:
        "Ini adalah teks contoh visual. Jadwal, status, dan isi pengumuman resmi akan berasal dari API pada phase konten.",
      dateLabel: "Contoh visual · 9 Oktober 2026",
      href: "#pengumuman",
    },
  ],
  gallery: [
    {
      id: "visual-galeri-1",
      title: "Contoh dokumentasi kegiatan",
      dateLabel: "Contoh visual",
      division: "Akademik",
      visualLabel: "Ilustrasi dokumentasi satu",
      tone: "navy",
    },
    {
      id: "visual-galeri-2",
      title: "Contoh kolaborasi mahasiswa",
      dateLabel: "Contoh visual",
      division: "BPH",
      visualLabel: "Ilustrasi dokumentasi dua",
      tone: "blue",
    },
    {
      id: "visual-galeri-3",
      title: "Contoh forum diskusi",
      dateLabel: "Contoh visual",
      division: "Contoh divisi",
      visualLabel: "Ilustrasi dokumentasi tiga",
      tone: "amber",
    },
    {
      id: "visual-galeri-4",
      title: "Contoh agenda fakultas",
      dateLabel: "Contoh visual",
      division: "FMIKOM",
      visualLabel: "Ilustrasi dokumentasi empat",
      tone: "slate",
    },
  ],
  registrations: [
    {
      id: "visual-pendaftaran",
      title: "Contoh informasi pendaftaran kegiatan",
      division: "Contoh divisi",
      openDate: "Contoh tanggal buka",
      closeDate: "Contoh tanggal tutup",
      status: "CONTOH VISUAL",
    },
  ],
  contact: {
    address: "Alamat sekretariat resmi belum dikonfirmasi.",
    email: "Alamat email resmi belum dikonfirmasi.",
    socialLinks: [],
  },
};

export function getLandingPageContent(enableDemoContent: boolean): LandingPageContent {
  return enableDemoContent ? visualReviewContent : defaultContent;
}

export function isDemoContentEnabled() {
  return process.env.BEM_DEMO_CONTENT === "true";
}
