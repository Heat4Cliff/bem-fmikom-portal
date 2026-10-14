# Phase 11 Report

## Scope

Phase 11 audit work was performed against the committed Phase 1–10 application. Phase 11A makes the scoped mobile responsive corrections described below; no other feature, analytics, deployment automation, or Phase 12 work was started. The source/design folders remain read-only.

## Phase 11A — Manual Mobile Landing Review

- The user manually reviewed `http://127.0.0.1:3290/` in Chrome at 360×800, 390×844, and 412×915 and reported **REJECTED**. Findings: hero artwork and its information card clipped, the right-side action clipped, fixed mobile navigation overlapped content, page composition was inconsistent, and hero spacing needed correction.
- The matching Stitch mobile `screen.png` was opened as the visual target. The paired `code.html` was treated only as structural/style reference and not copied into production.
- Corrections: hero content now has a fluid 4:3 mobile artwork frame, bounded responsive headline sizing and word wrapping, and a caption layout whose text can shrink and wrap beside its action. The page's mobile bottom padding now accounts for the fixed navigation's 48px links, top padding, border, and safe-area inset. Existing 16px mobile section gutters and single-column mobile composition were retained.
- Structural responsive regression assertions were added for the hero's fluid sizing/wrapping and navigation touch-target/safe-area classes. They do not measure rendered pixel geometry or replace manual Chrome review.
- **Current visual status: RE-REVIEW REQUIRED.** No updated screenshot or `scrollWidth <= innerWidth` measurement was captured in this environment. The original rejection remains the recorded manual result until the user repeats the review after the code change.
- SUPER_ADMIN UAT password setup is blocked by the repository's existing 12-character minimum for the requested 8-character value. A later create-admin attempt reported the preferred email already exists. No password value is included in this report.

## System Integration

- Repository database runtime started with the existing `pnpm db:up` and `infra/docker-compose.dev.yml`; no Compose configuration was created or changed.
- Docker daemon was reachable only through authorized local execution after the sandbox denied access to `npipe:////./pipe/dockerDesktopLinuxEngine`. Docker version: 29.8.2; Compose: 5.5.1; Compose image `mariadb:11.4`; health reported `healthy`.
- `pnpm api:integration` passed against the dedicated `_test` target. The integration suite migrated from version 0 through migration 7, performed rollback and re-apply, ran without skipping `TestDatabaseIntegration`, then cleaned the schema to version 0. Twelve named role/domain subtests passed, along with the existing integration assertions.
- `pnpm verify` passed: web lint and formatting (103 files each), TypeScript, frontend tests (8 files / 36 tests), production build, Go formatting (63 files), vet, Go tests, and Go build. Local runtime versions were Node.js 25.2.1 and Go 1.26.9; the repository/README pins Node.js 24.19.0, so local verification has a Node runtime mismatch even though the build passed.
- No runtime browser integration could be established. This prevents declaring the frontend-to-API user journeys and rendered output accepted for Phase 11.

## Route Audit

- The production build route table includes the required public routes (`/`, articles, announcements, gallery, registration list/detail/form, `/aspirasi`), admin routes (including editor, builder, submission detail, member forms, settings, users, Activity Log), plus `/robots.txt` and `/sitemap.xml`.
- All indexed screen routes were statically reconciled against the Next.js route tree. The gallery detail uses `[slug]` as documented in the route matrix. The category page `/admin/kategori-berita` has no dedicated Stitch reference and reuses the established CMS pattern.
- This is a build/static route audit only. Runtime status, internal links, dynamic not-found behavior, redirects, and broken-link behavior were not browser-checked; accidental runtime 404s therefore remain unexcluded.

## API Contract Audit

- Static enumeration found 94 registered Gin method/path pairs; 93 have explicit product API contract method/path matches. The only non-product endpoint is `GET /healthz`. No contract endpoint path was found missing from the registered router. The route-by-route consumer, authorization, contract-section, and unused-detail audit is in [PHASE_11_API_MATRIX.md](PHASE_11_API_MATRIX.md).
- Frontend feature paths match the current `/api/v1` domains; no stale legacy API path was found. Contract extensions already reconciled in `docs/backend/API_CONTRACT.md` include atomic `PUT /admin/registrations/{id}/fields` and user lifecycle actions under `/admin/users/{id}/{deactivate,reactivate,reset-password}`. These have matching frontend actions; no extra route behavior was added in this phase.
- No direct UI consumer was found for `GET /admin/periods/{id}`, `GET /admin/divisions/{id}`, `GET /admin/announcements/{id}`, or `GET /admin/users/{id}`. The first three editor states use their selected list record; `getUser` is defined but no screen calls it. These contract-supported detail reads do not currently break an exposed UI action.
- API response/error, browser network, request-loop, and real workflow parity were not observed in a browser. The contract audit is static plus existing automated coverage, not end-to-end approval.

## Role Integration

- Existing MariaDB integration and Phase 9 aggregate assertions cover database-current identity, SUPER_ADMIN/BPH cross-division scope, KEPALA_BIDANG own-division scope and denial, missing-division denial, and dashboard aggregate scope.
- Role-specific navigation, direct-URL UI denial, auth flash, and sensitive-data visibility were not rechecked in the browser for Phase 11. No UI-role UAT PASS is claimed.

## Public Workflows

Public route entries are present in the production build and their API/domain regressions ran through the existing automated suites. Landing populated/empty states, public news, announcement schedule/archive, gallery CTA, registration submission, aspiration feedback, contact links, and sitemap response were not exercised in a browser during this phase.

## Admin Workflows

Database and automated regressions remain green for the implemented auth, CMS, organization, aspirations, registration, private documents, settings, users, and Activity Log domains. Phase 11 browser workflows—create/edit/publish, form-builder mutation and export download, document upload/download/delete, settings preview/save, user lifecycle, audit filtering/detail, and cross-role navigation—were not run. The integration suite cleans the `_test` database; no persistent visual UAT dataset or plaintext test credentials were created.

## Visual Compliance

The actual inventory is 37 `screen.png` files and 37 paired `code.html` files; 8 screen references are explicitly mobile-labeled. The extra `modern_academic_horizon` folder contains tokens and is not a screen. The reference screenshots were inventoried in a contact-sheet overview. The implementation could not be opened after the in-app browser rejected `http://localhost:3000/admin/login` under its URL policy. No alternative browser, raw CDP, or command-line browser workaround was used. No implementation screenshot was captured. Every reference row in [PHASE_11_VISUAL_MATRIX.md](PHASE_11_VISUAL_MATRIX.md) is `BLOCKED` for implementation-side comparison; this records missing evidence and does not assert a visual defect.

## Responsive Compliance

The required 360, 390, and 412 px viewports, 390×844 representative state, table-to-card behavior, long-text stress cases, touch targets, and page-level `scrollWidth` were not measured on the implementation. Eight explicit mobile references are individually represented in the visual matrix but not marked compliant. Phase 5/7/8 manual responsive follow-ups remain open.

## Accessibility Review

The design reference defines Plus Jakarta Sans/Inter and a WCAG 2.1 AA target. This phase did not complete a component-by-component label/focus/contrast audit, and no browser keyboard, focus-return, zoom 125/150/200%, screen-reader, or axe audit ran. Accessibility acceptance is therefore not established.

## Browser Console / Network Review

The in-app browser tool rejected the loopback navigation and explicitly prohibited attempts to reach the same result through alternate browser surfaces or low-level workarounds. No console, hydration, CSP violation, network waterfall, request-loop, or role-screen observations are available. Production build success is not substituted for those browser checks.

## Security Regression

- `pnpm api:integration`: PASS against the repository test database; auth lifecycle, role/division, CMS/domain, and storage regressions ran without the database integration test being skipped.
- `pnpm verify`: PASS, including Go and frontend security-related tests.
- `pnpm audit`: PASS, no known vulnerabilities found.
- `go -C apps/api run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...`: PASS, 0 vulnerabilities in reachable application code; 1 module-only/non-reachable advisory remains as documented in Phase 10.
- `go -C apps/api run github.com/securego/gosec/v2/cmd/gosec@v2.29.0 -severity high -confidence high ./...`: PASS, 0 issues.
- No application code or security configuration was changed in Phase 11. Browser response headers, cookie roundtrip, XSS public rendering, and client token exposure were not manually re-probed.

## Performance Smoke

The optimized Next.js production build passed and emitted the expected route table. No browser LCP, layout shift, JS transfer, image loading, waterfall, or admin large-list measurement was possible; no performance threshold is claimed.

## Carried-Forward Verification Closed

None of the manually deferred visual/browser items from Phases 5, 7, 8, or 10 can be closed without the required browser inspection. Prior reports remain historically accurate and unchanged.

## Visual Differences Accepted

None accepted. There are no implementation-side captures to classify as CRITICAL, MAJOR, MINOR, CONTENT, or N/A. The matrix keeps these references blocked instead of asserting parity.

## PRD / Stitch Conflicts

- The PRD v1.4 endpoint table names `DELETE /admin/users/{id}` for soft deactivation. The Phase 9 API contract intentionally uses `POST /admin/users/{id}/deactivate`, `/reactivate`, and `/reset-password`; G-03 records this as resolved Phase 9 scope, and the UI/API do not expose a hard-delete. This is a documented wire-contract divergence, not a Phase 11 change.
- Detail endpoints added to the API contract beyond the full PRD table were reconciled with UI consumers in the API matrix. Member detail and gallery item listing have consumers; period/division/user detail and announcement detail do not have direct UI calls found in source. They remain registered contract operations, with consumer status explicit.
- The dashboard reference includes visitor/pageview presentation, while current Phase 9 scope and API contract defer visitor tracking. No analytics feature was added; dashboard pageview UI remains outside this phase.
- Announcement archive visibility remains the existing documented G-17 behavior: list/detail visibility follows `is_active`, while the active endpoint also applies its UTC schedule. This behavior was not silently changed.
- Extra controls outside PRD/API scope remain omitted. `screen.png` remains the visual target; `code.html` is not production architecture.

## Gaps Carried Forward

- Critical Phase 11 acceptance blocker: no permitted local browser session, leaving visual comparison, responsive measurements, public UAT, SUPER_ADMIN/BPH/KEPALA_BIDANG UAT, browser console/network, and interaction accessibility unverified.
- Phase 7/8 authenticated browser visual follow-ups and Phase 10 CSP console follow-up remain open.
- G-17 announcement archive semantics require stakeholder confirmation; current behavior is recorded above.
- G-15 activity-log retention remains a governance/deployment policy; no automatic deletion was introduced.
- Visitor/pageview analytics remain deferred post-MVP per Phase 9; no fake dashboard values or tracking were added.
- G-04 consumer audit: the API matrix identifies no direct screen calls for `GET /admin/periods/{id}`, `GET /admin/divisions/{id}`, `GET /admin/announcements/{id}`, and `GET /admin/users/{id}`. These contract-supported reads are not required by the current list-based editor flows; runtime API-client/network confirmation remains pending.

