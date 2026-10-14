# Phase 0 — Gaps and Questions

These are genuine contract or evidence gaps. They are kept explicit rather than resolved by guesswork. Criticality describes when the answer is required; these gaps do not prevent completion of the Phase 0 audit.

| ID | Gap / question | Impact | Criticality / answer needed before |
|---|---|---|---|
| G-01 | Should GET /public/gallery-categories/{slug} from full PRD remain canonical, replacing {id} in the API summary? | Public gallery links/client types can diverge. | Must resolve before gallery API/client implementation. |
| G-02 | Add the missing GET /admin/article-categories from v1.4 to the API summary? | Category list UI cannot reliably load. | Before content CMS implementation. |
| G-03 | Reconcile user deactivation: PRD DELETE /admin/users/{id} (soft deactivate) versus summary deactivate/reactivate/reset-password actions. Are reactivate and reset-password in scope? | Resolved for Phase 9 by its explicit POST deactivate/reactivate/reset-password scope and the existing API contract. The legacy PRD DELETE form is not exposed. | Resolved for Phase 9; broader contract cleanup remains a documentation task. |
| G-04 | Confirm added detail/list routes not present in the full PRD table: admin user/period/division/member detail GETs and gallery-items GET. | Edit/detail screens may rely on undocumented API behavior. | Before the corresponding API/client feature. |
| G-05 | Is GET /public/periods/active an approved additive endpoint, or should the public organization/settings response carry active-period metadata? | Duplicate/inconsistent period reads. | Before public organization integration. |
| G-06 | How are page views collected and deduplicated: backend endpoint/aggregation or approved analytics provider? Define privacy/retention and dashboard fields. | Visitor analytics are explicitly deferred post-MVP; Phase 9 operational dashboard metrics are implemented without visitor tracking. Reopen only if visitor analytics return to scope. | Deferred post-MVP; no Phase 9 blocker. |
| G-07 | What is the login screen visual target? PRD requires login, but the 37-screen index has no login reference. | Cannot claim full Stitch fidelity for /admin/login. | Before login visual implementation/review. |
| G-08 | Should /admin/kategori-berita be a standalone screen or part of article management? There is no screen-index entry and no dedicated Stitch reference. | Route/screen count and navigation remain unclear. | Before category UI implementation. |
| G-09 | What key combination enables duplicate registration checks per form (NIM, email, or both), and how should blank/normalized values compare? | Duplicate submission behavior is stakeholder-dependent. | Before registration submit/duplicate logic. |
| G-10 | Can fields change after first submission? If yes, are old submissions interpreted against a saved form/field snapshot or only the active definition? | Historical submission rendering/export may change or lose meaning. | Before form edit and submission persistence implementation. |
| G-11 | What happens to stored bytes and metadata on document delete (soft-delete, retention window, permanent removal)? | Data retention and restore/security behavior are undefined. | Before document deletion implementation. |
| G-12 | For cross-division IDs, should API return 403 or 404 consistently? | Error behavior may disclose resource existence; contract currently allows either. | Before finalizing authorization/error tests. |
| G-13 | Define refresh token rotation, revocation, and reuse detection; confirm local versus production Secure/SameSite behavior. | Session theft/logout guarantees and browser integration differ. | Before auth implementation. |
| G-14 | Define rate-limit windows/limits and response headers for public aspiration and registration writes. | Security requirement is testable only after thresholds are chosen. | Before public write implementation/release. |
| G-15 | Define activity-log retention and whether failed refresh attempts are logged as a named action/event. | Phase 9 applies recursive sensitive-key redaction and omits IP/User-Agent from responses; retention and failed-refresh vocabulary remain unspecified. | Retention decision required before production operations; event vocabulary before failed-refresh audit requirement is claimed. |
| G-16 | How should remaining 29/37 references adapt to mobile where no separate mobile screen exists? Validate against responsive acceptance at 360/390/412 widths. | Mobile layout cannot be visually inferred from nonexistent captures. | Before visual sign-off for each route. |
| G-17 | Confirm whether public archive lists expired/inactive announcements and exact visibility rules; summary says archive/list but lacks detailed publication policy. | Public list may expose or hide the wrong announcements. | Before announcement API/UI implementation. |
| G-18 | Define upload constraints/configuration overrides and whether browser preview is expected for small media versus document archive; PRD gives 10 MB baseline and file_path is internal. | Resolved for the explicit Phase 11B public image scope: shared picker/preview, JPEG/PNG, configurable 5 MiB default, 12 MP bound, public API media route; private documents remain separate. GIF/WebP are unsupported. | Resolved for Phase 11B; production storage/backup target is a deployment follow-up. |

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

