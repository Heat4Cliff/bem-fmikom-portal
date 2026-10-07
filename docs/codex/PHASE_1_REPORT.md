# Phase 1 Report

## Scope Implemented

Completed repository foundation and design-system scaffolding only. No business feature, database schema/migration, authentication, JWT issuance, role enforcement, division enforcement, or contract endpoint was implemented. Phase 2 and later plan checkboxes remain untouched.

## Repository Structure

- `apps/web`: Next.js App Router, strict TypeScript, semantic Tailwind tokens, public shell, admin shell, accessible reusable primitives, and Vitest tests.
- `apps/api`: Gin server, typed environment config, HTTP response foundation, request ID/logging/recovery middleware, health handler, and Go tests.
- `packages/contracts`: documented reservation for resolved OpenAPI/shared client contracts.
- `scripts`: cross-platform Go formatting check.
- `infra`: documented reservation; no database or storage provisioning.
- `docs/codex`: Phase 0 evidence plus this report.
- `source_docs` and `design_reference`: retained as read-only evidence.

## Runtime / Tool Versions

| Tool | Project/CI pin | Local version used |
|---|---:|---:|
| Node.js | 24.19.0 | 25.2.1 |
| pnpm | 11.25.0 | 11.25.0 |
| Go | 1.26.1 | 1.26.1 |
| Next.js | 16.4.0 | 16.4.0 |
| React / React DOM | 19.2.8 | 19.2.8 |
| TypeScript | 6.0.3 | 6.0.3 |
| Tailwind CSS | 4.3.3 | 4.3.3 |
| Biome | 2.5.15 | 2.5.15 |
| Gin | 1.12.0 | 1.12.0 |

CI reads the project pins. Local verification ran under Node 25.2.1, which is newer than but does not match the declared Node 24.19.0 pin.

## Design System Implemented

- Semantic CSS variables and Tailwind v4 theme mappings for navy/gold brand colors, surfaces, borders, body/heading text, and success/warning/danger/info states.
- Inter body/control font and Plus Jakarta Sans heading font, bundled locally and loaded with `next/font/local`.
- Radius, shadow, spacing, container, and responsive breakpoint tokens; reduced-motion and visible focus styles.
- Brand and layout patterns derived from `docs/design/DESIGN_SYSTEM.md` and the Stitch design reference. `screen.png` remains the visual authority; `code.html` was not copied into production code.

## Frontend Foundation

- App Router route groups for public and `/admin` shells.
- Core controls and display primitives: Button variants/states, form controls, Badge, Card, Alert, Separator, Avatar, Breadcrumb, PageHeader, Skeleton, Spinner, Empty/Error states, Pagination, SearchInput, FilterTrigger, FileInput, and table primitives.
- Radix-backed Dialog, confirmation dialog, Drawer, Dropdown Menu, Tooltip, and Toast.
- Responsive public header/drawer, mobile bottom navigation, footer/container, and admin sidebar/topbar/mobile drawer/content/breadcrumb. Admin shell accepts a typed role/division navigation context but performs no authorization.
- Landing and admin pages contain only explicit Phase 1 foundation placeholders, not business data.

## Backend Foundation

- Go module and Gin router with structured JSON request logs, generated or validated `X-Request-ID`, panic recovery, JSON 404, shared response envelope, and operational `GET /healthz`.
- Typed configuration for app environment/port, frontend URL, and future database/JWT/storage placeholders. Unused integration values do not prevent the Phase 1 server from starting.
- HTTP server timeouts and graceful shutdown.
- `/healthz` is operational infrastructure and is not a business endpoint in the API Contract.

## CI / Tooling

- pnpm workspace with exact direct package versions and lockfile.
- Root scripts for frontend lint/format/typecheck/test/build, Go fmt/vet/test/build, and aggregate `pnpm verify`.
- GitHub Actions workflow runs frozen pnpm install and the aggregate verification command without production secrets.
- README covers architecture, setup, environment templates, commands, authority of source documents, and phase status.

## Files Added / Changed