## Files Changed

- Added `docs/codex/PHASE_11_VISUAL_MATRIX.md`, `docs/codex/PHASE_11_API_MATRIX.md`, and `docs/codex/PHASE_11_INTEGRATION_MATRIX.md`.
- Added this report.
- Appended this Phase 11 evidence status to `docs/codex/GAPS_AND_QUESTIONS.md`, `docs/codex/IMPLEMENTATION_MATRIX.md`, and `docs/design/ROUTE_SCREEN_MATRIX.md`.
- No application source, migration, `README.md`, `DECISIONS.md`, or `PLANS.md` changes.

## Verification

### MARIA DB INTEGRATION

PASS. Existing repository Compose configuration only; `infra-mariadb-1` healthy; `pnpm api:integration` ran against the dedicated `_test` database and did not skip `TestDatabaseIntegration`. Migrations 0→7, rollback/re-apply, and cleanup to 0 completed. Twelve named subtests passed.

### FULL VERIFY

PASS. `pnpm verify` exit code 0: lint, format, strict TypeScript, 8 frontend test files / 36 tests, optimized Next build, Go format (63 files), vet, Go tests, and Go build.

### AUTH E2E

BLOCKED for Phase 11 manual/browser acceptance. Database auth regressions passed; browser unauthenticated redirect, invalid/valid login, refresh restore, inactive-user presentation, and logout/reload were not re-executed.

### SUPER_ADMIN UAT

BLOCKED. Database authorization coverage passed; navigation, complete workflow, direct URLs, real dashboard view, and activity/user UI were not browser-tested this phase.

### BPH UAT

BLOCKED. Database role policy passed; visible authorized operations and hidden forbidden menus/direct URLs were not browser-tested this phase.

### KEPALA_BIDANG UAT

BLOCKED. Database division scope passed; current-division visual label, own-division workflows, cross-division direct URLs, and no-flash behavior were not browser-tested this phase.

### PUBLIC UAT

BLOCKED. Build routes and API regression passed; populated/empty public pages and end-to-end submissions were not browser-tested this phase.

### ROUTE AUDIT

PARTIAL. Required Next route entries appear in the successful production build; runtime routes, dynamic states, internal links, and response status were not manually checked.

### API CONTRACT AUDIT

PARTIAL. Static router/frontend-path reconciliation and existing integration suites passed. User detail endpoint has no direct UI consumer; no stale frontend path was found. Per-request browser status/error/network parity remains unverified.

### DESKTOP VISUAL REVIEW

BLOCKED. No implementation screenshot; 1440×900 or reference-native comparison could not be performed.

### MOBILE 360 REVIEW

BLOCKED. No browser viewport or `scrollWidth` measurement.

### MOBILE 390 REVIEW

BLOCKED. No browser viewport or `scrollWidth` measurement.

### MOBILE 412 REVIEW

BLOCKED. No browser viewport or `scrollWidth` measurement.

### ACCESSIBILITY REVIEW

PARTIAL / BLOCKED. Static source/tests only; keyboard, focus, contrast, zoom, dynamic field grouping, and assistive technology review were not performed.

### CONSOLE REVIEW

BLOCKED by explicit in-app browser URL policy.

### NETWORK REVIEW

BLOCKED by explicit in-app browser URL policy; no browser network trace.

### SECURITY REGRESSION

PASS for automated gates: API integration, full verify, `pnpm audit` (no known vulnerabilities), govulncheck (0 reachable; 1 module-only non-reachable advisory), gosec high/high gate (0 findings). Manual browser security assertions were not re-probed.

### SOURCE INTEGRITY

PASS. Final `git diff --quiet -- source_docs design_reference` returned clean after the documentation updates. No command edited either source/reference tree.

## Known Issues

No new application defect was confirmed because no implementation screen or manual interaction could be observed. Phase 11 remains blocked by required browser/manual evidence, not by a failing automated test. The latest integration cleanup returned the dedicated `_test` database to version 0; `pnpm api:migrate:up` then restored it to migration version 7 for local UAT. It contains schema only; the integration cleanup removed test data, and an initial UAT administrator still needs to be bootstrapped. `PLANS.md` Phase 11 items remain incomplete where evidence is still outstanding; Phase 12 is untouched. No commit was made because the Phase 11 quality gate is not satisfied.

## Reverification — 2026-10-11

- The first sandboxed `pnpm verify` attempt failed because local TCP access to MariaDB and a storage rename were denied by the sandbox. The same repository command was rerun with authorized local execution and **passed with exit code 0**.
- `pnpm verify` evidence: lint and format checked 103 frontend files; strict TypeScript passed; 8 frontend test files / 36 tests passed; optimized Next build passed and emitted the required route table; Go formatting passed for 63 files; `go vet`, Go tests, and Go build passed.
- `pnpm api:integration` was rerun with `-count=1`. `TestDatabaseIntegration` actually ran (not skipped); all 12 named role/domain subtests passed. The migration round trip completed from version 0 through 7, down to 0, up again, and cleanup back to 0.
- `pnpm api:migrate:up` then restored `bem_fmikom_test` to migration version 7 for local UAT. No production database was targeted.
- Local API and optimized frontend were restarted for user-owned browser UAT: API port `8080`, Next production server bound to `0.0.0.0:3290`. An initial sandboxed Next process was not visible as a Windows host listener; the frontend was restarted with authorized local execution. Host `netstat` then showed `0.0.0.0:3290 LISTENING`, and `Test-NetConnection 127.0.0.1 -Port 3290` passed. The `localhost` probe reported no IPv6 `::1` listener before succeeding over IPv4, so `http://127.0.0.1:3290` is the unambiguous local URL. The interactive initial SUPER_ADMIN bootstrap is waiting for a password; no plaintext credential is recorded.
- No application source, migration, or feature was changed in this re-verification. The automated evidence is current; visual comparison, responsive measurements, accessibility review, browser console/network review, and public/SUPER_ADMIN/BPH/KEPALA_BIDANG manual UAT are still **MANUAL VERIFICATION REQUIRED**. They cannot be inferred from `pnpm verify` or database tests.

## Phase 11A Reverification — 2026-10-12

- Applied only mobile-responsive layout corrections to the public landing hero and fixed-navigation clearance; no API, database, migration, or Phase 12 changes.
- The targeted public web suite passed: 8 test files / 38 tests. New assertions cover fluid hero sizing and wrapping plus mobile navigation touch-target/safe-area structure; they do not claim browser pixel geometry.
- First full `pnpm verify` attempt was blocked by sandbox permissions for the local MariaDB socket and private-file rename. The exact repository command was rerun with authorized local execution and passed exit code 0: lint, format, strict TypeScript, 8 frontend test files / 38 tests, optimized Next build and route output, Go format (63 files), `go vet`, Go tests, and Go build.
- User-supplied manual findings remain REJECTED as the last actual visual result. The corrected app still requires user re-review at 360×800, 390×844, and 412×915, including `document.documentElement.scrollWidth <= window.innerWidth`; no refreshed browser capture or independent geometry reading is available here.
- The requested 8-character password is rejected by the existing 12-character minimum. A create-admin attempt later reported the preferred email already exists; this bootstrap command does not update an existing user's password.
- The user-facing web server is running from the production build on `0.0.0.0:3290`; authorized host HTTP verification returned `200` and included the landing hero. At that point, API startup was blocked because MariaDB was unavailable.
- `git diff --quiet -- source_docs design_reference` remains clean. No commit was made.

## Local Login Runtime Recovery — 2026-10-12

- Started Docker Desktop Engine and confirmed Docker Server 29.8.2. Ran only the existing `pnpm db:up`; `infra-mariadb-1` reports `healthy` with the repository's `mariadb:11.4` image.
- Restarted the API on port 8080. `GET /healthz` returned HTTP 200.
- Chrome uses origin `http://127.0.0.1:3290`, but local `.env` previously allowed only `http://localhost:3000`. Updated the ignored local `FRONTEND_ORIGINS` value to also allow `http://localhost:3290` and `http://127.0.0.1:3290`. Login preflight from `http://127.0.0.1:3290` returned 204 with matching `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials: true`.
- The screenshot's server connection problem is addressed at both the runtime and local CORS configuration. The user should reload the login page. A read-only database check confirmed `superadmin@bem-fmikom.test` exists as `SUPER_ADMIN`, `ACTIVE`, with no division assignment. The browser screenshot used `uat-super-admin@example.test`, which is absent from this local database. Credential validity remains unverified.
- The create-admin attempt shown by the user ended with `an account with that email already exists`; it did not change the existing password. The entered candidate is therefore not evidence of the existing account password.
- The `.env` file is ignored by Git. No new Compose configuration, credential, or production database change was made.

## Login CSP Regression — 2026-10-12

- Chrome DevTools showed two inline scripts blocked by the nonce CSP on `/admin/login`. The built route was static, so Next emitted script tags without the per-request nonce supplied by `src/proxy.ts`.
- Added a server layout for `/admin/login` that forces request-time rendering and reads request headers. The route now appears as dynamic in the optimized build. No `unsafe-inline` allowance was added.
- Added ignored `apps/web/.env.local` with `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1`; this lets the frontend build emit a CSP `connect-src` that includes the actual local API origin.
- Full `pnpm verify` passed after the application-code change: 8 web test files / 38 tests, build, Go fmt (63 files), vet, tests, and build. Frontend was rebuilt after the ignored local env setting.
- Host HTTP verification: `/admin/login` returned 200; all 15 script tags carried the exact nonce from that response's CSP; the next response received a different nonce; CSP contains no `unsafe-inline`; `connect-src` includes `http://localhost:8080`; API request from `http://127.0.0.1:3290` returned 200 with credentialed CORS headers.
- Browser console needs a fresh manual review after reload. These response checks do not claim Chrome has been re-reviewed or establish successful authentication; the existing SUPER_ADMIN password is unknown to this audit.

## Phase 11C — Public Navigation, Landing Polish, Admin Shell, React Bits

### Decisions and implementation

