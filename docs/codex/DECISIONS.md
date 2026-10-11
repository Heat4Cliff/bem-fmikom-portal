# Phase 0 — Decisions

## Fixed by user or authoritative source

| Area | Decision | Authority |
|---|---|---|
| Repository | Separate project directory named bem-fmikom-portal; keep preparation source separate. | User request |
| Source evidence | Treat source_docs/ and design_reference/ as read-only; copied unchanged into the new repo for a self-contained project. | User request |
| Scope order | PRD v1.4 controls behavior and scope; API/Security Contract controls wire/auth behavior subject to v1.4 reconciliation. | User request and v1.4 contract rule |
| Visual authority | Stitch screen.png is the target; code.html only informs layout/style; Stitch HTML is not production architecture. | User request and AGENTS.md |
| Design tokens | Use Stitch modern_academic_horizon/DESIGN.md as the token reference; screen images win visual conflicts. | User request and source design note |
| Frontend | Next.js App Router, TypeScript strict, Tailwind CSS, reusable component implementation. | User request / AGENTS.md |
| Backend | Go, Gin, REST API. | User request / AGENTS.md |
| Database | MySQL or MariaDB with versioned migrations in later phase. | User request / AGENTS.md |
| Authentication | JWT access token plus Secure HttpOnly refresh cookie on HTTPS. | User request / PRD v1.4 |
| Authorization | Backend is final authority; role and division scope must be enforced server-side. | AGENTS.md, Security Contract, PRD v1.4 |
| Period invariant | Exactly one active period; transitions preserve history and are atomic. | PRD v1.4 |
| Documents | Private/non-public storage and authorized streaming; no public direct file path. | PRD v1.4 / Security Contract |
| Phase boundary | Phase 0 only; no app scaffold, feature, migration, or Phase 1 checklist changes in this run. | User request |

## Audit resolutions (not product scope changes)

- Where summary endpoint inventory differs from the full v1.4 endpoint table, the full v1.4 contract is the authority; additions still need explicit reconciliation before implementation.
- The 37 Stitch files are evidence, not a requirement to reproduce raw HTML. Implement screens as reusable Next.js/TypeScript/Tailwind components in a later authorized phase.
- Explicitly mobile-named references are evidence for 8 screens. Other routes remain responsive requirements; do not imply that missing mobile captures exist.

## Not decided here

See GAPS_AND_QUESTIONS.md. Phase 0 does not choose new endpoint behavior, role access, analytics provider, retention policy, rate limit thresholds, or unprovided visual design.

## Phase 1 — Foundation choices

| Area | Decision | Reason / boundary |
|---|---|---|
| Workspace | pnpm 11.25.0 workspace with `apps/*` and `packages/*`; exact direct dependency versions are recorded in manifests and `pnpm-lock.yaml`. | Frozen lockfile install passed. A later offline supply-chain metadata check lacked a cached `@babel/helpers` record, so the final install check used the configured registry. |
| Frontend | Next.js 16.4.0 App Router, React 19.2.8, TypeScript 6.0.3 strict, Tailwind CSS 4.3.3, Biome 2.5.15 for lint/format. | ESLint 9 is out of support and ESLint 10 is outside the peer ranges of the Next React/import/a11y plugins. Biome keeps lint and format on a supported, peer-clean toolchain. TypeScript 7 was rejected after typescript-eslint reported its compiler API was unsupported. |
| Fonts | Inter and Plus Jakarta Sans are bundled from Fontsource 5.3.0 and loaded with `next/font/local`. | Keeps the specified typefaces while allowing offline builds without fetching Google Fonts. |
| Build bundler | Production build script uses `next build --webpack`; `next dev` keeps the Next.js default. | Turbopack's Windows worker failed to spawn under the restricted local build environment; Webpack production build passed. |
| Backend | Go 1.26.9 module with Gin 1.12.0, `slog` JSON logging, request ID, recovery, common JSON envelope, JSON 404, and operational `GET /healthz`. | Phase 10 govulncheck found reachable standard-library advisories on 1.26.1; pin the patched 1.26.9 toolchain. |
| Runtime pin | `.node-version` is 24.19.0 and `.go-version` is 1.26.9. | CI installs the declared versions. This workstation executed Node commands on Node 25.2.1; record the mismatch when interpreting local verification. |

