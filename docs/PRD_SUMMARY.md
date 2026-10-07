# PRD Summary v1.4 — BEM FMIKOM UNUGHA Portal

Current source of truth: `source_docs/PRD_BEM_FMIKOM_UNUGHA_v1.4_Revisi_API_Contract.*`.

## Product
Satu platform terintegrasi dengan dua surface:
1. Public Website/Company Profile.
2. Admin Management Dashboard/CMS.

## Roles
- `SUPER_ADMIN`: full system access; user management, period management, activity log.
- `BPH`: cross-division operational/content access; landing settings; announcements; org docs; registration/aspiration cross-division.
- `KEPALA_BIDANG`: restricted to own `division_id` for articles, gallery, registration, documents, relevant aspirations.
- Public/Student: no login required for public content, aspiration submission, registration submission.

## In-scope modules
- Landing/home settings: hero, about, vision/mission, contact/social.
- Articles/news + categories.
- Gallery categories/items using Google Photos URL.
- Announcements + active popup/schedule/archive.
- Periods, divisions, organization members/structure.
- Public aspirations with optional identity and division/category routing.
- Dynamic registration form builder + public submissions + XLSX export.
- Private document archive with authorized download.
- User/admin management.
- Activity log/audit trail.
- Simple website visit/pageview statistics.

## Out of scope v1.4
- Student login/SSO.
- Public forum/live chat.
- Voting/Pemira.
- Multi-faculty.
- Native mobile app.
- Automatic email notifications.
- Mass gallery upload to app storage.
- Complex correspondence workflow / digital signing / automatic letter numbering.
- Advanced form builder logic (branching/scoring/conditional/file upload from applicant).

## Core business rules
- Exactly one active period at a time.
- Old period/history must be preserved.
- Public article endpoint returns only PUBLISHED content.
- Announcement popup requires active flag and valid date range.
- Aspiration status: `PENDING -> READ -> RESOLVED`.
- Registration status: `DRAFT/OPEN/CLOSED`.
- Dynamic fields supported: text, textarea, email, number, select, radio, checkbox, date.
- Document baseline max size: 10 MB, allowed PDF/DOC/DOCX/XLS/XLSX.
- Documents should be private; authorized endpoint streams downloads.
- Activity log minimum actions: LOGIN, LOGOUT, CREATE, UPDATE, DELETE, PUBLISH, UNPUBLISH, ACTIVATE, DEACTIVATE, UPLOAD, DOWNLOAD, EXPORT.

## Definition of done highlights
- Responsive public/admin implementation follows audited Stitch design.
- API contract implemented or explicitly excluded before code freeze.
- 401 for unauthenticated; 403 for forbidden/cross-division.
- KEPALA_BIDANG IDOR/cross-division manipulation blocked.
- Dynamic registration works end-to-end including export.
- Private documents cannot be fetched through a public direct path.
- Public writes have server-side validation + rate limiting.
- No secrets in repository.
- Backup, deployment, rollback documented before production.
