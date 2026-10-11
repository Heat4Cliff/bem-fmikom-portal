# Phase 9 — Landing Settings, User Management, and Activity Log

**Status: PHASE 9: BLOCKED**

Phase 9 implementation scope is complete for the three directly specified domains. Final phase acceptance is blocked by the unresolved dashboard/pageview requirement in `PLANS.md` (G-06) and unavailable browser UAT. Phase 10 has not started.

## Scope delivered

- Reversible `00007_system_management.sql` adds one constrained `landing_settings` row with safe neutral defaults; earlier migrations are unchanged.
- `GET /api/v1/public/settings`, `GET/PATCH /api/v1/admin/settings`; the admin settings page edits hero, about, vision/missions, contact, and social content, previews shared public components, and warns about unsaved changes. Public `/` consumes the API with the existing neutral fallback on API failure.
- `GET/POST/PATCH /api/v1/admin/users`, deactivate/reactivate/reset-password actions; `/admin/users` provides responsive table/cards and reusable create/edit/reset dialogs. No hard-delete control exists.
- `GET /api/v1/admin/activity-logs` and detail; `/admin/activity-logs` provides read-only filters, desktop/mobile lists, and a detail drawer.
- Role authorization is enforced by API middleware and UI guards: settings SUPER_ADMIN/BPH; user management and Activity Log SUPER_ADMIN only; KEPALA_BIDANG is denied.
- Role/division validation requires an active division for KEPALA_BIDANG and no division for other roles. User lifecycle operations protect self-deactivation/demotion and the last active SUPER_ADMIN, serialize status changes, and revoke refresh sessions on deactivation/password reset.
- Activity details are recursively redacted by sensitive key; metadata contains changed field names, not full settings/password payloads. IP and User-Agent are omitted from the Activity Log API/UI. Log rows are read-only.
- Stitch `screen.png` guided the desktop/mobile layouts; `code.html` was not used as production architecture. Source evidence remains read-only.

### Stitch coverage

The repository screen index contains **37 references**, including **8 explicit mobile references**. Phase 9 targets found there are `pengaturan_landing_page_admin_bem`, `manajemen_pengguna_admin_bem`, `activity_log_audit_trail_admin_bem`, and mobile `activity_log_mobile_admin_bem`; public content also uses `beranda_publik_bem_fmikom_unugha` and `beranda_publik_mobile_bem_fmikom_unugha`. Other Phase 9 forms/detail states reuse components because the index has no separate captures for them. Browser screenshot comparison was blocked with the local browser connection failure recorded below.

## Contract and scope reconciliation

- PRD/API endpoint naming is retained: `/public/settings`, `/admin/settings`, `/admin/users...`, `/admin/activity-logs...`. The indexed UI routes remain `/admin/settings`, `/admin/users`, and `/admin/activity-logs`.
- API/Security role policy is aligned with implementation and tested: SUPER_ADMIN/BPH may edit landing settings; only SUPER_ADMIN may administer users or view audit logs; KEPALA_BIDANG has no permission for these features.
- User deactivation uses status actions and preserves user/audit references; reset updates bcrypt password hash and revokes refresh sessions. No password is included in a response.
- Public settings use `Cache-Control: public, max-age=60, stale-while-revalidate=300`; admin system routes use `private, no-store`.
- `PLANS.md` still requires dashboard summary metrics and a selected pageview mechanism. The detailed Phase 9 execution scope supplied for this run covers landing settings, users, and Activity Log but does not define dashboard fields, aggregation scope, pageview collection/deduplication, retention, or privacy policy. G-06 remains open; no analytics endpoint, fabricated counter, or unsupported dashboard behavior was added. This is a real phase scope conflict requiring a product decision.
- G-15 retention is still undefined. Phase 9 redacts display data and limits returned metadata but does not invent a retention period.
- `GET /api/v1/admin/users/{id}` remains available per API contract and is integration-tested; current edit UI uses the already loaded list record, so the detail endpoint has no explicit UI consumer yet. Preview and mission reorder are client-side actions and do not require separate endpoints.