## Phase 11B implementation gaps

- Admin mobile screenshots from the user's 360×800 inspection were rejected. The root shell layout is corrected and touched screens are marked `RE-REVIEW REQUIRED`; device re-review at 360×800, 390×844, and 412×915 is still mandatory.
- Public media supports JPEG/PNG only; WebP was intentionally not accepted without a decoder dependency. Revisit only if an approved product requirement needs it.
- URL columns are retained for compatibility. Asset use is recorded at upload but not relationally bound to content; upload or consumer-save failure can leave an orphan asset. Such assets are immutable and retained for manual cleanup so a referenced public URL is never removed.
- Local public media path must be provisioned/backed up with the database in deployment. Storage interface is replaceable, but S3-compatible implementation and production deployment are outside Phase 11B.

## Phase 5 disposition

| Contract/evidence item | Phase 5 outcome |
|---|---|
| Period/division/member endpoint route permissions absent from the earlier summary | Reconciled from PRD v1.4: SUPER_ADMIN alone writes periods; SUPER_ADMIN/BPH manage divisions and members; KEPALA_BIDANG reads only and is division-scoped. `API_CONTRACT.md` records the implemented route policy. |
| Active period and public structure payload shape | Fixed for implementation as minimal public DTOs: public active period returns 404 when none exists; public organization returns `period: null` and empty hierarchy; public division list returns an empty list. No NIM, internal status, or audit metadata is exposed. |
| Stitch member-form extras versus Phase 5 schema | The direct phase fields win: photo URL and NIM are optional; no uploads or extra study-program/year/contact/social fields were added. Stitch `screen.png` remains the visual target, and `code.html` is structural/style evidence only. |
| Division deletion and member deletion | Division deletion conflicts when referenced, with no cascade. Member deletion is soft INACTIVE and preserves history. Periods are archived, never deleted. |
| Exact-size mobile visual evidence | Phase 6 carry-forward review used the available viewport override: aspiration public/admin list/detail were reviewed at 360/390/412; Phase 5 `/admin/divisi` was reviewed at 360/390/412, and `/admin/pengurus` plus `/admin/struktur` at 390. BPH/KEPALA_BIDANG organization UI smoke was also performed. The Phase 5 implementation report remains unchanged; a complete screenshot matrix for every Phase 5 page/viewport remains release-wide follow-up. |

No new Phase 5 product blocker was found for implementation. The exact-size mobile visual review limitation is carried as verification risk to the full visual/UAT phase; do not claim mobile pixel parity from the available browser session.

## Phase 6 disposition

| Contract/evidence item | Phase 6 outcome |
|---|---|
| PRD target division/category vs. Stitch category label | Resolved to active-period divisions as target choices. No category taxonomy endpoint/model exists; no independent taxonomy was invented. |
| PRD optional name/NIM vs. mock form fields | Name/NIM remain optional. Optional email is included and validated if present; anonymous writes strip all three identity fields server-side and at the schema constraint. |
| Detail mock uploads, disposition/SLA/notes/export | Carried out of Phase 6 because PRD/API do not support those controls. No file, notification, export, comment, or SLA workflow was added. |
| Status transitions | Fixed to sequential PENDING → READ → RESOLVED. Direct PENDING → RESOLVED, same-state writes, and reopen are rejected; PRD defines no reopen semantics. |
| In-memory public limiter | Provides per-process abuse protection; deployment with multiple API instances will need shared/distributed rate-limit storage before scaling. |
| Exact 360/390/412 visual evidence | Carried forward from Phase 5; browser adapter viewport limits mean responsive source/test evidence cannot claim pixel parity. |
| MariaDB runtime unavailable during initial Phase 6 attempt | Resolved on 2026-10-10 using the repository's existing Compose configuration and dedicated `bem_fmikom_test`; final migration round trip, non-skipping integration suite, and Phase 2/4/5 regression assertions passed. |

