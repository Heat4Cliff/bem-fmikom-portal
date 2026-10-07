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