## Verification evidence

### Database / API

- Existing development Compose config and commands only: `infra/docker-compose.dev.yml`, `pnpm db:up`, and `pnpm api:integration`; no new MariaDB configuration was created.
- `TestDatabaseIntegration` ran against the dedicated `bem_fmikom_test` MariaDB schema and passed; it did not skip. The run migrated from version 0 through `00007`, rolled back to zero, migrated back to version 7, exercised assertions, and cleaned the schema back to version 0.
- `pnpm api:integration` passed again after the atomic user status-transition change. Coverage includes settings role/read/write/validation/singleton/public consumption, minimized audit fields, user list/detail/create/update and role/division checks, self/last-admin/concurrent status guards, reactivation, password reset/session revocation, inactive-login denial, activity filters/detail/read-only behavior, and nested sensitive-data redaction.
- The existing RBAC/division integration matrix also passed for SUPER_ADMIN and BPH cross-division access, KEPALA_BIDANG own-division access, cross-division denial, missing division denial, and wrong-role denial.

### Full repository verification

`pnpm verify` passed on 2026-10-10:

- Web lint, format, TypeScript typecheck: PASS.
- Frontend tests: 8 files, 35 tests passed (includes two Phase 9 system-management UI tests).
- Next.js production build: PASS; `/admin/settings`, `/admin/users`, and `/admin/activity-logs` are present in the route output.
- Go formatting check (59 files), `go vet`, all Go tests including real database integration, and Go build: PASS.

### Browser and source integrity

- Browser UAT was attempted after starting the Next dev server. The browser tool returned `net::ERR_CONNECTION_TIMED_OUT` for `localhost:3000`; a local HTTP smoke request also returned `HttpRequestException`. Therefore public visual smoke and role-authenticated browser workflows are NOT TESTED in this run. No browser PASS is claimed.
- `git diff --quiet -- source_docs design_reference` confirmed no source/design evidence changes.
- `git diff --check` passed.
- No Phase 10 work was started.

## Remaining blockers and risks

1. Resolve G-06: define dashboard metrics, role/division aggregation, pageview source/deduplication, privacy, retention, and API contract—or explicitly remove/defer these two Phase 9 checklist requirements.
2. Run browser UAT for public landing and SUPER_ADMIN/BPH/KEPALA_BIDANG route visibility and actions, then compare required desktop/mobile widths to Stitch `screen.png`. The current browser/network environment could not reach the local server.
3. Decide Activity Log retention and failed-refresh event vocabulary (G-15) before production audit-retention claims.
4. The user detail endpoint is retained by API contract but does not have a dedicated detail screen or current UI consumer.

## Recommended next step

Keep Phase 10 stopped. First settle the dashboard/pageview scope conflict (G-06), then rerun browser UAT in an environment that can reach the local server. If dashboard metrics/pageviews are formally deferred, record that scope decision, complete the remaining browser review, and reassess the Phase 9 gate.

**PHASE 9: BLOCKED**

## Phase 9B — Dashboard and Final Verification

This section preserves the original Phase 9 blocked record above and records the authorized follow-up. Visitor/pageview analytics are deferred post-MVP by product decision; no tracker, visitor counter, cookie/fingerprint, analytics provider, or visitor-event storage was added. The Stitch dashboard visit card is intentionally omitted.

### Dashboard contract and implementation