## Phase 7 disposition

| Earlier gap / evidence | Phase 7 resolution or carry-forward |
|---|---|
| G-09 duplicate identity policy | Resolved by the direct Phase 7 decision: required NIM, normalized trim/uppercase, database unique `(registration_form_id, normalized_nim)`, email is not a unique key. Same NIM is allowed on another form. |
| G-10 field mutation/history | Resolved without schema versioning: only DRAFT/no-submission schemas can change key/type/required/options/order; snapshots preserve each answer's field key/label/type/order. Post-submission field labels/placeholders may be corrected as presentation metadata; answer history keeps its original label. Core identity values are stored once. |
| G-14 public registration abuse limits | Resolved with a registration-only 5 requests/IP/30-minute configurable in-memory limiter. Distributed limiter remains a scale-out concern. |
| PRD public list only says OPEN; Phase 7 asks UPCOMING/OPEN/CLOSED | Reconciled explicitly in `API_CONTRACT.md`: the detailed direct Phase 7 instruction adds all three safe public effective states; DRAFT remains private. |
| PRD says duplicate NIM/email may be enabled by form; Phase 7 requires NIM once per form | The direct Phase 7 decision is stricter and overrides that baseline for this implementation; NIM is mandatory and email uniqueness is disabled. |
| Stitch form builder shows file upload, helper text/validation extras, and publish controls | No applicant file upload, arbitrary helper-text schema, quota/scoring, email/WhatsApp, or approval workflow is implemented. Only the eight fixed field types and the required schema controls are exposed. |
| Mobile-only registration administration mock; no separate Stitch detail-submission screen | `/admin/pendaftaran` desktop uses existing dense admin table patterns and mobile uses cards. Submission detail uses the design system and data-pendaftar visual language; no screen-specific pixel-parity claim is possible without a dedicated Stitch capture. |
| Public screenshots include contact/WhatsApp and file-upload examples | Those are visual mock elements, not authorization for notification/file workflows. Build matching hierarchy and responsive layout while keeping unsupported controls absent. |

Phase 8 implementation is tracked in `PHASE_8_REPORT.md`. Phase 9 scope and evidence are tracked in `PHASE_9_REPORT.md`. G-06 visitor/pageview analytics are deferred post-MVP; the operational dashboard does not collect or expose visitor data.

## Phase 8 disposition

| Contract/evidence item | Phase 8 outcome |
|---|---|
| PRD optional `division_id` vs. direct Phase 8 required division scope | Narrowed to required division ownership with a `(division_id, period_id)` foreign key. This is documented as a Phase 8-specific decision; a future product request for organization-wide documents would require a nullable ownership model and UI. |
| PRD/Security 10 MB vs. conditional 25 MiB fallback | 10 MiB configurable maximum is used because both source contracts already define 10 MB. |
| Delete atomicity across SQL and local filesystem | No distributed transaction exists. `delete_pending` and idempotent physical deletion support retry/repair; production restore must treat the DB and private file directory as one backup set. |
| Antivirus/content disarm | Not provided by this phase or existing contracts. Extension/MIME/signature/container checks reject malformed and trivially spoofed formats, but do not assert malware safety. Add approved scanning before production if organizational policy requires it. |
| Edit/delete-specific Stitch screens | No separate archive edit/delete references were found. Existing archive screens remain targets; those actions use reusable accessible dialogs and remain subject to manual visual sign-off. |
| Full exact viewport visual sign-off | Browser/viewport evidence and limitations are recorded in `PHASE_8_REPORT.md`; do not claim pixel parity for unreviewed widths/screens. |

Phase 7 implementation, database/security/export gates, browser workflows, and representative 390×844 public responsive review are complete. Final side-by-side visual sign-off for every relevant Stitch screen and the 360/412 breakpoints remains a manual release review item; see `PHASE_7_REPORT.md`.


## Phase 11 audit status — 2026-10-11