- Public desktop nav and mobile drawer now use the eight approved entries, including only one `Aspirasi` link to `/#aspirasi`. The duplicate “Kirim Aspirasi” CTA and “Portal Admin” link were removed; `/admin/login`, auth, RBAC, and bottom navigation remain unchanged.
- Landing eyebrow text remains as a simple uppercase kicker without a pill/badge/border/background. The hero headline uses the free React Bits Blur Text adaptation; description, CTA, hero artwork, landing sections, and media preview use the Fade Content adaptation.
- Removed the admin “Mode akses / Sesi terverifikasi” strip. The name/role and KEPALA_BIDANG division scope remain in the topbar. `AdminContent` minimum height now subtracts only the 4.5rem topbar instead of the former 7rem topbar-plus-strip height.
- Dashboard cards use the free React Bits Count Up adaptation with current API values; no sample metrics were added. Reduced-motion users see static copy and numbers.
- React Bits free TypeScript/Tailwind source was reviewed from the official component pages and repository. Hero/reveal styling is adapted to external CSS classes because the app's CSP has nonce-only `style-src`; no CSP directive was weakened. The only dependency added is pinned `motion@13.5.0` for the Count Up value/spring logic; it does not set element styles. `pnpm audit` reports no known vulnerabilities.
- Client boundaries are limited to the three small React Bits adapters and the already-client dashboard/media controls. Headline, supporting copy, sections, and CTA are visible in server HTML; lack of JavaScript or IntersectionObserver leaves them static and usable.

### Verification evidence

- Targeted nav/hero/auth/React Bits suites: **4 files / 22 tests passed**.
- Full `pnpm verify`: **PASS** when run with authorized local execution. Lint and format checked 111 web files; strict TypeScript passed; frontend **10 files / 44 tests** passed; optimized Next production build and route generation passed; Go formatting checked 69 files; `go vet`, Go tests, and Go build passed. The sandbox-only attempt was blocked by its MariaDB TCP and private storage rename restrictions; the same full command succeeded outside that restriction.
- `pnpm audit`: **PASS**, no known vulnerabilities.
- `git diff --check`: **PASS**. `git diff --quiet -- source_docs design_reference`: **clean**. The source/design evidence folders remain unchanged.
- The existing Node process on port 3290 was serving a stale pre-11C page; it was replaced with the current production build. Host-side HTTP verification now returns **200** for `/` and `/admin/login`; the home response includes the retained eyebrow and one Aspirasi nav link, omits “Kirim Aspirasi” and “Portal Admin”, has zero inline `style` attributes, and both responses' CSP headers contain no `unsafe-inline`.
- Backend/API contracts and implementation were not changed by Phase 11C, so `pnpm api:integration` was not rerun; the existing Phase 11B integration evidence remains recorded above.
- Automated unit tests assert required/forbidden public nav items on desktop and mobile, retained eyebrow without badge styling, absent admin status strip with role/scope and logout intact, reduced-motion static content, API-backed Count Up content, and no inline `style` attributes on initial animation markup.

### Manual review disposition

- **READY FOR PUBLIC UI RE-REVIEW** at `http://127.0.0.1:3290/`, desktop 1440×900 and mobile 360×800, 390×844, and 412×915. Check nav balance, eyebrow/headline hierarchy, section layout, reduced motion, overflow, hydration, and browser console/CSP.
- **READY FOR ADMIN UI RE-REVIEW** at `/admin`, `/admin/berita/new`, `/admin/galeri`, `/admin/pendaftaran`, `/admin/settings`, and `/admin/users` at 360, 390, and 412 px. Check topbar offset, no access strip, role context, media preview, dashboard counts, hydration, and browser console/CSP.
- Browser-side visual and console inspection could not be performed in this run because the available in-app browser had no local tabs. HTTP response checks do not establish viewport layout, hydration behavior, console cleanliness, or role-screen usability. No user acceptance or visual PASS is claimed. A manual review is still required; Phase 11 remains **BLOCKED pending that review**. No final Phase 11 commit, deployment, or Phase 12 work was performed.

## Phase 11B — Admin Mobile + Public Media Upload

### Manual rejection and affected screens

The user manually rejected admin mobile at 360×800 on `/admin/berita/new`, `/admin/galeri`, and `/admin/pendaftaran`. The first content blocks were reported hidden/overlapped at the topbar, with tight mobile spacing and cramped controls. Those results remain **REJECTED**; updated rows in `PHASE_11_VISUAL_MATRIX.md` are **RE-REVIEW REQUIRED**, never PASS. Additional form screens affected by image picker changes are listed there as well.

### Root cause and shared shell correction

The shared `AdminTopbar` used `position: sticky; top: 0` at every viewport, with a high stacking level; individual pages then began their own padded containers. The code audit found no negative top margins, fixed-height main content, or per-page fixed headers in the three reviewed flows. Sticky positioning preserves initial document flow but remains over content during scroll, matching the reported controls-under-header behavior. The shell now keeps the header in normal flow on mobile and enables sticky positioning only at the large-screen breakpoint. Shared wrappers are shrinkable (`min-w-0`), page-header title/actions wrap at small widths, and mobile form controls cannot exceed their parent. This attribution is a source-level diagnosis; the corrected screens still need the user's browser re-review.

### Image form inventory

| Form | Prior behavior | Required behavior | Implementation |
|---|---|---|---|
| Article cover (`/admin/berita/new`, edit) | URL string only; helper text said upload was not supported | Device picker, drop, preview, filename/size, replace/remove; optional URL | Shared `MediaUpload`, purpose ARTICLE_COVER |
| Announcement image (`/admin/pengumuman`) | Optional HTTPS URL | Same reusable upload UX; optional URL | Shared `MediaUpload`, purpose ANNOUNCEMENT_IMAGE |
| Gallery preview (`/admin/galeri`) | Optional preview URL | Upload UX separate from external album destination | Shared `MediaUpload`, purpose GALLERY_PREVIEW; Google Photos URL unchanged |
| Member photo (`/admin/pengurus/new`, edit) | Optional photo URL | Same reusable upload UX | Shared `MediaUpload`, purpose MEMBER_PHOTO |
| Landing hero (`/admin/settings`) | Optional hero URL with preview | Same reusable upload UX | Shared `MediaUpload`, purpose LANDING_HERO |

No other production admin image URL fields were found in `apps/web/src/features`; public-facing image URLs are read-only display models. The shared control uploads as soon as a file is selected, so a later form-save failure can leave an immutable orphan. DB/activity failures during the upload itself remove the newly written object. There is no asset delete API or automatic collector, preventing deletion of media still referenced by legacy URL columns.

### Database/API/storage

- Added next-only reversible migration `00008_media_assets.sql`; migrations `00001`–`00007` are unchanged. The table records uploader, bounded purpose, safe display filename, unique opaque key, JPEG/PNG MIME and extension, byte size, dimensions, SHA-256, timestamps, uploader FK, purpose/type/size checks, and indexes.
- Added authenticated `POST /api/v1/admin/media` and public `GET /api/v1/public/media/{id}`. Upload returns media ID and stable API-relative URL. API reads return validated image bytes with MIME, Content-Length, `nosniff`, sandbox CSP, and immutable one-year cache headers.
- Added `PublicMediaStorage` and `LocalPublicMediaStorage`; media bytes are held outside web static roots under `MEDIA_STORAGE_PATH` (default `../../var/media` from the API working directory). Storage keys are random hex plus extension. Existing URL columns remain unchanged and can store the configured absolute API URL; no destructive form-data migration was needed.
- The bounded allowlist is JPEG/PNG, up to configurable 5 MiB, 8,000 px per dimension and 12 MP. The backend verifies extension, declared MIME, decoded format, complete decode, body/file size, and dimensions. SVG/GIF/WebP are rejected. WebP remains an explicit unsupported-format note, not a silent conversion.
- SUPER_ADMIN/BPH can request all five purposes. KEPALA_BIDANG is restricted to ARTICLE_COVER/GALLERY_PREVIEW; any eventual resource save remains subject to the existing resource endpoint's own role/division checks. Upload audit details contain only actor, purpose, media ID, and size.
- `docs/backend/API_CONTRACT.md`, `SECURITY_CONTRACT.md`, `DATA_MODEL_NOTES.md`, `DECISIONS.md`, `GAPS_AND_QUESTIONS.md`, `IMPLEMENTATION_MATRIX.md`, and `ROUTE_SCREEN_MATRIX.md` now reflect the new decision and mappings. The Phase 4 URL-only implementation note is superseded for image forms by the explicit Phase 11B decision.

### Security and test evidence

- API role policy tests cover allowed/disallowed purposes. Storage tests cover root confinement, path-traversal keys, opaque object lifecycle; browser component tests cover picker request, preview, removal, and SVG rejection.
- `pnpm api:integration`: PASS against `bem_fmikom_test`; `TestDatabaseIntegration` ran with `-count=1` (not skipped). Migrations applied 0→8, assertions covered 21 tables, 3 named secondary/unique indexes, the uploader FK, and 4 CHECK constraints. DOWN removed `media_assets`, all migrations were rolled back to 0 and reapplied through 8, and deferred test cleanup returned the database to 0. SUPER_ADMIN, BPH, and scoped KEPALA_BIDANG uploads passed; anonymous upload was 401; forbidden KEPALA_BIDANG purpose was 403; invalid image content was rejected; public read bytes/MIME/cache/nosniff passed. Existing auth/RBAC/division and content regressions passed; the test reported 12 named role/domain subtests.
- `pnpm verify`: PASS, exit code 0. Lint/format checked 106 frontend files; strict TypeScript passed; 9 frontend test files / 40 tests passed; optimized Next build and route table passed; Go formatting passed for 69 files; `go vet`, all Go tests, and `go build ./...` passed.
- Local UAT runtime: existing `infra/docker-compose.dev.yml` MariaDB container reported healthy. Only development database `bem_fmikom_dev` was migrated through version 8; the configured `_test` database was not used for the running app. API health (`http://localhost:8080/healthz`), public landing page (`http://localhost:3290/`), and admin login (`http://localhost:3290/admin/login`) returned HTTP 200. CORS preflight from `http://127.0.0.1:3290` returned 204 with credentialed origin; public media unknown-ID read returned 404 as expected. The web and API are running for manual review.
- Login screenshot follow-up: an older `go run ./cmd/server` process still owned port 8080 and had not been replaced by the sandboxed API attempt. It was stopped after verifying its process identity, then the repository API was restarted with local-network access and the development database target. `/healthz` returned 200 and a diagnostic login request with a dummy password returned the expected 401 `UNAUTHENTICATED`; this verifies the API/auth path responds, not the user's credentials. Reload Chrome before retrying.
- `git diff --quiet -- source_docs design_reference`: PASS, both evidence directories remain unchanged. `git diff --check`: PASS.
- These HTTP checks do not provide browser visual sign-off. User must verify mobile layout and choose/drop/replace/remove/public rendering in Chrome.

