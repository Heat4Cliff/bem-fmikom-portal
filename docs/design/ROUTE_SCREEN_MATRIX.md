# Route ↔ Screen Implementation Matrix

This file is intentionally route-oriented. Route names may be refined during Phase 0 if Next.js route grouping provides a cleaner URL structure, but public URLs and API contract semantics must stay stable.

## Public route group
- `/` → public landing.
- `/berita` → published article list.
- `/berita/[slug]` → published article detail.
- `/pengumuman` → announcement archive/list.
- `/pengumuman/[slug]` → announcement detail.
- `/galeri` → gallery categories/events.
- `/galeri/[slug]` → gallery detail with Google Photos CTA (full PRD v1.4 is canonical).
- `/pendaftaran` → registration list.
- `/pendaftaran/[slug]` → registration info/detail.
- `/pendaftaran/[slug]/form` → dynamic form rendering + submission.
- `/aspirasi` → public aspiration submission.

## Admin route group
- `/admin` → overview.
- `/admin/berita` → articles.
- `/admin/berita/new`, `/admin/berita/[id]/edit` → editor.
- `/admin/kategori-berita` → categories (screen may share management patterns even when no dedicated Stitch folder exists; do not invent visual language).
- `/admin/galeri` → gallery categories/items.
- `/admin/pengumuman` → announcements.
- `/admin/struktur` → organization structure.
- `/admin/divisi` → divisions.
- `/admin/pengurus` → members.
- `/admin/pengurus/new`, `/admin/pengurus/[id]/edit` → member form.
- `/admin/periode` → periods.
- `/admin/pendaftaran` → registrations.
- `/admin/pendaftaran/[id]/builder` → form builder.
- `/admin/pendaftaran/[id]/submissions` → submission list/detail/export.
- `/admin/aspirasi` → aspirations.
- `/admin/aspirasi/[id]` → aspiration detail.
- `/admin/dokumen` → document archive.
- `/admin/settings` → landing settings.
- `/admin/users` → SUPER_ADMIN user management.
- `/admin/activity-logs` → SUPER_ADMIN audit trail.

## Role-aware navigation rule
Never show a menu item the current role cannot use, but never rely on hidden navigation for security. Backend permission remains mandatory.

## Phase 3 implementation status
- `/` is implemented as the responsive public landing page using the audited desktop reference `beranda_publik_bem_fmikom_unugha` and mobile reference `beranda_publik_mobile_bem_fmikom_unugha`.
- `/aspirasi` is a Phase 3 navigation placeholder only; it has no form, API call, or submission behavior.
- Other public content routes remain future-phase work; landing navigation targets stable section anchors until those routes are implemented.

## Phase 4 implementation status
- `/berita`, `/berita/[slug]`, `/pengumuman`, `/pengumuman/[slug]`, `/galeri`, and `/galeri/[slug]` are implemented against the public REST endpoints.
- `/admin/berita`, `/admin/berita/new`, `/admin/berita/[id]/edit`, `/admin/kategori-berita`, `/admin/galeri`, and `/admin/pengumuman` are implemented as role-aware authenticated CMS routes.
- `screen.png` is the visual target for the reviewed news, announcement, gallery, and landing screens. Corresponding `code.html` was used only as a structural/style reference. The category route has no dedicated Stitch screen and reuses the audited content-admin pattern.
- Admin navigation hides system permissions from BPH and KEPALA_BIDANG, while the backend remains authoritative for content and division scope.

## Phase 5 implementation status

- `/admin/periode` manages period metadata and activate/archive lifecycle; activation atomically archives the previous active period.
- `/admin/divisi` and `/admin/pengurus` provide searchable/filterable management with responsive mobile card variants; member create/edit forms require a division only for KEPALA_BIDANG and ANGGOTA.
- `/admin/struktur` renders the selected period's leadership, divisions, heads, and members. `/` consumes the public active-period organization DTO and preserves an explicit empty state when no active organization exists.
- Audited Stitch structure, division, member, and period `screen.png` files informed the visual target. Their `code.html` files remain layout/style evidence only. Admin visual browser review was performed at the available local browser viewport; exact 360/390/412 device emulation remains a verification limitation recorded in the Phase 5 report.
- Backend authorization is authoritative: SUPER_ADMIN manages periods; SUPER_ADMIN/BPH manage divisions and members; KEPALA_BIDANG is read-only and scoped to its database-assigned division.

## Phase 6 implementation status

- `/aspirasi` is a working public form using active-period division choices from the existing public divisions API. It supports optional identity, explicit anonymous submission, field validation, pending/success/rate-limit/server-error states, and no file upload.
- `/admin/aspirasi` shows filters and paginated results as a desktop table and mobile cards. `/admin/aspirasi/[id]` renders plain-text details and a simple timestamped status timeline.
- SUPER_ADMIN/BPH review across divisions; KEPALA_BIDANG is restricted to its current database division. Only the next sequential status action is available; resolved aspirations are terminal.
- The aspiration `screen.png` references are the visual target. Detail-only attachment, disposition, SLA, export, and analytics controls are excluded because they are not in the Phase 6 PRD/API behavior. `code.html` remains layout/style evidence only.
