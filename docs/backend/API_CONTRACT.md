# API Contract Baseline

Prefix: `/api/v1`

## Standard success
```json
{"success":true,"message":"Data berhasil diambil","data":{}}
```

## Standard list
```json
{
  "success": true,
  "message": "Data berhasil diambil",
  "data": [],
  "meta": {"page":1,"limit":20,"total":120,"total_pages":6}
}
```

## Standard validation error
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {"email":["Email wajib diisi"]}
}
```

## Auth
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/me`

## Public
- `GET /public/settings`
- `GET /public/periods/active`
- `GET /public/divisions`
- `GET /public/organization`
- `GET /public/articles`
- `GET /public/articles/{slug}`
- `GET /public/gallery-categories`
- `GET /public/gallery-categories/{slug}`
- `GET /public/announcements`
- `GET /public/announcements/active`
- `GET /public/announcements/{slug}`
- `POST /public/aspirations`
- `GET /public/registrations`
- `GET /public/registrations/{slug}`
- `POST /public/registrations/{id}/submissions`

## Admin dashboard
- `GET /admin/dashboard`

## Admin users — SUPER_ADMIN
- `GET /admin/users`
- `POST /admin/users`
- `GET /admin/users/{id}`
- `PATCH /admin/users/{id}`
- `POST /admin/users/{id}/deactivate`
- `POST /admin/users/{id}/reactivate`
- `POST /admin/users/{id}/reset-password`

## Periods
- `GET /admin/periods`
- `POST /admin/periods`
- `GET /admin/periods/{id}`
- `PATCH /admin/periods/{id}`
- `POST /admin/periods/{id}/activate`
- `POST /admin/periods/{id}/archive`

## Divisions
- `GET /admin/divisions`
- `POST /admin/divisions`
- `GET /admin/divisions/{id}`
- `PATCH /admin/divisions/{id}`
- `DELETE /admin/divisions/{id}`

## Organization members
- `GET /admin/members`
- `POST /admin/members`
- `GET /admin/members/{id}`
- `PATCH /admin/members/{id}`
- `DELETE /admin/members/{id}`

## Article categories
- `GET /admin/article-categories`
- `POST /admin/article-categories`
- `PATCH /admin/article-categories/{id}`
- `DELETE /admin/article-categories/{id}`

## Articles
- `GET /admin/articles`
- `POST /admin/articles`
- `GET /admin/articles/{id}`
- `PATCH /admin/articles/{id}`
- `DELETE /admin/articles/{id}`
- `POST /admin/articles/{id}/publish`
- `POST /admin/articles/{id}/unpublish`

## Gallery
- `GET /admin/gallery-categories`
- `POST /admin/gallery-categories`
- `PATCH /admin/gallery-categories/{id}`
- `DELETE /admin/gallery-categories/{id}`
- `GET /admin/gallery-items`
- `POST /admin/gallery-items`
- `PATCH /admin/gallery-items/{id}`
- `DELETE /admin/gallery-items/{id}`

## Announcements
- `GET /admin/announcements`
- `POST /admin/announcements`
- `GET /admin/announcements/{id}`
- `PATCH /admin/announcements/{id}`
- `DELETE /admin/announcements/{id}`
- `POST /admin/announcements/{id}/activate`
- `POST /admin/announcements/{id}/deactivate`

## Aspirations
- `GET /admin/aspirations`
- `GET /admin/aspirations/{id}`
- `PATCH /admin/aspirations/{id}/status`

## Registrations
- `GET /admin/registrations`
- `POST /admin/registrations`
- `GET /admin/registrations/{id}`
- `PATCH /admin/registrations/{id}`
- `DELETE /admin/registrations/{id}`
- `POST /admin/registrations/{id}/open`
- `POST /admin/registrations/{id}/close`
- `GET /admin/registrations/{id}/submissions`
- `GET /admin/registrations/{id}/submissions/{submissionId}`
- `GET /admin/registrations/{id}/export`

## Documents
- `GET /admin/documents`
- `POST /admin/documents` (`multipart/form-data`)
- `GET /admin/documents/{id}`
- `PATCH /admin/documents/{id}`
- `DELETE /admin/documents/{id}`
- `GET /admin/documents/{id}/download`

## Landing settings
- `GET /admin/settings`
- `PATCH /admin/settings`

## Activity logs — SUPER_ADMIN read-only
- `GET /admin/activity-logs`
- `GET /admin/activity-logs/{id}`