### Manual re-review handoff

**READY FOR ADMIN MOBILE RE-REVIEW**

**READY FOR MEDIA UPLOAD UAT**

Open `http://localhost:3290/admin/login` for UAT. Use `localhost` so the development refresh cookie remains same-site with the API on port 8080. The API health endpoint is `http://localhost:8080/healthz`.

Do not mark mobile visual rows PASS, commit final Phase 11, deploy, or start Phase 12 until the user approves the updated screenshots and upload workflows.

## Phase 11D — Public Media Rendering and Admin Actions

### Media investigation and correction

- Reproduced against the existing development database record for **Lukluk** (member ID 1). Its `photo_url` points to `http://localhost:8080/api/v1/public/media/1`; the referenced media row and opaque JPEG file exist. A direct host request to that public endpoint returned **200**, `image/jpeg`, 23,111 bytes, `X-Content-Type-Options: nosniff`, immutable public cache headers, and the isolated media-response CSP. This confirms the image and API record are present; the browser-visible failure was on the web page's resource policy/URL handling.
- Before the fix, the site's `img-src` admitted `self`, `data:`, and `https:` but not the configured development API origin. It also had no `blob:` source for local upload previews. The web CSP now adds only the configured API origin plus `blob:`; `unsafe-inline` and broad wildcard image access remain absent.
- Added `resolveMediaUrl` as the single policy for persisted image references: internal `media:<id>` and old-host `/api/v1/public/media/{id}` URLs are rewritten to the current configured API origin, valid legacy HTTPS URLs stay usable, unsafe protocols and arbitrary relative paths are rejected, and `blob:` is allowed only for temporary upload previews. No storage path/key is exposed to the UI.
- Added shared `MediaImage` with fixed intrinsic dimensions or fill sizing, accessible initials/icon fallbacks, and one error fallback without retry loops. It is now used for organization photos, article/announcement/gallery media, landing hero/sections, and upload previews. Member cards keep their accessible initials fallback.
- Added regression coverage for media ID and legacy host resolution, HTTPS compatibility, unsafe URL rejection, upload object previews, missing/broken-image fallback, and logged image failure behavior. The test verifies URL policy and fallback behavior; it does not replace Chrome's image rendering review.

### Shared admin action spacing

`PageHeader` now provides consistent bottom spacing; title and actions stack below the medium breakpoint; action controls fill the available width on narrow screens and return to content width at desktop; primary button labels do not wrap. Structure's period selector and create-member action use a responsive grid, and the member page now separates its filters from the heading/actions with a responsive toolbar.

| Route | Main action(s) | Shared layout / responsive treatment |
|---|---|---|
| `/admin/berita` | Tulis berita baru | `PageHeader`; action stacks below title on mobile |
| `/admin/berita/new`, `/admin/berita/[id]/edit` | Back link; save/publish form actions | shared header plus form controls with non-wrapping buttons |
| `/admin/kategori-berita` | Tambah kategori | shared header and responsive form row |
| `/admin/galeri` | Tambah kategori; simpan item | shared header; editor forms stack on narrow screens |
| `/admin/pengumuman` | Tambah pengumuman | `PageHeader`; full-width mobile action |
| `/admin/struktur` | Pilih periode; Tambah Pengurus | responsive selector/action grid |
| `/admin/divisi` | Tambah Divisi | `PageHeader`; full-width mobile action |
| `/admin/pengurus` | Tambah Pengurus; filter/search | shared heading/action, separate responsive filter toolbar |
| `/admin/periode` | Tambah Periode Baru | `PageHeader`; full-width mobile action |
| `/admin/pendaftaran` | Buat Pendaftaran | `PageHeader`; full-width mobile action |
| `/admin/pendaftaran/[id]/builder` | Buka/Tutup Pendaftaran; Data pendaftar | shared header with stacked mobile actions |
| `/admin/pendaftaran/[id]/submissions` | Ekspor XLSX | shared header; action remains one line |
| `/admin/arsip-dokumen` | Unggah Dokumen | `PageHeader`; full-width mobile action |
| `/admin/settings` | Pratinjau; Simpan Perubahan | shared header with mobile stacked actions |
| `/admin/users` | Tambah Pengguna | `PageHeader`; full-width mobile action |
| `/admin/aspirasi`, `/admin/activity-logs` | Search/filter/detail actions; no create action | shared title spacing; filters remain separate from title actions |

### Verification evidence

- Targeted web tests and full suite: **12 files / 51 tests passed**. Lint, format, strict typecheck, optimized production build, Go format (69 files), `go vet`, Go tests, and Go build all passed in the final `pnpm verify` run. The first attempt exposed an unnecessary React effect dependency and a `.next` file lock held by the running preview; the dependency was removed and the identified project preview was stopped before rerunning. No application test was skipped.
- `pnpm api:integration`: **PASS**, executed `go test -v ./... -run TestDatabaseIntegration -count=1` on the existing dedicated `_test` database. `TestDatabaseIntegration` ran (not skipped); migrations applied 0→8, rolled back to 0, reapplied to 8, and test cleanup returned the schema to 0. Twelve named role/domain subtests passed, including SUPER_ADMIN/BPH cross-division and KEPALA_BIDANG own/other/missing division denial cases.
- `git diff --check`: PASS. `source_docs/`, `design_reference/`, and `PLANS.md` remain unchanged. No Compose configuration was added. The code HTML references were not used as production architecture.
- The production `next start` preview is now running hidden on port 3290. After clearing the stale listener and rebuilding, host HTTP smoke returned `/` 200 and `/admin/login` 200. The login response CSP admits the exact configured `http://localhost:8080` image origin and `blob:`, contains no `unsafe-inline`, and all 15 script tags carry the response nonce. API health returned 200; `/api/v1/public/media/1` returned 200 (`image/jpeg`, 23,111 bytes). No Compose or port 3000 configuration was introduced.
- Browser automation was unavailable in this environment, so no Chrome console, visual measurement, or user-supplied screenshot is claimed for these changes. The HTTP/CSP and API media checks are supporting evidence only; they do not establish browser image rendering or visual alignment. Chrome re-review remains required.

### Manual re-review handoff

**READY FOR PHOTO RE-REVIEW** — reload the public landing page and member structure at `http://localhost:3290/` and its organization section, then check the Lukluk photo; also inspect article, announcement, gallery, hero, and upload previews where present.

**READY FOR GLOBAL ADMIN ACTION RE-REVIEW** — inspect the route table above at desktop and 360×800, 390×844, and 412×915. Confirm the title/action spacing, button labels stay on one line, and the structure selector/action and member filter toolbar do not collide or overflow.

Manual browser evidence is still pending. Phase 11 remains **BLOCKED pending user re-review**; no final Phase 11 commit, deployment, or Phase 12 work was performed.

## Phase 11E — Global Spacing and Media Card Polish

### Manual findings and root cause

- The user's latest manual review rejected `/admin/pengurus/new`'s Foto Profil card: the preview/file details and the three actions were compressed into a horizontal media layout at `sm`, even though the card's available width is controlled by the narrower photo column. This caused unnecessary text wrapping and poor preview/action hierarchy.
- The `/aspirasi` success state was rejected because the reference metadata and retry CTA were inline-level elements. The CTA's `mt-6` did not reliably create separation in the inline formatting context, so the reference and button appeared crowded together.
- Shared Button and PageHeader patterns already had non-wrapping button labels and token spacing, but custom media actions and this success footer did not use a reusable action-group rule.

### Shared spacing and media changes

- Added reusable `ActionGroup` in `apps/web/src/components/ui/primitives.tsx`: mobile-first vertical stacking, `gap-3` (12 px spacing token), then desktop wrapping flex with aligned controls. It is used for the media actions, member-form footer, and aspiration success actions.
- `MediaUpload` now uses a vertical selected-media card at every width: preview, filename and file size, then a bounded vertical action group. “Ganti gambar”, “Hapus”, and “Gunakan URL eksternal” each have at least 40 px height, adequate gap, visible focus treatment, and non-wrapping labels. This shared presentation applies to article, gallery, announcement, member, and landing-hero upload fields. External URL mode remains a labeled URL input with an accessible return action.
- Increased the member form photo column minimum from 15rem to 17rem while preserving the two-column layout only at `lg`; tablet and mobile stay one column. The form footer now uses the shared action group.
- The aspiration success card now orders success icon/title/description before reference metadata and the primary CTA. The action group stacks and gives the retry CTA full width on small mobile; from `sm` onward it uses a clean row with the reference as secondary metadata and the button kept on one line. The CTA retains the shared 40 px minimum height and is keyboard reachable.
- Audited member-form validation: it uses browser-native required-field constraints; its `Notice` is for API/save errors, not controlled required-field errors. No competing inline required-error component exists on this form, so validation behavior was left native rather than adding a second validation system. Aspirations already use controlled `noValidate` validation and were not changed.
- No motion behavior was added or used to disguise spacing; Fade Content still wraps static content and now forwards standard div attributes such as `data-slot`.

### Validation and disposition

- Targeted structural tests: **3 files / 11 tests passed**, covering aspiration reference/action grouping, CTA keyboard reachability, media selected state and action structure, URL-mode accessibility, and the shared mobile/desktop action-group classes. These tests do not claim pixel geometry.
- `pnpm verify`: **PASS**, exit 0. Lint and format checked 115 frontend files; TypeScript passed; frontend **12 files / 52 tests** passed; optimized Next production build passed; Go formatting checked 69 files; `go vet`, Go tests, and Go build passed.
- No backend/API/data behavior changed in Phase 11E, so `pnpm api:integration` was not rerun. The Phase 11D integration result above remains the latest database evidence.
- Production preview restarted at `http://localhost:3290`; HTTP smoke returned 200 for `/`, `/aspirasi`, and `/admin/pengurus/new`. These status checks do not count as visual review. `source_docs/`, `design_reference/`, and `PLANS.md` are unchanged; `git diff --check` passed. No commit, deployment, or Phase 12 work was performed.

**READY FOR ASPIRATION SPACING RE-REVIEW** — submit a manual test aspiration to show the success state and inspect at 1440×900, 360×800, 390×844, and 412×915.

**READY FOR MEMBER FORM LAYOUT RE-REVIEW** — inspect `/admin/pengurus/new` after login at the same sizes; select an image and verify preview, filename/size, Ganti gambar, Hapus, and external URL actions. Also spot-check the shared uploader on article, gallery, announcement, and landing settings.

Both findings remain **RE-REVIEW REQUIRED** until the user approves them. Phase 11 stays **BLOCKED pending manual acceptance**.

## CSP regression follow-up — 2026-10-12

