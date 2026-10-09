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
