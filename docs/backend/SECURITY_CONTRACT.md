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

## Public image media (Phase 11B)
- `POST /api/v1/admin/media` requires the current authenticated account. SUPER_ADMIN/BPH can upload the five bounded purposes; KEPALA_BIDANG is restricted to ARTICLE_COVER and GALLERY_PREVIEW. Later content/member/settings mutation endpoints remain responsible for resource ownership and database-current division checks; purpose selection alone never grants resource access.
- `GET /api/v1/public/media/{id}` is the only byte-serving path. Never mount media storage as a static directory or expose its filesystem path.
- Allow JPEG (`.jpg`, `.jpeg`, `.jfif`), PNG, WebP, static GIF, TIFF, and BMP only; reject SVG, animated GIF, HEIC/HEIF, AVIF, mismatched extensions/MIME/signature, and corrupt images. Verify decoded format and image dimensions. TIFF/BMP normalize to PNG. Max request default is 5 MiB; max dimension is 8,000 and max image area is 12 megapixels. Normalized output is capped too. `MEDIA_MAX_UPLOAD_BYTES` is configurable with an upper bound.
- Parse multipart under `http.MaxBytesReader`; accept exactly one file and one bounded purpose. Use generated opaque storage keys, sanitized display names, SHA-256 metadata, and the storage abstraction. Never log image bytes or raw upload tokens. Activity records actor, purpose, asset ID, and byte size only.
- Public image responses use the validated content type, `nosniff`, sandbox CSP, and immutable one-year caching. Assets are append-only so a URL never changes bytes. Deletion is not exposed; no cleanup may remove a referenced resource image. DB/activity failure removes the newly stored object; unattached uploads are retained for manual cleanup.

## Private registration IMAGE attachments (Phase 11H)

- Registration applicant images are private submission data and must never be written to the public `media_assets` store or served from `/api/v1/public/media/{id}`.
- An anonymous caller can upload only the bounded image parts attached to a currently OPEN registration submission. JSON `answers` remains the scalar payload; multipart file parts are mapped only to schema fields of type IMAGE. No base64, arbitrary file field, or applicant read/list endpoint is allowed.
- Current policy: up to five IMAGE fields/files per form/submission, 5 MiB per image, 25 MiB aggregate image budget plus bounded JSON/multipart overhead; dimension and pixel bounds match the validated Phase 11G raster pipeline. JPEG/JFIF, PNG, WebP, static GIF, TIFF, and BMP are accepted; TIFF/BMP normalize to PNG; animated GIF and SVG/HEIC/AVIF are rejected.
- Validate sanitized extension, declared MIME, decoded signature/format, complete decoding, file size, dimensions, and pixel area. Filenames are display metadata only; storage keys are random and opaque. Storage directory must be outside public/static roots and production path must be absolute.
- Stage bytes before final promotion. If validation, DB insert, storage promotion, or DB commit fails, delete staged and promoted objects best-effort while rolling back DB state. File data and applicant answers must not enter logs or activity metadata.
- `GET /api/v1/admin/registrations/{id}/submissions/{submissionId}/files/{fileId}` requires SUPER_ADMIN/BPH/KEPALA_BIDANG and reuses the form's database-current division authorization. The route verifies all parent identifiers, returns 404 for scoped-out/mismatched IDs, and responds with `private, no-store`, correct stored MIME, `nosniff`, and sandbox CSP. Attachment metadata exposes no storage key or public URL.

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
