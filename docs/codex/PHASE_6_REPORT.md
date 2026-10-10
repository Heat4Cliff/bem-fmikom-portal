# Phase 6 Report

## Scope Implemented

- Public submission at `/aspirasi`; admin list/detail at `/admin/aspirasi` and `/admin/aspirasi/[id]`.
- Public REST submission, admin list/detail/status endpoints, migration `00004_aspirations.sql`, independent submission limiter, validation, privacy-safe activity metadata, and role/division enforcement.
- Existing migrations `00001`–`00003` were not edited. Phase 7 was not started.

## Database Changes

`aspirations` references a required division, constrains workflow and anonymous identity, bounds identity/subject fields, and provides list/filter indexes. Migration was exercised on MariaDB 11.4.13 in the dedicated `bem_fmikom_test` database. The integration suite migrated from version 0 through 00004, ran 00004 DOWN/UP, and cleaned the schema back to version 0.

## Public Submission

The form loads active divisions for the active period. Name/NIM are optional; optional email is validated if supplied. Subject/message are plain text with limits of 200/5,000 characters. The API caps the JSON body, requires an explicit anonymous choice, rejects an untrusted supplied Origin, and returns only a safe receipt.

## Anonymous Behavior

When anonymous, the service omits name/NIM/email from the insert and the database constraint requires all three to be `NULL`. Integration tested a malicious payload that included identity values; stored values were `NULL`. The admin UI labels the sender “Anonim”.

## Status Workflow

Only `PENDING → READ → RESOLVED` is accepted; skipped, same-state, and reverse transitions are rejected. First `read_at`/`resolved_at` values are set once in UTC. Updates and their minimized activity events are transactional.

## Admin Management

The list supports pagination, search, division/status filters, and date range. Detail renders plain text and a simple status timeline. The UI exposes only the next allowed status action and uses desktop table/mobile card layouts.

## Authorization / Division Scope

SUPER_ADMIN and BPH have cross-division operational access. KEPALA_BIDANG is scoped to its current database-assigned division; out-of-scope detail returns 404 and mutation returns 403. Integration and browser role tests passed.

## Rate Limiting

Aspiration writes use an independent per-IP in-memory limiter, default 5 requests per 15 minutes, configured with `PUBLIC_ASPIRATION_RATE_LIMIT` and `PUBLIC_ASPIRATION_RATE_WINDOW`. Middleware/unit and database HTTP integration tests passed, including 429 and `Retry-After` behavior. Multi-instance deployments need shared limiter storage.

## Privacy / PII Controls

- Anonymous identity is removed server-side and constrained to SQL `NULL`.
- Public receipts contain only ID, status, and submission timestamp.
- Activity details contain aspiration ID, old/new status, and division ID only. Structured request logs omit payloads.
- Limits: name 160, NIM 32, email 254, subject 200, message 5,000 characters.
- Integration inspected database/activity/log content and found no submitted PII or message markers in activity metadata or structured request logs.

## Activity Log Events

Status changes emit `UPDATE` / `ASPIRATION` with only aspiration ID, old/new status, and division ID. MariaDB integration verified the persisted event shape.

## Decisions Made

- Active-period divisions are the destination choices; no independent category taxonomy was introduced without a contract endpoint/domain.
- Optional name/NIM follow the PRD; email remains optional and is validated if supplied.
- Transitions are sequential and RESOLVED is terminal.
- Attachment upload, disposition notes, SLA, export, and case-management controls shown in Stitch references are outside this phase's contract.
- `screen.png` is the visual target; `code.html` informed structure/style only and was not copied into production.

## Gaps Resolved

- Active division routing, anonymous stripping, transitions, IDOR policy, request limits, and PII-safe activity details are mapped in contracts and implementation documentation.
- Public/admin routes and endpoints are recorded in `docs/design/ROUTE_SCREEN_MATRIX.md`.
- Docker access became available through authorized local execution; the repository's existing Compose configuration ran MariaDB healthy for final verification.

## Gaps Carried Forward

- Rate limiting is process-local; shared storage is required before scaling to multiple API instances.
- Stitch-only attachment, separate taxonomy, disposition/SLA, export, and analytics controls remain out of scope.
- No Phase 6 acceptance blocker remains. Phase 5's release-wide evidence/limitations remain described in its original report; this phase added BPH/KEPALA_BIDANG organization UI smoke and exact viewport review without changing Phase 5 code.

## Verification

### STATIC VERIFICATION

PASS: `pnpm verify` completed lint, format, TypeScript typecheck, frontend production build, Go format (45 files), `go vet`, Go tests, and Go build.

### UNIT TEST

PASS: 26 frontend tests across 6 files, including 4 aspiration UI tests. Go unit/package tests passed as part of `pnpm verify`.

### DATABASE INTEGRATION