- The production build exposes the indexed public/admin route set and the 37 screen references (37 paired HTML references; 8 explicit mobile references) are inventoried in `PHASE_11_VISUAL_MATRIX.md`.
- In-app browser navigation to `http://localhost:3000/admin/login` was explicitly denied by browser URL policy. No alternate browser or low-level workaround was attempted. Implementation screenshot comparison, 360/390/412 measurements, keyboard/console/network checks, and Phase 11 public/cross-role UAT remain **BLOCKED**.
- Phase 7/8 manual follow-ups and Phase 10 browser CSP-console follow-up are not closed. Phase 9's visitor analytics deferral remains authoritative for current implementation; no analytics were added. G-17 and G-15 remain as previously described.
- Automated evidence on 2026-10-11: `pnpm api:integration`, `pnpm verify`, `pnpm audit`, pinned `govulncheck`, and the CI gosec severity/confidence gate passed. Static API comparison found 94 registered routes, 93 product-contract path matches, and only operational `GET /healthz` outside product scope. See `PHASE_11_REPORT.md` and `PHASE_11_API_MATRIX.md`; no Phase 11 product-code change or Phase 12 work was made.
- G-03 remains resolved as recorded for Phase 9: the implemented soft-deactivation lifecycle uses explicit POST actions rather than the legacy PRD DELETE method. G-04 detail/list operations are documented in the API contract; Phase 11 found several supported detail GETs have no direct UI caller, as listed in `PHASE_11_API_MATRIX.md`.

## Phase 11H — Resolved defects and remaining acceptance gaps (2026-10-13)

| Item | Status | Evidence / remaining work |
|---|---|---|
| Uploaded content image crop/zoom | Implementation corrected | Shared bounded contain frames and avatar exception are automated/build verified; user must review varied aspect ratios and modal layouts at target viewports. |
| Landing organization “Lihat struktur” | Implementation corrected | Accessible in-place expansion uses the existing organization payload; user must review hierarchy, keyboard behavior, and mobile overflow. |
| Registration IMAGE applicant data | Implemented | Private attachment domain/storage and authorized route are integration verified; user must review builder/form/admin detail and submit a real UAT image. |
| OPEN form absent from landing | Root cause fixed | The landing previously did not fetch the public registrations endpoint. It now uses a no-store API fetch; user must confirm open/close actions update the landing and CTA without restart/rebuild. |
| Registration time display | Policy fixed | API persists/evaluates in UTC and UI localizes to Asia/Jakarta; boundary behavior is covered by backend integration. Do not manually offset timestamps. |
| Phase 11H manual visual acceptance | Pending | Every affected visual row remains RE-REVIEW REQUIRED. This is the remaining release acceptance work; it is not an automated implementation blocker. |

Phase 11H identified no new product-contract question. Existing gap G-17 and prior Phase 11/Phase 7/8 manual UAT items remain open as documented above.

## Phase 11 finalization disposition — 2026-10-14

- The owner manually accepted the current public `/struktur` design, explicitly at 1440×900 and generally for the current layout. This closes the `/struktur` owner visual-acceptance gap; it does not create separate manual viewport measurements for 1280×800 or mobile sizes.
- Automated responsive observations for `/struktur` at 360×800, 390×800, and 412×800 remain separate evidence. The broader G-16 requirement to validate every responsive route/reference is not globally resolved by acceptance of this route.
- Remaining unreviewed screens, exact viewport captures, and authenticated role workflows stay as documented follow-ups with their existing statuses. No screen outside the owner's explicit acceptance is marked PASS by this update. G-15 and G-17 remain open governance/product-policy questions and are not critical Phase 11 runtime/security blockers.

## Phase 11 organization hierarchy corrective — 2026-10-13

- Organization/aspiration classification is resolved through `divisions.show_in_public_structure`. This flag only affects `GET /public/organization`; the active public divisions and aspiration write remain status/period scoped and continue to include an ACTIVE Fakultas.
- Existing canonical `Fakultas` and `BPH` rows are classified false once by migration `00011_public_structure_visibility.sql`; runtime presentation uses metadata, not names. Future records use the explicit flag exposed in division administration.
- `/struktur` hierarchy, dialog contents, and 360/390/412/desktop visual acceptance remain pending manual user re-review. This is not overall Phase 11 acceptance.
