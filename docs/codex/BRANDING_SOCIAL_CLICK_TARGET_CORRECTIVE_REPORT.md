# Branding, Social Media, dan Click Target — Corrective Report

**Tanggal:** 2026-10-15
**Status:** READY FOR OWNER BRANDING & NAVIGATION RE-REVIEW

## Ringkasan perubahan

### Navigasi Hero, Tentang, Visi & Misi

Root cause berada pada stretched-link berita: pseudo-element link absolut dapat menangkap klik di luar kartu ketika kartu tidak membentuk containing/stacking context. Kartu berita kini memakai `position: relative` dan `isolation`, sehingga area link tertahan di dalam kartu. Audit source hanya menemukan implementasi stretched-link tersebut pada kartu berita.

Hero, gambar, judul, dan whitespace bukan link. CTA Hero tetap menjadi satu-satunya target navigasi Hero. Area Tentang dan Visi & Misi juga bukan link; hanya CTA **Lihat selengkapnya** yang menuju `/tentang`. Kartu/CTA berita tetap menuju detail berita. Regression test memeriksa elemen non-CTA tidak dibungkus link, CTA memakai href yang dimaksud, dan link berita tetap berada di batas kartu.

### Social media

Settings mendukung daftar terurut `{ platform, label, url, order, active }`. Admin memilih platform, mengisi label dan URL HTTPS, mengaktifkan/nonaktifkan item, serta melihat preview icon. Ikon Instagram, TikTok, YouTube, Facebook, X, WhatsApp, dan LinkedIn memakai paket icon brand lokal; platform yang belum dikenal memakai ikon tautan generik. Public UI menampilkan icon dan label aksesibel, hanya merender item aktif, memakai link eksternal aman (`noopener noreferrer`), dan mengikuti token theme portal.

### Global brand logo

`navbar_logo_url` yang sudah ada menjadi sumber Global Brand Logo. Provider settings dipasang di layout admin agar logo tidak diminta ulang pada tiap halaman. Komponen `BrandLogo` reusable menampilkan logo pada public navbar, admin sidebar/mobile header, dan admin login, dengan fallback BEM lokal ketika logo kosong atau media gagal dimuat. Logo kabinet hanya digunakan pada `/tentang`.

Pengaturan bentuk `original`, `rounded`, `circle` dan fit `contain`/`cover` bersifat presentasional; file asli tidak diubah. Footer secara default mengikuti Global Brand Logo dan dapat memakai upload override dengan toggle. Perubahan bentuk/fallback mempertahankan ukuran container, aspect ratio, dan alt text.

## API, database, media, dan keamanan

- Migration `00013_branding_social_controls.sql` menambah kolom nullable JSON untuk social links dan field shape/fit serta footer toggle dengan default kompatibel. Existing footer asset dipertahankan sebagai override ketika migrasi dijalankan.
- Migration `00012_landing_brand_identity.sql` dan `00013_branding_social_controls.sql` ikut diverifikasi dalam urutan migrasi yang ada. Integration suite menjalankan migrate up, down to zero, lalu up kembali; hasil akhirnya deterministik dan schema test dikembalikan ke versi 0 oleh cleanup suite.
- Integrasi mencakup upload dan penggantian asset Global Brand Logo, persistence settings, normalisasi URL media, validasi URL sosial HTTPS, serta penolakan bentuk logo tidak valid.
- Authorization settings tetap mengikuti role yang sudah diizinkan. Test integrasi RBAC/division scope lulus. Validasi upload/media, isolasi storage, CORS, dan policy autentikasi tidak dilonggarkan.
- API contract diperbarui untuk field branding dan social links.

## Tema, responsive, dan aksesibilitas

Komponen memakai theme tokens aktif dan tidak mengubah warna pixel logo yang diunggah. Fallback berupa tile netral bertuliskan BEM menjaga keterbacaan tanpa asset. Social links tetap menampilkan label, dapat difokuskan keyboard, dan memakai focus-visible styling; shape selection memakai radio controls.

Frontend lint, unit test, typecheck, dan build mencakup perubahan komponen/layout. Smoke HTTP mengembalikan 200 untuk beranda, login, API health, dan public settings pada runtime lokal. Pemeriksaan visual manual terperinci pada ukuran 360×800, 390×844, 412×915, 768+, 1280×800, dan 1440×900 **belum dilakukan dalam run ini**. Owner branding/navigation review tetap diperlukan; hasil automated check tidak dianggap sebagai penerimaan visual manual.

## Verifikasi aktual

| Gate | Hasil |
|---|---|
| `pnpm api:integration` | PASS — `TestDatabaseIntegration` RUN terhadap MariaDB test; migration 00001–00013 dan subtest RBAC/division scope lulus. |
| `pnpm verify` | PASS — lint, format, typecheck, 27 file / 122 test frontend, Next production build, Go format (71 file), vet, Go tests, dan Go build. |
| `pnpm audit` | PASS — no known vulnerabilities found. |
| `git diff --check` | PASS. Git hanya melaporkan peringatan line-ending LF→CRLF untuk file yang berubah. |
| `source_docs/`, `design_reference/` | Tidak berubah. |
| Local HTTP smoke | PASS — HTTP 200: `127.0.0.1:8081/healthz`, `/api/v1/public/settings`, `127.0.0.1:3290/`, dan `/admin/login`. |

API lokal berjalan pada port **8081** karena port 8080 bentrok dengan listener localhost lain. Web berjalan pada **http://127.0.0.1:3290**; override port dan API URL hanya berlaku pada proses runtime ini, tidak mengubah konfigurasi project.

## Temuan dan risiko tersisa

1. Manual owner review untuk upload/simpan/reload logo pada seluruh surface, footer override, ikon/social link, click target aktual, dan keenam viewport masih menunggu.
2. Browser runtime dapat smoke-tested, tetapi smoke HTTP bukan bukti visual atau acceptance.
3. Jangan anggap corrective pass ini sebagai persetujuan final branding atau release. ZIP, GitHub Release, tag, deployment, commit, dan push tidak diubah/dijalankan.

## Permintaan review owner

Silakan tinjau Global Brand Logo (termasuk shape/fit), footer inherit/override, kabinet pada `/tentang`, daftar social media, dan navigasi Hero/Tentang/Visi & Misi. Setelah owner menyetujui tampilan dan target klik, hasil dapat diteruskan sesuai instruksi berikutnya.

**READY FOR OWNER BRANDING & NAVIGATION RE-REVIEW**
