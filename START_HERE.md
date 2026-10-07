# START HERE — BEM FMIKOM UNUGHA Portal

Paket ini menyiapkan repository agar dapat dikerjakan Codex dengan pola kerja bertahap seperti project LabFlow: source-of-truth jelas, phase plan, quality gates, visual review, dan laporan verifikasi sebelum berpindah fase.

## Urutan baca wajib
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

## Source of truth priority
Jika ada konflik, gunakan prioritas berikut:
1. PRD v1.4 = kebutuhan produk dan business rules.
2. API Contract + Security Contract = kontrak backend/frontend dan authorization.
3. Stitch `screen.png` = target visual.
4. Stitch `code.html` = referensi struktur/spacing/style, **bukan production code yang harus dicopy mentah**.
5. `DESIGN_SYSTEM.md` = tokens dan pola komponen.
6. Implementasi mengikuti source di atas; jangan mengarang fitur baru.

## Cara mulai
Gunakan prompt pada `CODEX_MASTER_PROMPT.md`. Prompt pertama hanya menjalankan **Phase 0 — Repository & Source Audit**. Jangan langsung membuat seluruh aplikasi.
