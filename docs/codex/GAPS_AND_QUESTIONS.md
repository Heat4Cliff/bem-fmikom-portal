# Phase 0 — Gaps and Questions

These are genuine contract or evidence gaps. They are kept explicit rather than resolved by guesswork. Criticality describes when the answer is required; these gaps do not prevent completion of the Phase 0 audit.

| ID | Gap / question | Impact | Criticality / answer needed before |
|---|---|---|---|
| G-01 | Should GET /public/gallery-categories/{slug} from full PRD remain canonical, replacing {id} in the API summary? | Public gallery links/client types can diverge. | Must resolve before gallery API/client implementation. |
| G-02 | Add the missing GET /admin/article-categories from v1.4 to the API summary? | Category list UI cannot reliably load. | Before content CMS implementation. |
| G-03 | Reconcile user deactivation: PRD DELETE /admin/users/{id} (soft deactivate) versus summary deactivate/reactivate/reset-password actions. Are reactivate and reset-password in scope? | User management actions may have no authoritative semantics; accidental scope expansion. | Before user-management implementation. |
| G-04 | Confirm added detail/list routes not present in the full PRD table: admin user/period/division/member detail GETs and gallery-items GET. | Edit/detail screens may rely on undocumented API behavior. | Before the corresponding API/client feature. |
| G-05 | Is GET /public/periods/active an approved additive endpoint, or should the public organization/settings response carry active-period metadata? | Duplicate/inconsistent period reads. | Before public organization integration. |
| G-06 | How are page views collected and deduplicated: backend endpoint/aggregation or approved analytics provider? Define privacy/retention and dashboard fields. | Dashboard visit metric has no data source contract. | Before statistics implementation. |
| G-07 | What is the login screen visual target? PRD requires login, but the 37-screen index has no login reference. | Cannot claim full Stitch fidelity for /admin/login. | Before login visual implementation/review. |
| G-08 | Should /admin/kategori-berita be a standalone screen or part of article management? There is no screen-index entry and no dedicated Stitch reference. | Route/screen count and navigation remain unclear. | Before category UI implementation. |
| G-09 | What key combination enables duplicate registration checks per form (NIM, email, or both), and how should blank/normalized values compare? | Duplicate submission behavior is stakeholder-dependent. | Before registration submit/duplicate logic. |
| G-10 | Can fields change after first submission? If yes, are old submissions interpreted against a saved form/field snapshot or only the active definition? | Historical submission rendering/export may change or lose meaning. | Before form edit and submission persistence implementation. |
| G-11 | What happens to stored bytes and metadata on document delete (soft-delete, retention window, permanent removal)? | Data retention and restore/security behavior are undefined. | Before document deletion implementation. |
| G-12 | For cross-division IDs, should API return 403 or 404 consistently? | Error behavior may disclose resource existence; contract currently allows either. | Before finalizing authorization/error tests. |
| G-13 | Define refresh token rotation, revocation, and reuse detection; confirm local versus production Secure/SameSite behavior. | Session theft/logout guarantees and browser integration differ. | Before auth implementation. |
| G-14 | Define rate-limit windows/limits and response headers for public aspiration and registration writes. | Security requirement is testable only after thresholds are chosen. | Before public write implementation/release. |
| G-15 | Define activity-log retention, detail redaction, and whether failed refresh attempts are logged as a named action/event. | Audit usefulness versus privacy/storage; PRD mentions refresh failure, minimum vocabulary does not. | Before activity-log implementation. |
| G-16 | How should remaining 29/37 references adapt to mobile where no separate mobile screen exists? Validate against responsive acceptance at 360/390/412 widths. | Mobile layout cannot be visually inferred from nonexistent captures. | Before visual sign-off for each route. |
| G-17 | Confirm whether public archive lists expired/inactive announcements and exact visibility rules; summary says archive/list but lacks detailed publication policy. | Public list may expose or hide the wrong announcements. | Before announcement API/UI implementation. |
| G-18 | Define upload constraints/configuration overrides and whether browser preview is expected for small media versus document archive; PRD gives 10 MB baseline and file_path is internal. | Upload UX and deployment storage configuration may vary. | Before upload implementation. |

## Contract reconciliation gate