PASS: Docker 29.8.2, Compose 5.5.1, Docker Engine 29.8.2, MariaDB 11.4.13. Used `pnpm db:up` and the existing `infra/docker-compose.dev.yml`; container health was `healthy` and `mariadb-admin ping` returned `mysqld is alive`. `pnpm api:integration` ran `TestDatabaseIntegration` without skipping against `bem_fmikom_test`; final run passed. The suite covers Phase 2 auth/RBAC, Phase 4 content, Phase 5 organization/periods, and Phase 6 aspirations.

### MIGRATION ROUND TRIP

PASS: empty schema → migrations 00001–00004 UP → 00004 DOWN → 00004 UP. Integration teardown returned the test schema to version 0. No non-test database was used.

### PUBLIC WRITE TEST

PASS: identified and anonymous submissions, minimal receipt, required subject, invalid email, overlong message, unknown/inactive/archived-period divisions, FK and workflow constraints.

### ANONYMOUS STRIPPING TEST

PASS: supplied identity values on an anonymous request were stripped and verified as SQL `NULL`.

### RATE LIMIT TEST

PASS: independent configured limiter enforced a one-request integration threshold and returned 429 with `Retry-After`; default remains 5/15m.

### STATUS WORKFLOW TEST

PASS: allowed and rejected transitions, UTC first timestamps, timestamp order, and terminal RESOLVED behavior verified in MariaDB. Browser UAT advanced one record to READ and RESOLVED and another to READ.

### RBAC / IDOR TEST

PASS: SUPER_ADMIN/BPH cross-division, KEPALA_BIDANG own-division, cross-division denial, concealed detail, forbidden mutation, and missing-division cases passed integration. Browser UAT confirmed BPH saw both divisions; KEPALA_BIDANG saw only Divisi A and a direct Divisi B detail request was denied.

### PII / LOGGING TEST

PASS: integration confirmed PII/message markers were absent from activity details and structured request logs; activity metadata contained only safe status/division identifiers.

### PHASE 5 regression

PASS: period/division/member/public organization integration assertions passed in the shared database suite. Carried-forward browser smoke: BPH saw both test divisions and organization controls; KEPALA_BIDANG saw only its division and no division mutation controls; BPH's period screen was read-only. Phase 5 implementation files/report were not changed.

### PHASE 4 regression

PASS: article/category, gallery, and announcement integration assertions passed in the shared database suite.

### PHASE 2 auth regression

PASS: auth/session, database-current role/status/division, refresh rotation/reuse, and role policy assertions passed in the shared suite and `pnpm verify`.

### PUBLIC BROWSER UAT

PASS: active-period division choices loaded. Identified and anonymous test aspirations submitted successfully and returned safe confirmation receipts. Admin displayed the anonymous sender as “Anonim”.

### SUPER_ADMIN UAT

PASS: saw aspirations across divisions, opened anonymous detail, and advanced it through READ and RESOLVED. Existing System navigation remained visible.

### BPH UAT

PASS: saw both division records, opened a Divisi A aspiration, and changed it to READ. User Management and Activity Log were not visible.

### KEPALA_BIDANG UAT

PASS: scope badge showed Divisi A; list contained only Divisi A. Direct Divisi B detail was concealed. User Management, Activity Log, and Landing Settings were not visible.

### DESKTOP VISUAL REVIEW

PASS: data-populated public form, admin list/table, and detail/timeline reviewed against audited Stitch `screen.png` targets at desktop size. Layout hierarchy, filters, table, anonymous identity, timeline, and action panels were visible. No Stitch HTML was copied.

### MOBILE VISUAL REVIEW

PASS: public form, admin list/cards, and detail/timeline visually reviewed at 360×844, 390×844, and 412×844. Mobile cards replaced the table; no visible horizontal overflow was observed. Temporary viewport override was reset.

## Known Issues

- The first integration attempts uncovered two test-fixture issues: auth regression left the seeded administrator role changed to BPH, and an Origin assertion used a helper that always sent the trusted Origin. Both fixtures were corrected; subsequent integration runs passed.
- The limiter is process-local, as described above.
- Test-only browser fixtures were removed by the final integration cleanup; the database schema returned to version 0.

## Security Notes

Database-backed authorization, IDOR, anonymous persistence, status transactions, activity privacy, and public rate limiting passed integration. Browser UAT confirmed public submission and role-scoped access. No critical Phase 6 security failure remains in the tested scope.

## Phase Result

Phase 6 implementation and verification gates passed on 2026-10-10, including the non-skipping MariaDB integration suite, migration round trip, full `pnpm verify`, role-based browser UAT, desktop/mobile review, and Phase 2/4/5 regression. `source_docs/` and `design_reference/` are unchanged. Phase 7 was not started.

**PHASE 6: PASS**
