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
- `GET /public/gallery-categories/{id}`
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
