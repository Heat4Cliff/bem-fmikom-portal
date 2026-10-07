# Phase 0 — Source Audit

## Scope and method

Audit date: 2026-10-07. This report records a read-only audit of the preparation package before application work. A separate repository was created at bem-fmikom-portal; the preparation package was copied into it as a baseline so the project has its own source evidence. The original BEM_FMIKOM_Codex_Prep/ tree was not edited, renamed, or cleaned up. The copied source_docs/ and design_reference/ directories are evidence and must be treated as immutable.

No application was scaffolded, no feature or migration was implemented, and no Phase 1+ checklist item was changed. No application environment probe was needed for this source-only phase.

## Required source reading order

All twelve requested documents were read in this order:

1. AGENTS.md — binding rules, stack, architecture, visual and security constraints.
2. MANIFEST.md — source inventory and authority pointers.
3. docs/PRD_SUMMARY.md — v1.4 scope, roles, core business rules.
4. docs/ARCHITECTURE.md — monorepo, frontend/backend layering and topology.
5. docs/backend/API_CONTRACT.md — baseline endpoints, responses and filters.
6. docs/backend/SECURITY_CONTRACT.md — authentication, RBAC, scope, file, CORS and test requirements.
7. docs/design/DESIGN_SYSTEM.md — token and component guidance.
8. docs/design/SCREEN_INDEX.md — 37 screen-to-route entries.
9. docs/design/ROUTE_SCREEN_MATRIX.md — required public/admin routes and menu policy.
10. PLANS.md — phase boundaries and checklists (left unchanged).
11. docs/codex/IMPLEMENTATION_PROTOCOL.md — verification and phase exit policy.
12. docs/codex/VISUAL_REVIEW_CHECKLIST.md — visual review criteria.

Supplemental evidence read for the audit:

- source_docs/PRD_BEM_FMIKOM_UNUGHA_v1.4_Revisi_API_Contract.docx (261 paragraphs, 38 tables) was extracted and reviewed, including its role matrix, full API endpoint tables, stories, data model, and acceptance criteria.
- The paired v1.4 PDF is present (23 pages; text extraction returned 38,986 characters). The DOCX was used for structured table comparison. source_docs/Product_Requirements_Document_1.pdf is historical v1.3 and is not authoritative.
- Workflow notes: docs/workflows/AUTH_RBAC.md, REGISTRATION.md, ASPIRATION.md, DOCUMENTS.md, PERIOD_AND_ORGANIZATION.md; also docs/backend/DATA_MODEL_NOTES.md and docs/acceptance/ACCEPTANCE_CRITERIA.md.
- Stitch package and design token source: design_reference/stitch_bem_fmikom_unugha_portal/modern_academic_horizon/DESIGN.md.

## Source authority and conflicts

The source-of-truth order supplied for this run is applied as follows: PRD v1.4 controls behavior and scope; API/Security Contract controls request/response and authorization except where the more specific original v1.4 contract contradicts its summary baseline (the baseline itself says v1.4 wins); screen.png is the visual target; code.html is only layout/style evidence; DESIGN.md is the design-token reference. Stitch HTML is not production architecture.

### PRD v1.4 versus API Contract

The API baseline is broadly aligned to modules, but it is not an exact transcription of the full v1.4 endpoint tables:

