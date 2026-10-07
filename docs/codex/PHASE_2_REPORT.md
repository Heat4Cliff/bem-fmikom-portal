# Phase 2 Report

## Scope Implemented

Implemented the database foundation, admin authentication lifecycle, current-database RBAC context, division-scope policy primitives, login screen, auth client/provider, admin route guard, and role-aware navigation. No product CRUD or Phase 3 work was started.

## Database Schema

The foundation migration creates only `periods`, `divisions`, `users`, `refresh_sessions`, and `activity_logs`. It includes foreign keys, relevant indexes, normalized unique email, period-scoped division-name uniqueness, and a generated unique active-period slot to enforce at most one active period. `KEPALA_BIDANG` division assignment and role/division consistency are also validated by application services. Timestamps persist in UTC. No product-domain tables were added. Live MariaDB verification and an UP/DOWN/UP round trip are recorded in the Phase 2B results below.

## Migrations

Goose v3.28.0 runs embedded, versioned SQL with UP and DOWN operations. The migration CLI supports up, one-step down, and down-to-zero. The local MariaDB 11.4 Compose service, test initialization, root scripts, and CI MariaDB service are configured. Migration execution and schema assertions against MariaDB 11.4 are verified below.

## Authentication Architecture

The Go API uses `database/sql` with `go-sql-driver/mysql`, explicit repositories, service logic, Gin handlers, and composable middleware. Implemented routes are `POST /api/v1/auth/login`, `/refresh`, `/logout`, `GET /api/v1/auth/me`, plus the existing `/healthz`. Login normalizes email and returns a generic credential error. Protected requests verify the access JWT and then reload the active user, role, and division from the database. Inactive accounts are denied at login, refresh, and protected-request lookup.

Access JWTs use HS256, `iss/sub/iat/exp/jti`, a configurable 15-minute default, and an environment secret of at least 32 bytes; insecure production placeholders fail fast. Passwords use bcrypt with cost 12 by default, a 12-character minimum, and a 72-byte maximum. The interactive `pnpm api:create-admin` command creates an active `SUPER_ADMIN` without a committed or default credential.

## Refresh Token Lifecycle

Refresh tokens are opaque 32-byte cryptographically random values. The database stores only lowercase SHA-256 hashes. Refresh validates expiry and revocation, rotates the token within its family, and returns a new access token and cookie. Reuse of a revoked token denies access, revokes the active family, and writes a structured security event without logging token material. Logout is idempotent, revokes the related family when identifiable, records LOGOUT only for a known actor, and clears the cookie.

## RBAC Implementation

The database is the authority for current account status, role, and division. `RequireAuth`, `RequireRoles(...)`, and policy helpers are separate and composable. `SUPER_ADMIN` has full access and no division; `BPH` has cross-division operational scope and no user/activity-log privileges; `KEPALA_BIDANG` must have one division. The frontend navigation hides user management and activity logs from non-super-admin roles, allows landing settings for BPH, and hides them for KEPALA_BIDANG. Navigation visibility is only a UX aid; backend authorization remains authoritative.

## Division Scope

Reusable policy helpers deny a KEPALA_BIDANG principal without a current division and allow a division-scoped resource only when its division matches the current database-backed principal. SUPER_ADMIN bypasses the division restriction and BPH remains cross-division. No resource CRUD endpoints were added in this phase.

## Frontend Authentication

Added `/admin/login` using the existing navy/amber tokens, fonts, and reusable form primitives. It includes email/password input, show/hide password, validation, server error, and loading states. Status: **DESIGN SYSTEM CONSISTENT**; there is **NO DEDICATED STITCH REFERENCE** for login, so no pixel-parity claim is made.

The typed API client keeps access JWT only in module memory, includes credentials for cookies, supports one refresh retry after 401, and clears auth state on refresh failure. `AuthProvider` bootstraps with refresh then `/me`; the admin guard shows a loading state before redirecting unauthenticated users. The sidebar is role-aware.

## Security Controls

