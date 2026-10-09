# PLANS.md — Execution Plan

> Rule: Codex may update checkboxes only after implementation **and** required verification for that item are complete. Execute one phase at a time unless the user explicitly asks otherwise.

## Phase 0 — Repository & Source Audit
- [ ] Create/confirm isolated project root `bem-fmikom-portal/`.
- [ ] Inventory source docs, Stitch screens, stack requirements, and environment assumptions.
- [ ] Produce `docs/codex/SOURCE_AUDIT.md`.
- [ ] Produce `docs/codex/DECISIONS.md` for choices not fixed by PRD.
- [ ] Produce `docs/codex/GAPS_AND_QUESTIONS.md` only for genuine blockers/ambiguities.
- [ ] Produce implementation matrix: requirement → API → route/screen → test.
- [ ] Do **not** start Phase 1 in the same run.

## Phase 1 — Repository Foundation & Design System
- [x] Scaffold monorepo structure `apps/web`, `apps/api`, `packages/contracts`, `scripts`, `infra`.
- [x] Initialize Next.js TypeScript strict + Tailwind.
- [x] Initialize Go module + Gin + config + structured logging.
- [x] Establish environment templates (`.env.example`) without secrets.
- [x] Implement design tokens from Stitch `DESIGN.md`.
- [x] Implement core reusable UI primitives and public/admin shells.
- [x] Add lint/format/typecheck/test scripts.
- [x] Add baseline CI workflow.

## Phase 2 — Database, Auth & Authorization Foundation
- [ ] Finalize migrations for users, periods, divisions, members, articles/categories, galleries, announcements, aspirations, registration forms/fields/submissions, documents, landing settings, activity logs.
- [ ] Seed deterministic local dev data for SUPER_ADMIN, BPH, KEPALA_BIDANG.
- [x] Implement login/refresh/logout/me.
- [x] Implement account-active checks, role policy, division scoping.
- [x] Security tests: 401, 403, cross-division denial at the authorization-policy layer (real CRUD IDOR routes remain out of Phase 2 scope).

## Phase 3 — Public Foundation & Landing
- [x] Public layout/nav/footer responsive.
- [ ] Landing settings/public home.
- [ ] Active-period organization structure.
- [x] SEO metadata/sitemap/OpenGraph baseline.
- [x] Visual review against desktop/mobile Stitch refs.

## Phase 4 — Content CMS
- [x] Article categories + articles CRUD/publish/unpublish.
- [x] Public article list/detail.
- [x] Gallery categories/items with Google Photos URL.
- [x] Public gallery list/detail.
- [x] Announcements CRUD/scheduling/active popup/public list/detail.
- [x] RBAC/division scope tests.

## Phase 5 — Organization & Period Management
- [ ] Period CRUD/activate/archive with single-active invariant.
- [ ] Divisions CRUD.
- [ ] Members/organization management + sort order.
- [ ] Public active-period structure.
- [ ] Historical preservation tests.

## Phase 6 — Aspirations
- [ ] Public aspiration form + rate limit + validation.
- [ ] Division/category routing.
- [ ] Admin list/detail/status PENDING → READ → RESOLVED.
- [ ] Scope tests for KEPALA_BIDANG.
- [ ] Responsive visual review.

## Phase 7 — Dynamic Registration
- [ ] Registration form CRUD and status DRAFT/OPEN/CLOSED.
- [ ] Form fields: text, textarea, email, number, select, radio, checkbox, date.
- [ ] Form builder reorder + validation.
- [ ] Public form render/detail/submit.
- [ ] Duplicate rule support.
- [ ] Submission list/detail/filter.
- [ ] XLSX export + activity log EXPORT.
- [ ] End-to-end integration tests.

## Phase 8 — Private Document Archive
- [ ] Private file storage abstraction.
- [ ] Upload allow-list + size limit + safe filenames.
- [ ] Metadata CRUD.
- [ ] Authorized streaming download endpoint.
- [ ] UPLOAD/DOWNLOAD/DELETE activity logs.
- [ ] IDOR/path traversal/security tests.

## Phase 9 — Settings, Users, Activity Log & Dashboard Metrics
- [ ] Landing settings admin editing/preview.
- [ ] User management + role/division validation + deactivate/reactivate/reset-password flow.
- [ ] Activity log list/detail read-only for SUPER_ADMIN.
- [ ] Dashboard summary metrics.
- [ ] Simple visit/pageview mechanism chosen and documented.

## Phase 10 — Hardening
- [ ] Public write rate limiting.
- [ ] CORS, secure cookies, headers, sanitization.
- [ ] Input/file validation review.
- [ ] Authorization matrix integration suite.
- [ ] Database indexes/constraints audit.
- [ ] Performance/SEO/accessibility pass.
- [ ] Dependency/security scan.

## Phase 11 — Full Integration & Visual Compliance
- [ ] All API contract routes verified.
- [ ] Desktop visual review against Stitch.
- [ ] Mobile 360/390/412 review against Stitch patterns.
- [ ] Loading/empty/error/permission states verified.
- [ ] No out-of-scope features introduced.
- [ ] Create defect log and resolve P0/P1 issues.

## Phase 12 — UAT & Deployment Readiness
- [ ] Deterministic reset/seed procedure.
- [ ] Manual UAT checklist by role: Public, KEPALA_BIDANG, BPH, SUPER_ADMIN.
- [ ] Backup/restore procedure.
- [ ] Deployment docs: Vercel frontend + approved Go/MySQL host.
- [ ] Rollback procedure.
- [ ] Production env checklist and secret inventory.
- [ ] Final release report: automated vs manual verification clearly separated.