## Query conventions
List endpoints should support as relevant:
- `page`, `limit`
- `search`
- `status`
- `division_id`
- `period_id`
- date range (`start_date`, `end_date`)
- module-specific category/type filters

## Contract rule
Before implementation begins, Phase 0 must compare this baseline against the full v1.4 source document. If the source document contains a more specific route/payload/error rule, the v1.4 source wins and this file must be reconciled before Phase 1/2.

## Phase 4 implementation contract

The Phase 4 routes below are implemented under `/api/v1`. List responses use the standard list envelope and pagination metadata; mutation responses use the standard success envelope. Validation failures use 400, missing records use 404, duplicate slugs/unique conflicts use 409, unauthenticated requests use 401, and role/scope denials use 403. Scoped reads may return 404 to conceal cross-division resource existence.

| Consumer/action | Endpoint | Access |
|---|---|---|
| Public article list/detail | `GET /public/articles`, `GET /public/articles/{slug}` | Public; only published articles whose publication time has arrived |
| Public gallery list/detail | `GET /public/gallery-categories`, `GET /public/gallery-categories/{slug}` | Public; active categories and non-deleted album items |
| Public announcement archive/detail/active | `GET /public/announcements`, `GET /public/announcements/{slug}`, `GET /public/announcements/active` | Public; archive/detail visibility follows the currently implemented `is_active` flag; active endpoint additionally filters the UTC start/end window |
| Article category management | `GET/POST /admin/article-categories`, `PATCH/DELETE /admin/article-categories/{id}` | SUPER_ADMIN/BPH manage; KEPALA_BIDANG reads only |
| Article management | `GET/POST /admin/articles`, `GET/PATCH/DELETE /admin/articles/{id}`, `POST /admin/articles/{id}/publish`, `POST /admin/articles/{id}/unpublish` | SUPER_ADMIN/BPH cross-division; KEPALA_BIDANG is bound to its current DB division |
| Gallery category/item management | `GET/POST /admin/gallery-categories`, `PATCH/DELETE /admin/gallery-categories/{id}`, `GET/POST /admin/gallery-items`, `PATCH/DELETE /admin/gallery-items/{id}` | SUPER_ADMIN/BPH manage categories and items across divisions; KEPALA_BIDANG reads own-division categories and can manage items only under its own-division parent; gallery categories are read-only for KEPALA_BIDANG |
| Announcement management | `GET/POST /admin/announcements`, `GET/PATCH/DELETE /admin/announcements/{id}`, `POST /admin/announcements/{id}/activate`, `POST /admin/announcements/{id}/deactivate` | SUPER_ADMIN/BPH write; KEPALA_BIDANG read-only |

Article bodies and announcement bodies are sanitized on the server using the configured HTML allowlist. Gallery album links must be HTTPS Google Photos URLs; preview/image URLs, when supplied, must also use HTTPS. Content images are URL-backed; no upload endpoint or private media storage is introduced in Phase 4. Date filters on announcement lists are `date_from` and `date_to`; API timestamps are persisted as UTC.

The full PRD v1.4 specifies the public gallery detail route by slug, so this contract replaces the earlier `{id}` placeholder. Announcement public archive visibility for scheduled/expired records remains a documented clarification item in `GAPS_AND_QUESTIONS.md`.

## Phase 5 organization implementation contract

Organization routes below are implemented under `/api/v1`. Lists use the shared paginated envelope and accept `page`, `limit`, `search`, `status`, and relevant `period_id`, `division_id`, or `position` filters. Invalid input returns 422; missing/scoped-out detail records return 404; unique-head/dependency conflicts return 409; unauthenticated requests return 401; denied roles or mutations return 403.