- Path mismatch: PRD v1.4 defines GET /public/gallery-categories/{slug}; docs/backend/API_CONTRACT.md uses {id}. The response should be category plus its gallery items. Resolve the path before client typing/routing.
- Missing PRD route in baseline: PRD defines GET /admin/article-categories; the baseline section omits it although category listing is an admin UI need.
- User deactivation mismatch: PRD defines DELETE /admin/users/{id} as deactivate/soft-delete. The baseline instead has POST .../deactivate, POST .../reactivate, and POST .../reset-password, and does not list the PRD DELETE operation. Reactivate/reset are not specified by PRD v1.4 and need explicit scope approval or removal.
- Baseline additions absent from PRD endpoint tables: detail GETs for admin users, periods, divisions, and members; GET /admin/gallery-items; and some action routes are useful for the mapped edit/detail screens, but are additions rather than routes listed by the full PRD contract. Confirm and reconcile them in the API source before implementation.
- Public active period: the baseline adds GET /public/periods/active; PRD requires the public site to show the active-period organization, while the PRD endpoint table names GET /public/organization but not this standalone path. Decide if this is an explicit additive endpoint or if active-period metadata is part of organization/settings.
- Statistics write/collection path: PRD requires simple page-view/visit statistics and permits backend aggregation or an approved analytics service. No collection endpoint or provider contract appears in the baseline. GET /admin/dashboard is the consumer, but source/retention/deduplication behavior is not fixed.

The source precedence resolves behavior in favor of v1.4, but it does not authorize undocumented additions. Reconcile the canonical endpoint list and payload/error schemas before feature implementation.

### API Contract versus UI

- The UI matrix includes /admin/kategori-berita, while no corresponding route appears in SCREEN_INDEX.md; the API summary also lacks GET /admin/article-categories.
- The PRD requires an admin login, and auth endpoints are listed, but none of the 37 Stitch references is a login screen and the route matrix does not name /admin/login. A functional login route is still required; its visual target is missing.
- The user management screen may expose deactivation/reactivation/reset actions; API semantics differ as described above. Keep only actions with approved endpoint behavior.
- Period, division, member, and gallery-item detail/edit experiences require read operations. Some reads occur only in the summary baseline and not in the original PRD endpoint tables. Confirm as additive contract routes or design edit flows around list payloads.
- Dashboard visit statistics have a visible consumer but no finalized source/collection route. Decide backend aggregation versus approved provider before dashboard metrics are implemented.
- Core consumer mapping otherwise exists for public settings/content, organization, submissions, documents, settings, and audit log. See IMPLEMENTATION_MATRIX.md for detailed mapping and unowned actions/endpoints.

## Stitch screen and responsive coverage

- Found exactly 37 screen.png files and 37 paired code.html files under the Stitch reference tree; the counts match SCREEN_INDEX.md.
- Exactly 8 references are explicitly mobile-specific: public home and public registration form (2); admin dashboard, articles, registrations, aspirations, documents, and activity log (6). The other 29 references do not have a separately named mobile screen.
- The screen index maps public landing, news list/detail, announcements list/detail, gallery list/detail, registration list/detail/form, aspiration, plus admin overview, content, organization, registration, aspiration, document, settings, users and activity-log experiences.
- Public active-period structure is represented as a requirement/consumer of the home organization section, not a standalone indexed public structure screen.
- The screen index has no dedicated login screen or article-category screen. Dedicated Stitch references also do not exist for every responsive admin route/state. Responsive behavior remains required by PRD/design docs; it must be implemented from the reviewed visual target and responsive rules, with gaps tracked rather than claiming one-to-one mobile evidence.
- screen.png is the visual target. code.html is not a component source and must not be copied into production. No rendered product screens exist in Phase 0; no visual review is claimed.

## Cross-cutting requirement audit

