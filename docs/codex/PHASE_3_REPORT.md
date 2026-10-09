# Phase 3 — Public Website Foundation & Landing Page

**Current status: PASS** — Phase 3B below supersedes the initial BLOCKED verification attempt.

## Scope

Implemented the public presentation foundation and landing page only. No CMS, content CRUD, product-domain migration, aspiration submission, registration submission, or Phase 4 work was started. `source_docs/` and `design_reference/` were not modified.

## Implemented

- Replaced the public home placeholder with reusable Next.js/TypeScript section components for hero, overview, vision/mission, organization, news, announcements, gallery, registration, aspirations, contact, and footer.
- Added responsive public navigation, mobile drawer, bottom navigation, skip link, and semantic main-content focus target.
- Added announcement preview dialog behavior and a navigation-only `/aspirasi` placeholder. No write form or API call is presented as working.
- Added typed public-home view data. Production defaults have no dynamic content or sample statistics. Clearly labeled visual fixtures require development mode or the server-only `BEM_DEMO_CONTENT=true` flag.
- Added metadata, optional canonical origin, `robots.txt`, and a root-only sitemap. The absolute origin remains unset until `NEXT_PUBLIC_SITE_URL` is approved.
- Used token-based CSS artwork and the existing text brand mark because the source package has no approved production photos/logo. No Stitch HTML or remote Stitch image URL was copied into production code.
- Updated README, decisions, gaps, implementation matrix, and route/screen matrix with Phase 3 scope and evidence boundaries.

## Stitch references

Reviewed `DESIGN.md` and the desktop/mobile `screen.png` plus corresponding `code.html` for the public homepage:

- `design_reference/stitch_bem_fmikom_unugha_portal/modern_academic_horizon/DESIGN.md`
- `design_reference/stitch_bem_fmikom_unugha_portal/beranda_publik_bem_fmikom_unugha/screen.png`
- `design_reference/stitch_bem_fmikom_unugha_portal/beranda_publik_mobile_bem_fmikom_unugha/screen.png`

`screen.png` was treated as the visual target; `code.html` was used only to understand layout/style. Desktop and mobile references include photo-led artwork and official-looking content that is not available as approved source data. The implementation uses visibly labeled placeholders rather than reproducing those photos or inventing official facts.

## Phase 3A verification evidence (initial gate)

| Check | Result | Evidence / limitation |
|---|---|---|
| Web lint | PASS | `pnpm --filter @bem-fmikom/web lint` |
| Web format | PASS | `pnpm --filter @bem-fmikom/web format:check` |
| Web typecheck | PASS | `pnpm --filter @bem-fmikom/web typecheck` |
| Frontend tests | PASS | 4 files, 19 tests |
| Production build | PASS | Included in final `pnpm verify`; routes `/`, `/aspirasi`, `/admin`, `/admin/login`, `robots.txt`, and `sitemap.xml` built successfully. |
| Go fmt / vet / unit tests / build | PASS | Included in final `pnpm verify`; Go tests completed successfully. |
| Full verification | PASS | `pnpm verify` passed. It required the authorized broader local filesystem execution because sandboxed Next.js build workers intermittently received `EPERM` while creating `.next` output directories. No project configuration workaround was retained. |
| Docker client | PASS | `docker --version`: Docker 29.8.2; `docker compose version`: v5.5.1. |
| MariaDB daemon | BLOCKED | `docker info` could not open `npipe:////./pipe/dockerDesktopLinuxEngine`; `docker desktop start` could not find the Docker Desktop registry key. Retrying `docker info` with broader local execution produced the same missing-pipe error. |
| Database integration/auth regression | BLOCKED | `pnpm api:integration` did not skip. `TestDatabaseIntegration` failed connecting to `127.0.0.1:3306` with Windows `connectex: An attempt was made to access a socket in a way forbidden by its access permissions`. No migration or DB-backed auth result is claimed for this run. |
| Browser desktop | PASS | Production preview inspected at 1440×900. Document width 1425px; no horizontal overflow. Hero and first statistics row rendered. |
| Browser mobile | PASS | Production preview inspected at 390×844. No horizontal overflow. Mobile menu opened and closed with Escape; fixed bottom navigation remained visible. |
| Announcement interaction | PASS | Demo fixture's preview button opened an accessible dialog; close action worked. This verifies preview UI only, not an API-backed announcement. |
| Unauthenticated admin route | PASS | `/admin` redirected to `/admin/login`. No credential-based or role-based UAT was run because DB/auth integration could not be repeated. |
| Production fixture isolation | PASS | Preview started without `BEM_DEMO_CONTENT`; no demo banner, `450+` sample statistic, or news fixture cards appeared. |
| Read-only evidence | PASS | `git diff --quiet -- source_docs design_reference` returned exit code 0. |

