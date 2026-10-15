# Phase 12.1 — Frontend / Backend Deployment Separation & Handoff

## Status

**PHASE 12.1: PASS WITH FOLLOW-UPS**

Standalone frontend and backend handoff packages were created, built, tested, archived, and verified from extracted ZIP contents. Production deployment was not performed. Remaining provider and operational decisions are listed below.

## Scope and source boundaries

- Read the project instructions, Phase 12 deployment handoff context, package/workspace manifests, lockfile, environment templates, backend configuration, deployment runbook, and prior Phase 12 acceptance evidence before export.
- Did not change product behavior, API contracts, database schema/migrations, app UI, root package scripts, or the Phase 12 checklist.
- `source_docs/` and `design_reference/` remain unchanged (`git diff --quiet -- source_docs design_reference`: clean).
- The pre-existing untracked `docs/codex/PHASE_12_FINAL_ACCEPTANCE.md` was left untouched.

## Export contents

| Artifact | Contents | ZIP verification |
|---|---|---|
| `deployment_handoff/frontend/` | Standalone Next.js app, standalone pnpm workspace/lock, Biome config, Vitest tests, public assets, safe `.env.example`, local Inter and Plus Jakarta Sans font files | 141 files; extracted SHA-256 match for every entry |
| `deployment_handoff/backend/` | Standalone Go/Gin module, source/tests, `go.mod`/`go.sum`, embedded migration SQL, safe production-shaped `.env.example` | 87 files; extracted SHA-256 match for every entry |
| `deployment_handoff/DEPLOYMENT_HANDOFF.md` | Runtime contract, independent commands, environment responsibilities, deployment decisions and scope limits | Included in handoff folder |
| `bem-fmikom-frontend.zip` | Frontend package rooted at `frontend/` | 315,493 bytes; SHA-256 `B9AB24753629A60C5C7FC5AB07335294438310C66465BE5E93886CEA557F80D8` |
| `bem-fmikom-backend.zip` | Backend package rooted at `backend/` | 181,517 bytes; SHA-256 `35BED9712C670A0D60D168CD0E6541C0F25A9EB39B86A249D2B81064020E0F80` |

The frontend export changes only its copied `next/font/local` asset paths so builds resolve included font files without reaching into monorepo `node_modules`. The backend export changes only its copied development dotenv loader so the standalone package reads its own `.env`, not a parent repository file. The application source remains unchanged.

## Verification evidence

### Frontend export

- Frozen install from the standalone frontend lock completed with pnpm 11.25.0 using the local package cache. The host sandbox blocked pnpm's project-store symlink and offline policy metadata was unavailable; the lock from the repository was therefore used in trusted-lock mode for this local check. The export does not require trusted-lock mode in its documented normal connected install command.
- Biome lint: PASS, 127 files.
- Biome format check: PASS, 127 files.
- TypeScript strict typecheck: PASS.
- Vitest: PASS, 19 files / 89 tests.
- Next.js production build: PASS; all public/admin routes generated successfully.

### Backend export

- Module toolchain: Go 1.26.9 auto-selected per `go.mod`; host default was Go 1.26.1.
- `go test ./...`: PASS, including `TestDatabaseIntegration` against the dedicated configured database with `_test` suffix; migrations and RBAC/division-scope cases ran.
- `go vet ./...`: PASS.
- `go build ./...`: PASS.
- Integration access and Windows storage filesystem tests needed execution outside the restricted sandbox; they passed in the authorized local run.
- `govulncheck` was not installed, and a `go mod verify` attempt did not complete within the host execution window. No Go vulnerability scan or module-cache integrity result is claimed.

### Source repository regression

- `pnpm api:integration`: PASS; `TestDatabaseIntegration` ran, migrations 00001–00011 applied and rolled back, and role/division coverage passed against the dedicated test DB.
- `pnpm verify`: PASS; lint, format, typecheck, 19 frontend test files / 89 tests, Next.js production build, Go format (70 files), Go vet, Go tests, and Go build completed.
- `pnpm audit`: PASS; no known JavaScript dependency vulnerabilities found.
- `git diff --check`: PASS.
- `git diff --quiet -- source_docs design_reference`: PASS.
- ZIP extraction validation: PASS; all entries matched source package SHA-256 values. Scans found no actual `.env`, private key, local machine path, `.next`, `node_modules`, or build-cache files in either archive. Each archive contains only its safe `.env.example` template.

## Remaining follow-ups before deployment

1. Run `govulncheck ./...` (or the campus-approved Go advisory scanner) and complete `go mod verify` in a normal, unblocked Go environment.
2. Select and document API/database providers, region, database engine/version, DNS/TLS/reverse-proxy topology, storage durability, backup retention, RPO/RTO, monitoring, and rollback/on-call owners.
3. Supply production URLs and secrets through the deployment platform's secret/config system. Set the frontend API/site URLs before the Next.js build; set exact HTTPS `FRONTEND_ORIGINS` and production-secure cookie settings on the API.
4. Provision a dedicated non-root database account and private absolute document/registration storage paths that are not publicly served. Review backup/restore and perform a release-owner manual smoke/UAT after deployment configuration is selected.

## Files and release boundary

- Added `deployment_handoff/` and the two ZIP archives as local handoff artifacts.
- Added this report.
- No commit or push was made.
- No production deployment, production credentials, provider infrastructure, or Phase 13 work was performed.
