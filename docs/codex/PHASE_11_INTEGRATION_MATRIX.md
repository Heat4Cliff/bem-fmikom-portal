# Phase 11 Integration Matrix

## Evidence convention

`API/DB PASS` means the existing MariaDB `TestDatabaseIntegration` and prior phase regression assertions ran against the repository's dedicated `_test` database. It does not imply browser UAT. Most role-specific visual/navigation workflows remain unreviewed. The owner has separately accepted the current public `/struktur` layout on 2026-10-14; that narrow acceptance does not cover authenticated admin workflows.

| Workflow | Public | SUPER_ADMIN | BPH | KEPALA_BIDANG | Data dependency | Result | Notes |
|---|---|---|---|---|---|---|---|
| Authentication | N/A | API/DB PASS | API/DB PASS | API/DB PASS | users, refresh_sessions, activity_logs | PARTIAL | Existing auth lifecycle and role/division integration ran; Phase 11 login/reload/logout browser flow and inactive-account UI denial not run. |
| Dashboard | N/A | API/DB PASS | API/DB PASS | API/DB PASS | scoped article, registration, aspiration, document, announcement aggregates; active period | PARTIAL | Phase 9 MariaDB assertions compare aggregates with database values. Dashboard screen and navigation were not inspected in browser. Visitor/pageview card remains explicitly deferred. |
| CMS — articles/categories | Published list/detail API covered | API/DB PASS (prior) | API/DB PASS (prior) | scoped API/DB PASS (prior) | articles, categories, division, activity_logs, ARTICLE_COVER media | AUTOMATED PASS; MANUAL UAT PENDING | Real MariaDB `TestDatabaseIntegration` passed for JPG, JFIF, PNG, WebP, static GIF, TIFF, and BMP article covers through create/reload/edit/publish/public article and public media reads. Browser picker/preview and rendered public image still require manual review. |
| Gallery | Public API covered | API/DB PASS (prior) | API/DB PASS (prior) | own-division API/DB PASS (prior) | gallery_categories, gallery_items, Google Photos URL, purpose-bound media asset | AUTOMATED PASS; MANUAL UAT PENDING | Real MariaDB `TestDatabaseIntegration` passed category create/list, Google Photos HTTPS destination, uploaded preview persistence/reload, edit, and public gallery read; existing division policy assertions passed. Browser dialog, image rendering, and responsive UX still require manual review. |
| Announcements | Public API covered | API/DB PASS | API/DB PASS | read policy API/DB PASS | announcements, optional image, purpose-bound media asset, active flag, UTC schedule | PARTIAL | Phase 11F covers no-image create, uploaded and legacy image, invalid reference/date errors, start-only schedule, update/reload, activation/deactivation, and public active read. G-17 archive visibility remains unchanged; manual `/admin/pengumuman` create/reload/render remains required. |
| Organization | Public active organization API covered; `/struktur` layout OWNER ACCEPTED | API/DB PASS | API/DB PASS | read-only scoped API/DB PASS | periods, divisions, members | **API/DB PASS; public `/struktur` OWNER ACCEPTED; admin UAT follow-up** | Period invariant/history, visibility separation, serializer privacy and authorization are integration-tested. Owner accepted current public `/struktur` layout; admin workflows and role-specific navigation were not included in that acceptance. |
| Aspirations | Submit API covered | API/DB PASS | API/DB PASS | own-division API/DB PASS | aspirations, divisions, activity_logs | PARTIAL | Identified/anonymous and status/scope regression is automated; Phase 11 public/admin form, confirmation, filters, and redaction UI not browser-checked. |
| Registration | List/detail/submit API covered | API/DB PASS | API/DB PASS | own-division API/DB PASS | forms, fields, submissions, unique NIM, export | PARTIAL | Dynamic form/schema lock/export DB regression passes; Phase 11 builder/list/detail browser flow and XLSX download confirmation not rerun. |
| Private documents | No public file route by design | API/DB PASS | API/DB PASS | own-division API/DB PASS | documents metadata + confined private storage | PARTIAL | Phase 8 upload/signature/stream/hash/authorization regressions are included in the suite; Phase 11 upload/filter/edit/download/delete browser UAT not rerun. |
| Landing Settings | Public settings API covered | API/DB PASS | API/DB PASS | DENIED by policy | landing_settings | PARTIAL | Role policy and public API regressions exist; save/preview and actual public-page synchronization were not checked in a browser this phase. |
| Users | N/A | API/DB PASS | DENIED by policy | DENIED by policy | users, refresh_sessions, activity_logs | PARTIAL | Lifecycle, scope, and invariant regression is automated; role navigation and direct-URL denial UI not browser-checked this phase. |
| Activity Log | N/A | API/DB PASS | DENIED by policy | DENIED by policy | activity_logs, users; redaction | PARTIAL | Read-only/API redaction regression exists; Phase 11 event filters/details and UI redaction were not visually inspected. |

