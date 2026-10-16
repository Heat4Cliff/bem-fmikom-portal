# Theme System Revision, Hero Rollback, and Admin Integration

## Status

**READY FOR OWNER MULTI-THEME VISUAL REVIEW**

Implementation and automated checks are complete. Owner visual acceptance is still pending. The preview is listening at `127.0.0.1:3290`; HTTP smoke checks returned 200 for `/` and `/admin/login`, and browser accessibility snapshots loaded both routes. This does not claim manual visual QA or acceptance.

## Scope and Safety

- The rejected Hero News Carousel was removed. The Hero is again a stable framed visual and is not driven by featured-article data. It retains the existing optional configured hero image; without one, it shows the decorative placeholder.
- No API contract, database schema, authentication/RBAC, or business logic changed.
- `source_docs/` and `design_reference/` remain unchanged.
- No production deployment, release, tag movement, commit, or handoff ZIP regeneration occurred.

## Hero Rollback

The right side of the single existing Hero component renders a responsive framed composition: when a configured `hero.imageUrl` exists it remains available as an optional visual; otherwise the component shows layered outlines, subtle accent details, and a bottom content card with a calendar icon. It uses `hero.visualCaption`, including “Dokumentasi kegiatan akan ditampilkan setelah aset resmi tersedia.” The left side retains the existing eyebrow, heading, description, and “Jelajahi informasi” action. Hero no longer fetches or varies with featured-article data. The existing article fetch remains used by the homepage news section.

The former carousel component, its tests, and its animation CSS were removed. Responsive aspect ratios and existing typography/spacing remain.

## Theme Architecture

The existing shared Next.js component tree now uses five values on the root `data-theme` attribute. Theme changes update semantic CSS variables; public and admin share the same selection and persistence. Admin sidebar colors use dedicated sidebar tokens. Existing shared semantic classes continue to theme surfaces, text, forms, buttons, badges, alerts, dialogs, tables, upload surfaces, and navigation.

The picker is a compact Palette-icon + “Tema” button that opens an accessible Radix dialog, not a dropdown. It displays five visual cards in the requested order. Selection is immediate. Dialog focus trapping, Escape close, focus return, and keyboard trigger behavior come from the shared Radix dialog primitives. Cards have keyboard focus styling, `aria-pressed`, and a visible selected check.

The pre-hydration `/theme-init.js` accepts only `nawasena`, `classic`, `salt`, `matcha`, or `mint` from `bem-fmikom-theme`; it reads the previous `bem-theme` key for compatibility. Invalid/missing values fall back to Nawasena. Theme is stored locally, not in the API/database, and remains shared across routes and login/logout.

## Palette Tokens

Values below are exact current source values. Status meaning remains consistent across all themes (success green, warning amber, danger red, info blue).

| Theme | Canvas / Surface / Muted / Raised | Primary / Foreground | Accent | Sidebar |
|---|---|---|---|---|
| Nawasena | `#090806` / `#15130E` / `#201C12` / `#342B18` | `#FAC90A` / `#000000` | `#FAC90A`; muted `#5B4919` | `#15130E`, active `#5B4919` |
| Classic | `#F8F9FF` / `#FFFFFF` / `#EFF4FF` / `#DCE9FF` | `#00236F` / `#FFFFFF` | `#904D00`; original blue/amber values preserved | `#EFF4FF`, active `#00236F` |
| Salt | `#F7F5F0` / `#FFFFFF` / `#EFEDE6` / `#DFD9CC` | `#49453B` / `#FFFFFF` | `#715D43`, container `#E9DFCF` | `#EFEDE6`, active `#DFD9CC` |
| Matcha | `#F3F5EE` / `#FCFDF9` / `#E8EDE2` / `#D2DDCB` | `#35533D` / `#FFFFFF` | `#49684F`, container `#E1EADC` | `#E8EDE2`, active `#D2DDCB` |
| Mint | `#F2F8F6` / `#FCFEFD` / `#E6F1EE` / `#CCDFD9` | `#24594F` / `#FFFFFF` | `#397B6D`, container `#DEEEE9` | `#E6F1EE`, active `#CCDFD9` |

Classic's primary, surface, text, border, and amber palette were retained. Added themes are derived from cool/warm neutral, botanical, and fresh mint directions. Nawasena gold remains an accent; large vision/aspiration surfaces use the restrained panel token.

## Coverage and Review

- **Public:** all routes share the global semantic tokens, including landing, structure, news, gallery, announcements, registration, aspiration, footer, and shared dialogs.
- **Admin:** the same root theme selection applies to login and authenticated routes; sidebar has theme-specific background/hover/active tokens. Forms, tables, buttons, badges, alerts, upload components, dialogs, filters, and pagination use shared surface/status tokens where implemented.
- **Responsive picker:** one column below 420px and two columns at 420px and above; dialog content is bounded and scrollable on short screens.
- **Visual matrix:** implementation has not been manually inspected at each requested viewport/theme. Local HTTP/browser smoke evidence confirms the homepage and admin login load, but does not establish visual acceptance. Owner review remains required for public and admin representative screens in all five themes at 360×800, 390×844, 412×915, 768+, 1280×800, and 1440×900.
- **Accessibility:** automated interaction tests cover accessible control name, visible labels, selected state, selection for all five themes, persistence/remount, invalid-value fallback, Escape, focus restoration, and keyboard activation. This is not a formal WCAG contrast audit; representative visual/contrast review remains open.

## Verification

- `pnpm api:integration`: **PASS** — live MariaDB `TestDatabaseIntegration` ran; 25 subcases passed.
- `pnpm verify`: **PASS on full rerun** — lint, format, typecheck, 20 frontend test files / 99 tests, Next production build, Go formatting (70 files), vet, Go tests, and build. An earlier run had one transient gallery assertion failure; that test passed in isolation and all tests passed on the full rerun.
- `pnpm audit`: **PASS** — no known vulnerabilities.
- `git diff --check`: **PASS** after report save.
- Source evidence directories: verified unchanged after implementation.
- API contract/auth/RBAC and database schema: unchanged by this theme correction.

## Remaining Owner Review

Use the local preview at `http://127.0.0.1:3290/`. Review the landing Hero rollback, theme picker, representative public pages, `/admin/login`, and authenticated admin surfaces in all five themes. The local API health endpoint returned HTTP 200 at `http://127.0.0.1:8080/healthz`. Automated verification is not owner visual acceptance or deployment readiness.
