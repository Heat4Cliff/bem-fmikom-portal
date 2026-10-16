# Dual Brand Theme Implementation Report

> Historical two-theme handoff. Current five-theme implementation and verification are documented in [MULTI_THEME_REVISION_REPORT.md](MULTI_THEME_REVISION_REPORT.md).

## Status

**READY FOR OWNER DUAL-THEME VISUAL REVIEW**

Implementation and automated regression gates pass. Manual owner acceptance is still pending. This report does not declare the work final or released.

## Theme Architecture

- The existing component tree and routes use semantic CSS variables; no duplicate theme-specific component tree or new theme dependency was added.
- Theme is represented by `data-theme="nawasena"` or `data-theme="classic"` on the root document.
- Before hydration, `apps/web/public/theme-init.js` reads the non-sensitive `bem-theme` local-storage preference. First visit and invalid values use Nawasena. The root inline bootstrap file is loaded with Next.js `beforeInteractive` to avoid rendering the default theme before restoration.
- The reusable selector updates the root theme and saves the choice. Storage errors do not prevent changing the current tab's theme.

## Nawasena

- Uses the requested legacy brand colors through semantic tokens: gold `#FAC90A`, white `#FFFFFF`, dark brown `#5B4919`, and black `#000000`.
- Current page layouts and behavior are retained. The visual system uses near-black surfaces, light text, gold emphasis, and restrained brown supporting surfaces.
- Existing public content, forms, dialogs, admin surfaces, tables, uploaders, and status states now use shared semantic surface and status tokens where a fixed white surface or fixed status color would have broken theme switching.
- Gold is paired with dark text on primary controls; body text uses light neutral colors on dark surfaces. Brown remains a supporting tone rather than small body text.

## Classic

- The existing blue/light token values are preserved under `data-theme="classic"`; the existing page layout, navigation, cards, and component hierarchy remain shared.
- Intentional difference: the theme selector appears in both themes because it is needed to switch back. No legacy layout or navigation was brought back.

## Theme Switcher

- Public: desktop navigation and mobile drawer.
- Admin: login, desktop admin header, and mobile admin drawer.
- Accessible name: `Ubah tema`; options are `Nawasena` and `Classic`.
- No public Admin/Login navigation link was added.

## Responsive Review

Automated regression confirms shared responsive components remain in use. Browser spot checks covered the public homepage and admin login at the available desktop preview size (1280×720). The requested exact viewport matrix has not received owner visual acceptance; the rows below remain pending rather than being reported as passed.

| Viewport | Nawasena | Classic | Overflow | Notes |
|----------|----------|---------|----------|-------|
| 1440×900 | Pending owner review | Pending owner review | Pending | Not captured at this exact size |
| 1280×800 | Pending owner review | Pending owner review | Pending | Spot check was 1280×720, not this exact height |
| 360×800 | Pending owner review | Pending owner review | Pending | Mobile owner check required |
| 390×844 | Pending owner review | Pending owner review | Pending | Mobile owner check required |
| 412×915 | Pending owner review | Pending owner review | Pending | Mobile owner check required |

Representative pages requested for owner review: public landing, organization structure, registration and form, aspiration; admin dashboard, representative forms, uploader, registration builder, and organization management. Admin login and public landing have had a visual smoke check in both themes. Authenticated admin pages and remaining public routes still need owner review.

## Accessibility

- The theme control is a native labeled select, operable by keyboard, and announces its selected value.
- Focus remains visible; primary gold controls use black text. Contrast checks for key token pairs: gold/black 13.42:1, near-white `#FFFFFF` on `#090806` 20.02:1, muted `#C8C1AD` on `#090806` 11.14:1, and border `#806726` against `#090806` 3.70:1. Gold on white and brown as small body text are avoided.
- Existing reduced-motion behavior is retained. Automated tests cover the theme control and persistence; exhaustive assistive-technology review is pending owner acceptance.

## Regression

- `pnpm api:integration`: PASS against the existing MariaDB development/test runtime; `TestDatabaseIntegration` ran, with migrations 00001–00011 and 25 subcases.
- `pnpm verify`: PASS — Biome lint and format (129 files), TypeScript, 20 frontend test files / 97 tests, Next production build, Go formatting, vet, tests, and build.
- `pnpm audit`: PASS — no known vulnerabilities found.
- `git diff --check`: PASS.
- `source_docs/` integrity: unchanged.
- `design_reference/` integrity: unchanged.
- Theme component tests: 8 passed, including default, switching, persistence/restoration, invalid-value fallback, and accessible label.

## Files Changed

All source changes are in `apps/web/`:

- `src/app/globals.css`, `src/app/layout.tsx`
- `public/theme-init.js`
- `src/components/ui/theme-switcher.tsx`, `src/components/ui/theme-switcher.test.tsx`
- `src/components/ui/primitives.tsx`, `src/components/ui/overlays.tsx`
- `src/components/public/public-navigation.tsx`, `src/components/public/public-shell.tsx`
- `src/components/admin/admin-shell.tsx`, `src/components/admin/media-upload.tsx`
- `src/app/(admin)/admin/login/page.tsx`, `src/app/(admin)/admin/aspirasi/[id]/page.tsx`
- `src/app/(public)/aspirasi/page.tsx`, `src/app/(public)/berita/page.tsx`, `src/app/(public)/berita/[slug]/page.tsx`, `src/app/(public)/galeri/page.tsx`, `src/app/(public)/galeri/[slug]/page.tsx`, `src/app/(public)/pengumuman/page.tsx`, `src/app/(public)/pengumuman/[slug]/page.tsx`
- `src/features/aspirations/aspiration-admin.tsx`, `src/features/aspirations/public-aspiration-form.tsx`
- `src/features/content/announcement-manager.tsx`, `src/features/content/article-editor.tsx`, `src/features/content/article-manager.tsx`, `src/features/content/category-manager.tsx`, `src/features/content/gallery-manager.tsx`, `src/features/content/public-content.tsx`
- `src/features/documents/document-archive.tsx`
- `src/features/organization/organization-admin.tsx`, `src/features/organization/public-organization-page.tsx`
- `src/features/public/home/components/about-section.tsx`, `announcement-section.tsx`, `contact-sections.tsx`, `gallery-section.tsx`, `hero-section.tsx`, `news-section.tsx`, `organization-section.tsx`, `registration-section.tsx`, `section-primitives.tsx`
- `src/features/registrations/public-registration.tsx`, `src/features/registrations/registration-admin.tsx`
- `src/features/system/activity-admin.tsx`, `src/features/system/settings-admin.tsx`, `src/features/system/users-admin.tsx`

## Release State

- Existing `v1.0.0-campus-handoff` tag remains on the original commit `6c35140832a6c3c32f2ad472aedcba36dd86f86f`.
- No production deployment was performed.
- No new tag or GitHub Release was created.
- Existing handoff ZIP files and `deployment_handoff/` were not regenerated or overwritten.
- No commit was created.

## Manual Review

Local URL: **http://127.0.0.1:3290/**

Admin login: **http://127.0.0.1:3290/admin/login**

Review both Nawasena and Classic on the representative public and admin pages listed above. The API health check is available at `http://127.0.0.1:8080/healthz`. The local frontend and API are left running for owner review.
