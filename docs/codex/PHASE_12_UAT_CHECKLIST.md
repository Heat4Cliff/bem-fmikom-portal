# Phase 12 — Manual UAT Checklist

## Run record

- Environment / build commit:
- Browser and version:
- Desktop viewport(s):
- Mobile viewport(s): 360 / 390 / 412 CSS px
- UAT schema and seed marker:
- Tester / date:
- Result per row: `PASS`, `FAIL`, or `NOT TESTED`; attach sanitized evidence and defect ID.

Use synthetic fixtures from `PHASE_12_RESET_SEED.md`. Do not record passwords, refresh cookies, access tokens, applicant identity, or private document content. Automated test results do not substitute for these manual checks.

## Public (anonymous)

- [ ] Home, organization `/struktur`, articles, announcements, gallery, registration list/detail, and aspiration route load without authentication.
- [ ] Public pages show only published/active data according to each endpoint; draft registration/article data is not disclosed.
- [ ] Organization shows current active period and public structure only; no internal IDs, NIM, or audit fields are visible.
- [ ] Public registration form validates required fields and options; a unique synthetic submission receives a success reference; duplicate NIM is rejected according to the form rule.
- [ ] Registration IMAGE upload accepts a valid supported synthetic image, displays validation errors for rejected input, and does not expose bytes via public media URLs.
- [ ] Aspirasi accepts valid input, displays success/error/rate-limit states, and does not expose admin-only details.
- [ ] Public media renders; unavailable APIs and empty content show deliberate fallback/empty states.
- [ ] Desktop and widths 360/390/412 have no horizontal overflow, clipped actions, or broken image/card bounds; keyboard focus and reduced-motion behavior remain usable.

## KEPALA_BIDANG (one account each for DIV-A and DIV-B)

- [ ] Login, reload/session refresh, `/auth/me`-backed current role/division display, and logout work.
- [ ] Own-division articles, gallery items, members, aspirations, registration forms/submissions, and documents are visible/actionable only where the role contract allows.
- [ ] Cross-division direct URL/detail/download/export attempts are denied or concealed per contract; repeat with both divisions.
- [ ] A user with no active division cannot be used as a valid KEPALA_BIDANG principal; server denies protected operations.
- [ ] Landing settings, user management, and Activity Log are absent from navigation and direct endpoint access is denied.
- [ ] Private documents and registration image attachments are accessible only through authorized scoped endpoints; no direct storage path works.
- [ ] Division filters or IDs changed in URL/body do not widen scope.
- [ ] Mobile admin navigation/forms/actions remain usable at 360/390/412; empty/error/permission states are legible.

## BPH

- [ ] Login, reload/session refresh, current role display, and logout work.
- [ ] Cross-division operational content, registration, aspiration, organization, and private-document tasks allowed by the API contract work.
- [ ] Landing Page Settings is visible and can be edited/saved; public view reflects safe updated content.
- [ ] User Management and Activity Log are not available in navigation; direct API requests are denied.
- [ ] BPH does not gain SUPER_ADMIN-only period activation or user/security administration.
- [ ] Uploads and private file access follow purpose, format, and authenticated endpoint rules.

## SUPER_ADMIN

- [ ] Login, reload/session refresh, `/auth/me`, and logout work; a reload after logout remains unauthenticated.
- [ ] User Management and Activity Log are visible; BPH/KB-only restrictions remain enforced by the API even if routes are entered directly.
- [ ] Create/update/deactivate/reactivate/reset-password flows validate role/division rules, preserve the last active SUPER_ADMIN guard, and revoke stale sessions as expected.
- [ ] Activity Log filters/detail show safe metadata; password/token/raw IP/User-Agent and private file bytes are absent.
- [ ] Period activation keeps the single-active-period invariant and archives the prior active period.
- [ ] Cross-division operational actions work where authorized; invalid division assignments are rejected.
- [ ] Dashboard values are operational aggregates only; no visitor/pageview analytics are implied.

## Cross-role security and release checks

- [ ] Wrong role receives 403 (or documented concealment 404); unauthenticated protected access receives 401/login redirect.
- [ ] Untrusted Origin is rejected; trusted HTTPS Origin and credentialed CORS behave as configured.
- [ ] Production refresh cookie is `HttpOnly; Secure`; configured SameSite policy is valid (`None` requires Secure).
- [ ] Login/public-write throttling gives bounded, understandable errors and does not leak credentials or tokens to logs.
- [ ] Browser console/network review has no CSP violations, missing production assets, unexpected 4xx/5xx requests, or leaked secrets.
- [ ] Backup/restore drill is completed in a separate isolated restore target before launch; see `PHASE_12_BACKUP_RESTORE.md`.

## Acceptance

Manual rows are accepted only with tester/date/evidence. Mark unexecuted rows `NOT TESTED`; automated regression or route HTTP 200 alone cannot be recorded as manual acceptance. Phase 11 follow-ups remain separate and are not silently closed by this checklist.