- Configurable HttpOnly refresh cookie scoped to `/api/v1/auth`; development defaults to Secure=false and SameSite=Lax, production defaults Secure=true, and SameSite=None requires Secure.
- Exact comma-separated `FRONTEND_ORIGINS` allow-list with credentialed CORS; state-changing auth requests require a trusted Origin.
- Configurable per-IP in-memory login rate limiter; Gin proxy trust is disabled by default. This is single-instance protection, not distributed limiting.
- No plaintext password, raw refresh token, access token, or JWT secret is intentionally logged or stored in the database.
- Test database integration requires a dedicated database whose name ends with `_test`; tests reset and clean that database.

## Files Added / Changed

- Backend configuration, SQL database connection and Goose migration, domain types, repositories, security helpers, auth services, Gin auth handlers/middleware/router, admin bootstrap CLI, migration CLI, unit/integration tests, and Go module dependencies under `apps/api/`.
- Frontend login route, API client, auth provider/guard/helpers, auth tests, and role-aware admin shell/navigation under `apps/web/`.
- MariaDB Compose and initialization files under `infra/`; CI database service and root DB/auth commands in `.github/workflows/ci.yml` and `package.json`.
- Updated `.env.example`, `apps/api/.env.example`, `.gitignore`, `README.md`, `docs/codex/DECISIONS.md`, `docs/codex/GAPS_AND_QUESTIONS.md`, and `docs/codex/IMPLEMENTATION_MATRIX.md`.
- Added this report. Phase 1 snapshot is commit `6bdf2ba` (`chore: complete phase 1 foundation`); Phase 2 changes remain a separate working-tree change set.

## Decisions Made

- Followed the fixed Phase 2 five-table schema scope, despite the broader generic table list in `PLANS.md`; no future product tables were created.
- Used Goose with embedded SQL, MySQL-compatible generated active-period uniqueness, and the existing `DATABASE_URL` DSN configuration.
- Implemented access expiry at 900 seconds. This follows the explicit Phase 2 decision; PRD v1.4 login example says 1800 seconds. The discrepancy is recorded in `DECISIONS.md`.
- Used full `/api/v1/auth/*` route paths, consistent with the Phase 2 route list and API contract prefix.
- Used structured security logging for refresh reuse, separate from product activity rows. LOGIN and known-actor LOGOUT are the Phase 2 activity events.

## Phase 0 Gaps Resolved

- G-13 is resolved for the authentication foundation: token lifetimes, opaque refresh rotation/reuse behavior, cookie defaults, trusted-origin handling, database-current identity, and auth route prefix are fixed.
- G-07 remains a visual evidence limitation: login has no dedicated Stitch screen. The implementation uses the established design system without claiming pixel parity.

## Gaps Carried Forward

Other Phase 0 gaps remain open for their corresponding phases, including gallery/category route reconciliation, analytics source/retention, registration duplicate and schema-snapshot rules, document deletion/retention, 403/404 disclosure, public write rate limits, mobile screen evidence, announcement visibility, and upload configuration. The Phase 2B section below records the live database and auth UAT evidence.

## Verification

### STATIC VERIFICATION

Command: `pnpm verify` on Node.js 24.19.0; the command runs frontend lint, format check, typecheck, tests, production build, Go formatting verification, `go vet`, Go tests, and Go build.

Result: **PASS** (exit code 0). Next.js production build completed and prerendered `/`, `/_not-found`, `/admin`, and `/admin/login`. The final verification ran with workspace permissions elevated after the restricted attempt received Windows `EPERM` while creating `.next` output directories.

### UNIT TEST

Command: Included in `pnpm verify` (`pnpm test` and `go test ./...`).

Result: **PASS**. Frontend suite: 2 files, 10 tests passed. Go unit/package tests passed. The database-dependent test was skipped and is reported separately below.

### DATABASE INTEGRATION TEST (initial attempt)

Command: `go test -v ./internal/http/router -run '^TestDatabaseIntegration$'` (also exposed as `pnpm api:integration`).

Result at the initial Phase 2 verification: **SKIPPED / NOT VERIFIED** because database runtime was unavailable. This earlier result is superseded by the passing runtime retry below.

### FRONTEND AUTH TEST

Command: `pnpm test` (included in the successful `pnpm verify`).