- The user's Chrome screenshot showed `/admin` stalled on its loading skeleton and DevTools reported inline scripts blocked by the strict nonce-based CSP. This is new manual evidence; the prior HTTP checks had covered `/` and `/admin/login`, not the protected admin shell.
- Root cause: only `/admin/login` forced request-time rendering for nonce propagation. `/admin` and other admin routes could be prerendered, leaving Next.js inline bootstrap scripts without the nonce required by the production CSP.
- `apps/web/src/app/(admin)/admin/layout.tsx` now forces dynamic rendering and reads request headers, matching the nonce-aware login layout. The production build lists all `/admin` routes as dynamic (`ƒ`), not static.
- `pnpm verify`: **PASS**, exit 0 after this fix. Lint and format checked 111 web files; strict TypeScript passed; frontend **10 files / 44 tests** passed; optimized production build passed; all `/admin` routes were dynamic; Go formatting checked 69 files; `go vet`, Go tests, and `go build ./...` passed.
- `pnpm audit`: **PASS**, no known vulnerabilities. The sandbox registry request first failed DNS; the approved local retry completed successfully.
- The stale Next server was replaced with the latest production build on port 3290. Two requests to `/admin` returned HTTP 200 with distinct per-request CSP nonces; all 16 script tags on the response, including both inline bootstrap tags, matched the response nonce. `unsafe-inline` is absent. This verifies response markup and headers; it does not establish a clean Chrome console or successful client hydration.
- The supplied screenshot is recorded as CSP **REJECTED** before the fix. The corrected `/admin` and affected admin screens are **RE-REVIEW REQUIRED** until Chrome is reloaded and the user confirms the skeleton resolves with no CSP errors. Phase 11 remains blocked on browser/UAT and other carry-forward visual acceptance; no commit, deployment, or Phase 12 work was performed.

## Phase Result

**PHASE 11: BLOCKED — MANUAL REVIEW REQUIRED**

Phase 11B implementation, real database integration, and full automated verification are complete. The admin mobile result remains rejected until the user re-reviews the corrected UI at 360×800, 390×844, and 412×915; media picker/drop/replace/remove/public rendering also need the user's browser UAT. Prior auth/cross-role manual Phase 11 gates and the SUPER_ADMIN credential confirmation remain carried forward. No final Phase 11 commit was made. Do not start Phase 12 until manual review and remaining acceptance gates are approved.

## Phase 11E manual re-review update — 2026-10-12

- **MANUAL RE-REVIEW: PASS**
- **USER ACCEPTANCE: 2026-10-12**
- Accepted: aspiration success action spacing and success-card hierarchy at `/aspirasi`.
- Accepted: member create form media-card/action layout at `/admin/pengurus/new`, including preview, file metadata, Ganti gambar, Hapus, and external URL.
- This dated disposition supersedes the earlier Phase 11E handoff status only for those named UI states. Historical rejection and corrective implementation evidence above is retained.
- The initial aspiration form, member edit route, and all other unreviewed states remain pending. **Overall Phase 11 remains open in manual UAT; this update does not mark Phase 11 PASS.**

## Phase 11F — Announcement and Gallery Create Regression

### Manual defect and reproduction evidence

- User-reported dispositions: **Announcement Create — REJECTED — FUNCTIONAL DEFECT** at `/admin/pengumuman`; **Gallery Create — REJECTED — FUNCTIONAL DEFECT** at `/admin/galeri`.
- Local service pre-flight: frontend routes on `http://localhost:3290` returned HTTP 200; API `/healthz` returned HTTP 200. The development MariaDB container was healthy. No production database or production data was used.
- No Chrome Network request was available in the connected browser state. One local admin login attempt returned HTTP 401, so no authenticated create request was sent to the already-running API and its original browser status/body cannot be claimed as directly captured. The request contract below was reconstructed from the live form code and then exercised through the same authenticated handlers/payload shapes in the MariaDB integration harness.
- Announcement form request: `POST http://localhost:8080/api/v1/admin/announcements`, `Content-Type: application/json`, JSON fields `title`, `slug`, `content_html`, `image_url`, `start_at` (ISO 8601), and `end_at` (ISO 8601 or `null`). A selected image is uploaded separately as multipart `POST /api/v1/admin/media` with `purpose=ANNOUNCEMENT_IMAGE` and one `file`; the form then submits the API-origin URL resolved from returned `public_url`.
- Gallery category request: `POST /api/v1/admin/gallery-categories`, JSON fields `title`, `slug`, `description`, numeric-or-null `period_id`, numeric-or-null `division_id`, and `event_date` (date or `null`). Integration confirmed category create and list reload succeed.
- Gallery item request: `POST http://localhost:8080/api/v1/admin/gallery-items`, `Content-Type: application/json`, JSON fields `category_id` (selected numeric ID), `title`, `description`, `google_photos_url` (HTTPS destination), `preview_image_url`, and numeric `sort_order`. Preview upload is multipart `POST /api/v1/admin/media` with `purpose=GALLERY_PREVIEW` and one `file`.
- Exact pre-fix backend cause: announcement `image_url` and gallery `preview_image_url` were sent as `http://localhost:8080/api/v1/public/media/{id}` after `MediaUpload` resolved the returned internal media path. Both save services called `validateOptionalURL`, which accepted only absolute HTTPS URLs. The existing handler maps that validator error to HTTP **422**, code `VALIDATION_ERROR`, message `Validation failed`, and no field details. This is the code-derived baseline response, not a captured original Chrome Network response: the live admin session could not be authenticated in this run. The corrected integration requests now return HTTP 201 for valid uploaded-image creates. Gallery category create itself has no matching defect in the integration run; the failing workflow is item creation when the uploaded preview is attached.
- The relevant validation path does not write a server stack/SQL error for this rejected media reference. Other invalid dates had the same generic catch UI; malformed timestamp JSON previously failed binding as a body error.

### Fix and data policy

- Added a server-side image-reference policy for announcement images and gallery previews. It accepts `null`/empty, legacy HTTPS URLs, `media:{id}`, the exact `/api/v1/public/media/{id}` path, or an HTTP(S) absolute URL whose path matches that public media route. Internal references must resolve to an existing `media_assets` row with the matching `ANNOUNCEMENT_IMAGE` or `GALLERY_PREVIEW` purpose. Accepted internal absolute references are normalized before persistence to `/api/v1/public/media/{id}`. Other relative paths, malformed URLs, unsafe schemes, missing media IDs, and wrong-purpose media are rejected; arbitrary HTTPS legacy images remain supported.
- The Google Photos album URL remains a separate HTTPS external destination. Preview media remains independently stored in `preview_image_url`; no database field or migration was changed.
- Announcement date decoding now reports `start_at`/`end_at` field errors for malformed ISO 8601 values. End-before-start reports `end_at`; valid timestamps are normalized to UTC. `end_at` and both image fields remain optional; empty image values persist as SQL `NULL`.
- The forms now use the API’s structured field errors with user-facing field labels and safe fallbacks. Save is disabled while media upload is pending, and payload builders reject `blob:` values. Gallery create continues to submit the selected category ID and keeps Google Photos URL separate from preview media.

### Verification evidence

- Files changed for Phase 11F: `apps/api/internal/service/content.go`, `apps/api/internal/http/handler/content.go`, `apps/api/internal/http/router/auth_integration_test.go`, `apps/api/internal/http/router/media_integration_test.go`, `apps/web/src/lib/api/client.ts`, `apps/web/src/components/admin/media-upload.tsx`, `apps/web/src/components/admin/media-upload.test.tsx`, `apps/web/src/features/content/announcement-manager.tsx`, `apps/web/src/features/content/gallery-manager.tsx`, new `content-payloads.ts` and `form-errors.ts` helpers/tests, `docs/backend/API_CONTRACT.md`, `docs/backend/DATA_MODEL_NOTES.md`, and the Phase 11 visual/integration matrices.
- MariaDB integration coverage adds category create/list reload, optional preview, uploaded purpose-matched preview, absolute API media normalization, persistence/reload, legacy HTTPS preview, invalid path rejection with `preview_image_url` field feedback, invalid Google Photos HTTP URL, and public gallery API values retaining the separate album destination and image reference. Existing division-scope assertions continue to pass.
- Announcement integration coverage adds no-image SQL `NULL`, optional `end_at`, uploaded image create and canonical SQL persistence, admin detail reload, public activation/list rendering, deactivation, legacy HTTPS update, unsafe image rejection with `image_url` feedback, invalid date order, and malformed start timestamp feedback. Existing lifecycle/RBAC assertions continue to pass.
- `pnpm api:integration`: **PASS** on `bem_fmikom_test`; `TestDatabaseIntegration` ran (not skipped). Migration 0→8, down/up round trip, content/media/auth/RBAC integration, and cleanup back to version 0 completed. The first sandbox execution could not open the local MariaDB socket; the authorized local execution completed successfully. A follow-up rerun after isolating upload fixtures to the integration test's configured temporary media directory also passed.
- Targeted frontend tests: **3 files / 7 tests PASS** for media upload pending state, exact announcement/gallery payload construction (including optional image, selected category, separated album URL, and no blob URLs), and field-aware validation messaging.
- `pnpm verify`: **PASS**, exit 0. Lint and format checked 119 frontend files; strict TypeScript passed; frontend **14 files / 57 tests** passed; optimized Next build passed; Go formatting checked 69 files; `go vet`, all Go tests, and Go build passed.
- Public read in the integration server returned the persisted announcement image and gallery preview; the media endpoint returned the stored valid image bytes. These are API/DB proofs, not a browser page-rendering claim.
- Final `git diff --check`: **PASS**. `git diff --quiet -- source_docs design_reference`: **clean**; the untracked/updated Phase 11 documentation also passed a trailing-whitespace scan.
- Current local UAT services were restarted from the verified code: `http://localhost:3290/admin/pengumuman` and `/admin/galeri` returned HTTP 200; API `http://localhost:8080/healthz` returned 200; credentialed CORS preflight for `http://localhost:3290` returned 204 with the exact allowed origin and credentials enabled. This confirms the preview is reachable, not that authenticated create works in the browser.

### Phase 11F re-verification — 2026-10-12