These choices do not resolve Phase 0 product/API/security gaps or introduce authentication, authorization, database schema, migrations, or business endpoints.

## Phase 2 — Database, authentication, and authorization foundation

| Area | Decision | Reason / boundary |
|---|---|---|
| Phase snapshot | Phase 1 was committed as `6bdf2ba chore: complete phase 1 foundation` before Phase 2 changes. `source_docs/` and `design_reference/` remain unchanged from baseline. | User's Phase 2 handoff required a distinct verified snapshot. |
| Database driver/config | `database/sql` with `github.com/go-sql-driver/mysql` v1.10.1; retain Phase 1 `DATABASE_URL` MySQL DSN and configure UTC session time, parseTime, pool limits, and connection lifetime. | Explicit SQL avoids ORM behavior and works with MySQL/MariaDB. |
| Migrations | Goose v3.28.0, embedded SQL migration with UP/DOWN, and a unique generated `periods.active_slot` value so MySQL/MariaDB can enforce at most one ACTIVE period. | Deterministic, versioned migrations; one database constraint protects the later single-active-period workflow. The destructive DOWN is for local/test rollback only. |
| Phase 2 schema scope | Create only `periods`, `divisions`, `users`, `refresh_sessions`, and `activity_logs`. Do not create Phase 3+ product-domain tables. | The explicit Phase 2 handoff narrows the generic broader Phase 2 migration list in `PLANS.md`; product tables remain with their feature phases. |
| Passwords and access JWT | bcrypt cost 12 by default, admin password policy 12 characters minimum and 72 UTF-8 bytes maximum; HS256 access JWT defaults to 15 minutes, `iss/sub/iat/exp/jti`, and a minimum 32-byte environment-only secret. Production rejects known placeholder/low-diversity secrets. | Implements the fixed Phase 2 security requirements. JWT contains no authoritative role/division state. |
| Refresh sessions | 32-byte cryptographic opaque token; store only lowercase SHA-256 hex hash; default seven-day expiry; rotate transactionally while preserving `family_id`; revoked-token replay revokes the active family and emits a structured security event without token values. | Satisfies rotation/reuse behavior and makes stolen/replayed values non-reversible in the database. |
| Authorization | Load the active user and current role/division from MySQL for every protected request. `SUPER_ADMIN` and `BPH` have no division assignment; `KEPALA_BIDANG` must have exactly one. Role middleware and reusable division helper are separate. | Database state is authoritative and deactivation/role changes take effect without waiting for JWT expiry. |
| Auth API paths | Register `POST /api/v1/auth/login`, `/refresh`, `/logout`, `GET /api/v1/auth/me`; retain `/healthz`. The API contract's `/auth/...` entries are interpreted under its `/api/v1` prefix. | The routes reconcile with both the explicit Phase 2 contract and PRD v1.4; no unrelated product endpoints were added. |
| Access lifetime discrepancy | PRD v1.4's login example shows `expires_in: 1800`; the explicit Phase 2 handoff fixes the default at 15 minutes/900 seconds. Implement 900 and record the conflict for contract maintenance. | Newer, direct phase instruction resolves this otherwise material response example mismatch. |
| Login visual | Add a simple `/admin/login` screen using current navy/amber tokens, fonts, and form primitives. Status: `DESIGN SYSTEM CONSISTENT`; no dedicated Stitch login reference and no pixel-parity claim. | G-07 remains an evidence limitation rather than a reason to invent a new visual language. |
| Cookie/CORS | Cookie defaults: development Secure=false, production Secure=true, SameSite=Lax, HttpOnly always, path `/api/v1/auth`; `SameSite=None` requires Secure. Use comma-separated exact `FRONTEND_ORIGINS`, credentialed allow-list CORS, and require a trusted Origin for auth POSTs. | Covers refresh and logout while limiting the cookie to the auth routes. |
| Login rate limit | Configurable, per-IP fixed-window in-memory limiter; Gin proxy trust is disabled by default to prevent client-supplied forwarded headers from bypassing the limit. | Simple single-instance protection without Redis; distributed enforcement stays future hardening. |
| Admin bootstrap | `pnpm api:create-admin` invokes an interactive Go CLI; password entry is hidden, confirmed, validated, normalized/hashed, and creates only an ACTIVE `SUPER_ADMIN` with no division. | Avoids hardcoded credentials and avoids passing a password in command history. |
| Activity events | Store LOGIN and known-identity LOGOUT as product activity rows. Refresh failures/reuse use structured security logs only; no password, access token, refresh token, or JWT secret is logged. | Matches the Phase 2 activity schema scope and explicit exclusion of failed login as a product log; the PRD's broader audit vocabulary remains relevant in later phases. |
| Integration environment | Local database is MariaDB 11.4 via Compose. Integration suite requires a dedicated database name ending in `_test`, resets it through DOWN/UP, and cleans it after the run. GitHub Actions configures a MariaDB 11.4 service. | Prevents accidental destructive tests against development/production. The initial Phase 2 check was blocked; the later live rerun passed as recorded in `PHASE_2_REPORT.md`. |

