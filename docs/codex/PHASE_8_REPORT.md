# Phase 8 — Private Document Archive

**Status: PHASE 8: PASS WITH MANUAL VERIFICATION REQUIRED**

Phase 8 adds private, division-scoped document storage, metadata management, and authenticated streaming downloads. No Phase 9 implementation was started.

## Scope delivered

- New reversible migration `00006_documents.sql`; migrations `00001`–`00005` remain unchanged.
- Private filesystem storage abstraction with opaque keys, `.staging` uploads, path traversal confinement, streaming SHA-256, and bounded file size.
- Allow-listed PDF/DOC/DOCX/XLS/XLSX upload with extension, declared MIME, signature/container checks, safe original filename handling, and default 10 MiB configurable limit.
- Metadata list/filter, detail (loaded when opening edit), update, and delete; authenticated download sends attachment/no-store/nosniff headers.
- SUPER_ADMIN and BPH cross-division archive access; KEPALA_BIDANG is forced to the current database division. Scoped details conceal cross-division existence with 404.
- UPLOAD, DOWNLOAD, UPDATE, and DELETE activity events with minimized metadata. No raw file contents, stored paths, or tokens are added to activity details or request logs.
- Responsive `/admin/arsip-dokumen` table/dialog and stacked mobile-card experience based on archive Stitch `screen.png` targets.

## Decisions and contract reconciliation

- PRD v1.4 and SECURITY_CONTRACT define a 10 MB baseline; implementation uses configurable 10 MiB. The conditional 25 MiB fallback does not apply.
- PRD makes `division_id` optional, while the direct Phase 8 scope requires period/division ownership and fixed KB division. Phase 8 narrows document ownership to a required division with a composite `(division_id, period_id)` FK. See `DECISIONS.md` and `GAPS_AND_QUESTIONS.md`.
- UI route is `/admin/arsip-dokumen`; existing API contract paths remain `/api/v1/admin/documents...`.
- Delete uses a `delete_pending` retry state because local filesystem and SQL cannot share a transaction. Integration coverage injects both a physical-delete failure and audit/finalization failure, then retries.
- `code.html` is not production code. `screen.png` remains the visual target. No separate Stitch screens exist for edit/delete dialogs.
- Extension/MIME/signature checks are not antivirus scanning or content disarm. Windows local storage inherits directory ACLs; POSIX objects are created with `0600` permissions.

## Verification evidence

### MariaDB and migration

- Docker CLI: `Docker version 29.8.2, build 7fc2dff`.
- Docker Compose: `Docker Compose version v5.5.1`.
- MariaDB: `11.4.13-MariaDB-ubu2404`, container healthy; dedicated `bem_fmikom_test` schema existed before the run.
- Existing command/configuration only: `pnpm db:up` using `infra/docker-compose.dev.yml`; no new Compose file was created.
- `pnpm api:integration` passed. `TestDatabaseIntegration` ran (not skipped): migrations UP from zero, DOWN for `00006` and prior rollback checks, UP again, then full auth/Phase 4–7 and document integration assertions. Test cleanup returned the dedicated schema to zero migrations.
- Document integration exercised valid PDF/DOC/DOCX/XLS/XLSX, extension/MIME/signature rejection, metadata update, authorized byte-identical download, SHA-256, list filters, cross-division RBAC, activity rows, storage-delete failure recovery, finalization failure recovery, and delete retry.

### Full verification

`pnpm verify` passed after the final integration-test changes:

- Frontend lint, formatting, and typecheck.
- Frontend tests: 7 files, 33 tests passed.
- Next.js production build, including `/admin/arsip-dokumen`.
- Go format check, vet, all Go tests (including the non-skipping database integration test), and build.

### Browser and visual review

- Local production frontend started successfully. Navigating unauthenticated to `/admin/arsip-dokumen` redirected to `/admin/login`.
- API/DB integration verifies SUPER_ADMIN/BPH cross-division access and KEPALA_BIDANG own-division access/denial.
- Authenticated browser workflows for each role and side-by-side desktop/390 px Stitch visual sign-off remain manual follow-up. The available browser automation policy does not permit automating the login flow. Therefore no authenticated visual-parity claim is made.

### Source integrity

- `source_docs/` and `design_reference/` were treated as read-only and have no Phase 8 changes.
- `PLANS.md` Phase 9–12 checklists remain untouched.

## Remaining manual review and risks

- Review archive upload, filter, edit, download, and delete at desktop and 390 px against Stitch screenshots using a human-authenticated browser session.
- Confirm whether future product requirements need organization-wide documents without a division; current schema deliberately requires a division.
- Production operations must back up and restore SQL metadata and `DOCUMENT_STORAGE_PATH` as one recovery set. Add approved malware scanning if organizational policy requires it.

## Phase result

PHASE 8: PASS WITH MANUAL VERIFICATION REQUIRED
