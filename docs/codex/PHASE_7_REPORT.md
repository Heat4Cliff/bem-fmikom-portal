# Phase 7 Report

## Scope Implemented

- Public registration list, detail, and dynamic form routes.
- Admin registration CRUD/status controls, dynamic field builder, paginated submissions/detail, and XLSX export.
- Eight allow-listed dynamic field types, required core identity fields, server-side validation, normalized NIM duplicate prevention, and immutable structural schema after first submission.
- Role/division authorization, public write rate limiting, private response/cache handling, and minimized registration activity metadata.
- No Phase 8 document archive or later-phase user/settings/activity dashboard features were started.

## Database Changes

- Added reversible `00005_registration.sql`; existing migrations were not edited.
- MariaDB 11.4.13 integration verified all five migrations from an empty `bem_fmikom_test` schema, rolled migration 00005 down/up, and cleaned the schema back to version 0.
- Tables include registration forms, fields, field options, submissions, and historical answer snapshots, with foreign keys, indexes, status/date checks, and unique `(registration_form_id, normalized_nim)` protection.

## Registration Lifecycle

- Admin transitions are DRAFT → OPEN → CLOSED; CLOSED cannot be reopened.
- Public derived status is UPCOMING/OPEN/CLOSED, with DRAFT never public. The OPEN window is UTC and inclusive at both ends.
- Open validates form timing and schema before the transition. Submissions outside the OPEN window are rejected.

## Dynamic Schema

- Supported types: TEXT, TEXTAREA, EMAIL, NUMBER, SELECT, RADIO, CHECKBOX, DATE.
- New forms have exactly one required `full_name`, `nim`, `email`, and `phone` core field. Schema/options are validated and saved transactionally; choice answers are checked against current options.
- Duplicate JSON answer keys are rejected during request decoding instead of silently collapsing into a Go map. Requests are size bounded; fields and text lengths are bounded.

## Schema Lock Policy

- Structural fields (key, type, required, options, and order) can change only for DRAFT forms with no submissions. After the first submission, structural changes return 409 and answer history remains intact.
- Presentation labels/placeholders can be corrected without rewriting stored answer snapshots. Core identity values are stored once on the submission, not duplicated in answer rows.

## Duplicate Submission Policy

- NIM is required, trimmed, uppercased for comparison, and unique per form at the database level. Display NIM remains a string; email is not a unique key.
- Integration verified normalized duplicate rejection, reuse on another form, and a concurrent race where exactly one request succeeds and one receives a conflict.

## Public Registration

- `/pendaftaran`, `/pendaftaran/[slug]`, and `/pendaftaran/[slug]/form` consume safe public DTOs and render responsive reusable Next.js components.
- Public list filters UPCOMING/OPEN/CLOSED, sorts deterministically, and never exposes DRAFT or administrative metadata.
- Submission validation runs on both client and server; the receipt does not echo applicant PII.

## Form Builder

- `/admin/pendaftaran/[id]/builder` supports core fields, the eight field types, option editing, ordering with keyboard-accessible move controls, metadata editing, preview, and status transitions.
- A form created through the browser UI navigated to its builder; schema save and opening were verified. The submit handler reads the actual form values, including native datetime-local controls.

## Submission Management

- `/admin/pendaftaran/[id]/submissions` and `/admin/pendaftaran/[id]/submissions/[submissionId]` provide paginated/searchable admin data and snapshot-based detail. Inaccessible forms hide submission/export controls.
- Forms with submissions cannot be deleted; historical answers remain readable after safe label edits.

## Excel Export

- XLSX export is capped at 1,000 rows, uses explicit string cells for all submitted values, and returns `Cache-Control: private, no-store`.
- Integration reopened the generated workbook and verified its 12 headers/columns, two data rows, Indonesian field labels, checkbox serialization (`web; data`), date value, and an applicant formula string stored without an executable formula.
- Export activity metadata contains form/division/row count only. A browser download completed successfully; the local UAT file contained synthetic test data.

## Authorization / Division Scope

- SUPER_ADMIN and BPH have cross-division operational access. KEPALA_BIDANG is scoped to its current database division; create cannot select another division.
- Database integration verified both directions of cross-division denial, divisionless KEPALA_BIDANG denial, allowed own-division access, BPH/SUPER_ADMIN cross-division access, and denial on nested submission/export routes.
- Browser UAT verified role-sensitive navigation for BPH and KEPALA_BIDANG, the KB scope badge, absence of User Management/Activity Log for BPH, absence of those modules and Landing Settings for KB, and no cross-division form visibility.

## Privacy / PII

- No raw client IP is stored. Public responses, request logs, and activity details do not include applicant identity or answer payloads.
- Submission core identity is stored once; answer rows retain only custom answers and field snapshots. Export access is division checked and no-store.
- Integration verified the receipt, structured logs, export activity metadata, answer storage, and formula-safe workbook content for PII/sensitive-value leakage.

## Rate Limiting

- Public registration writes use a separate configurable per-process fixed-window limiter: default 5 attempts per IP per 30 minutes. Integration verified 429 and `Retry-After`.
- Multi-instance deployments require a shared limiter store; this is an operational scaling follow-up, not a local integration failure.

## Activity Log Events

- Registration form create/update/delete/open/close/schema changes and export produce minimized activity events. Export details contain no applicant data.
- The Activity Log dashboard is outside Phase 7.

## Decisions Made