## Initial verification gaps and risks (historical; see Phase 3B below)

1. MariaDB/Docker runtime is unavailable in this environment, so the required database integration and auth regression evidence is missing. This is the critical phase gate and the reason this report is BLOCKED.
2. SUPER_ADMIN/BPH/KEPALA_BIDANG authenticated browser UAT and division behavior were not re-run in this Phase 3 session. Previous Phase 2 evidence remains in `PHASE_2_REPORT.md` but does not substitute for the blocked requested regression run.
3. No public content endpoint or product-domain schema is currently implemented. Production news, announcement, gallery, registration, organization-period, vision/mission, contact, and statistics areas therefore show empty/neutral states; the visual fixtures are strictly for local review.
4. Official logo, photos, contact/social links, and public origin are not in the source package. These prevent full image/content parity with Stitch and must be supplied/configured before release.
5. The aspiration route is informational only; submission and storage remain future work.

## Phase 3 disposition

### Source read

Used the Phase 3 authority and design references already recorded in the source audit, plus the public home Stitch token file and desktop/mobile screen references listed above. The exact implementation decisions and unresolved content/API gaps are recorded in `DECISIONS.md`, `GAPS_AND_QUESTIONS.md`, `IMPLEMENTATION_MATRIX.md`, and `ROUTE_SCREEN_MATRIX.md`.

### Technical decisions fixed

- Public sections use a typed view model and reusable components; no fake network client or unimplemented API calls were added.
- Production dynamic content stays empty until an approved backend source exists; local visual fixtures are visibly labeled and runtime-gated.
- Section links use stable anchors; `/aspirasi` is an explicit placeholder.
- Canonical URLs are emitted only when a valid public origin is configured.
- Placeholder artwork uses design tokens and local CSS; approved imagery/identity assets remain outstanding.

### Recommendation for next phase

Restore a working Docker Desktop daemon, confirm the `bem_fmikom_test` database is empty and reachable, then re-run `pnpm api:integration` and the auth/RBAC regression suite without skips. Once that gate passes, plan the public content/domain API needed to populate the landing page and confirm official vision/mission, period, contacts, social links, and approved media. Do not treat the current empty-state presentation as a complete content integration.

At this initial verification, the database/auth gate did not pass, so no Phase 3 PASS or implementation commit was claimed. That result is retained as history and is superseded by the Phase 3B evidence below. Phase 4 was not started.

**Initial status (superseded): PHASE 3: BLOCKED**

## Phase 3B — Database/Auth Regression Verification

This verification supersedes the initial database/auth BLOCKED result above. It used the repository's existing `infra/docker-compose.dev.yml`, the existing `pnpm db:up` command, and only the dedicated `bem_fmikom_test` database. No new MariaDB configuration was created. The manually created failed container mentioned in the handoff was not used.

### Docker and database runtime

- `docker --version`: **Docker 29.8.2**, build `7fc2dff`.
- `docker compose version`: **Docker Compose v5.5.1**.
- The restricted shell initially returned `permission denied` for `npipe:////./pipe/dockerDesktopLinuxEngine`; repeating `docker info` and `docker ps` with broader authorized local execution reached Docker Desktop Engine **29.8.2**.
- `pnpm db:up`: **PASS**, using the existing root script `docker compose --env-file .env -f infra/docker-compose.dev.yml up -d`; Compose reported `infra-mariadb-1 Running`.
- `docker compose ... ps` and Docker inspect health: container `infra-mariadb-1`, image `mariadb:11.4`, state **healthy**, bound only to `127.0.0.1:3306`.
- Authenticated local database query returned MariaDB **11.4.13-MariaDB-ubu2404** and confirmed `bem_fmikom_test` exists. No password or local `.env` value was printed.