Before implementing affected APIs, update docs/backend/API_CONTRACT.md in this repository to match v1.4 or record an approved additive change. Also add route-level permissions, request/response schemas, status/error behavior, and required filters where absent. Do not implement unresolved behavior by inference.

## Source evidence gaps

- SCREEN_INDEX.md claims 37 mapped references and inventory confirms 37 pairs, but the list has no explicit /admin/login or article-category reference.
- Exactly 8 screens have explicit mobile reference names; absence of a separate mobile screen is not evidence that the route is desktop-only.
- This report did not claim human visual review of the screens or run UAT.

## Phase 2 disposition

- **G-13 — Resolved for the authentication foundation.** Phase 2 fixes access JWT lifetime at 15 minutes and refresh lifetime at seven days; refresh tokens are opaque, hashed with SHA-256, rotated transactionally, and replay revokes the session family. Cookie defaults/configuration, production Secure behavior, trusted-origin checks, active-user reload, and `/api/v1/auth/*` route prefix are recorded in `DECISIONS.md`. The API Contract's example `expires_in: 1800` conflicts with the user-fixed 900 seconds; implementation uses 900 and the conflict is explicit.
- **G-07 — Carried forward as a visual evidence limitation.** `/admin/login` uses the established design tokens and form primitives (`DESIGN SYSTEM CONSISTENT`); there is no dedicated Stitch login image and no pixel-parity claim.
- **Other Phase 0 gaps remain open.** Gallery/article-category routes, analytics, registration duplicate/snapshot rules, document retention, 403/404 disclosure, public write thresholds, mobile evidence, announcement visibility, and upload configuration do not block the Phase 2 database/auth/RBAC foundation and were not silently decided here.
- **Initial database integration was unverified.** The first Phase 2 attempt had no Docker runtime; that historical result was superseded by the live MariaDB rerun in `PHASE_2_REPORT.md`.

## Phase 3 public-site gaps

| ID | Gap | Impact / handling |
|---|---|---|
| G-19 | No approved standalone BEM/FMIKOM logo, hero photo, organization portrait, news thumbnail, or gallery image was found in the copied source package. Stitch `code.html` references remote generated media, which is not a production asset source. | CSS artwork and the Phase 1 text mark are used as clearly labeled visual placeholders. Replace them when approved local assets and reuse rights are supplied; this is the main remaining visual-parity gap. |
| G-20 | Official contact address, email, phone, and social-media URLs are absent from the source data. | Production shows a clear empty state; no contact details or social links are invented. |
| G-21 | An approved public origin/canonical host is not specified. | `NEXT_PUBLIC_SITE_URL` remains optional; configure it before publishing canonical and sitemap URLs. |
| G-22 | The current API server has no public content endpoints, and product content tables belong to later phases. | Dynamic home sections stay empty in production; typed demo fixtures are isolated behind a local-only flag. Public API/database integration remains future work. |

The later live MariaDB result and Phase 2B/2C evidence are recorded in `PHASE_2_REPORT.md`.

## Phase 4 disposition

| Gap | Phase 4 outcome |
|---|---|
| G-01 gallery detail key | Resolved from the full PRD v1.4: public gallery detail uses `{slug}`. `API_CONTRACT.md` and `ROUTE_SCREEN_MATRIX.md` now agree. |
| G-02 article category list endpoint | Resolved: `GET /api/v1/admin/article-categories` is implemented and consumed by category/article administration. |
| G-08 category screen reference | The route is implemented using existing content-admin design patterns. No dedicated Stitch category screen exists, so screen-specific visual parity remains unclaimed. |
| G-12 cross-division read status | Resolved policy: mutations return 403; scoped detail reads may return 404 to conceal resource existence. Both deny access. |
| G-17 public announcement visibility | Still requires product confirmation. Current archive/detail expose only records whose `is_active` flag is true (including derived SCHEDULED/EXPIRED); `/active` additionally requires the UTC time window. Confirm whether scheduled and expired entries belong in the public archive before release. |
| G-18 media/upload configuration | Phase 4 explicitly uses HTTPS URL fields and does not implement file upload or storage. Revisit storage constraints only if a later authorized scope adds uploads. |

The unresolved G-17 rule is a release clarification, not a database/auth integration failure. See `PHASE_4_REPORT.md` for the exercised behavior and verification evidence.