- After correcting the integration fixture to use `cfg.MediaStoragePath` (the isolated temporary media root configured for the integration server), `pnpm api:integration` passed again. `TestDatabaseIntegration` ran, migration round trip and cleanup completed, and no fixture writes use the development media directory.
- `pnpm verify` was rerun after that test change and passed with exit code 0: 119 frontend files linted/formatted, strict TypeScript passed, 14 frontend test files / 57 tests passed, optimized Next.js build passed, Go formatting covered 69 files, and Go vet/tests/build passed.
- Existing local services were confirmed active rather than duplicated: admin pengumuman and gallery routes returned HTTP 200, API `/healthz` returned HTTP 200, and credentialed CORS preflight returned 204 with `Access-Control-Allow-Origin: http://localhost:3290` and credentials enabled.
- No authenticated browser create was performed in this re-verification. Manual announcement and gallery acceptance remains with the user; overall Phase 11 remains open.

### Manual re-test and disposition

- **Announcement Create: FUNCTIONAL RE-REVIEW REQUIRED** — use `/admin/pengumuman`, create with no image, then create with a newly uploaded image; save, reload the list, edit the record, activate it, and inspect the public announcement page. Check 1440×900, 360×800, 390×800, and 412×915.
- **Gallery Create: FUNCTIONAL RE-REVIEW REQUIRED** — create a category, confirm it remains selected/reappears after reload, create an item with an HTTPS Google Photos URL and uploaded preview, reload, edit, and inspect the public category/item. Confirm the CTA uses the album URL and the preview uses the stored image. Check the same desktop/mobile sizes.
- Keep both routes **FUNCTIONAL RE-REVIEW REQUIRED** until the user manually confirms success. No final Phase 11 commit, deployment, unrelated screen work, or Phase 12 work was performed.

## Phase 11G — Multi-format image pipeline and gallery UX (2026-10-13)

### Image policy and implementation

- Accepted inputs are JPEG (`.jpg`, `.jpeg`, `.jfif`), PNG, WebP, single-frame GIF, TIFF (`.tif`, `.tiff`), and BMP. The picker lists only these types; backend extension, declared MIME, actual decoder format, complete decode, byte limit, and dimensions are checked server-side.
- JPEG, PNG, WebP, and static GIF preserve their image encoding. TIFF and BMP are decoded then normalized to PNG. Original MIME and stored MIME are separately recorded; `mime_type` remains the stored-MIME alias. Normalized output stays under the configured 5 MiB cap.
- Animated GIF is explicitly rejected with `GIF animasi belum didukung.` It is not silently reduced to one frame. SVG, HEIC/HEIF, and AVIF remain unsupported.
- Pixel safety remains 8,000 px per edge and 12 megapixels total. No native binaries or CGO dependency were added. `golang.org/x/image` v0.47.0 supplies TIFF, BMP, and WebP decoders; the fixture is attributed in its testdata README.
- Added only migration `00009_multi_format_media.sql` after the existing media migration to record original MIME and widen accepted input/stored format constraints. Existing `00008_media_assets.sql` was not edited.

### Gallery management UX

- Before: category navigation, full category form, item list, and full item/album/preview form competed on one page; period/division required raw IDs and long album links dominated item content.
- After: compact horizontally scrollable category chips with clear active context, live category/documentation summaries, scan-friendly preview cards and Google Photos link labels, and responsive dialogs for category/item create/edit. Secondary actions use menus; deletion uses confirmation dialogs. Empty state offers an add-documentation CTA.
- Period and division selectors display API-backed names while requests still carry IDs; divisions are filtered by selected period. Item creation preselects the active category. Google Photos remains a separate required HTTPS destination; preview remains the shared MediaUpload field.
- Changed gallery component: `apps/web/src/features/content/gallery-manager.tsx`; new regression tests: `apps/web/src/features/content/gallery-manager.test.tsx`. Responsive classes avoid a squeezed two-column master/detail on mobile. Real-device/browser layout remains unreviewed.

### Verification status

- Frontend targeted tests: **2 files / 10 tests PASS** (MediaUpload supported-format accept/copy/selection/error/preview/remove and gallery navigation/dialog/preselection/upload payload/edit/delete-confirm/empty state).
- Go targeted media service tests: **PASS** for JPEG/JFIF, PNG, WebP, static/animated GIF policy, TIFF/BMP normalization, and invalid/security boundaries.
- `pnpm audit --prod`: **PASS**, no known vulnerabilities found.
- `pnpm api:integration`: **BLOCKED — EXECUTED, FAILED before migration or test setup**. `TestDatabaseIntegration` ran (not skipped) and could not connect to `127.0.0.1:3306` because the connection was actively refused. Pre-flight found no `npipe:////./pipe/dockerDesktopLinuxEngine`; the existing `pnpm db:up` command failed because the daemon pipe did not exist. `docker desktop start` then reported the installed launcher could not find `SOFTWARE\\Docker Inc.\\Docker Desktop`. No new Compose or manually configured database was created.
- `pnpm verify`: **BLOCKED — exit 1 at `pnpm api:test`**. Lint, format, typecheck, frontend **15 files / 64 tests**, optimized Next production build, Go formatting (**70 files**), and `go vet ./...` passed. The Go test packages without the integration database passed; `TestDatabaseIntegration` failed at the same refused `127.0.0.1:3306` connection. Since the sequential verify script stopped there, `pnpm api:build` was run separately and **PASS**.
- New article regression code creates, reloads, edits, publishes, and public-reads one article cover per accepted format and verifies public stored MIME. It must run against MariaDB before it counts as passing evidence.
- Manual visual/functional review has not been performed by the user for `/admin/berita/new`, article edit, `/admin/galeri`, or public rendering after these changes. These rows remain **RE-REVIEW REQUIRED**; overall Phase 11 remains open. No final commit, deployment, or Phase 12 work was performed.

### Phase 11G disposition

- **READY FOR MULTI-FORMAT MEDIA RE-REVIEW: PENDING** — first restore Docker/MariaDB, run the real migration and `pnpm api:integration`, then re-review upload/save/edit/public rendering in the browser.
- **READY FOR GALLERY UX RE-REVIEW: UI READY; FUNCTIONAL UAT PENDING** — the redesigned page and dialogs can be visually inspected, but category/item persistence and uploaded preview must be rechecked after MariaDB is available.
- **PHASE 11G: BLOCKED** on the unavailable database runtime and user's manual re-review. Overall Phase 11 remains open.

## Phase 11G — Database Recovery and Integration Reverification (2026-10-13)

This dated section supersedes the prior database-runtime blocker above. The earlier failed attempts remain as historical evidence; the manual UAT limitations remain active.

- **Docker pre-flight:** Docker Engine reachable through authorized local execution. Docker version 29.8.2, Docker Compose 5.5.1. Existing container `infra-mariadb-1` (`mariadb:11.4.13`) reported `running (healthy)`; TCP `127.0.0.1:3306` passed.
- **Database startup/target:** Existing `pnpm db:up` command completed successfully with `infra/docker-compose.dev.yml`. The integration target was confirmed as `bem_fmikom_test`, a dedicated `_test` database. No development/UAT data was wiped.
- **Migration 00009:** `pnpm api:integration` started with an empty test schema, applied migrations 00001–00009, verified migration 00009 DOWN restored the previous `media_assets` columns and check constraints, verified migration 00008 DOWN removed the media table, re-applied migrations through 00009, then cleaned the dedicated test schema back to version 0. Migration 00001–00008 files were not changed.
- **Database integration:** `pnpm api:integration` **PASS**, exit 0; `TestDatabaseIntegration` ran, was not skipped, and reported 26 named subtests passing. It covered supported/rejected media inputs, article cover create/reload/edit/publish/public article and media reads across JPG, JFIF, PNG, WebP, static GIF, TIFF, and BMP, plus auth/content/RBAC/division-scope and gallery operations.
- **Test harness repair:** Integration assertions had assumed the previous migration count and expected migration 00009 DOWN to drop a table created by 00008. Assertions now validate the 00009 rollback separately from the 00008 table removal. Since accepted multi-format test fixtures cannot satisfy the old-format constraints during final downgrade, reset/teardown clears only `media_assets` fixtures in the guarded `_test` database before migrating down. No media policy or migrations 00001–00008 were changed.
- **Gallery regression:** Real MariaDB/API flow passed category creation and list reload, selected category use, Google Photos HTTPS destination, uploaded internal preview persistence and reload, item edit, and public gallery/category read. Integration evidence separates the album URL from internal public preview media.
- **Full verification:** `pnpm verify` **PASS**, exit 0: lint and format over 120 frontend files, strict TypeScript, frontend 15 files / 64 tests, optimized Next.js build, Go format (70 files), `go vet`, full Go tests (including MariaDB integration), and Go build.
- **Security audit:** `pnpm audit --prod` **PASS**, no known vulnerabilities found.
- **Local UAT runtime:** API and frontend started using the repository's existing scripts/configuration. HTTP 200 verified at `http://localhost:8080/healthz`, `http://127.0.0.1:3290/`, `/admin/login`, `/admin/berita/new`, and `/admin/galeri`. The admin route checks confirm HTTP reachability only, not authenticated browser workflow success.
- **Source integrity:** `git diff --check` **PASS**; `git diff --quiet -- source_docs design_reference` returned clean. No development/UAT records were cleared, no final commit was created, and no deployment or Phase 12 work was started.
- **UAT handoff:** Article multi-format file picker/preview/public rendering and gallery dialog/layout/public rendering remain **RE-REVIEW REQUIRED**. Other unreviewed Phase 11 states also remain open. No visual/manual row is marked PASS based on automation; no final Phase 11 commit, deployment, or Phase 12 work was performed.

**Current Phase 11G status: DATABASE/AUTOMATED VERIFICATION PASS; MANUAL UAT PENDING. Overall Phase 11 remains open.**

## Phase 11G — Manual UAT finding: development schema alignment (2026-10-13)

- The user's Chrome screenshot of `/admin/berita/1/edit` showed `POST /api/v1/admin/media` returning HTTP 500 and a prior `POST /api/v1/auth/refresh` returning HTTP 401. API request logs show a subsequent login returned HTTP 200 and admin article reads returned HTTP 200 before the failed media upload.
- Root cause for the media 500: the local API used `bem_fmikom_dev`, whose schema had not yet applied migration 00009 although the running Phase 11G code expects `media_assets.original_mime_type`. The screenshot also showed public media requests returning 500 under that stale schema.
- Remediation: ran the existing `pnpm api:migrate:up` against the confirmed development database `bem_fmikom_dev`; it applied only `00009_multi_format_media.sql` and reported migration version 9. This is an additive schema migration; no development/UAT records were deleted and no database configuration was changed.
- Post-migration evidence: unauthenticated public reads of existing media IDs 1–4 returned HTTP 200 with `image/jpeg` and non-empty image bytes. API and frontend remained available on ports 8080 and 3290.
- **Manual upload retry remains pending.** The authenticated browser POST was not repeated by automation. User should reload `/admin/berita/1/edit`, sign in again if refresh reports 401, and retry an image upload. Keep article media and visual rows **RE-REVIEW REQUIRED** until that browser workflow and preview are confirmed.