## Acceptance limitation

The owner acceptance of `/struktur` is a visual/manual acceptance of that route, not a cross-role end-to-end browser `PASS`. Automated contract/database evidence and the remaining manual role workflows stay separately identified.

## Phase 11G — Initial blocked evidence (2026-10-13)

- Targeted MediaUpload and GalleryManager tests: **2 files / 10 tests PASS**.
- Go media decoder/format unit tests: **PASS** for supported raster inputs and malformed, oversized, extreme-dimension, unsupported, mismatched-MIME, and animated GIF cases.
- `pnpm api:integration` ran `TestDatabaseIntegration` (not skipped) but failed before migration setup because `127.0.0.1:3306` actively refused the connection. Docker Desktop's Linux Engine pipe was absent; the installed launcher also could not start Docker Desktop because its expected registry key was missing. The new article cover matrix has not run against MariaDB.
- `pnpm verify` passed lint, format, typecheck, 15 frontend test files / 64 tests, optimized Next build, Go formatting (70 files), and `go vet`; it stopped at `TestDatabaseIntegration` on the same refused connection. `pnpm api:build` was run separately and passed.
- New UI/API behaviors remain subject to browser UAT. This section is automated evidence only and does not grant manual acceptance.

## Phase 11G — Database recovery and integration re-verification (2026-10-13)

- Docker Desktop Engine reachable through authorized local execution: Docker 29.8.2, Compose 5.5.1. Existing `infra-mariadb-1` (`mariadb:11.4.13`) is healthy; `127.0.0.1:3306` accepts TCP connections. Existing `pnpm db:up` was used; no Compose, container, port, or database credentials were created or changed.
- Test target verified as `bem_fmikom_test` (dedicated `_test` database). No development/UAT database cleanup was run.
- `pnpm api:integration`: **PASS**, exit 0. `TestDatabaseIntegration` ran (not skipped); 26 named subtests passed, including media validation, all seven article cover formats (JPG, JFIF, PNG, WebP, static GIF, TIFF, BMP), auth/content/RBAC/division-scope assertions, and existing gallery create, uploaded preview persistence/reload/edit, Google Photos destination, and public gallery reads.
- Migration round trip: empty test schema → migrations 00001–00009 UP → migration 00009 DOWN verified the prior `media_assets` shape/constraints → migration 00008 DOWN verified table removal → migrations re-applied through 00009 → integration workflows → cleanup back to version 0. The test harness assertions were corrected to account for migration 00009 and cleanup now clears only test media fixtures before downgrading the `*_test` schema, preventing the new accepted extensions from violating the old-format check during teardown.
- `pnpm verify`: **PASS**, exit 0. Biome lint/format covered 120 frontend files; strict TypeScript passed; frontend 15 files / 64 tests passed; optimized Next.js build passed; Go format covered 70 files; `go vet`, all Go tests (including the real MariaDB test), and Go build passed.
- `pnpm audit --prod`: **PASS**, no known vulnerabilities found.
- Local UAT runtime started from existing scripts and development configuration. `http://localhost:8080/healthz`, `http://127.0.0.1:3290/`, `/admin/login`, `/admin/berita/new`, and `/admin/galeri` each returned HTTP 200. Protected page HTTP reachability does not prove authenticated access or manual workflow success.
- A subsequent Chrome screenshot showed media upload HTTP 500 and refresh HTTP 401. The API was using `bem_fmikom_dev` without migration 00009; existing `pnpm api:migrate:up` applied 00009 to that development database without deleting records. Existing public media IDs 1–4 then returned HTTP 200. Authenticated upload retry and browser preview remain pending.
- Manual visual/functional acceptance remains pending for multi-format article upload and public rendering, gallery dialogs/UX and public rendering, and previously unreviewed Phase 11 rows. This update changes only automated integration evidence; it does not mark any visual row PASS or close overall Phase 11.