### Phase 3 — Public website foundation and landing page

| Area | Decision | Reason / boundary |
|---|---|---|
| Public home model | Present sections consume a typed `LandingPageContent` view model, not future API response types. | Content APIs are not implemented in this phase; later API adapters can map into this model without coupling UI to wire contracts. |
| Dynamic previews | Organization, news, announcements, gallery, registration, vision/mission, and statistics use empty production defaults. Visual fixtures are enabled only in development or by server-only `BEM_DEMO_CONTENT=true`; the page marks demo mode visibly. | Avoids fake APIs, official-looking sample content, and demo statistics in production. |
| Public navigation | Landing sections use stable fragment anchors. `/aspirasi` is a no-form placeholder linked from the public CTA; it is not an aspiration feature. | Prevents Phase 3 links from navigating to unimplemented public routes while keeping the planned route usable as a clear placeholder. |
| Images and identity | Use palette-based CSS artwork placeholders and the existing text BEM mark. Do not reuse screenshots, hotlink Stitch image URLs, or invent a logo. | The source package contains no approved production logo/photo asset or image license record. Replace placeholders when approved assets are supplied. |
| Canonical origin | Canonical/sitemap absolute URLs are emitted only after `NEXT_PUBLIC_SITE_URL` is configured; sitemap lists only `/`. | No approved public domain was found in the source package and unimplemented routes must not be advertised. |
| Color tokens | Align the accent foreground/background mapping to Stitch `DESIGN.md` (`secondary #904d00`, `on-secondary #ffffff`, secondary container `#fe932c`). | The prior app token mapped white text to the lighter `#D97706`; matching the audited YAML improves contrast and keeps the public palette token-based. |

### Phase 2 source reconciliation notes

- PRD v1.4 confirms roles and high-level access: SUPER_ADMIN full; BPH cross-division but not user/activity-log management; KEPALA_BIDANG scoped, with periods/divisions/members read-only and announcements read-only. Phase 2's more specific period-mutation rule is SUPER_ADMIN-only.
- The API contract lists auth routes relative to `/api/v1`; the implemented full paths match the Phase 2 endpoint list.
- The PRD's 1800-second login example conflicts with the user-fixed 900-second default; implementation follows 900 seconds.
- `PLANS.md` has a broad Phase 2 schema checklist naming future domain entities. The user-fixed Phase 2 scope limits migrations to the five foundation tables; no later feature tables were created and Phase 3+ checklist text was not changed.
- The Phase 0 gap G-13 is resolved for auth: access/refresh lifetimes, opaque refresh rotation/reuse, cookie behavior, current-DB identity, and auth route prefix are fixed. Missing visual login reference remains documented as G-07.

