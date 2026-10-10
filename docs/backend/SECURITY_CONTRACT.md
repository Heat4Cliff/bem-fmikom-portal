# Security & Authorization Contract

## Authentication
- Admin-only authentication via email/password.
- Short-lived access JWT (target 15–30 minutes).
- Refresh token target 7 days via Secure HttpOnly cookie in production HTTPS.
- Account must be active.

## RBAC
### SUPER_ADMIN
- Full access.
- Exclusive: user management, activity log, critical administration, period activation.

### BPH
- Cross-division operational/content access.
- Landing settings, announcements, organization docs, registrations and aspirations across divisions.
- No SUPER_ADMIN privilege management.

### KEPALA_BIDANG
- Must have `division_id`.
- Data queries and writes scoped to own division for relevant modules.
- Must not escape scope by URL/body/query manipulation.

## IDOR prevention
Never authorize solely from UI state or request-provided `division_id`.
For scoped resources, derive allowed division from authenticated principal and apply it in service/repository query. Cross-division resource identifiers must return 403 or not-found according to finalized disclosure policy.

## Public writes
- Server-side validation mandatory.
- Rate limit public aspiration and registration submission.
- Optional CAPTCHA only if spam pressure requires it.

## Rich text
Sanitize stored/rendered rich text. Do not trust admin editor HTML blindly.

## Documents
- Allowed baseline: PDF/DOC/DOCX/XLS/XLSX.
- Baseline max 10 MB.
- Validate extension + MIME as defense-in-depth.
- Generate safe stored filename; never trust client path/name.
- Private/non-public storage preferred.
- Download only through authorized endpoint and stream response.
- Log UPLOAD/DOWNLOAD/DELETE.
- Default to 10 MiB, configurable with `DOCUMENT_MAX_UPLOAD_BYTES`; production `DOCUMENT_STORAGE_PATH` must be absolute and outside public/static web roots.
- A KEPALA_BIDANG can only upload/list/download/manage the current database-assigned division. SUPER_ADMIN/BPH may access across divisions. Scoped detail/download denial may be concealed with 404.
- Stream to private staging, generate opaque storage keys, compute SHA-256, and verify extension plus declared MIME and file signature/container. Never use the original filename as a path.
- Download with attachment disposition, `Cache-Control: private, no-store`, and `X-Content-Type-Options: nosniff`.

## Secrets & logging
- `.env` ignored; `.env.example` contains placeholders only.
- Never log passwords, access/refresh tokens, reset secrets, full private document content, or unnecessary applicant payloads.
- Activity logs store operation metadata and minimal change detail.

## CORS / Cookies
- Explicit allowed origins from environment.
- Never wildcard origin with credentials.
- Secure/SameSite configuration documented for local vs production.

## Security tests required
- unauthenticated protected route -> 401.
- wrong role -> 403.
- KEPALA_BIDANG cross-division read/write -> denied.
- inactive account -> denied.
- direct private document path not available.
- path traversal filename attempts rejected.
- disallowed file/MIME/oversize upload rejected.
- public endpoint rate-limit behavior verified.