- Roles: PRD provides a module/role matrix. SUPER_ADMIN has full/exclusive user, period activation and activity-log powers. BPH has cross-division content/operations and landing settings, but not super-admin account powers. KEPALA_BIDANG requires division_id and is restricted to own-division content, gallery, registrations/submissions/exports, documents and routed aspirations; announcement/period/division reads are limited as described by the role matrix. Summary endpoint labels such as Scoped are not a sufficient permission specification on their own.
- Division scope: enforce in authenticated backend policy/service/repository queries; never trust client division_id. Route and query manipulation must not widen scope. Test read and write paths, exports and downloads.
- Dynamic registration: builder, versioned field definitions, public rendering/submission, validation against active field definitions, optional duplicate NIM/email policy, scoped submissions and XLSX export are in scope. Advanced branching/scoring and applicant file uploads are out. Duplicate identity combination, field edit/lock behavior after submissions, and definition snapshot semantics need decisions.
- Private documents: use non-public storage, allow PDF/DOC/DOCX/XLS/XLSX, baseline 10 MB configurable, MIME+extension validation, safe server filenames, scope checks and authorized stream downloads. No public direct URL. Retention behavior on delete and policy for cross-scope 403 versus 404 are unresolved.
- Activity log: required minimum actions include login/logout, CRUD, publish/unpublish, activate/deactivate, upload/download/export; only SUPER_ADMIN reads logs. Logs must not contain secrets, document bytes, or unnecessary private submission data. Refresh failure logging is mentioned by PRD authentication behavior but not normalized into the minimum action vocabulary. Redaction and retention rules remain open.
- Single active period: exact-one-active is a hard invariant; activation must preserve history and be atomic. MySQL implementation requires transaction/locking or another verified integrity strategy; concurrency behavior needs an integration test.
- Responsive: required for public and admin surfaces; mobile tables become structured cards and admin nav becomes off-canvas at tablet/mobile per design rules. Only 8/37 screens are explicitly mobile references, so responsive acceptance cannot be inferred from desktop screenshots alone.
- Security: short-lived access JWT, refresh cookie, active-account check, explicit CORS allowlist, public write validation/rate limits, rich-text sanitization, private storage, authorization at data access, secret-free repo, secure error handling, and role/IDOR/file/rate-limit tests are in scope. Cookie rotation/reuse behavior, concrete rate thresholds, 403/404 disclosure policy, and reset-password lifecycle need decisions.
- Unmapped items: no dedicated login visual; article-category UI route/reference mismatch; statistics ingestion/provider; required public/mobile responsive states beyond explicit mobile captures; and endpoint additions/mismatches listed above.

## Phase 0 verification record

- STATIC VERIFICATION: source inventory confirmed 37 screen.png + 37 code.html; source v1.4 DOCX/PDF availability confirmed; endpoint and role matrices compared.
- AUTOMATED TEST: not applicable; no application code was created.
- INTEGRATION TEST: not run; no API or database exists in this phase.
- VISUAL REVIEW: not run; audit inventories screen evidence but does not certify a future UI against it.
- MANUAL UAT: not run.

## End-of-phase summary

1. Source read: all 12 required documents in order, PRD v1.4 DOCX and paired PDF, relevant workflow/model/acceptance references, and Stitch DESIGN.md.
2. Stitch screen count: 37 images, 37 paired HTML references.
3. Fixed technical decisions: Next.js App Router + strict TypeScript + Tailwind; Go + Gin REST; MySQL/MariaDB; JWT access + Secure HttpOnly refresh cookie; modular reusable UI; monorepo boundaries from the architecture note.
4. Unfixed decisions: canonical/additive routes, dashboard statistics source, login visual, category screen strategy, duplicate submission key, form edit/snapshot behavior, document deletion retention, cookie lifecycle details, rate limits, audit retention/redaction, and 403/404 disclosure.
5. Gaps/blockers: API summary needs reconciliation with the full v1.4 endpoint table; role labels need route-level matrices; 29 screens lack dedicated mobile references; login/category visuals are missing. These are documented, not silently guessed. No missing-source blocker prevents completion of the audit.
6. Implementation risks: contract drift between frontend/backend; division IDOR/export/download leaks; destructive period/document behavior; dynamic form schema drift; incomplete mobile fidelity; private data leaking through logs; user-management actions not backed by an approved contract.
7. Phase 1 recommendation: when instructed, start with repository foundation only, keep reference assets immutable, and record the API reconciliation and missing visual references as gates before implementing their affected features. Preserve PLANS.md Phase 1+ checklist as copied.

PHASE 0: PASS
