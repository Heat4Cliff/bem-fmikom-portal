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
- Access: SUPER_ADMIN, BPH, and KEPALA_BIDANG. Response is `{success,message,data}` with `Cache-Control: private, no-store`.
- `data` contains `generated_at`, `scope` (`organization` or `division` plus the database-resolved division identity), nullable `active_period`, and aggregate objects: `articles` (`total`, `published`, `draft`), `registrations` (`active_forms`, `active_form_submissions`), `aspirations` (`total`, `pending`, `read`, `resolved`), `documents.total`, and `announcements.total`.
- Articles exclude soft-deleted rows. Registration counts include forms whose stored status is OPEN and whose UTC open/close window currently includes the request; submission count includes only those forms. Aspirations count by current status. Document count excludes `delete_pending` records. Active announcements follow the public active endpoint's UTC schedule predicate.
- SUPER_ADMIN and BPH see cross-division aggregate counts. KEPALA_BIDANG division is loaded from the authenticated database principal; articles, forms/submissions, aspirations, and documents are restricted to that division. The announcement table has no division ownership and is readable cross-division by the existing KEPALA_BIDANG announcement-read policy, so its active count remains global. No caller-supplied division filter is accepted.
- Dashboard summary never returns names, applicant identity, aspiration content, document metadata, or activity details. Visitor/pageview metrics and collection are explicitly deferred post-MVP; no tracking endpoint or visitor data is introduced.

## Admin users — SUPER_ADMIN
- `GET /admin/users`
- `POST /admin/users`
- `GET /admin/users/{id}`
- `PATCH /admin/users/{id}`
- `POST /admin/users/{id}/deactivate`
- `POST /admin/users/{id}/reactivate`
- `POST /admin/users/{id}/reset-password`

## Phase 9 — landing settings, users, and activity logs

- `GET /public/settings` returns the singleton public landing content. It is anonymous and uses `Cache-Control: public, max-age=60, stale-while-revalidate=300`.
- `GET/PATCH /admin/settings` reads or replaces the typed singleton. SUPER_ADMIN and BPH may edit; KEPALA_BIDANG is denied. Public strings are plain text; external media/social/location URLs require HTTPS; CTA paths are same-origin paths or fragments. Save writes only changed field names to activity metadata.
- User routes above are SUPER_ADMIN only. `GET /admin/users` accepts `page`, `limit`, `search`, `role`, `status`, and `division_id`; response uses the standard list envelope. Create/update accepts `name`, `email`, `role`, nullable `division_id`, `status`, and create-only `password`. Role KEPALA_BIDANG requires an active division; other roles require null division. Passwords never appear in responses. Deactivate/reactivate are idempotent status transitions; reset-password accepts `password` and matching `confirm_password`. Status/reset revoke refresh sessions. Self-deactivation/demotion and deactivation of the last active SUPER_ADMIN return 409.
- `GET /admin/activity-logs` and `GET /admin/activity-logs/{id}` are SUPER_ADMIN only and read-only. List accepts `page`, `limit`, `search`, `actor_user_id`, `action`, `entity_type`, `date_from`, and `date_to`. Sensitive nested detail keys are redacted. IP and User-Agent are not returned.
- Validation failures use 422, role denial uses 403, missing resources use 404, and guarded conflicts use 409. Admin settings responses are `private, no-store`.
- `/admin`, `/admin/settings`, `/admin/users`, and `/admin/activity-logs` are the corresponding authenticated UI routes. Pageview analytics are deferred post-MVP and are not implied by the dashboard contract.

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

Document list filters: `page`, `limit`, `search`, `category`, `period_id`, `division_id`, `file_type`, `date_from`, and `date_to`. Upload fields are `name`, `category` (`SK`, `PROPOSAL`, `LPJ`, `OTHER`), `period_id`, `division_id`, and one `file` part. Metadata PATCH accepts `name` and `category`. Allowed formats are PDF/DOC/DOCX/XLS/XLSX; max upload defaults to 10 MiB and is configurable. File bytes are private and accessible only by the authenticated download route. The Phase 8 UI route is `/admin/arsip-dokumen`.

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