### Phase 4 — Content CMS

| Area | Decision | Reason / boundary |
|---|---|---|
| Schema scope | Add migration `00002_content_cms.sql` for article categories/articles/map, gallery categories/items, and announcements; keep foundation migration unchanged. | Phase 4 feature data is isolated in its own reversible migration. |
| Rich text | Sanitize article/announcement HTML server-side with an allowlist (`p`, `h2`, `h3`, `strong`, `em`, `ul`, `ol`, `li`, `blockquote`, `a`, `br`); strip active-content subtrees, event attributes, and unsafe URL schemes. | Reference HTML is not trusted production code or content. |
| Images and gallery | Use URL fields only; do not upload files. Google Photos album links and optional image/preview links require HTTPS. | Explicit Phase 4 handoff decision; no storage service or media permission was provided. |
| Article roles | SUPER_ADMIN/BPH operate across divisions. KEPALA_BIDANG division is always taken from the current database principal; cross-division access is denied. | Prevents request-body division spoofing and IDOR. |
| Category roles | Article categories are shared taxonomy: KEPALA_BIDANG reads only. Gallery album categories are managed by SUPER_ADMIN/BPH; KEPALA_BIDANG reads its division's albums and can CRUD items only within those parent categories. | Keeps taxonomy centrally managed while allowing division-scoped documentation work. |
| Announcement roles/status | SUPER_ADMIN/BPH write; KEPALA_BIDANG reads. Persist `is_active` and UTC schedule bounds; report status as INACTIVE/SCHEDULED/ACTIVE/EXPIRED. The active endpoint requires the current UTC time window. | No scheduled task is required to transition the derived status. Public archive visibility still needs product confirmation (G-17). |
| Public caching | Public CMS Server Components call the REST API with `cache: "no-store"`; admin requests use the existing typed authenticated client. | Avoids serving stale publication state during this phase. |
| Gallery detail key | Implement `/api/v1/public/gallery-categories/{slug}` and `/galeri/{slug}`. | Full PRD v1.4 route table takes precedence over the API summary's `{id}` placeholder; contract and route matrix are reconciled. |
| Division denial response | Mutations return 403 for unauthorized scope; scoped detail reads may return 404 to conceal record existence. | The security contract requires denial, and the existing G-12 permits 403 or 404. |
| Category screen | `/admin/kategori-berita` uses the audited content-management component language; no dedicated Stitch category reference exists. | Reuses the existing design system without claiming screen-specific pixel parity (G-08). |

Phase 4 integration evidence and known limitations are recorded in `PHASE_4_REPORT.md`.

### Phase 5 — Organization and periods

| Area | Decision | Reason / boundary |
|---|---|---|
| Schema | Reversible migration `00003_organization.sql` adds `divisions.display_order` and `organization_members`; foundation and content migrations stay unchanged. | Keeps Phase 5 data additive and permits a migration round trip. |
| Role matrix | SUPER_ADMIN alone creates/edits/activates/archives periods. SUPER_ADMIN and BPH manage divisions and members across divisions. KEPALA_BIDANG reads periods and only its current DB-assigned division/members; all organization writes are denied. | Follows PRD v1.4 roles; authorization is enforced by the API and is not inferred from hidden UI controls. |
| Period lifecycle | No period deletion. Activation is transactional, archives the prior ACTIVE period, preserves all history, and maintains exactly one ACTIVE period. | Period history is used by members and dependent records; the database constraint protects concurrent transitions. |
| Division lifecycle | Division deletion is rejected when dependent records exist; no cascade deletion. Division order is explicit and public listing uses a stable order. | Prevents orphaning historical members/users/content and keeps public hierarchy deterministic. |
| Member lifecycle | DELETE is soft deactivation to INACTIVE. NIM and photo URL are optional; no upload workflow. Head/member positions require a same-period division; other leadership positions may have none. | Matches v1.4 history/role intent and the direct Phase 5 field/schema instruction. |
| Active head | At most one ACTIVE KEPALA_BIDANG per division/period; validate in service and enforce with a MariaDB generated-column unique key. | Preserves correctness under concurrent requests. |
| Public organization | Public homepage consumes the minimal organization DTO for the current ACTIVE period. Inactive members and internal/audit fields are omitted. No active period produces `period: null` and an empty hierarchy; active-period metadata endpoint returns 404. | Prevents leaking internal records and gives clients deterministic no-data behavior. |
| Stitch reconciliation | Use the requested optional URL-backed photo and optional NIM; do not add upload, student program, academic year, contact/social, export, or print features solely because extra controls appear in Stitch references. | The explicit Phase 5 scope/schema is authoritative for behavior; `screen.png` guides visuals and `code.html` is not production architecture. |
| Mobile evidence | Admin routes use responsive cards/forms and source Stitch files were reviewed, but the current in-app browser adapter cannot set an exact 360/390/412 viewport. | Report this manual visual verification limitation explicitly; do not claim device-sized visual parity. |