Result: **PASS**. Covers login validation/loading/error, auth bootstrap, unauthorized redirect, one refresh retry, logout clearing memory state, and role-aware navigation. Also confirms BPH sees landing settings and KEPALA_BIDANG does not.

### SECURITY TEST (initial attempt)

Command: Unit/security packages included in `go test ./...`; frontend auth client/provider suite included in `pnpm test`.

Result at the initial Phase 2 verification: **UNIT CHECKS PASS; DATABASE SECURITY INTEGRATION UNVERIFIED**. This earlier result is superseded by the passing integration evidence below.

### VISUAL REVIEW

Result at the initial verification: **DESIGN SYSTEM CONSISTENT** by source-level review of login and auth-loading/navigation components against the existing tokens and primitives. The later browser run verified interaction states and captured the local login view, but no dedicated Stitch login reference exists and no pixel-parity review against a login screen can be claimed.

### MANUAL UAT (initial attempt)

Result at the initial Phase 2 verification: **NOT RUN**. This earlier result is superseded by the live local browser UAT evidence below.

## Known Issues at Initial Verification (superseded)

- The initial local build attempt hit a Windows filesystem `EPERM` while writing `.next`; the aggregate verification subsequently passed after workspace permission elevation.
- Browser preview and manual interaction checks were unavailable during the initial verification; the local browser UAT was later completed below.
- The initial Phase 2 auth and policy checklist items were left unchecked until database-backed verification; they are now updated below.
- Remote GitHub Actions CI was not run from this workspace.

## Security Notes

No production credential or generated admin password was added. Local database values are placeholders in environment examples. Keep root/API `.env` files untracked and use a dedicated disposable `_test` database for the destructive integration suite. The in-memory limiter does not coordinate across API instances.

## Initial Phase Result (superseded by Phase 2B below)

**PHASE 2: IMPLEMENTATION COMPLETE — VERIFICATION REQUIRED** at the time of the original report. The final result is recorded at the end of this report.

## Phase 2B Verification — 2026-10-07 Runtime Retry

This section supersedes the earlier BLOCKED attempt above. All database work used the existing `infra/docker-compose.dev.yml` and the dedicated `bem_fmikom_test` database. The ignored local `.env` contains workstation-only generated secrets and is not part of the repository change set.

### Pre-flight and runtime

- `docker --version`: Docker version **29.8.2**.
- `docker compose version`: Docker Compose version **v5.5.1**.
- `docker info`: daemon active; Docker Desktop server version **29.8.2**. The regular workspace shell could not open the Desktop named pipe, so runtime inspection and Docker commands used the available elevated host access.
- Existing command discovery confirmed `pnpm db:up`, `pnpm db:down`, and `pnpm api:migrate:*`; no new Compose file was introduced for this retry. `db:up`/`db:down` now pass the ignored root `.env` explicitly because Compose did not load it from the shell's launch directory.
- MariaDB container health: **healthy**. TCP 127.0.0.1:3306 accepted a database query. Server version: **11.4.13-MariaDB-ubu2404**. Dedicated database `bem_fmikom_test` existed and was the only integration target.

### Schema and migration

- `pnpm api:migrate:up` on the empty `_test` schema: **PASS**.
- Queried `information_schema`: all five expected InnoDB tables exist (`periods`, `divisions`, `users`, `refresh_sessions`, `activity_logs`); inspected foreign keys, expected indexes/unique keys, and UTC `DATETIME(6)` timestamps.
- Single-active-period invariant: inserting a second `ACTIVE` row was rejected by `uq_periods_single_active`; archiving the first then allowed a new active row.
- Round trip: `pnpm api:migrate:down` (the only migration, equivalent to down-to-zero), confirmed zero foundation tables, then `pnpm api:migrate:up`: **PASS**, version 1 reapplied deterministically.
- After the final integration suite reset the test schema, `pnpm api:migrate:up` reapplied the schema; the MariaDB container remains healthy with the five-table schema present.

### Database integration and auth security