Article bodies and announcement bodies are sanitized on the server using the configured HTML allowlist. Gallery album links remain separate HTTPS Google Photos destination URLs. Announcement `image_url` and gallery `preview_image_url` are optional image references: they accept legacy HTTPS image URLs or an uploaded public-media reference whose persisted asset exists and has the matching `ANNOUNCEMENT_IMAGE` or `GALLERY_PREVIEW` purpose. Internal references are stored canonically as `/api/v1/public/media/{id}`; absolute API URLs and `media:{id}` are normalized to that path. Empty values normalize to `null`. Other relative paths and unsafe/malformed schemes are rejected. Announcement `start_at` is required as an ISO 8601 timestamp, `end_at` is optional, timestamps are normalized to UTC, and an end before the start is rejected with a field-level validation error. Date filters on announcement lists are `date_from` and `date_to`.

The full PRD v1.4 specifies the public gallery detail route by slug, so this contract replaces the earlier `{id}` placeholder. Announcement public archive visibility for scheduled/expired records remains a documented clarification item in `GAPS_AND_QUESTIONS.md`.

## Phase 5 organization implementation contract

Organization routes below are implemented under `/api/v1`. Lists use the shared paginated envelope and accept `page`, `limit`, `search`, `status`, and relevant `period_id`, `division_id`, or `position` filters. Invalid input returns 422; missing/scoped-out detail records return 404; unique-head/dependency conflicts return 409; unauthenticated requests return 401; denied roles or mutations return 403.

| Consumer/action | Endpoint | Access and response scope |
|---|---|---|
| Public active period | `GET /public/periods/active` | Public; returns the active period, or 404 if no period is active. |
| Public division list | `GET /public/divisions` | Public; all ACTIVE divisions in the active period (including aspiration-only destinations), ordered deterministically by display order, name, and ID; empty list if no period is active. The public-structure visibility flag does not filter aspiration destinations. |
| Public organization hierarchy | `GET /public/organization` | Public; minimal active-period, leadership, public work-division, head/member DTO; no internal IDs, NIM, status, or audit fields. Leadership is selected by position regardless of division assignment; only divisions with `show_in_public_structure=true` are rendered as work divisions. Returns an empty organization when no period is active. |
| Period list/detail | `GET /admin/periods`, `GET /admin/periods/{id}` | SUPER_ADMIN/BPH/KEPALA_BIDANG read; KB has no write access. |
| Period create/update/activate/archive | `POST /admin/periods`, `PATCH /admin/periods/{id}`, `POST /admin/periods/{id}/activate`, `POST /admin/periods/{id}/archive` | SUPER_ADMIN only. Activation transactionally archives the previously active period; history is retained. |
| Division list/detail | `GET /admin/divisions`, `GET /admin/divisions/{id}` | SUPER_ADMIN/BPH cross-division; KB list/detail is restricted to its current database division. |
| Division create/update/delete | `POST /admin/divisions`, `PATCH /admin/divisions/{id}`, `DELETE /admin/divisions/{id}` | SUPER_ADMIN/BPH. Delete conflicts when period/member/user/content references prevent safe deletion; no cascade removal. |
| Member list/detail | `GET /admin/members`, `GET /admin/members/{id}` | SUPER_ADMIN/BPH cross-division; KB reads only its current DB-assigned division. Out-of-scope detail is concealed as 404. |
| Member create/update/deactivate | `POST /admin/members`, `PATCH /admin/members/{id}`, `DELETE /admin/members/{id}` | SUPER_ADMIN/BPH only. DELETE is soft deactivation to INACTIVE; records and history remain. KB mutations receive 403. |

