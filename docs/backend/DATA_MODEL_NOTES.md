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
- `name`, `nim`, and `email` are nullable; `subject` is bounded to 200 characters; `message` is plain text and bounded to 5,000 characters.
- `is_anonymous=true` requires all identity columns to be NULL, enforced by both the service and a database CHECK constraint.
- `submitted_at`, `read_at`, `resolved_at`, `created_at`, and `updated_at` use UTC. Status is constrained to `PENDING`, `READ`, or `RESOLVED`.
- Indexes cover `(division_id,status,created_at)` and `(status,submitted_at)` for routed lists and filtering.

## Registrations (Phase 7)

- `registration_forms` stores period, optional same-period division, creator, title/slug/description, UTC `open_at`/`close_at`, DRAFT/OPEN/CLOSED, and `deleted_at`. It has a globally unique slug, `open_at < close_at`, period/division composite FK, and indexes for period/division/status/window.
- `registration_fields` belongs to a form and has a unique stable snake_case `field_key`, label, one of the eight supported types, nullable placeholder, required flag, and deterministic display order. `registration_field_options` stores unique values per SELECT/RADIO/CHECKBOX field only.
- Each form creates exactly one required core field for full_name, nim, email, and phone. NIM/email/phone use strings; normalization never numeric-casts NIM.
- `registration_submissions` stores core identity once, `normalized_nim`, submitted_at, and nullable `ip_hash`; MariaDB enforces `UNIQUE(registration_form_id, normalized_nim)` and form/submitted indexes. Raw client IP is not stored.
- `registration_submission_answers` stores field ID when still present, field key/label/type/order snapshots, and either custom scalar text or structured checkbox JSON. Core snapshot rows omit identity values; detail/export joins them back to the single set of core columns.
- First submission locks key/type/required/options/order permanently. A transaction locks the form row during submission and schema/lifecycle mutation so a field edit or close cannot race validation/insertion. Answer snapshots remain readable without versioning the whole form.
- Forms with submissions cannot be deleted; otherwise DELETE soft-deletes. Activity records form and export operations without applicant PII. Export returns a private/no-store XLSX with a 1,000-submission cap.