- NIM is the one-per-form identity; email uniqueness is disabled.
- Schema structural lock begins after the first submission; no schema versioning engine is introduced.
- Public listing includes UPCOMING, OPEN, and CLOSED as directed by the detailed Phase 7 request, while DRAFT stays private.
- Export limit is 1,000 rows. Registration rate limiting is isolated from login and aspiration limits.
- Stitch `screen.png` remains the visual target; Stitch `code.html` was used only as a structural/style reference, never copied as production code.

## Gaps Resolved

- Phase 7 duplicate policy, dynamic schema types/options, answer history, schema lock, form lifecycle, public write abuse limit, XLSX formula/privacy behavior, public/admin API/UI mapping, and role/division behavior are implemented and verified.
- `API_CONTRACT.md`, `DECISIONS.md`, `GAPS_AND_QUESTIONS.md`, `IMPLEMENTATION_MATRIX.md`, and `ROUTE_SCREEN_MATRIX.md` were reconciled for Phase 7.

## Gaps Carried Forward

- Per-process rate limiting must move to shared storage before horizontal API scaling.
- A stakeholder should perform final visual comparison against every relevant Stitch `screen.png` at desktop and 360/390/412 widths. Browser evidence in this run confirms representative 390×844 public list/form and admin workflows, not pixel parity for every screen/breakpoint.
- File uploads, notifications, applicant approval/selection, and Phase 8 archive remain out of scope.

## Verification

### STATIC VERIFICATION

PASS — `pnpm verify`: Biome lint/format, TypeScript strict typecheck, Next production build, Go fmt, Go vet, Go tests, and Go build.

### UNIT TEST

PASS — Web: 7 files, 33 tests. Go package tests PASS.

### DATABASE INTEGRATION

PASS — `pnpm api:integration`, actual MariaDB 11.4.13, dedicated `bem_fmikom_test`; `TestDatabaseIntegration` ran (not skipped).

### MIGRATION ROUND TRIP

PASS — 00001–00005 UP; 00005 DOWN/UP; integration cleanup returned schema to version 0.

### FORM LIFECYCLE TEST

PASS — DRAFT omitted from public list, UPCOMING public state, OPEN submit allowed, CLOSED submit rejected, CLOSED reopen rejected.

### DYNAMIC FIELD TEST

PASS — all eight types accepted with valid values; tests reject missing required values, wrong JSON type, unknown fields, duplicate answer keys, forged options, and overlong TEXTAREA values.

### SCHEMA LOCK TEST

PASS — structural changes after submission return 409; safe label edits preserve the original answer snapshot.

### DUPLICATE SUBMISSION TEST

PASS — trim/case-normalized NIM duplicate rejected; same NIM on another form allowed; concurrent duplicate race yields one create and one conflict.

### RBAC / IDOR TEST

PASS — SUPER_ADMIN/BPH cross-division access; KEPALA_BIDANG own-division access only; cross-division detail/mutation/submission/export and divisionless KB requests denied.

### RATE LIMIT TEST

PASS — independent limiter returned 429 and `Retry-After` after configured test threshold.

### EXCEL EXPORT TEST

PASS — XLSX download and workbook parsing; row/column count, headers, custom columns, checkbox/date serialization, private/no-store headers verified.

### FORMULA INJECTION TEST

PASS — formula-shaped applicant input remained a literal string cell with no formula in the reopened workbook.

### PRIVACY / LOGGING TEST

PASS — receipt, structured request logs, database answer shape, and export audit metadata contained no unintended PII duplication/leak.

### PHASE 6 REGRESSION

PASS — existing aspiration anonymous/validation/status/rate-limit/division integration assertions passed within the database suite.

### PHASE 5 REGRESSION

PASS — period/division/member authorization and single-active-period regression assertions passed within the database suite.

### PHASE 4 REGRESSION

PASS — existing content CMS tests passed in integration and Go package suites.

### PHASE 2 AUTH REGRESSION

PASS — auth/session/RBAC regression tests passed in integration and Go package suites.

### PUBLIC UAT

PASS — public list, detail, dynamic form, successful synthetic submission, and normalized duplicate NIM rejection. A custom TEXTAREA submission was also verified earlier in this Phase 7 browser run.

### SUPER_ADMIN UAT

PASS — browser created a form, navigated to the builder, saved schema/opened a form, viewed submissions/detail, and downloaded XLSX.

### BPH UAT

PASS — cross-division operational form/submission access and role-hidden User Management/Activity Log navigation verified.

### KEPALA_BIDANG UAT

PASS — own-division scope and navigation verified; no alternate-division selector; crafted cross-division browser routes denied. API integration verified server-side scope.

### DESKTOP VISUAL REVIEW

PASS WITH FOLLOW-UP — representative admin registration/builder/submission screens inspected at the browser's default desktop viewport; exhaustive Stitch-by-Stitch pixel comparison remains manual.

### MOBILE VISUAL REVIEW

PASS WITH FOLLOW-UP — public registration list and dynamic form inspected with `window.innerWidth=390` and `innerHeight=844`; admin scope/layout was also inspected at 390. No horizontal overflow was visible in representative captures. Remaining breakpoint/screen comparisons are listed above.

## Known Issues

- No known blocking database, security, privacy, export, or functional defect remains in the tested Phase 7 scope.
- Full visual sign-off against all Stitch references and all requested viewport widths remains a human review item.

## Security Notes

- The test fixture and downloaded XLSX used synthetic values only. Browser UAT, fixture rows, and account were confined to local services and `bem_fmikom_test`.
- Production must use unique strong JWT secrets, approved frontend origins, HTTPS, and a shared distributed public write limiter when horizontally scaled.

## Phase Result

PHASE 7: PASS WITH MANUAL VERIFICATION REQUIRED