Phase 5 integration and UAT evidence are recorded in `PHASE_5_REPORT.md`.

### Phase 6 — Aspiration management

| Area | Decision | Reason / boundary |
|---|---|---|
| Identity | Name and NIM remain optional as PRD 7.6 says. Email is also optional and is validated only when supplied, matching the public Stitch field and Phase 6 handoff. When anonymous, service and DB constraint force name/NIM/email to NULL. | Do not make identifying fields required based on mock content; enforce anonymity server-side. |
| Target/routing | The public target field uses existing active-period divisions as its values. `division_id` is required in the API and verified against an ACTIVE division joined to the ACTIVE period at submit time. | PRD requires routing to a division/category, but no category taxonomy API/model exists; use the supported division routing without inventing taxonomy. |
| Plain text and bounds | Subject 200, message 5,000, name 160, NIM 32, email 254 Unicode/codepoint characters; subject/message are trimmed and stored/rendered as plain text. | Bounded request/resource use and existing domain conventions; React escapes admin display. |
| Status workflow | Only sequential `PENDING → READ → RESOLVED`; no direct resolution from PENDING, no same-state update, no reopen. First `read_at` and `resolved_at` are set once in UTC. | This follows the PRD's explicitly ordered lifecycle and keeps RESOLVED terminal where no reopen rule exists. |
| Activity data | Only admin status transitions emit an `UPDATE`/`ASPIRATION` activity event with aspiration ID, old/new status, and division ID. Public submit identity/message is not copied into activity rows or structured request logs. | Minimize sensitive student data; activity read dashboard remains Phase 9. |
| Abuse / Origin | Dedicated configurable fixed-window limiter defaults to 5 writes per 15 minutes per client IP. Reject a supplied untrusted Origin on the public write without making Origin mandatory for non-browser clients. | Separate from login throttling while retaining public REST use and explicit origin protection. |
| Stitch-only controls | Do not implement attachment upload, separate category taxonomy, disposition notes, SLA metrics, export, or case-management workflow. | These controls are visible in reference mockups but not specified by Phase 6 PRD/API contract; uploads and export belong to later authorized scope. |

### Phase 7 — Dynamic registration management

