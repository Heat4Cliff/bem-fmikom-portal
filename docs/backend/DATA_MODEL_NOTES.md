# Data Model Notes

Logical entities required by PRD:
- users
- periods
- divisions
- members / organization members
- article_categories
- articles
- article_category_map
- gallery_categories
- gallery_items
- announcements
- aspirations
- registration_forms
- registration_fields
- registration_submissions
- documents
- landing_settings
- activity_logs

Minimum uniqueness/index notes:
- UNIQUE users.email
- UNIQUE articles.slug
- UNIQUE announcements.slug
- UNIQUE registration_forms.slug
- articles(status, published_at)
- aspirations(status, division_id, created_at)
- registration_forms(status, open_at, close_at)
- registration_submissions(form_id, submitted_at)
- documents(period_id, division_id, category)
- activity_logs(user_id, action, created_at)

Single active period must be transactionally enforced if MySQL/MariaDB cannot express the desired invariant as an appropriate partial unique constraint.

Historical data should prefer archive/soft-delete semantics over destructive cascades when the PRD requires preservation.

## Aspirations (Phase 6)

- Each aspiration requires one `division_id` referencing a division. Submission is accepted only for an ACTIVE division in the ACTIVE period; historical aspirations remain attached to their original division.
- `divisions.show_in_public_structure` is independent of division status and aspiration eligibility. It controls only whether an ACTIVE division appears as a work-unit chart in `GET /public/organization`; `GET /public/divisions` continues to include ACTIVE aspiration destinations with either flag value.
- `name`, `nim`, and `email` are nullable; `subject` is bounded to 200 characters; `message` is plain text and bounded to 5,000 characters.
- `is_anonymous=true` requires all identity columns to be NULL, enforced by both the service and a database CHECK constraint.
- `submitted_at`, `read_at`, `resolved_at`, `created_at`, and `updated_at` use UTC. Status is constrained to `PENDING`, `READ`, or `RESOLVED`.
- Indexes cover `(division_id,status,created_at)` and `(status,submitted_at)` for routed lists and filtering.

## Registrations (Phase 7)

- `registration_forms` stores period, optional same-period division, creator, title/slug/description, UTC `open_at`/`close_at`, DRAFT/OPEN/CLOSED, and `deleted_at`. It has a globally unique slug, `open_at < close_at`, period/division composite FK, and indexes for period/division/status/window.
- `registration_fields` belongs to a form and has a unique stable snake_case `field_key`, label, one of the nine supported types (TEXT, TEXTAREA, EMAIL, NUMBER, SELECT, RADIO, CHECKBOX, DATE, IMAGE), nullable placeholder, required flag, and deterministic display order. `registration_field_options` stores unique values per SELECT/RADIO/CHECKBOX field only. IMAGE is structural and follows the existing post-submission schema lock.
- Each form creates exactly one required core field for full_name, nim, email, and phone. NIM/email/phone use strings; normalization never numeric-casts NIM.
- `registration_submissions` stores core identity once, `normalized_nim`, submitted_at, and nullable `ip_hash`; MariaDB enforces `UNIQUE(registration_form_id, normalized_nim)` and form/submitted indexes. Raw client IP is not stored.
- `registration_submission_answers` stores field ID when still present, field key/label/type/order snapshots, and either custom scalar text or structured checkbox JSON. Core snapshot rows omit identity values; detail/export joins them back to the single set of core columns.
- `registration_submission_files` stores one IMAGE attachment per `(submission_id, field_key_snapshot)`, including nullable field FK, safe display filename, opaque unique storage key, original/stored MIME, size, decoded dimensions, SHA-256, and timestamp. File bytes remain outside the database in a dedicated private storage root; no public URL is stored.
- First submission locks key/type/required/options/order permanently. IMAGE references are separate from answer `value_text/value_json`; binary data is never stored in answers. A transaction locks the form row during submission and schema/lifecycle mutation so a field edit or close cannot race validation/insertion. Staged file writes use promotion/compensation around DB commit. Answer snapshots remain readable without versioning the whole form.
- Forms with submissions cannot be deleted; otherwise DELETE soft-deletes. Activity records form and export operations without applicant PII. Export returns a private/no-store XLSX with a 1,000-submission cap.

### Phase 11H private registration image policy

- At most five IMAGE fields/files per form/submission, 5 MiB per image. Multipart request size is bounded to the aggregate image budget plus bounded JSON and fixed overhead.
- Accepted JPEG/JPG/JFIF, PNG, WebP, static GIF, TIFF, BMP. TIFF/BMP normalize to PNG; animated GIF is rejected; SVG, HEIC/HEIF, and AVIF are unsupported. Decoder policy caps each edge at 8,000 px and total area at 12 megapixels.
- Only the authenticated `/api/v1/admin/registrations/{id}/submissions/{submissionId}/files/{fileId}` route serves attachments and it enforces parent-form role/division scope. No anonymous attachment read route exists.

## Public media (Phase 11B)

- `media_assets` stores immutable public raster assets with an auto-increment ID, nullable uploader FK (`ON DELETE SET NULL`), allow-listed purpose, sanitized original filename, original MIME type, unique opaque storage key, stored MIME type, extension, size, width, height, SHA-256, and timestamps.
- CHECK constraints bound purpose to ARTICLE_COVER, ANNOUNCEMENT_IMAGE, GALLERY_PREVIEW, MEMBER_PHOTO, or LANDING_HERO; original inputs to JPEG/PNG/WebP/GIF/TIFF/BMP; stored outputs to JPEG/PNG/WebP/static GIF; and require non-zero size/dimensions. TIFF/BMP are normalized to PNG. Indexes cover uploader/purpose/time and SHA-256.
- Existing CMS/settings/member image columns remain URL strings for compatibility. For announcements and gallery previews, new internal references are normalized and persisted as `/api/v1/public/media/{id}` after the API verifies the asset exists and its purpose matches; legacy HTTPS image URLs remain accepted. The public media route is immutable, and the frontend resolves the API-relative reference using its configured API origin. No nullable `media_asset_id` columns were added because those URL fields already resolve the same endpoint without a destructive data migration. Gallery Google Photos album URLs remain independent external destinations.
- No resource FK or delete route is present. Upload actor/purpose are auditable; resource write APIs continue enforcing their own division scope. Unused assets are retained for manual cleanup rather than deleted while reference safety cannot be established across legacy URL fields.