## Phase 11H — Media, organization, and registration corrective scope (2026-10-13)

### Findings and implementation

- **Media fitting:** audit found content images shared cover-style fitting/fixed-height assumptions in several public and admin consumers, including the announcement popup. Added reusable bounded MediaFrame/MediaImage presentation; article, announcement, gallery, landing, and registration content images use centered object-contain within fixed/bounded frames. Avatar photos remain an explicit cover-cropped exception with stable fallback. The announcement popup constrains its image to viewport height and keeps the full image visible on desktop/mobile.
- **Public structure action:** the route matrix and public API use the existing landing #struktur section; no dedicated public structure route was required. The prior “Lihat struktur” action had no interaction. It now toggles the active-period leadership/division hierarchy in place with an accessible button, aria-expanded, and aria-controls; no duplicate organization fetch is added. Runtime smoke then exposed that this stateful component needed a client boundary; adding the client directive removed the actual landing HTTP 500.
- **Registration IMAGE:** added the ninth structural field type, bounded image selection/preview/replace/remove, multipart answers JSON plus image:{field_key} file parts, and an authenticated admin detail stream. Applicant bytes are in a dedicated private storage root using opaque keys and registration_submission_files metadata; no public media URL or binary answer value is stored. Migration 00010_registration_image_attachments.sql is reversible and is the first migration after 00009. It uses staging/promotion with best-effort cleanup on validation/DB failure. Existing post-first-submission schema lock remains in force.
- **Accepted formats:** JPEG/JPG/JFIF, PNG, WebP, static GIF, TIFF, and BMP. TIFF/BMP normalize to PNG. Animated GIF, SVG, HEIC/HEIF, and AVIF are rejected. Uploads remain bounded at 5 MiB each, up to five files, 8,000 px per edge, and 12 megapixels.
- **OPEN registration visibility root cause:** / previously never requested /api/v1/public/registrations?status=OPEN; the landing content therefore always received an empty registrations array. Landing now fetches the existing backend public endpoint with cache: no-store and renders returned open forms with availability/division data and a real /pendaftaran/{slug}/form CTA. API remains the authority for status and time-window filtering. Local dev API read returned HTTP 200 with one active OPEN registration; after the client-boundary fix the page returned HTTP 200 and rendered both “Daftar sekarang” and “Lihat struktur”. No fixture data was inserted.
- **Time policy:** persisted/API open_at/close_at and availability comparisons remain UTC, including inclusive open-window boundaries. UI presentation uses Asia/Jakarta; no manual +7 offset is added. Admin and public availability share the same backend predicate.

### Verification and UAT handoff

- pnpm api:migrate:up: PASS on configured development database bem_fmikom_dev; applied only 00010_registration_image_attachments.sql to version 10. No database/Compose configuration was added or changed.
- pnpm api:integration: PASS, exit 0. TestDatabaseIntegration ran (not skipped); empty dedicated bem_fmikom_test migrated UP through 00010, exercised migration 00010 DOWN and re-applied the full schema, ran existing auth/RBAC/division/media regressions and registration IMAGE/private access assertions, then cleaned the test schema. No development data was deleted.
- pnpm verify: PASS, exit 0 after correcting a test race and the landing Client Component boundary. Lint, format, strict typecheck, frontend 16 files / 70 tests, optimized Next.js build, Go formatting (70 files), go vet, Go tests, and Go build all passed. The parallel gallery regression now waits for the actual item action before interacting.
- pnpm audit --prod: PASS, no known vulnerabilities found.
- UAT services now use existing repository commands on API localhost:8080 and production frontend 127.0.0.1:3290. Final HTTP smoke: /healthz, /, and /admin/login each returned 200. The landing response contains a current registration CTA and structure control. These are route/data smoke checks, not authenticated, viewport, or visual/manual acceptance.
- git diff --check: PASS. git diff --quiet -- source_docs design_reference: clean. Phase 11H visual rows for content images, announcement modal, structure interaction, registration list/form/builder/submission detail, and mobile states remain RE-REVIEW REQUIRED. No final Phase 11 commit, deployment, or Phase 12 work was performed.

**Phase 11H implementation is ready for manual re-review. Overall Phase 11 remains open pending the user's visual and authenticated UAT; Phase 12 has not started.**

## Phase 11I — Public registration multipart/image regression (2026-10-13)

### Finding and correction

- Manual browser submission at `/pendaftaran/famget/form` returned HTTP 422 with `Bagian jawaban tidak valid`, despite valid fields, a valid PNG, and a browser-generated multipart boundary. The frontend sends the `answers` part as the JSON answer object itself; the multipart handler incorrectly parsed it as the JSON-only `{ "answers": { ... } }` envelope. This request-shape mismatch was the root cause; the file part, MIME, and boundary were valid.
- The multipart handler now wraps the direct answer object in the service input envelope while retaining strict JSON duplicate-key/trailing-data checks. JSON-only requests continue to use the documented envelope. Invalid answers and IMAGE uploads return safe structured validation errors; the public UI maps field errors to their inputs and maps the existing `DUPLICATE_SUBMISSION` API code to an actionable NIM message. The API contract's duplicate code remains unchanged.
- IMAGE remains optional when the registration schema says optional and required when marked required. Accepted image validation and private attachment storage continue to be enforced server-side. The image metadata lives in `registration_submission_files`; image bytes are not placed in `registration_submission_answers` or public media URLs.

### Verification evidence

- **MariaDB integration:** `pnpm api:integration` **PASS**, exit 0. `TestDatabaseIntegration` ran (not skipped) against the dedicated `bem_fmikom_test` database, migrated the empty schema through 00010, exercised DOWN/UP migration round-trip and integration flows, then cleaned the test schema. Its 26 named subtests passed; registration IMAGE multipart, required/optional IMAGE, private detail/access, and storage assertions also passed in the integration flow.
- **Browser UAT:** the actual public form submitted the requested test values and a verified PNG through the browser. The API returned a successful receipt and the UI showed success reference `#2`. A second attempt with the same NIM was rejected, the inline NIM error stated that it was already registered, and no second submission was accepted.
- **Development DB inspection:** receipt 2 exists as one submission row; it has one private `image/png` attachment record, 784 bytes, 320×240, with a 64-character SHA-256. The IMAGE answer row contains no binary/text value. The database inspection did not print applicant contact details or credentials. This is one local development UAT record and was retained.
- **Full verification:** `pnpm verify` **PASS**, exit 0 after the final code changes. Biome lint/format covered 123 frontend files; strict TypeScript passed; frontend 17 files / 73 tests passed; optimized Next.js build passed; Go format covered 70 files; `go vet`, all Go tests, and Go build passed.
- **Security/source checks:** `pnpm audit --prod` **PASS**, no known vulnerabilities. Final `git diff --check` **PASS** and `git diff --quiet -- source_docs design_reference` clean. No Phase 12 work, deployment, or final Phase 11 commit was performed.

### Manual review boundary

- The public form functional correction is ready for the user's review. The desktop and mobile Stitch comparisons, responsive layout, accessibility/keyboard path, admin authenticated submission detail/file preview, and all other pending Phase 11 rows still require manual re-review. Automated and local browser evidence does not close overall Phase 11.

**Phase 11I implementation is ready for public registration submission re-review. Overall Phase 11 remains open; Phase 12 has not started.**

## Phase 11 public registration grid and organization detail corrective review — 2026-10-13

### Findings and correction

- **Registration grid root cause:** equally important OPEN registration cards used an asymmetric `1.3fr / 0.7fr` desktop grid, implicitly giving the first card featured treatment. They now use equal grid tracks at desktop (`lg:grid-cols-2`), one column on mobile/tablet, `w-full`, and row stretching. No index-based span is used, so an odd last card remains a normal single track.
- **Landing organization behavior:** the previous “Lihat struktur” control expanded all leadership/division/member hierarchy inline and changed to “Sembunyikan struktur”. It is now a summary preview with up to four leadership cards and a Next.js Link to `/struktur`; no accordion state or full roster is included in the landing props. Navbar and footer use the same canonical route.
- **Public detail route:** `/struktur` is a dynamic server-rendered route using existing `GET /api/v1/public/organization`. It renders active-period metadata, leadership, divisions, division heads and members in API response order. The frontend public DTO omits NIM. No endpoint, migration, mock official roster, or production code copied from Stitch `code.html` was added.
- **Image and empty/error behavior:** member portraits use a fixed square/circular frame, centered `object-cover`, and the shared fallback when absent or broken. Missing active period, empty leadership/divisions, unassigned division head, empty division members, and API failure each have explicit neutral messages.
- **Visual reference gap:** `beranda_publik_bem_fmikom_unugha` and its mobile counterpart inform the landing preview and registration layout. No public organization detail Stitch screen is indexed. The admin `struktur_organisasi_pengurus_admin_bem` screen is not substituted as a public target; `/struktur` reuses public design tokens and remains subject to manual review.

### Verification and manual review boundary

- Corrective frontend tests: **PASS**, 4 files / 20 tests (organization detail, registration grid even/odd card semantics, landing preview navigation/no accordion, and navigation route consistency).
- Full `pnpm verify`: **PASS**, exit 0 on the final elevated local run after sandbox restrictions blocked MariaDB localhost and storage rename operations. Evidence: lint, format, strict typecheck; frontend **19 files / 79 tests**; optimized Next.js build with dynamic `/struktur`; Go fmt (**70 files**), vet, all Go tests, and build.
- Responsive browser review on the running local app: registration cards at 360, 390, 412, and 768 CSS px each use one normal-width column; at 1280 and 1440 they use two equal tracks (measured 577px each at 1280 and 632px each at 1440). Organization detail was opened at 1440×900 and 360×800, 390×800, and 412×800; `documentElement.scrollWidth` matched `clientWidth` at each size (no horizontal overflow). Landing registration and organization preview were opened at 390×844. Detail page, active data, person photos, and empty division/head states rendered. The detail mobile checks used 800px viewport height.
- Browser navigation: landing CTA reached `/struktur`, its heading rendered, and browser Back returned to `/`. Navbar/footer use the same `/struktur` destination. No “Sembunyikan struktur” text/control remains. Browser HTTP checks returned 200 for `/struktur` with `Struktur Organisasi` in HTML and 200 for API `/healthz`.
- These are agent-run browser observations, not user visual acceptance. All added/affected visual rows in `PHASE_11_VISUAL_MATRIX.md` remain **RE-REVIEW REQUIRED** for the user, especially landing registration and structure preview at 1440×900/390×844 and detail at 1440×900/360×800/390×844/412×915. No row is marked PASS from these checks.
- Documentation traceability is updated in `ROUTE_SCREEN_MATRIX.md`, `IMPLEMENTATION_MATRIX.md`, and `DECISIONS.md`. `source_docs/` and `design_reference/` remain evidence-only and untouched.