| Area | Decision | Reason / boundary |
|---|---|---|
| Duplicate identity | Require the `nim` core field and enforce `UNIQUE(registration_form_id, normalized_nim)` in MariaDB. Trim and uppercase without numeric conversion; preserve the submitted display value. Email is not a duplicate key. | The direct Phase 7 instruction settles G-09 more specifically than PRD 7.7's stakeholder-selectable NIM/email baseline. |
| Core fields | Every form has exactly one required `full_name` (TEXT), `nim` (TEXT), `email` (EMAIL), and `phone` (TEXT) field. New forms receive these defaults; they cannot be removed or made optional. | The direct Phase 7 schema rule defines identity fields; NIM and phone remain strings to preserve leading zeroes and symbols. |
| Dynamic schema | Use the eight allow-listed field types only. Persist options in a normalized options table and answers as bounded text or JSON. Save complete schema changes in one transaction. | Avoids arbitrary types and partial builder writes. |
| Submission history/schema lock | The first submission permanently locks field key, type, required flag, options, and order. Only a DRAFT form with zero submissions can change those structural attributes. Labels/placeholders may be corrected as presentation metadata; answer rows retain label/type/key/order snapshots for historical display and export. Core identity values live only on the submission row; core answer snapshots do not duplicate those values. | Resolves G-10 without a versioning engine, preserves meaning/history, and minimizes duplicate PII. Create another form for a different schema. |
| Form lifecycle | Admin transitions are DRAFT → OPEN → CLOSED only; CLOSED cannot be reopened. Public state is derived in UTC as UPCOMING/OPEN/CLOSED from status and `open_at`/`close_at`, with both endpoints inclusive for OPEN. OPEN forms may be opened before `open_at` and appear as UPCOMING. | Implements the direct Phase 7 state definitions without adding reopen behavior. |
| Public registration visibility | Public listing returns safe non-DRAFT UPCOMING, OPEN, and CLOSED forms with deterministic priority OPEN, UPCOMING, then recently CLOSED. DRAFT is never exposed. | Direct Phase 7 instruction extends PRD 10.5, which only names OPEN registrations; this explicit request resolves the narrower listing behavior. |
| Ownership and division | `period_id` is required; `division_id` remains nullable as PRD 7.7 permits. If a KEPALA_BIDANG creates a form, the server forces the current database division and rejects a conflicting submitted division. Cross-division operations are permitted for SUPER_ADMIN/BPH only. | Follows PRD plus the direct role instructions; request filters never grant scope. |
| Delete | Soft-delete a form only when it has no submissions. A form with submissions returns 409 and retains all history. | Honors the PRD's “soft delete bila aman” and the direct no-submissions guard. |
| Public submission bounds | Limit JSON requests to 64 KiB, fields to 50 including the four core fields, TEXT to 500 characters, TEXTAREA to 5,000, and identity/contact strings to their documented bounds. Validate all values/options server-side. Store no raw client IP; `ip_hash` remains NULL because the limiter needs only transient IP state. | Bounded work and data minimization without applicant file uploads or unsupported field types. |
| Registration rate limit | Independent per-process fixed window: 5 submission attempts per client IP per 30 minutes by default; configurable via `PUBLIC_REGISTRATION_RATE_LIMIT` and `PUBLIC_REGISTRATION_RATE_WINDOW`. | Resolves G-14 for registrations without reusing login or aspiration thresholds; distributed deployments need a shared limiter. |
| XLSX export | Use Excelize v2.11.0 and explicit string cells for all applicant-controlled values. Cap exports at 1,000 submissions after filters; return a clear limit error above the cap. Export audit details contain only form ID, division ID, and row count. | Prevents formula execution and bounds memory/response size for sensitive exports. |
| Logging/privacy | Public registration payloads and answers are never copied into request logs or activity rows. Submission rows contain core identity once; answer snapshots contain values only for custom fields. Export is private/no-store and requires the same form scope as submissions. | Minimizes student PII exposure. |
| Activity | Emit CREATE/UPDATE/DELETE/ACTIVATE/DEACTIVATE for REGISTRATION_FORM as applicable; schema saves use minimized UPDATE metadata. Emit EXPORT with only form ID, division ID, and submission count. | Satisfies PRD audit requirements without logging schemas or applicant answers. |
| Stitch reconciliation | Build responsive public/admin screens from reusable Next/TypeScript/Tailwind components. Omit file upload, quota/scoring/selection, auto-save, email/WhatsApp, and applicant approval workflow controls shown in some mocks. | The direct Phase 7 scope and PRD v1.4 exclude these behaviors; `screen.png` remains the visual target and `code.html` is reference only. |

### Phase 8 — Private document archive

