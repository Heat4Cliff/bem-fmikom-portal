# Stitch Screen Index

Base path: `design_reference/stitch_bem_fmikom_unugha_portal/<folder>/`

| Surface | Suggested route | Stitch folder | Purpose |
|---|---|---|---|
| Public | `/` | `beranda_publik_bem_fmikom_unugha` | Desktop/public landing |
| Public | `/ (mobile)` | `beranda_publik_mobile_bem_fmikom_unugha` | Mobile public landing |
| Public | `/berita` | `halaman_daftar_berita_informasi_bem_fmikom` | News list |
| Public | `/berita/[slug]` | `detail_berita_workshop_ui_ux_next.js` | News detail |
| Public | `/pengumuman` | `halaman_pengumuman_resmi_bem_fmikom` | Announcement list |
| Public | `/pengumuman/[slug]` | `detail_pengumuman_krs_beasiswa_aspirasi` | Announcement detail |
| Public | `/galeri` | `halaman_galeri_kegiatan_bem_fmikom` | Gallery list |
| Public | `/galeri/[id-or-slug]` | `detail_galeri_fmikom_tech_fest_2024` | Gallery detail |
| Public | `/pendaftaran` | `halaman_daftar_pendaftaran_bem_fmikom` | Registration list |
| Public | `/pendaftaran/[slug]` | `detail_informasi_pendaftaran_oprec_fmikom_tech_fest` | Registration detail |
| Public | `/pendaftaran/[slug]/form` | `formulir_pendaftaran_online_oprec_tech_fest` | Public dynamic form desktop |
| Public | `/pendaftaran/[slug]/form (mobile)` | `formulir_pendaftaran_mobile_bem_fmikom` | Public dynamic form mobile |
| Public | `/aspirasi` | `layanan_aspirasi_mahasiswa_fmikom` | Public aspiration form |
| Admin | `/admin` | `dashboard_overview_admin_bem` | Dashboard overview |
| Admin | `/admin (mobile)` | `dashboard_admin_mobile_bem_fmikom_unugha` | Dashboard mobile |
| Admin | `/admin/berita` | `manajemen_berita_artikel` | Article management |
| Admin | `/admin/berita (mobile)` | `manajemen_berita_mobile_admin_bem` | Article management mobile |
| Admin | `/admin/berita/new or /[id]/edit` | `article_editor_tulis_edit_berita` | Article editor |
| Admin | `/admin/pendaftaran/[id]/submissions` | `data_pendaftar_submisi_admin_bem` | Registration submissions |
| Admin | `/admin/pendaftaran/[id]/builder` | `registration_form_builder_admin_bem` | Form builder |
| Admin | `/admin/pendaftaran (mobile)` | `manajemen_pendaftaran_mobile_admin_bem` | Registration management mobile |
| Admin | `/admin/aspirasi` | `manajemen_aspirasi_mahasiswa_admin_bem` | Aspirations management |
| Admin | `/admin/aspirasi (mobile)` | `manajemen_aspirasi_mobile_admin_bem` | Aspirations mobile |
| Admin | `/admin/aspirasi/[id]` | `detail_aspirasi_mahasiswa_admin_bem` | Aspiration detail |
| Admin | `/admin/divisi` | `manajemen_divisi_admin_bem` | Division management |
| Admin | `/admin/pengurus` | `data_pengurus_admin_bem` | Member management |
| Admin | `/admin/pengurus/new or /[id]/edit` | `tambah_edit_pengurus_admin_bem` | Member form |
| Admin | `/admin/struktur` | `struktur_organisasi_pengurus_admin_bem` | Organization structure |
| Admin | `/admin/periode` | `periode_kepengurusan_admin_bem` | Period management |
| Admin | `/admin/galeri` | `manajemen_galeri_admin_bem` | Gallery management |
| Admin | `/admin/pengumuman` | `manajemen_pengumuman_admin_bem` | Announcement management |
| Admin | `/admin/dokumen` | `arsip_dokumen_internal_admin_bem` | Document archive |
| Admin | `/admin/dokumen (mobile)` | `arsip_dokumen_mobile_admin_bem` | Document archive mobile |
| Admin | `/admin/settings` | `pengaturan_landing_page_admin_bem` | Landing settings |
| Admin | `/admin/users` | `manajemen_pengguna_admin_bem` | User management |
| Admin | `/admin/activity-logs` | `activity_log_audit_trail_admin_bem` | Activity log |
| Admin | `/admin/activity-logs (mobile)` | `activity_log_mobile_admin_bem` | Activity log mobile |

Total mapped screen references: **37**.

Each folder contains `screen.png` and `code.html`. Codex must visually inspect `screen.png` before implementing the corresponding route.