Members use the six PRD positions (`KETUA`, `WAKIL_KETUA`, `SEKRETARIS`, `BENDAHARA`, `KEPALA_BIDANG`, `ANGGOTA`). A head/member must reference a division in the same period; division is optional for other leadership positions. At most one ACTIVE `KEPALA_BIDANG` is allowed per division and period. A generated-column unique constraint is the database concurrency guard. Mutations emit activity events without sensitive data. Division admin DTOs include `show_in_public_structure`; create requests that omit it default to `true`, while update requests that omit it preserve its stored value. This metadata only controls the public organization chart and does not affect division status or aspiration eligibility.

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
| Public submission | `POST /public/registrations/{id}/submissions` | No auth. JSON remains `{ "answers": {"field_key": value} }` for text-only forms. Forms with IMAGE fields use `multipart/form-data`: one JSON `answers` part plus zero-or-one `image:{field_key}` file part for each IMAGE field. Requires effective OPEN at transaction time. Returns only a safe reference and submitted_at. Duplicate form/NIM returns 409 with `DUPLICATE_SUBMISSION`; throttling returns 429. JSON body limit 64 KiB; multipart body is bounded to five 5 MiB images plus JSON and multipart overhead; independent default 5/IP/30m limiter. |
| Admin form list/create | `GET/POST /admin/registrations` | SUPER_ADMIN/BPH cross-division; KEPALA_BIDANG current DB division only. Filters: `page`, `limit`, `search`, stored `status`, `division_id`, `period_id`, `date_from`, `date_to`. Create accepts title, optional description, period_id, optional division_id, optional slug, open_at, close_at; starts DRAFT and creates required core fields transactionally. KB division is forced from principal. |
| Admin form detail/update/delete | `GET/PATCH/DELETE /admin/registrations/{id}` | Same role scope. PATCH updates metadata; period/division/schema-changing metadata is locked once submissions exist. DELETE soft-deletes only with zero submissions; otherwise 409 `REGISTRATION_HAS_SUBMISSIONS`. Out-of-scope detail may be concealed as 404. |
| Open/close | `POST /admin/registrations/{id}/open`, `POST /admin/registrations/{id}/close` | Same scope. Only DRAFT→OPEN and OPEN→CLOSED; CLOSED cannot reopen. Open validates title, period/division relationship, UTC open<close, all core fields and option/type schema. Open may precede open_at; public state is UPCOMING until the window. |
| Atomic form builder | `PUT /admin/registrations/{id}/fields` | Same scope. Replaces the complete ordered schema in one transaction. Only a DRAFT with no submissions can structurally change; after first submission field_key/type/required/options/order are immutable (409 `REGISTRATION_SCHEMA_LOCKED`). Labels/placeholders can be corrected as presentation metadata and existing submission snapshots remain unchanged. |
| Submission list/detail | `GET /admin/registrations/{id}/submissions`, `GET /admin/registrations/{id}/submissions/{submissionId}` | Same form scope; list filters: `search`, `date_from`, `date_to`, `page`, `limit`. List contains core identity and submitted_at only. Detail contains schema-snapshot labels/types and answers. No answer editing. |
| XLSX export | `GET /admin/registrations/{id}/export` | Same scope. Optional search/date filters; max 1,000 matched submissions. Columns: No., Nama, NIM, Email, No HP, Waktu Pendaftaran, then custom fields by snapshot display order. Formula-like/user strings are written as string cells. Response is attachment, `private, no-store`; successful export writes EXPORT/REGISTRATION_FORM with only form ID, division ID, and count. |

Supported field types: TEXT, TEXTAREA, EMAIL, NUMBER, SELECT, RADIO, CHECKBOX, DATE, IMAGE. Every form requires exactly one each of `full_name` (TEXT), `nim` (TEXT), `email` (EMAIL), and `phone` (TEXT); all are required and cannot be removed. IMAGE is a structural field type with no scalar value/options and at most five IMAGE fields per form. A required IMAGE needs exactly one matching multipart file; optional IMAGE may be absent. NIM is trimmed/uppercased for uniqueness without numeric conversion; email is trimmed/lowercased; phone remains a bounded string. SELECT/RADIO accept exactly one current option; CHECKBOX accepts only an array of current options and is stored as JSON. Other types cannot have options. Unknown/duplicate keys, missing required values, invalid types/options/date/email/phone, and overlong values are rejected server-side.

Public state is derived in UTC: DRAFT is hidden; OPEN before open_at is UPCOMING; OPEN within inclusive `[open_at, close_at]` is OPEN; OPEN after close_at and stored CLOSED are CLOSED. A submission transaction locks its form row, rechecks status/window and loads the active schema before inserting the unique `(registration_form_id, normalized_nim)` submission and snapshots. Open/close/schema mutations use the same form-row lock, preventing close/schema races. A unique database constraint is the concurrency-safe duplicate guard.

