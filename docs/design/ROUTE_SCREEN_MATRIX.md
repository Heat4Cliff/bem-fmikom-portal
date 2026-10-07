# Route ↔ Screen Implementation Matrix

This file is intentionally route-oriented. Route names may be refined during Phase 0 if Next.js route grouping provides a cleaner URL structure, but public URLs and API contract semantics must stay stable.

## Public route group
- `/` → public landing.
- `/berita` → published article list.
- `/berita/[slug]` → published article detail.
- `/pengumuman` → announcement archive/list.
- `/pengumuman/[slug]` → announcement detail.
- `/galeri` → gallery categories/events.
- `/galeri/[id-or-slug]` → gallery detail with Google Photos CTA.
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