### Database/auth integration

- Command: `pnpm api:integration` (`go -C apps/api test -v ./... -run TestDatabaseIntegration -count=1`).
- Result: **PASS**, exit code 0. Output included `=== RUN TestDatabaseIntegration`; it was not skipped.
- The existing test exercised the foundation migration lifecycle against the empty test schema, auth bootstrap and credential cases, inactive-user denial, database-current `/auth/me` state, refresh rotation/reuse and expiry, logout/idempotency, activity rows and token-storage/logging protections. The six explicit division/RBAC subtests all passed:
  - SUPER_ADMIN cross-division.
  - BPH cross-division.
  - KEPALA_BIDANG own division allowed.
  - KEPALA_BIDANG other division denied (both division directions).
  - KEPALA_BIDANG without a division denied.
- After browser UAT, the temporary test accounts and seed data were removed by the existing `pnpm api:migrate:down` command pointed to `bem_fmikom_test` only.

### Full verification

- Command: `pnpm verify`, run with broader local filesystem execution because the restricted Windows sandbox intermittently denied Next.js `.next` output creation.
- Result: **PASS**, exit code 0. Frontend lint, format, typecheck, **4 files / 19 tests**, Next.js production build, Go formatting (**34 files**), Go vet, Go tests, and Go build passed.
- Node pin check: project specifies Node.js **24.19.0**, but that runtime was not available on PATH or via `nvm`; verification used **Node.js 25.2.1**. This is recorded as an environment mismatch, not hidden as a Node 24 result.

### Public landing smoke

- Production/default `/`: **PASS**. The landing heading rendered; the 390 px viewport had no horizontal overflow. `documentWidth` was 375 px versus `innerWidth` 390 px; no element exceeded the viewport.
- Production/default content isolation: **PASS**. No demo banner, sample statistics, or news fixture cards appeared.
- Desktop 1440×900 public navigation: clicking `Tentang` navigated to `/#tentang`; no horizontal overflow.
- Mobile 390×844 navigation: drawer exposed the public section links and closed with Escape; the fixed bottom navigation remained available.
- Demo-mode announcement: **PASS**. With the existing `BEM_DEMO_CONTENT=true` runtime flag, the labeled example announcement opened an accessible dialog and the close button dismissed it. The preview was then restarted without the flag and production/default isolation was rechecked.

### Admin auth smoke

- The API and production web preview were run with the API `DATABASE_URL` overridden for that process to `bem_fmikom_test`; repository `.env` files were not modified.
- `/admin` without a session redirected to `/admin/login`: **PASS**.
- A temporary SUPER_ADMIN created by the repository's interactive bootstrap CLI logged in to `/admin`; Dashboard, Manajemen Pengguna, and Activity Log were visible: **PASS**.
- Reload retained the authenticated dashboard session through the refresh-cookie flow: **PASS**. Logout returned to `/admin/login`; another reload stayed unauthenticated: **PASS**.
- A temporary KEPALA_BIDANG test principal in `bem_fmikom_test` showed `Scope: Phase3B Akademik` after login and did not show Landing Settings, User Management, or Activity Log: **PASS**. The role/division authorization matrix also passed in the integration suite.
- Temporary accounts, period, and division existed only in `bem_fmikom_test` and were removed when its existing migration was rolled down after UAT.

### Source evidence and checklist

- `git diff --quiet -- source_docs design_reference`: **PASS**, exit code 0. No source evidence was changed.
- Phase 3 checklist items for responsive public layout, SEO baseline, and desktop/mobile visual review are checked. Landing settings/content integration and active-period organization data remain unchecked because their API/domain data is not part of this phase's implemented scope.
- Phase 4 and later checklist items were not changed.

All Phase 3B regression gates requested in the handoff passed. Phase 4 was not started.

**PHASE 3: PASS**