No Phase 12 work, deployment, or commit is performed. Overall Phase 11 remains open pending manual visual acceptance.

## Phase 11 — Public organization hierarchy and member-detail corrective review (2026-10-13)

### Architecture correction

- Public work-division visibility is now explicit metadata (`divisions.show_in_public_structure`), added by reversible migration `00011_public_structure_visibility.sql`. The one-time backfill hides canonical Faculty and BPH divisions from the public work chart; runtime filtering uses the flag and never infers visibility from division names. Admin division create/update can control this flag.
- Public aspiration destinations continue to include every active division, independent of the structure visibility flag. The local API smoke returned HTTP 200 and five aspiration division options after the migration.
- Current-period leadership is selected by position independently of assigned division. Leadership members assigned to BPH are returned once in the leadership hierarchy and are not duplicated in the work chart. Public response objects expose names, positions, photos, division names, and period only; database IDs and NIM are omitted.
- `/struktur` uses a distinct leadership tier under Ketua, then separately renders public work divisions. Division heads precede members; API order is preserved. The layout adds responsive connectors and bounded portrait frames. Person cards open a Radix accessible detail dialog with public-safe fields only, Escape dismissal, and focus restoration.
- Admin division editor exposes the new visibility control and explains that it does not affect Aspirasi eligibility.

### Verification evidence

- `pnpm api:migrate:up`: **PASS** on the configured local development database `bem_fmikom_dev`; applied `00011_public_structure_visibility.sql` to schema version 11.
- `pnpm api:integration`: **PASS**, exit 0. `TestDatabaseIntegration` ran against the dedicated test database, migrated the empty schema through 00011, verified the new migration's DOWN/UP and metadata backfill, and completed 26 named integration subtests. It did not skip.
- `pnpm verify`: **PASS**, exit 0. Biome lint and format, strict TypeScript, frontend **19 files / 80 tests**, optimized Next.js build, Go formatting (**70 files**), `go vet`, Go tests, and Go build passed. The focused public organization UI suite also passed **5/5**.
- Local UAT smoke after restarting the stale API process: `GET /healthz`, `GET /api/v1/public/organization`, `GET /api/v1/public/divisions`, and `GET http://127.0.0.1:3290/struktur` all returned HTTP 200. Runtime payload inspection confirmed public member keys omit ID/NIM and the aspirations endpoint still returns five divisions. API is available on port 8080 and web on port 3290.
- `source_docs/` and `design_reference/` remain unchanged. Automated/runtime checks do not constitute manual visual acceptance.

### Manual review boundary

- **READY FOR MANUAL RE-REVIEW:** `http://127.0.0.1:3290/struktur` and admin division visibility editor. Please review `/struktur` at 1440×900, 360×800, 390×844, and 412×915; inspect leadership/work hierarchy, portrait crop, card detail dialog (mouse and keyboard), Escape/focus return, and mobile overflow. Confirm that hiding a division from public structure leaves it available in Aspirasi destinations.
- Related landing preview and aspiration flow remain subject to their existing visual matrix rows. No affected visual row is marked accepted by this automated run.

**This corrective implementation is ready for manual re-review. Overall Phase 11 remains open; Phase 12, deployment, and commit have not started.**

## Phase 11 — Public `/struktur` division chart layout corrective (2026-10-13)

### Root cause and correction

- The public division list used an `xl:grid-cols-2` layout, so two unrelated divisions could appear side by side. Connector sizing was tied to that card layout, and an earlier leadership-to-work-area line had no corresponding division head.
- Each division now renders as its own sequential, full-width chart. The division head and member cards are centered; the member branch width follows the actual child group, with at most four members per desktop row. Mobile uses one centered vertical member list without horizontal branch lines. Missing-head and empty-member states render no invented leadership connectors. The orphan leadership-to-divisions line has been removed.
- Member cards still use the same fixed image frames and accessible public-safe profile dialog. API order, role labels, privacy-safe dialog fields, and the existing `/struktur` route and CTA destinations are unchanged.

### Verification and manual review boundary

- Focused `public-organization-page.test.tsx`: **PASS**, 14/14 tests, covering sequential odd division counts, one/two/three/five member grouping, API member order, empty and missing-head states, public division filtering, and the public-safe member detail dialog.
- Frontend lint, format check, and strict typecheck: **PASS** after the latest edits. Optimized Next.js production build: **PASS**.
- Local runtime smoke: `/struktur` returned HTTP 200 with its heading; the current build's CSS returned HTTP 200; API `/healthz` returned HTTP 200. Web is running at `http://127.0.0.1:3290/struktur` and API at port 8080. A stale Next.js process initially served CSS 500 after rebuild; it was replaced with the current build and CSS now returns 200.
- `git diff --check`: **PASS**. `git diff --quiet -- source_docs design_reference`: clean; those evidence folders remain unchanged.
- Source/design evidence remains read-only. No API/domain or database change was needed for this layout correction.
- **READY FOR MANUAL RE-REVIEW:** `http://127.0.0.1:3290/struktur`. Review at 1440×900, 1280×800, 412×915, 390×844, and 360×800. Confirm sequential divisions, centered cards/connectors, no orphan lines or mobile overflow, and preserved profile interaction/image framing.

The public structure visual rows remain **RE-REVIEW REQUIRED**. Overall Phase 11 remains open; Phase 12, deployment, and commit have not started.

## Phase 11 Finalization — 2026-10-14

### Final Status

**PHASE 11: PASS WITH DOCUMENTED NON-BLOCKING FOLLOW-UPS**

### Owner Manual Acceptance

The owner manually re-reviewed and **ACCEPTED** the current public `/struktur` layout and behavior: BPH/leadership hierarchy, sequential full-width division charts, division heads and members, valid connectors, fixed member image/card sizing, privacy-safe member detail interaction, and the overall page layout. The owner explicitly accepted the desktop presentation at **1440×900**. No separate manual geometry measurement was reported for **1280×800** or individual mobile viewports. Those device-specific measurements remain distinct from the owner's acceptance of the current design.

### Final Regression

- **Database/runtime preflight:** Docker 29.8.2 and Compose 5.5.1 reachable; existing `infra-mariadb-1` MariaDB container healthy. Existing `.env` points integration to the dedicated `bem_fmikom_test` database; no new Compose/database was created.
- **`pnpm api:integration`: PASS.** `TestDatabaseIntegration` ran against `bem_fmikom_test` (not skipped), passed 26 named subtests, applied migrations 00001–00011 from the empty schema, and exercised DOWN/UP migration round-trip and cleanup to version 0.
- **`pnpm verify`: PASS.** Web lint, format, strict typecheck, 19 frontend test files / 89 tests, optimized production build, Go formatting (70 files), `go vet`, Go tests, and Go build all passed.
- **`pnpm audit`: PASS.** No known vulnerabilities found.
- **Security/domain regression:** MariaDB integration passed auth lifecycle, SUPER_ADMIN/BPH cross-division policy, KEPALA_BIDANG own/other/no-division scope, media upload/format/privacy, CMS, organization serialization/visibility, aspiration destinations, registration IMAGE and private attachment scope, and private-document authorization assertions. Public organization objects omit internal IDs/NIM; hidden public-structure divisions remain independent of active aspiration destinations.
- **Source integrity:** `git diff --check` passed; `git diff --quiet -- source_docs design_reference` was clean. The evidence folders remain unchanged.

### Public Workflow Status

Public HTTP smoke passed with HTTP 200 for `/`, `/struktur`, `/berita`, `/pengumuman`, `/galeri`, `/pendaftaran`, and `/aspirasi`. The final smoke also confirmed API `/healthz` and the current build's `/struktur` CSS returned HTTP 200. HTTP smoke confirms route reachability, not visual or submission UAT.

### Admin Workflow Status

HTTP smoke returned HTTP 200 for `/admin/login`, `/admin`, `/admin/struktur`, `/admin/pengurus`, `/admin/divisi`, `/admin/berita`, `/admin/galeri`, `/admin/pengumuman`, `/admin/pendaftaran`, `/admin/arsip-dokumen`, `/admin/settings`, `/admin/users`, and `/admin/activity-logs`. Auth/RBAC and division-scope backend integration passed. These unauthenticated route smoke responses do not claim authenticated browser workflows or role-specific navigation were manually accepted.

### Organization Structure Status

`/struktur` is owner accepted for the current design. Leadership and division hierarchy, sequential chart layout, valid connectors, safe empty states, fixed image/card sizing, and the privacy-safe member dialog are covered by automated tests and/or database/API assertions. The landing CTA and navbar continue to target `/struktur`; Fakultas is not rendered as a BEM work unit and remains available as an aspiration destination when active in the domain model.

### Responsive / Visual Status

The owner acceptance is explicit at 1440×900 and applies to the current `/struktur` design. It is not a manual viewport measurement for 1280×800 or individual mobile sizes. Prior automated browser geometry found no horizontal overflow at 360×800, 390×800, and 412×800. The mobile 390×844 and 412×915 measurements were not separately user-reported. Other visual-matrix rows retain their individual states; no additional screens are marked user accepted by this finalization.

### Remaining Non-Blocking Findings

- Exact viewport-specific manual measurements beyond the explicit 1440×900 acceptance, including 1280×800 and individual mobile viewport geometry.
- Authenticated browser UAT/navigation for the broader SUPER_ADMIN, BPH, and KEPALA_BIDANG screens. Automated authorization/division tests pass; no user acceptance is inferred.
- Visual re-review of other unaccepted Stitch routes and responsive states; their per-row BLOCKED/RE-REVIEW REQUIRED statuses remain unchanged.
- Broader governance questions G-15 (activity-log retention/event vocabulary), G-17 (announcement archive visibility), and G-16 (all-route responsive sign-off) remain documented follow-ups. They do not correspond to a failing Phase 11 integration/build/security gate.

### Remaining True Blockers

None found in the executed Phase 11 integration, security, build, source-integrity, or HTTP route gates. No critical/major runtime, security, data-integrity, or broken-route defect remains open.

### Git Status

The finalization commit contains Phase 11 application, tests, and documentation only. It excludes environment secrets, database/media runtime data, temporary screenshots, and generated `apps/web/tsconfig.tsbuildinfo`. No push was performed.

### Phase Boundary

**Phase 12 has not started.** No deployment or production-environment change was made.