- Added `GET /api/v1/admin/dashboard` with a typed success envelope and `Cache-Control: private, no-store`.
- The dashboard reports total/published/draft articles; currently open forms and their submissions; aspiration totals by PENDING/READ/RESOLVED; available documents; currently scheduled active announcements; active period; generation timestamp; and resolved scope.
- SUPER_ADMIN and BPH receive organization-wide counts. KEPALA_BIDANG scope comes from the authenticated database principal and filters articles, registration forms/submissions, aspirations, and documents by its current division. No request query can widen scope. The announcements table has no division key and is globally readable for KEPALA_BIDANG, so only that metric remains global.
- Queries use bounded scalar SQL aggregates in one summary round trip (plus division name lookup for KEPALA_BIDANG), omit soft-deleted articles/forms and delete-pending documents, and return no student identities, aspiration text, document details, activity data, or visitor data. Unsupported Stitch approvals/storage-capacity mock panels remain omitted.
- `/admin` renders five responsive metric cards, the active period and resolved scope, and links to the existing routes. Frontend test covers data rendering, visible division scope, and the no-visitor-collection notice.

### Database integration evidence

- Runtime: Docker CLI 29.8.2, Docker Compose 5.5.1; Docker Desktop Server 29.8.2. Existing `infra/docker-compose.dev.yml` service `infra-mariadb-1` reported `running (healthy)` on MariaDB 11.4.13. No Compose/database configuration was added. Initial restricted-sandbox TCP access was denied; authorized broader local execution allowed Docker and MariaDB checks.
- Database: repository integration command `pnpm api:integration`; dedicated `TEST_DATABASE_URL` is `bem_fmikom_test`. `TestDatabaseIntegration` ran (not skipped), applied migration versions 1–7 from version 0, rolled down through version 0, migrated back to 7 for assertions, then cleaned to version 0. The router integration suite passed, including dashboard role and aggregate assertions.
- Dashboard tests verify unauthenticated denial, organization scope for SUPER_ADMIN/BPH, DB-resolved own-division identity and aggregate parity for KEPALA_BIDANG A/B, no-division service denial, and that the API does not expose visitor analytics. Aggregate assertions compare returned values to the corresponding MariaDB queries for articles, forms/submissions, aspirations, and documents.
- `pnpm verify`: PASS after dashboard integration. Lint, format, strict TypeScript check, 8 frontend test files / 36 tests, production build, Go formatting check (61 files), `go vet`, all Go tests including database integration, and Go build all passed.

### Stitch and responsive reconciliation

- Source evidence remains 37 Stitch screen references, including 8 explicit mobile references. Dashboard target references are desktop `dashboard_overview_admin_bem` and mobile `dashboard_admin_mobile_bem_fmikom_unugha`.
- Implemented counts align with available domain tables and access rules. The dashboard does not present unsupported approval/storage values or pageview values. Mobile layout uses a two-column card grid that collapses at narrow widths; screenshot-by-screenshot visual sign-off remains a Phase 11 check.

### Final verification, remaining limitations

- `pnpm api:integration`: PASS against real MariaDB; test was not skipped. Earlier unrestricted run before dashboard changes failed only because sandbox TCP was denied; authorized rerun passed.
- Browser UAT via the Codex in-app browser passed on 2026-10-10 using the test database: public `/` loaded; SUPER_ADMIN login opened the dashboard and showed User Management/Activity Log; BPH saw organization scope and Landing Settings while User Management/Activity Log were absent; KEPALA_BIDANG saw its database division scope while Landing Settings/User Management/Activity Log were absent. Reload restored the BPH session through the refresh cookie; logout and reload returned to login. The dashboard at the available 1101px browser viewport had no horizontal body overflow. Test-only accounts and period were created only in `bem_fmikom_test`; a subsequent integration run resets that schema.
- Manual visual comparison against Stitch desktop/mobile `screen.png` remains outstanding; exact 360/390/412 viewport emulation was unavailable through this browser surface. This is the only browser-review limitation and remains in the Phase 11 visual-review scope.
- Source integrity: `git diff --quiet -- source_docs design_reference` passed and `git status --short -- source_docs design_reference` returned no changes. Only Phase 9 checkboxes in `PLANS.md` were updated; Phase 10+ checklist content is untouched.
- G-06 is closed for Phase 9 by deferring visitor analytics; G-15 Activity Log retention and failed-refresh event vocabulary remain post-implementation operational questions and are not Phase 9 acceptance blockers.

**PHASE 9: PASS**
