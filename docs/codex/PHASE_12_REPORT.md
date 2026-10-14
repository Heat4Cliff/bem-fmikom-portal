# Phase 12 Report

## Scope

Phase 12 is **UAT & Deployment Readiness**. Work is limited to deterministic isolated UAT preparation, role-based manual UAT planning, database/private-storage backup and restore, Vercel + approved Go/MySQL host deployment readiness, rollback, production environment/secret inventory, and a release evidence template. No production deployment is included.

## Authority / Contract Review

- `PLANS.md` defines seven Phase 12 deliverables.
- PRD v1.4 requires deployment, backup, and rollback to be documented before production.
- API/Security/Data Model contracts define the database, authorization, storage paths, cookie, CORS, upload, and sensitive-data controls reflected in the runbooks.
- No product behavior, API route, database schema, or design contract changed.
- The exact approved API/database providers, region, retention, RPO/RTO, and production owner remain deployment decisions; the runbook makes them explicit prerequisites.

## Implementation

- Added an isolated UAT schema/bootstrap and deterministic fixture recipe. It uses the existing local Compose service, migrations, and interactive `api:create-admin`; it introduces no destructive reset utility or checked-in credential.
- Added a role-by-role manual UAT checklist with Public, KEPALA_BIDANG (both divisions), BPH, SUPER_ADMIN, cross-role security, and responsive checks.
- Added recovery-set backup/restore instructions for MariaDB and private document, registration attachment, and public media storage.
- Added Vercel + approved Go/MySQL host deployment readiness and rollback runbook, production environment/secret-name inventory, and go-live gate.
- Linked Phase 12 artifacts from README.

## Database Changes

None planned. Migrations 00001–00011 are unchanged. UAT procedure provisions a fresh isolated schema and does not reset existing development data.

## API Changes

None.

## UI Changes

None.

## Security

Procedures prohibit plaintext credentials, production use of local/test credentials, public storage mounts, wildcard credentialed CORS, root database accounts, unsafe reset against retained data, and unreviewed destructive migration rollback. Backup instructions pair database and all three storage roots.

## Regression Evidence

PASS. `pnpm api:integration` exited 0 on the existing dedicated `bem_fmikom_test` database. `TestDatabaseIntegration` ran (not skipped), migrated empty schema through 00011, exercised down/up and returned to version 0; all 26 named subtests passed, including RBAC/division-scope, authorization, upload/storage, and organization cases.

`pnpm verify` exited 0: Biome lint/format checked 127 files, strict TypeScript passed, frontend **19 files / 89 tests** passed, optimized Next.js production build passed, Go format checked 70 files, `go vet`, Go tests, and Go build passed. A first aggregate run showed an intermittent gallery test failure without final diagnostics; the gallery suite passed in isolation (5/5), and the full aggregate rerun then passed.

`pnpm audit`: PASS, no known vulnerabilities. `git diff --check`: PASS. `git diff --quiet -- source_docs design_reference`: clean. No migration or source/design reference was changed.

## Browser / Manual UAT

Checklist provided; user/operator execution remains pending and is not inferred from automation. No browser UAT was performed. Phase 11 manual follow-ups remain distinct.

## Remaining Findings

- Actual provider/region/engine compatibility, RPO/RTO, backup retention, and operational ownership must be selected by the release owner.
- No general fixture-seed CLI exists; current repeatable bootstrap uses the secure interactive SUPER_ADMIN command and fixture creation through the authenticated UI/API.
- No restore drill, production deploy, or user-owned role-by-role manual UAT is claimed by this report. These remain go-live gates.

## Files Changed

- `README.md`
- `PLANS.md`
- `docs/codex/DECISIONS.md`
- `docs/codex/GAPS_AND_QUESTIONS.md`
- `docs/codex/PHASE_12_RESET_SEED.md`
- `docs/codex/PHASE_12_UAT_CHECKLIST.md`
- `docs/codex/PHASE_12_BACKUP_RESTORE.md`
- `docs/codex/PHASE_12_DEPLOYMENT_RUNBOOK.md`
- `docs/codex/PHASE_12_REPORT.md`

## Git Status

Changes are limited to the Phase 12 plan, README links, decisions/gaps, runbooks, checklist, and report. No application, migration, local environment, runtime data, build output, or read-only source/design evidence is included.

## Final Phase Result

**PASS WITH DOCUMENTED NON-BLOCKING FOLLOW-UPS** — the Phase 12 procedures and required automated gates are complete. This status does not assert manual UAT acceptance, a successful restore drill, provider approval, or production deployment; all are explicit pre-launch gates in the runbooks.