| Area | Decision | Reason / boundary |
|---|---|---|
| Archive route and contract mapping | Admin UI uses `/admin/arsip-dokumen`; retain `/api/v1/admin/documents` contract routes. | Reconciles the direct Phase 8 route with API names and Stitch archive screens. |
| Division ownership | Documents require a `division_id` belonging to the selected `period_id`; KEPALA_BIDANG is forced to the current database division. | PRD v1.4 allows optional division, while direct Phase 8 instructions specify period/division ownership and fixed KB division. This phase narrows the schema and behavior accordingly. |
| Upload maximum | Default is 10 MiB (`10 * 1024 * 1024`), configurable with `DOCUMENT_MAX_UPLOAD_BYTES`. | PRD v1.4 and SECURITY_CONTRACT explicitly set 10 MB; the conditional 25 MiB fallback does not apply. |
| Private storage | Local private filesystem implements a storage interface, uses generated keys, private staging, `0600` POSIX object permissions, and `os.Root` path confinement. Production requires an absolute `DOCUMENT_STORAGE_PATH` outside public/static/web roots. | No public/static file route exists; bytes are served only through the authenticated download endpoint. Windows inherits the storage directory ACL because Windows does not expose POSIX permission semantics. |
| File checks | Allow PDF/DOC/DOCX/XLS/XLSX; validate normalized extension and declared MIME, then inspect signatures or OOXML ZIP members; stream and hash SHA-256. | Client filename/MIME are untrusted. Format structure checks do not provide antivirus scanning or content disarm. |
| Delete repair | Mark `delete_pending`, remove the private object idempotently, then delete metadata and write activity in a final transaction. Failed object deletion attempts clear the pending state; final transaction failures leave a hidden retryable row. | This makes the filesystem/database boundary recoverable without claiming a distributed transaction. |
| Screens | Desktop uses a metadata table and upload dialog; narrow screens use stacked cards/actions. Existing archive desktop/mobile `screen.png` are visual targets; no separate edit/delete reference exists. | Use reusable Next.js, TypeScript, and Tailwind components. Stitch `code.html` stays layout/style evidence only. |

Phase 8 verification and limitations are recorded in `PHASE_8_REPORT.md`.

### Phase 9 — Landing settings, user management, and activity log

| Area | Decision | Reason / boundary |
|---|---|---|
| API compatibility | Retain `/api/v1/public/settings`, `/api/v1/admin/settings`, `/api/v1/admin/users...`, and `/api/v1/admin/activity-logs...` from the API contract. | Phase 9 admin route names map to those existing API resources. |
| Landing settings | One `landing_settings` row (id=1); plain-text content; HTTPS external URLs; same-origin CTA paths/fragments; public response cache is 60s with stale-while-revalidate 300s; admin response is no-store. | Bounded safe text and explicit public cache behavior; no uploads, HTML authoring, or unverified contact facts. |
| Roles | Settings write is SUPER_ADMIN/BPH; user administration and activity-log read are SUPER_ADMIN only; KEPALA_BIDANG cannot access these resources. | Enforced on the API and repeated as UI route guards; navigation hiding is not the security control. |
| User lifecycle | No hard delete. Status transitions are transactional and idempotent; protect self-deactivation/demotion and last active SUPER_ADMIN; require an active division only for KEPALA_BIDANG. Password reset hashes with bcrypt and revokes refresh sessions. | Preserves account/audit references and avoids stale concurrent update races. |
| Activity privacy | Read-only list/detail with pagination and filters; recursively redact sensitive keys; omit IP/User-Agent; record changed field names and safe metadata only. Reset uses existing UPDATE action vocabulary. | No new action enum or sensitive payload exposure; retention policy remains G-15. |
| Dashboard summary / visitor analytics | Implement operational counts from articles, currently open registration forms/submissions, aspiration statuses, available documents, and currently scheduled announcements. Use the authenticated database role/division for scope. Visitor/pageview analytics are deferred post-MVP and no tracker, cookie, fingerprint, analytics provider, or visitor-event table is introduced. | Provides source-backed Phase 9 metrics without collecting visitor data; global announcement count follows its existing divisionless domain and read policy. |

Phase 9 implementation and evidence are tracked in `PHASE_9_REPORT.md`.