## Phase 11H — Automated/database evidence and UAT boundary (2026-10-13)

| Workflow | Endpoint/data | Automated result | Manual state | Notes |
|---|---|---|---|---|
| Content media fitting | Public content APIs, public media assets | Component tests/build PASS | RE-REVIEW REQUIRED | Bounded contain frames and announcement modal viewport cap; avatar cover exception remains. Image aspect-ratio visual comparison pending. |
| Public organization preview and detail | GET /api/v1/public/organization; active-period leadership/divisions/heads/members | Web preview navigation, active/empty/error, order, image-fallback, and responsive-class tests PASS; production build PASS; local browser route and viewport checks recorded in Phase 11 report | RE-REVIEW REQUIRED | Landing is preview-only; canonical `/struktur` renders the existing endpoint response. The former inline disclosure mapping is superseded. No dedicated public detail Stitch screen; user visual acceptance remains open. |
| Registration IMAGE builder/form | PUT /api/v1/admin/registrations/{id}/fields; POST /api/v1/public/registrations/{id}/submissions | Web tests, MariaDB TestDatabaseIntegration, and browser submit/duplicate UAT PASS | RE-REVIEW REQUIRED | Phase 11I fixed the multipart `answers` object/envelope mismatch that returned 422 for valid browser submissions. Actual browser submit returned receipt #2; a repeated NIM was rejected with a field-level message. A real dev DB inspection found one private PNG metadata row and no binary IMAGE answer value. Keep visual/responsive and admin authenticated detail review open. |
| Registration IMAGE private detail | GET /api/v1/admin/registrations/{id}/submissions/{submissionId}/files/{fileId} | MariaDB authorization/private-storage assertions PASS | RE-REVIEW REQUIRED | Authenticated stream checks role/form/division scope; no anonymous read route. Browser download/preview not manually rechecked. |
| OPEN landing registration | GET /api/v1/public/registrations?status=OPEN; landing server fetch | API returned HTTP 200 with one active form; landing HTTP 200 and emitted Daftar sekarang CTA | RE-REVIEW REQUIRED | No fixture inserted; no-store server fetch delegates time/status filtering to API. CTA direct-form/mobile behavior awaits user review. |

API/DB evidence does not establish image appearance, responsive layout, keyboard/browser behavior, authenticated user navigation, or visual acceptance. The overall Phase 11 remains open pending the user's manual re-review.

## Phase 11 final disposition — 2026-10-14

- `TestDatabaseIntegration` ran against the dedicated `bem_fmikom_test` database, did not skip, and passed migrations 00001–00011, DOWN/UP round-trip, authentication, RBAC/division scope, CMS/media, organization visibility/serialization, aspirations, registration IMAGE/privacy, and private-document assertions.
- Owner manual acceptance closes the current public `/struktur` layout review only. It does not claim manual completion of the unreviewed role-specific admin workflows listed above.
- Those remaining visual/authenticated-browser items are documented non-blocking follow-ups for this Phase 11 finalization because all automated, security, build, and route smoke gates pass and no critical/major product failure was found. They remain visibly marked partial/pending and are not claimed as user accepted.
