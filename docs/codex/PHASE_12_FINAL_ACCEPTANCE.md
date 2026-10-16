# Phase 12 Final Acceptance

## Environment

- Date: 2026-10-14 (Asia/Jakarta)
- Docker Desktop: 29.8.2; Docker Compose: v5.5.1; daemon reachable.
- Database: existing repository Compose service `infra-mariadb-1`, MariaDB 11.4, healthy.
- Web: `http://127.0.0.1:3290` returned HTTP 200 for the public routes and `/admin/login` after the API switch. This is HTTP smoke evidence only, not visual acceptance.
- API: `/healthz` and the public API routes returned HTTP 200 on port 8080 after switching to the isolated UAT database. The prior repository API process was stopped to avoid accidentally using its unknown database target.
- UAT runtime and storage are isolated under the operating system temp directory, not the repository. Production, `bem_fmikom_dev`, and `bem_fmikom_test` were not selected for UAT setup.

## Runtime Recovery — 2026-10-15

- After device restart, Docker Desktop was started and the existing `pnpm db:up` command brought `infra-mariadb-1` back to **healthy**. No Compose configuration was added or changed.
- API is running on port 8080 with the isolated UAT environment; `/healthz` and `/api/v1/public/settings` returned HTTP 200.
- Next.js is running from `apps/web` on port 3290; `/`, `/admin/login`, `/struktur`, and the emitted CSS asset returned HTTP 200.
- The SUPER_ADMIN bootstrap is still pending. The app is available for public-page UAT, but authenticated admin UAT cannot pass until the user creates the UAT admin through the temporary protected launcher.

## Commit

- Readiness documentation commit: `89e91e2` (`docs: complete phase 12 deployment readiness`).
- UAT acceptance work is not committed yet.

## UAT Fixture

- Database `bem_fmikom_uat`: **PASS** — newly created on the existing MariaDB service; dedicated `bem_uat` principal is limited to this schema.
- Migrations: **PASS** — existing `pnpm api:migrate:up` applied versions 00001–00011 using the UAT connection.
- Separate documents, registration-attachment, and public-media storage roots: **PASS** — created outside the repository under a fresh temp directory.
- Synthetic organization/content/accounts/uploads: **NOT TESTED** — not created yet.
- SUPER_ADMIN bootstrap: **PENDING USER INPUT** — the password was entered into a separate PowerShell prompt as a command, not into the repository CLI; no UAT account was created. Treat that typed value as exposed and do not reuse it. A current-user-only ACL launcher in the OS Temp directory now sets the isolated UAT environment and runs the repository CLI; it removes itself on successful bootstrap. The API is running again on port 8080.

## Public UAT

- Home and route HTTP smoke (`/`, `/struktur`, `/berita`, `/pengumuman`, `/galeri`, `/pendaftaran`, `/aspirasi`, `/admin/login`): **PASS** — every route returned HTTP 200. Visual rendering and UAT data flows remain **NOT TESTED**.
- Public lifecycle, privacy-safe payload, Fakultas aspiration destination, image bounds, registration submission/duplicate NIM/IMAGE, aspiration, and empty/error behavior: **NOT TESTED**.
- Desktop and mobile visual review at 360×800, 390×844, and 412×915: **NOT TESTED**.

## KEPALA_BIDANG DIV-A UAT

**NOT TESTED** — account/fixtures are not yet present. Own-division, cross-division denial, database-current role/division, menu restrictions, private files, session refresh, and logout require interactive UAT.

## KEPALA_BIDANG DIV-B UAT

**NOT TESTED** — account/fixtures are not yet present. Symmetric scope regression requires interactive UAT.

## BPH UAT

**NOT TESTED** — account/fixtures are not yet present. Cross-division operations, settings, denied user/activity/period actions, files, and session lifecycle require interactive UAT.

## SUPER_ADMIN UAT

**NOT TESTED** — bootstrap is pending. Dashboard, user lifecycle/guards, activity-log redaction, period activation, full admin workflows, session rotation/revocation, and logout require interactive UAT.

## Browser Security Review

**NOT TESTED** — Chrome console, hydration, CSP, Network panel, protected status codes, cookie flags, and token/private-file leakage were not inspected during this run. Production cookie requirements remain HttpOnly + Secure over HTTPS; verify SameSite against the deployed topology.

## Responsive Review

**NOT TESTED** — no new visual measurements or manual page review have been performed for this acceptance run. Existing Phase 11 accepted `/struktur` at 1440×900 remains separate evidence and does not close the broader viewport checklist.

## Backup Evidence

**NOT TESTED** — no recovery-set backup has been captured yet. Required set: UAT database plus UAT document, registration-attachment, and media storage roots.

## Restore Drill Evidence

**NOT TESTED** — no restore target has been created. Must restore to a fresh `bem_fmikom_restore_test`-style schema and fresh storage directories without overwriting `bem_fmikom_uat`, then verify representative data and file references.

## Regression

- `pnpm api:integration`: **PASS** — `TestDatabaseIntegration` ran against the repository test database (not skipped); output includes passing SUPER_ADMIN/BPH cross-division and both KEPALA_BIDANG scope cases, plus denial cases for cross-division and missing division.
- `pnpm verify`: **PASS** — frontend lint/format/typecheck; 19 frontend test files and 89 tests; Next.js production build; Go formatting (70 files), vet, tests, and build.
- `pnpm audit`: **PASS** — no known vulnerabilities found.
- `git diff --check`: **PASS**. `source_docs/` and `design_reference/`: **PASS** — no differences. The build-updated tracked `apps/web/tsconfig.tsbuildinfo` was restored; only this report remains untracked.
- These automated gates do not substitute for the outstanding authenticated manual UAT or restore drill.

## Bugs Found and Fixed

**NOT TESTED** — no authenticated/manual UAT flow has completed, so no new product bug is claimed or fixed.

## Remaining Risks

- Manual browser UAT and the full database-plus-storage restore drill remain the critical pending acceptance gates.
- Phase 11 follow-up visual/manual rows remain separate and are not closed by Phase 12 setup.
- Provider selection, engine compatibility, region, backup retention, RPO/RTO, and operational ownership remain open deployment decisions.
- The local web build points at `http://localhost:8080/api/v1`; browser UAT should use a host spelling/origin covered by the local CORS allow-list.

## Provider Decisions Still Required

**NOT APPLICABLE to this local drill; required before production:** approved Go/API and MySQL/MariaDB providers, exact engine/version, region, DNS/TLS, private networking, storage durability, backup encryption/retention, RPO/RTO, and operational/on-call owners.

## Release Recommendation

**PHASE 12: PASS WITH MANUAL ACCEPTANCE PENDING.** Automated integration, verification, audit, and public route HTTP smoke gates pass; UAT schema is isolated and migrated. The SUPER_ADMIN CLI currently waits for the user to enter a password in the Codex terminal. Do not recommend production release until the account/fixtures are created, role/browser and responsive UAT is manually accepted, a matched DB+all-storage restore drill passes, and provider/operational decisions are approved. No production deployment or Phase 13 work is authorized/performed by this run.