| Consumer/action | Endpoint | Access and response scope |
|---|---|---|
| Public active period | `GET /public/periods/active` | Public; returns the active period, or 404 if no period is active. |
| Public division list | `GET /public/divisions` | Public; active-period divisions ordered deterministically by display order, name, and ID; empty list if no active period. |
| Public organization hierarchy | `GET /public/organization` | Public; minimal active-period, leadership, active-division, head/member DTO; no NIM, status, or audit fields. Returns an empty organization when no period is active. |
| Period list/detail | `GET /admin/periods`, `GET /admin/periods/{id}` | SUPER_ADMIN/BPH/KEPALA_BIDANG read; KB has no write access. |
| Period create/update/activate/archive | `POST /admin/periods`, `PATCH /admin/periods/{id}`, `POST /admin/periods/{id}/activate`, `POST /admin/periods/{id}/archive` | SUPER_ADMIN only. Activation transactionally archives the previously active period; history is retained. |
| Division list/detail | `GET /admin/divisions`, `GET /admin/divisions/{id}` | SUPER_ADMIN/BPH cross-division; KB list/detail is restricted to its current database division. |
| Division create/update/delete | `POST /admin/divisions`, `PATCH /admin/divisions/{id}`, `DELETE /admin/divisions/{id}` | SUPER_ADMIN/BPH. Delete conflicts when period/member/user/content references prevent safe deletion; no cascade removal. |
| Member list/detail | `GET /admin/members`, `GET /admin/members/{id}` | SUPER_ADMIN/BPH cross-division; KB reads only its current DB-assigned division. Out-of-scope detail is concealed as 404. |
| Member create/update/deactivate | `POST /admin/members`, `PATCH /admin/members/{id}`, `DELETE /admin/members/{id}` | SUPER_ADMIN/BPH only. DELETE is soft deactivation to INACTIVE; records and history remain. KB mutations receive 403. |

Members use the six PRD positions (`KETUA`, `WAKIL_KETUA`, `SEKRETARIS`, `BENDAHARA`, `KEPALA_BIDANG`, `ANGGOTA`). A head/member must reference a division in the same period; division is optional for other leadership positions. At most one ACTIVE `KEPALA_BIDANG` is allowed per division and period. A generated-column unique constraint is the database concurrency guard. Mutations emit activity events without sensitive data.

## Phase 6 aspiration implementation contract

| Consumer/action | Endpoint | Contract / access |
|---|---|---|
| Public destination options | `GET /api/v1/public/divisions` | Existing Phase 5 endpoint; returns ACTIVE divisions in the ACTIVE period only. |
| Submit aspiration | `POST /api/v1/public/aspirations` | Anonymous/public JSON write, separate fixed-window IP rate limit; required `division_id`, `subject`, `message`, and `is_anonymous`; optional `name`, `nim`, `email`. Name/NIM are optional per PRD; email is validated if supplied. |
| List aspirations | `GET /api/v1/admin/aspirations` | SUPER_ADMIN/BPH cross-division; KEPALA_BIDANG database division only. Supports `page`, `limit` (max 100), `search`, `status`, `division_id`, `date_from`, `date_to`. |
| Aspiration detail | `GET /api/v1/admin/aspirations/{id}` | Same role scope; out-of-scope KEPALA_BIDANG detail is returned as 404. |
| Update status | `PATCH /api/v1/admin/aspirations/{id}/status` | Same role scope; body `{ "status": "READ" | "RESOLVED" }`. Cross-division write is 403. |

Subject is capped at 200 Unicode characters; plain-text message at 5,000; optional name/NIM/email at 160/32/254 characters. Anonymous writes persist every identity column as NULL. Destination validation joins the requested division to the current ACTIVE period at write time. Status transitions are sequential only (`PENDING → READ → RESOLVED`); reopening and direct PENDING-to-RESOLVED are rejected. First `read_at` and `resolved_at` timestamps are preserved and stored in UTC. Public create returns only ID, PENDING status, and `submitted_at`; admin activity details contain only aspiration ID, status change, and division ID. Search is parameterized and literal-substring based. A supplied untrusted browser Origin is rejected for this public write route; requests without Origin remain usable by non-browser clients. No auth cookie is required.

## Phase 7 registration implementation contract

All paths below are under `/api/v1`. Lists use the standard `{success,message,data,meta}` envelope with `page` and `limit` (maximum 100). Validation errors are 422, duplicate NIM/unique conflicts and locked schema/deletion conflicts are 409, unauthenticated requests are 401, denied role/scope is 403, concealed out-of-scope details may be 404, and public throttling is 429 with `Retry-After`. Public responses never include submission data.

