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
