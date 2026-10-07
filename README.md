# BEM FMIKOM UNUGHA Portal

Repository terpisah untuk pembangunan portal publik dan dashboard admin BEM FMIKOM UNUGHA. Repository ini memulai **Phase 0 — Source Audit** saja.

## Status

- Phase 0 report: `docs/codex/SOURCE_AUDIT.md`
- Phase 0 status: **PASS**
- Phase 1 has not started. No application scaffold, feature, or migration was created.

## Source material

`source_docs/` dan `design_reference/` disalin apa adanya dari folder persiapan agar repository ini mandiri. Keduanya adalah read-only evidence; jangan edit atau rapikan file sumber. PRD v1.4 mengatur behavior dan scope. Stitch `screen.png` adalah visual target; `code.html` hanya referensi struktur/style, bukan production architecture. Token ada di `design_reference/stitch_bem_fmikom_unugha_portal/modern_academic_horizon/DESIGN.md`.

Dokumen baseline produk, API, keamanan, desain, workflow, `AGENTS.md`, dan `PLANS.md` juga tersedia pada salinan folder ini. Checklist Phase 1+ tetap sama dengan sumber.

## Phase 0 deliverables

- `docs/codex/SOURCE_AUDIT.md`
- `docs/codex/DECISIONS.md`
- `docs/codex/GAPS_AND_QUESTIONS.md`
- `docs/codex/IMPLEMENTATION_MATRIX.md`

Belum ada aplikasi yang bisa dijalankan. Lihat report untuk audit API, role/scope, 37 Stitch screens, responsive coverage, keamanan, risiko dan rekomendasi.

## Stack baseline untuk fase berikutnya

- Frontend: Next.js App Router, TypeScript strict, Tailwind CSS
- Backend: Go, Gin, REST
- Database: MySQL/MariaDB
- Authentication: JWT access token + Secure HttpOnly refresh-token cookie

Ikuti `AGENTS.md` dan jalankan satu phase pada satu waktu. Tunggu instruksi sebelum memulai Phase 1.
