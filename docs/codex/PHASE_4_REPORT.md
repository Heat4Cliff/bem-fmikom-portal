# Phase 4 — Content CMS

**Status: PASS** — Phase 4 is complete. Phase 5 has not started.

## Scope delivered

- Article category and article administration, rich-text sanitization, publish/unpublish lifecycle, and public list/detail pages.
- Gallery category and album-item management, HTTPS Google Photos URLs, public category/detail pages.
- Announcement CRUD, scheduling, activation/deactivation, derived schedule state, public archive/detail/active APIs, and landing-page announcement/news/gallery reads.
- Role-aware admin navigation and server-enforced role/division checks.
- Reversible migration `00002_content_cms.sql`; the foundation migration was not modified.
- Source references remain read-only. Stitch `screen.png` was treated as the target, `code.html` as layout/style evidence only, and production UI remains reusable Next.js/TypeScript/Tailwind components.

Out-of-scope domains were not implemented: Phase 5 organization/period UI, aspirations, registration/form builder, private document archive, landing settings backend, user management, activity-log dashboard, analytics, and file upload/storage.

## Source and visual audit

- Read the repository's current phase instructions/contracts, Phase 3 report, PRD v1.4, API/Security Contracts, database/domain notes, design system, route/screen matrix, implementation protocol, and visual checklist before implementation.
- The source inventory contains **37 Stitch screen references** (37 `screen.png` / `code.html` pairs); **8** references are explicitly mobile-labeled. Reviewed the Phase 4 news list/detail/admin desktop/mobile/editor, announcement list/detail/admin, gallery list/detail/admin, and public landing desktop/mobile references (13 relevant screen images). Admin CMS and public routes were also checked in browser at desktop 1280×900 and mobile 390×844.
- The category management screen has no dedicated Stitch reference and reuses the audited admin content patterns. No pixel-parity claim is made for this unreferenced screen.
- Contract discrepancy found and reconciled: the full PRD v1.4 uses `/public/gallery-categories/{slug}`, while the API summary had `{id}`. The public API, client, route matrix, and API contract now use `{slug}`.
- `git diff --quiet -- source_docs design_reference` returned clean after implementation and verification.

## Fixed decisions

1. MariaDB migration adds `article_categories`, `articles`, `article_category_map`, `gallery_categories`, `gallery_items`, and `announcements`; foundation migration remains unchanged.
2. Article and announcement HTML is sanitized on the server using an allowlist. Active-content elements, event attributes, and unsafe URL schemes are stripped; oversized bodies are rejected.
3. Images are URL-backed only; no file uploads or private media storage. Google Photos album URLs and optional image URLs require HTTPS.
4. SUPER_ADMIN and BPH can manage shared content across divisions. KEPALA_BIDANG division is read from the current DB principal, never trusted from the request; it can manage its own articles and gallery items under centrally managed own-division album categories. Article taxonomy and announcements are read-only for KEPALA_BIDANG.
5. Announcement status is derived as INACTIVE, SCHEDULED, ACTIVE, or EXPIRED from `is_active` and UTC schedule bounds; the public active endpoint additionally requires the current time window.
6. Public content Server Components request `cache: "no-store"`; admin requests use the typed authenticated API client.
7. Scoped detail reads can return 404 to conceal record existence; forbidden mutations return 403.

## Verification evidence

| Gate | Result | Evidence |
|---|---|---|
| Docker preflight | PASS | Docker 29.8.2; Docker Compose v5.5.1; `docker info` reached Docker Desktop Engine (Linux, 12 CPUs); `docker ps` showed the repository's `infra-mariadb-1`. |
| Existing DB command/config | PASS | Inspected `package.json`, README, and `infra/docker-compose.dev.yml`; started only with `pnpm db:up`. No new Compose/MariaDB configuration was created. |
| MariaDB readiness | PASS | MariaDB 11.4.13; container health `healthy`; `healthcheck.sh --connect --innodb_initialized` passed; the integration suite opened the configured `bem_fmikom_test` database. |
| Migration integration | PASS | `pnpm api:integration` ran `TestDatabaseIntegration` (not skipped). Goose applied 00001 and 00002 from schema version 0; test rolled content migration down, reapplied it, and suite cleanup returned the dedicated test schema to zero. |
| Database/auth/CMS integration | PASS | Test verified required foundation/content tables, single-active-period invariant, bootstrap admin, bcrypt hash, successful/invalid/inactive authentication, `/auth/me`, refresh rotation/reuse/expiry/logout/idempotency, login/logout activity, token-secret redaction, article/category/gallery/announcement create/update/delete and lifecycle actions, public visibility, duplicate slug, sanitization, and division/role denials. Six named role/scope cases passed: SUPER_ADMIN cross-division, BPH cross-division, KEPALA_A own division, KEPALA_A other division, KEPALA_B other division, KEPALA without division. |
| Frontend tests | PASS | 5 files; 22 tests. Includes public content card/link/pagination cases. |
| Full verification | PASS | `pnpm verify`: lint, format, TypeScript typecheck, frontend tests, Next.js production build, Go fmt, vet, tests, and build all exited successfully. Build emitted the public and admin CMS routes. |
| Browser public smoke | PASS | `/`, `/berita`, `/pengumuman`, `/galeri` rendered from the API's empty test dataset with correct empty states; no demo fixtures. Public APIs were requested by server rendering and returned successfully. |
| Browser admin/auth UAT | PASS | `/admin` unauthenticated redirected to `/admin/login`; wrong password showed the generic invalid-credentials message; SUPER_ADMIN login opened dashboard and exposed User Management/Activity Log; reload restored the session; logout returned to login and reload stayed unauthenticated. BPH exposed Landing Page Settings but not User Management/Activity Log. KEPALA_BIDANG displayed `Scope: Divisi UAT A`, had no Landing Settings/User Management/Activity Log, article taxonomy was read-only, and gallery item controls were visible only for a seeded own-division album while gallery category mutation controls were absent. |
| Browser viewport | PASS | Public and CMS list routes checked at desktop 1280px and mobile 390×844; `scrollWidth` did not exceed viewport width. The final gallery role correction was rechecked at 390×844. |
| Immutable evidence | PASS | No diff in `source_docs/` or `design_reference/`. |

The first integration attempt found that a scoped article detail deliberately returned 404 while a test expected only 403. The test was corrected to accept 403 or the contract-permitted hiding 404; the unchanged server policy continued to deny access, and the full integration suite then passed.

## Gaps and risks

- **G-17 remains a product clarification:** current public announcement archive/detail expose records with `is_active=true`, even when their derived status is SCHEDULED or EXPIRED; `/active` applies the UTC window. Confirm whether scheduled/expired records should be public before release. The ambiguity does not prevent the requested CMS/integration gates from passing.
- There is no dedicated Stitch category screen. The category manager follows the existing admin content language; it has no screen-specific visual acceptance claim.
- The source package lacks approved official photos/logo/social/contact content. URL-backed image fields accept HTTPS but production owners must supply approved assets and reuse rights.
- Browser UAT used temporary accounts and an empty dedicated `_test` database, not production-like content. Backend integration tested real content data; visual content parity with approved news/media remains a follow-up.
- Integration/CLI access to the local Docker TCP port required authorized broader local execution in this Windows sandbox; no project setting was weakened.

## Recommendations for the next authorized phase

This work did not begin Phase 5. Before a later release, confirm announcement archive visibility (G-17), supply approved media/content, and repeat visual review with real approved records. Continue to keep Phase 5 and later scope separate until explicitly authorized.

**PHASE 4: PASS**
