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