- `pnpm api:integration` sets up a real MariaDB database, runs the embedded migration and auth flow, and cleans the `_test` schema to zero. Final rerun result: **PASS, no integration skip**; one parent integration test and all six named division-matrix subtests passed. Go's `no tests to run` notices for packages outside the selected integration test are expected; the router output shows `=== RUN TestDatabaseIntegration` and its six PASS cases.
- Integration verified bootstrap SUPER_ADMIN and bcrypt hash, duplicate-admin conflict, successful login, generic wrong-password/unknown-email denial, inactive-user denial, database-current role/status/division through `/auth/me`, last-login update, refresh rotation, old-token reuse denial with active-family revocation and no new access token, expired refresh denial, logout and idempotent logout, LOGIN/LOGOUT activity rows, and no raw refresh token/password/JWT secret in stored rows/logs.
- Token storage inspection: `refresh_sessions` contains SHA-256 hashes only; no raw refresh token, access JWT, JWT secret, or plaintext password is stored. The password column contains only a bcrypt hash.
- RBAC: SUPER_ADMIN and BPH cross-division allowed; SUPER_ADMIN-only route denies BPH; KEPALA_BIDANG must have exactly one current division. Division policy matrix passed: SUPER_ADMIN cross-division, BPH cross-division, KB-A own-division allowed, KB-A other-division denied, KB-B other-division denied, and KB without division denied. No product resource CRUD endpoints exist in Phase 2, so HTTP-level IDOR tests remain for the phases that introduce those endpoints.
- Rate limit and origins: per-IP login threshold returned 429; trusted Origin received credentialed CORS allow-origin; untrusted Origin was denied without CORS credentials/allow-origin. Cookie assertions verified HttpOnly, path scope, development Secure=false, production Secure=true, and SameSite=None only with Secure=true.

### Full verification

- `pnpm verify`: **PASS**, exit code 0 on Node.js **24.19.0**. Lint, formatting, typecheck, frontend tests (**2 files / 10 tests**), frontend production build, Go fmt (34 files), Go vet, Go tests (including the real DB integration), and Go build all passed. The final full verification was run with workspace permissions elevated after Windows denied `.next` directory creation in the restricted attempt.

### Live browser auth UAT

- Local Go API (`http://localhost:8080`) and Next.js frontend (`http://localhost:3000`) were started against `bem_fmikom_test`.
- `/admin` without a session redirected to `/admin/login`: **PASS**.
- Wrong credentials showed a generic error; SUPER_ADMIN login reached the dashboard: **PASS**.
- SUPER_ADMIN could see User Management and Activity Log. Reload restored the authenticated session through the refresh cookie: **PASS**.
- With the same test principal reassigned only inside `_test`, BPH saw Landing Page Settings and did not see User Management or Activity Log: **PASS**.
- KEPALA_BIDANG with Divisi A saw the scoped-role navigation and did not see Landing Settings, User Management, or Activity Log: **PASS**. The Phase 2 dashboard shell does not yet render a human-readable division label; `/auth/me` division data and division authorization were verified by the DB integration. This is a UI visibility gap to address when a division-aware module is implemented, not a cross-division authorization failure.
- Logout returned to login; reload remained unauthenticated: **PASS**.

### Checklist and remaining boundaries

- `PLANS.md` Phase 2 auth, current-status/role/division, and authorization-policy test items are checked after live DB verification. Broad product-table migrations and a committed deterministic developer seed remain unchecked because Phase 2 scope only creates the five foundation tables and no checked-in seed command was added. Phase 3 and later checklist content was not changed.
- `README.md` now shows the tested `pnpm api:create-admin` argument syntax (without forwarding a literal `--`).
- Remaining implementation risks: division scope is not yet exercised through CRUD/resource endpoints because those routes are not part of Phase 2; rate limiting is in-memory and per process; the dashboard does not display the KB division name; remote GitHub Actions CI was not run.
- No Phase 3 work was started.

## Phase Result

**PHASE 2: BLOCKED** — database integration ran without skips and full `pnpm verify` passed, with no critical backend auth/RBAC failure. The requested browser UAT criterion “KEPALA_BIDANG: division scope terlihat” is not met: the current Phase 2 shell does not display the assigned division name/scope. `/auth/me` and the database-backed policy matrix prove the division assignment and enforcement, but not visible scope in the UI. Resolve this UI evidence gap in a separately authorized follow-up; Phase 3 was not started.
