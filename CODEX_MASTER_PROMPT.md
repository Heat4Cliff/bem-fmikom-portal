# CODEX MASTER PROMPT — First Run

Gunakan folder persiapan ini sebagai source material untuk membangun project **BEM FMIKOM UNUGHA Portal**.

Buat project pada folder TERPISAH bernama:

`bem-fmikom-portal`

Jangan mengubah, menimpa, atau merapikan folder source reference. Anggap `source_docs/` dan `design_reference/` sebagai read-only evidence.

Sebelum melakukan coding, baca dengan urutan:

1. `AGENTS.md`
2. `MANIFEST.md`
3. `docs/PRD_SUMMARY.md`
4. `docs/ARCHITECTURE.md`
5. `docs/backend/API_CONTRACT.md`
6. `docs/backend/SECURITY_CONTRACT.md`
7. `docs/design/DESIGN_SYSTEM.md`
8. `docs/design/SCREEN_INDEX.md`
9. `docs/design/ROUTE_SCREEN_MATRIX.md`
10. `PLANS.md`
11. `docs/codex/IMPLEMENTATION_PROTOCOL.md`
12. `docs/codex/VISUAL_REVIEW_CHECKLIST.md`

Source-of-truth rules:
- PRD v1.4 menentukan behavior dan scope.
- API/Security Contract menentukan request/response dan authorization.
- Stitch `screen.png` menentukan visual target.
- Stitch `code.html` hanya referensi layout/style; jangan copy sebagai production architecture.
- Jika ada konflik, laporkan. Jangan menebak diam-diam.

Stack baseline:
- Frontend: Next.js App Router + TypeScript strict + Tailwind CSS.
- Backend: Go + Gin + REST API.
- Database: MySQL/MariaDB.
- Auth: JWT access token + refresh token Secure HttpOnly cookie.
- Deploy target: frontend Vercel; backend/database pada environment yang mendukung Go + MySQL/MariaDB.

Untuk RUN PERTAMA ini, kerjakan **PHASE 0 SAJA** dari `PLANS.md`.

Output Phase 0 yang wajib dibuat di project baru:
- `docs/codex/SOURCE_AUDIT.md`
- `docs/codex/DECISIONS.md`
- `docs/codex/GAPS_AND_QUESTIONS.md`
- `docs/codex/IMPLEMENTATION_MATRIX.md`
- `README.md` awal yang menjelaskan struktur target dan cara menjalankan project setelah scaffold nanti.

`IMPLEMENTATION_MATRIX.md` minimal memetakan:
`Requirement/User Story -> API Endpoint -> DB/Domain -> UI Route/Screen -> Role -> Test type`.

Pada Phase 0:
- Jangan scaffold aplikasi dulu kecuali diperlukan untuk memeriksa environment.
- Jangan implement feature.
- Jangan membuat database migration.
- Jangan mulai Phase 1.
- Jangan mengubah checklist Phase 1+.

Di akhir, berikan ringkasan:
1. sumber yang berhasil dibaca,
2. jumlah screen Stitch yang ditemukan,
3. keputusan teknis yang sudah fixed vs belum fixed,
4. gap/blocker nyata,
5. rekomendasi langkah Phase 1,
6. status `PHASE 0: PASS` atau `PHASE 0: BLOCKED` beserta alasannya.

Jika tidak ada blocker kritis, STOP dan tunggu instruksi saya untuk melanjutkan Phase 1.