JSON submission remains bounded to 64 KiB; multipart permits at most five 5 MiB images plus bounded JSON/overhead. Schema remains capped at 50 fields. TEXT max is 500 Unicode characters; TEXTAREA max is 5,000; name 160, NIM 32, email 254, phone 32. No raw client IP is persisted (`ip_hash` stays NULL); the rate limiter uses the client IP only in transient process memory. Public payload values never enter HTTP/request or activity logs. Core identity is stored once in registration_submissions; answer rows keep core key/label/type/order snapshots without repeating those values, while custom answers store text or JSON. IMAGE bytes never enter answer values. Private file metadata is in `registration_submission_files`; bytes use opaque keys in private storage and are only served through the authenticated admin route below. Email/WhatsApp notification, review/selection workflow, scoring, quota management, and Phase 8 archive are not part of this contract.

### Private registration IMAGE attachments — Phase 11H

- `GET /admin/registrations/{id}/submissions/{submissionId}/files/{fileId}` requires SUPER_ADMIN, BPH, or KEPALA_BIDANG. It verifies the attachment belongs to that form/submission and applies the existing database-current division scope; cross-division IDs are concealed as 404.
- Admin detail returns safe file metadata and the authenticated API path, never a storage key or public media URL. Bytes use the validated stored MIME with `Cache-Control: private, no-store`, `X-Content-Type-Options: nosniff`, and sandbox CSP.
- Multipart accepts only an `answers` JSON part and `image:{field_key}` file parts mapped to IMAGE fields. Anonymous users may submit images only as part of a valid OPEN form and receive no attachment read/list route.

## Phase 11B — Public image media

All paths below are under `/api/v1`. Image uploads are distinct from Phase 8 private documents. Admin uploads require an access JWT; public reads expose only an opaque numeric asset ID and approved raster image bytes.

| Consumer/action | Endpoint | Contract / access |
|---|---|---|
| Upload image | `POST /admin/media` | `multipart/form-data` with exactly one `purpose` and one `file`. SUPER_ADMIN/BPH may upload all bounded purposes. KEPALA_BIDANG may upload only `ARTICLE_COVER` and `GALLERY_PREVIEW`; consuming resource writes still enforce current division scope. Returns 201 with `{id,purpose,original_filename,original_mime_type,stored_mime_type,mime_type,file_extension,size_bytes,width,height,public_url,created_at}`. `mime_type` remains an alias for `stored_mime_type`. |
| Read image | `GET /public/media/{id}` | Public, immutable asset read. Returns raw bytes with the validated stored JPEG/PNG/WebP/GIF Content-Type, Content-Length, `X-Content-Type-Options: nosniff`, sandbox CSP, and `Cache-Control: public, max-age=31536000, immutable`; missing/invalid IDs return 404. |

Accepted inputs are JPEG (`.jpg`, `.jpeg`, `.jfif`), PNG (`.png`), WebP (`.webp`), static GIF (`.gif`), TIFF (`.tif`, `.tiff`), and BMP (`.bmp`). TIFF/BMP are decoded and stored as PNG; other supported formats retain their original encoding. Animated GIF is rejected with an explicit validation error to preserve animation rather than silently flattening it. SVG, HEIC/HEIF, and AVIF are not supported. `MEDIA_MAX_UPLOAD_BYTES` defaults to 5 MiB and accepts a bounded configured value; normalized output is subject to the same cap. Each dimension is at most 8,000 pixels and total image pixels at most 12 megapixels. The server checks extension, declared MIME, decoded format/signature, complete image decoding, size, and dimensions. The multipart request is capped before parsing. Validation returns 422 `VALIDATION_ERROR`; over-limit body/file returns 413 `MEDIA_TOO_LARGE`; unauthenticated upload is 401; disallowed role/purpose is 403. Unknown purpose, extra file/field, corrupt image, empty/unsafe filename, and type mismatch are rejected; directory components in uploaded filenames are stripped before storage. Article, announcement, and gallery image fields accept legacy HTTPS URLs or a purpose-matched persisted media reference; persisted internal references are canonical API-relative paths.

Metadata stores actor, bounded purpose, sanitized display filename, opaque generated storage key, MIME/extension, byte size, dimensions, SHA-256, and timestamps. `public_url` is an API-relative `/api/v1/public/media/{id}` path; the browser prefixes the configured API origin. Local bytes live outside the Next.js public/static tree behind the API read handler. Existing URL string fields remain compatible; the gallery's Google Photos album URL remains a separate external destination. Upload records are immutable and have no delete route. Failed metadata/activity transactions remove the just-written object; uploaded assets that are never attached to a saved resource are retained for manual storage maintenance. Resource save authorization continues to enforce role/division scope.