- Root: `.env.example`, `.gitignore`, `.node-version`, `.go-version`, `.npmrc`, `package.json`, `pnpm-workspace.yaml`, `pnpm-lock.yaml`, `README.md`.
- CI/tooling: `.github/workflows/ci.yml`, `scripts/gofmt-check.mjs`.
- Frontend: `apps/web/` manifests/configuration, local font dependencies, tokens, route shells, components, and tests. Next.js generated `apps/web/AGENTS.md` during local development.
- Backend: `apps/api/` Go module, env template, config, router, handler, middleware, and tests.
- Supporting docs: `packages/contracts/README.md`, `infra/README.md`, this report, and the Phase 1 addition to `docs/codex/DECISIONS.md`.
- `PLANS.md`: only verified Phase 1 items were marked complete. Phase 2+ is unchanged.

## Decisions Made

- pnpm workspace and exact resolved dependency versions.
- TypeScript 6.0.3 and Biome 2.5.15: the initial latest resolution selected TypeScript 7, which typescript-eslint did not support. ESLint 9 was out of support, while ESLint 10 fell outside the peer ranges of Next's React/import/a11y plugins. Biome provides lint and formatting without those conflicts; `pnpm peers check` is clean.
- Local Fontsource fonts with Next local font loading, so builds do not require access to Google Fonts.
- Production Next build uses Webpack after a Turbopack Windows worker process failed under the restricted environment. The default Next dev command remains unchanged.
- Full baseline commit is `eaccb9b chore: establish audited project baseline`; Phase 1 changes are uncommitted.

## Gaps Carried Forward

All Phase 0 gaps G-01 through G-18 remain open and unchanged. Contract reconciliation, role matrix, division scope, registration duplicate/version policies, document retention, token rotation, rate limits, activity-log retention, analytics, and missing screen coverage remain later-phase decisions. No endpoint or authorization behavior was inferred to fill a gap.

## Verification

### STATIC VERIFICATION

Command: `pnpm install --frozen-lockfile`

Result: PASS. Lockfile is up to date and supply-chain checks passed. An offline repeat could not check cached registry metadata for `@babel/helpers`, so the final frozen install used the configured registry.

Command: `pnpm verify`

Result: PASS on local Node 25.2.1 / pnpm 11.25.0 / Go 1.26.1. This covered frontend lint and formatting, typecheck, tests, production build, Go formatting, vet, tests, and build.

### AUTOMATED TEST

Command: `pnpm test`; `go -C apps/api test ./...`

Result: PASS. Vitest: 3 tests passed. Go config and router tests passed; packages without isolated behavior report no test files.

### INTEGRATION TEST

NOT APPLICABLE. Product APIs and database integrations are intentionally outside Phase 1.

### VISUAL REVIEW

MANUAL VERIFICATION REQUIRED for final responsive acceptance. Actual browser captures of the public landing shell and admin shell at the available desktop viewport were compared with the corresponding Stitch home/dashboard references. Palette, typography, desktop navigation, admin sidebar width/density, and shell hierarchy are aligned at the foundation level. The implemented pages are placeholders and do not reproduce full Stitch feature screens. A live 390px viewport capture was unavailable in the connected browser, so mobile screenshot parity has not been claimed; responsive breakpoints and off-canvas/mobile navigation are present in code.

The Phase 0 inventory contains 37 Stitch screen images and 37 paired `code.html` references, including 8 explicitly mobile-labeled references. The Phase 1 comparison covers only the public/admin shell, not all 37 feature screens.

### MANUAL UAT

NOT APPLICABLE. There is no authentication or product workflow in this phase.

## Security Check

- Environment templates contain placeholders only; `DATABASE_URL` and `JWT_SECRET` are blank.
- `.env*` files are ignored except `.env.example` templates.
- Request ID input is restricted to safe correlation characters; request logs omit headers, cookies, and bodies; panic output does not expose panic values.
- No auth, RBAC, division scope, database, or private file workflow is active yet. These are not represented as complete.
- `source_docs/` and `design_reference/` show no Git diff from the Phase 0 baseline.
- No Phase 2 feature or migration was added.

## Known Issues

- Local Node was 25.2.1 while CI/project pins Node 24.19.0. The full local quality gate passed on the newer runtime; the pinned Node 24 workflow has not run remotely in GitHub Actions.
- Turbopack production build failed to spawn a worker in this Windows sandbox; the documented Next `--webpack` build path passed. CI uses the same build script.
- Mobile visual screenshot parity remains for manual verification as stated above.

## Phase Result

PHASE 1: PASS