| Consumer/action | Endpoint | Contract / access |
|---|---|---|
| Public registration list | `GET /public/registrations` | Public filters: `status` (UPCOMING/OPEN/CLOSED), `division_id`, `period_id`, `page`, `limit`. Excludes DRAFT/deleted; default order OPEN, UPCOMING, then newest CLOSED. PRD 10.5 names OPEN-only listing; the direct Phase 7 instruction explicitly requests all three public effective states and governs this implementation. |
| Public registration detail | `GET /public/registrations/{slug}` | Public safe fields: public `id` (needed by the contracted submit path), title, slug, description, period/division names, UTC open/close times, effective public status, and ordered field key/label/type/placeholder/required/options. No created_by, internal field/option IDs, submission count, or answers. DRAFT/deleted/not-found returns 404. |
| Public submission | `POST /public/registrations/{id}/submissions` | No auth. Body `{ "answers": [{"field_key":"...","value":...}] }`; one answer per known key. Requires effective OPEN at transaction time. Returns only a safe reference and submitted_at. Duplicate form/NIM returns 409 with `DUPLICATE_SUBMISSION`; throttling returns 429. Body limit 64 KiB; independent default 5/IP/30m limiter. |
| Admin form list/create | `GET/POST /admin/registrations` | SUPER_ADMIN/BPH cross-division; KEPALA_BIDANG current DB division only. Filters: `page`, `limit`, `search`, stored `status`, `division_id`, `period_id`, `date_from`, `date_to`. Create accepts title, optional description, period_id, optional division_id, optional slug, open_at, close_at; starts DRAFT and creates required core fields transactionally. KB division is forced from principal. |
| Admin form detail/update/delete | `GET/PATCH/DELETE /admin/registrations/{id}` | Same role scope. PATCH updates metadata; period/division/schema-changing metadata is locked once submissions exist. DELETE soft-deletes only with zero submissions; otherwise 409 `REGISTRATION_HAS_SUBMISSIONS`. Out-of-scope detail may be concealed as 404. |
| Open/close | `POST /admin/registrations/{id}/open`, `POST /admin/registrations/{id}/close` | Same scope. Only DRAFT→OPEN and OPEN→CLOSED; CLOSED cannot reopen. Open validates title, period/division relationship, UTC open<close, all core fields and option/type schema. Open may precede open_at; public state is UPCOMING until the window. |
| Atomic form builder | `PUT /admin/registrations/{id}/fields` | Same scope. Replaces the complete ordered schema in one transaction. Only a DRAFT with no submissions can structurally change; after first submission field_key/type/required/options/order are immutable (409 `REGISTRATION_SCHEMA_LOCKED`). Labels/placeholders can be corrected as presentation metadata and existing submission snapshots remain unchanged. |
| Submission list/detail | `GET /admin/registrations/{id}/submissions`, `GET /admin/registrations/{id}/submissions/{submissionId}` | Same form scope; list filters: `search`, `date_from`, `date_to`, `page`, `limit`. List contains core identity and submitted_at only. Detail contains schema-snapshot labels/types and answers. No answer editing. |
| XLSX export | `GET /admin/registrations/{id}/export` | Same scope. Optional search/date filters; max 1,000 matched submissions. Columns: No., Nama, NIM, Email, No HP, Waktu Pendaftaran, then custom fields by snapshot display order. Formula-like/user strings are written as string cells. Response is attachment, `private, no-store`; successful export writes EXPORT/REGISTRATION_FORM with only form ID, division ID, and count. |

Supported field types: TEXT, TEXTAREA, EMAIL, NUMBER, SELECT, RADIO, CHECKBOX, DATE. Every form requires exactly one each of `full_name` (TEXT), `nim` (TEXT), `email` (EMAIL), and `phone` (TEXT); all are required and cannot be removed. NIM is trimmed/uppercased for uniqueness without numeric conversion; email is trimmed/lowercased; phone remains a bounded string. SELECT/RADIO accept exactly one current option; CHECKBOX accepts only an array of current options and is stored as JSON. Other types cannot have options. Unknown/duplicate keys, missing required values, invalid types/options/date/email/phone, and overlong values are rejected server-side.

Public state is derived in UTC: DRAFT is hidden; OPEN before open_at is UPCOMING; OPEN within inclusive `[open_at, close_at]` is OPEN; OPEN after close_at and stored CLOSED are CLOSED. A submission transaction locks its form row, rechecks status/window and loads the active schema before inserting the unique `(registration_form_id, normalized_nim)` submission and snapshots. Open/close/schema mutations use the same form-row lock, preventing close/schema races. A unique database constraint is the concurrency-safe duplicate guard.

The submission request is bounded to 64 KiB and schema to 50 fields. TEXT max is 500 Unicode characters; TEXTAREA max is 5,000; name 160, NIM 32, email 254, phone 32. No raw client IP is persisted (`ip_hash` stays NULL); the rate limiter uses the client IP only in transient process memory. Public payload values never enter HTTP/request or activity logs. Core identity is stored once in registration_submissions; answer rows keep core key/label/type/order snapshots without repeating those values, while custom answers store text or JSON. No applicant file upload, email/WhatsApp notification, review/selection workflow, scoring, quota management, or Phase 8 archive is part of this contract.
