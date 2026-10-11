# Phase 10 Report

## Scope

Security audit and hardening of the Phase 1–9 implementation only. No business feature, Phase 11 integration/visual work, or Phase 12 UAT/deployment work was started. Source/design reference folders remained read-only.

## Threat Model

Threats reviewed for admin credentials, access and refresh tokens, registration PII, aspiration identity, private documents, activity logs, CMS content, database records, and environment secrets. Actors included unauthenticated/external attackers, low-privilege admins, compromised BPH/KEPALA_BIDANG accounts, malicious applicants, and document uploaders. See [THREAT_MODEL.md](../security/THREAT_MODEL.md) for STRIDE-like entry-point and residual-risk mapping.

## Authentication Hardening

- Passwords remain bcrypt (default cost 12); values beyond 72 bytes are rejected rather than truncated. Generic invalid-credential response, inactive-user denial, and DB-current identity remain in place.
- JWT parsing now requires expiration and issued-at claims, uses only HS256, validates issuer, positive subject, and JTI.
- Production configuration requires a strong unique 32-byte-minimum JWT secret and rejects known placeholders/low-diversity secrets.

## Session / Refresh Hardening

- Refresh credentials remain 256-bit opaque random values, with only SHA-256 hashes stored. Rotation/reuse family revocation, expiry, logout, inactive-user denial, and secret redaction are covered by database integration.
- Refresh token syntax is validated before hashing/database lookup and the refresh endpoint now has a bounded configurable per-IP limiter.
- The refresh cookie remains HttpOnly and auth-path-scoped. Production requires Secure; `SameSite=None` requires Secure. Cookie domain remains explicit configuration.

## RBAC / IDOR Audit

Database-current role/status/division remain authoritative. MariaDB integration verified SUPER_ADMIN/BPH cross-division operational access, KEPALA_BIDANG own-division access, cross-division denial for both division heads, missing-division denial, wrong-role 403s, and existing user/document/content route scope checks. No authorization bypass was found.

## CORS / CSRF

CORS uses exact configured origins with credentials and never wildcard/reflection. Integration denied an untrusted Origin and verified trusted Origin headers. Auth lifecycle POSTs require a trusted Origin. No extra CSRF token was added: cookie-authenticated refresh/logout are SameSite-protected and exact-Origin-gated; bearer-authenticated API mutations require Authorization. Reassess if cookie-authenticated business mutations are introduced.

## Security Headers

Next production response includes `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, strict Referrer Policy, Permissions Policy, COOP, and CSP. Admin login response was HTTP 200 with `Cache-Control: private, no-store`. The Go API middleware header tests pass. Local HTTP correctly omitted HSTS; frontend sets HSTS only when the configured site URL is HTTPS.

## CSP

Production CSP is nonce-based, with explicit script/style/image/font/connect directives and `object-src 'none'`, `base-uri 'self'`, `frame-ancestors 'none'`, and `form-action 'self'`; no `unsafe-eval` or `unsafe-inline` was added. The production build passed and direct local HTTP confirmed a per-response nonce policy. Browser console violation capture could not be completed because the in-app browser could not reach the loopback server; see Browser Security Smoke.

## Input Validation

Global API guards now cap JSON bodies at 2 MiB, queries at 8 KiB, `search` at 256 bytes, page size at 100, and offset at 10,000,000; duplicate query parameters and unsupported body media types are rejected. Multipart is only accepted for private-document upload, whose existing configured upload cap remains authoritative. Server deadlines and maximum headers are bounded.

## SQL Injection Review

Request values are bound through SQL placeholders. Dynamic query assembly consists of static predicate fragments, static ordering, and a generated placeholder list sized from validated IDs. Gosec G202 reports were manually reviewed against these construction sites; no user-controlled SQL identifiers or raw query fragments were found. MariaDB integration and SQL-bearing tests pass.

## XSS Review

The server-side rich-text sanitizer retains its allow-list and unsafe element/attribute/URL defenses. React escapes plain text. The only `dangerouslySetInnerHTML` call renders sanitized CMS content. Existing sanitizer and content tests pass; no new HTML sink was introduced.

## URL / SSRF Review

Content URL validation accepts HTTPS for external media. Next images use unoptimized browser-side loading; no `remotePatterns`, ImageResponse/OG route, backend HTTP fetch of supplied URLs, or query-driven open redirect was found. External media hosts can observe visitor network requests; trusted-editor content remains a privacy residual risk.

## Rate Limiting

Login, refresh, public aspiration, and registration are limited per IP in memory. New-key storage is capped at 4096 entries and stale entries are purged before admission when full. Gin trusted proxies are disabled. Limits are process-local; horizontally scaled deployments need an edge/shared limiter.

## File Upload Hardening

The private-document upload path retains configured size limits, allow-listed types/signatures, bounded metadata, safe names, and cleanup on failure. Upload and private-document regression tests pass. No antivirus or CDR is present.

## Private Storage Review

Storage remains outside public/static assets and uses `os.Root` confinement, opaque keys, restrictive POSIX modes, streaming size/hash checks, and authorized downloads. OOXML ZIP inspection bounds entry count and expanded size and reads needed entries without extracting archive paths. Traversal, authorization, integrity, and document integration tests pass.

## Logging / Privacy

Request IDs are validated; request logs omit bodies/query strings/IP/User-Agent; panic responses/logs are generic and omit panic contents. Auth structured-log regression checks found no password, raw refresh token, or JWT secret. Activity Log API/UI omit IP/User-Agent and recursively redact sensitive metadata. Auth/security audit rows still retain source IP; policy retention/access control remains a deployment decision.

## Secrets / Configuration

Production startup rejects unsupported APP_ENV, missing explicit trusted HTTPS origins, localhost/private origins, root DB principal, weak/default JWT secret, insecure refresh cookie, and relative private storage path. `.env` remains ignored; tracked-source secret-pattern review found no matching secret material. Local example secrets remain placeholders and were not treated as production credentials.

## HTTP / Database Hardening

HTTP server now uses 5 s ReadHeaderTimeout, 3 min ReadTimeout/WriteTimeout, 60 s IdleTimeout, and 32 KiB MaxHeaderBytes. Database pool now applies configured max idle time. Existing migration constraints/indexes and transaction protections remained unchanged and exercised against MariaDB; application guidance requires a least-privilege non-root production principal.

## Dependency Audit

- `pnpm audit`: PASS; 309 frontend dependencies including dev/optional, 0 advisories in all severities.
- Final Go scan after patching reachable findings reports zero code-reachable vulnerabilities. One module-only `golang.org/x/crypto/openpgp` advisory has no fixed version; OpenPGP is not imported by this application. The required `x/crypto` bcrypt path remains in use.
- Reachable advisories identified during the audit were resolved by pinning Go 1.26.9, `golang.org/x/net` 0.60.0, and `github.com/quic-go/quic-go` 0.59.1.

## Static Security Analysis

`gosec` full scan produced 40 findings (2 medium G124 cookie flags, 11 medium G202 static-query assembly, and 27 low G104 cleanup/close errors). The medium findings were manually reviewed and are documented in [SECURITY_AUDIT.md](../security/SECURITY_AUDIT.md); no `nosec` suppressions were added. The configured high-severity/high-confidence gate passed with 0 findings.

## CI Security

The existing read-only CI now runs full `pnpm audit`, pinned `govulncheck`, and pinned `gosec` at high severity/high confidence before `pnpm verify`. Each command was run locally against the same sources and passed at its configured gate.

## Findings Fixed

Required JWT claims, opaque refresh validation and refresh limiting, production config constraints, global JSON/query/pagination bounds, limiter map bounds, HTTP timeouts/header cap, DB idle time, frontend nonce CSP/security headers, and Go dependency patches are implemented and regression-tested.

## Findings Deferred

- `GO-2026-5932` applies to unused `x/crypto/openpgp` packages; there is no known fixed version. It is non-reachable and the Go vulnerability scanner reports zero affected code paths.
- LOW gosec cleanup/close findings remain; MEDIUM G124/G202 scanner reports are reviewed false positives with config/static-fragment evidence.
- Source-IP retention/access policy and antivirus/CDR are deployment/policy limitations.
- Browser DevTools CSP-console review was unavailable; direct response header inspection and production compilation succeeded.

## Known Limitations

Rate limits are process-local. External HTTPS media may disclose visitor IP to the chosen host. No antivirus/CDR, SIEM, WAF, distributed limiter, or cloud secret manager exists. Log/activity retention is not specified. OWASP ZAP was not available. Browser console inspection and authenticated role-based visual smoke were not completed because the in-app browser could not connect to local loopback.

## Verification

### AUTH SECURITY TEST

**PASS.** `pnpm api:integration` exercised valid and invalid login, missing/inactive account denial, `/auth/me`, refresh rotation, old-token replay denial with no access token, family revocation, expired refresh denial, logout/idempotent logout, cookie attributes, and hash-only DB storage.

### RBAC / IDOR TEST

**PASS.** MariaDB integration role/division matrix and existing document/content authorization regressions passed.

### CORS TEST

**PASS.** Trusted Origin credentialed allow-list accepted; untrusted Origin denied without CORS allow headers.

### CSRF REVIEW

**PASS.** Refresh/logout exact-Origin checks plus SameSite cookie; no unprotected cookie-authenticated business mutation found.

### SECURITY HEADERS TEST

**PASS.** API middleware test passed. Production Next login response returned 200 with CSP, `nosniff`, frame denial, Referrer Policy, and `private, no-store`.

### CSP TEST

**PASS WITH EVIDENCE LIMIT.** Production build and loopback response confirmed nonce CSP and strict directives. Browser console capture was not available because the app browser could not reach host loopback. HSTS was absent on local HTTP as intended.

### XSS TEST

**PASS.** Sanitizer regression suite and source review passed; only sanitized rich content uses an HTML injection sink.

### SQL INJECTION REVIEW

**PASS.** Placeholder binding and static dynamic-clause construction reviewed; G202 reports triaged as false positives.

### RATE LIMIT TEST

**PASS.** Login, refresh, aspiration, and registration limiter tests pass; limiter key capacity remains bounded.

### FILE UPLOAD TEST

**PASS.** Private document validation and integration regression pass; malware scanning is not present.

### PATH / ZIP SAFETY TEST

**PASS.** Confined-root/path traversal and bounded OOXML tests pass; ZIP members are never extracted.

### LOG / SECRET REDACTION TEST

**PASS.** Structured auth logs omit password, raw refresh token, and JWT secret; activity display omits raw IP/User-Agent. Tracked-source pattern scan found no secret pattern.

### DEPENDENCY AUDIT

**PASS.** `pnpm audit` (0 advisories); `govulncheck` (0 reachable vulnerabilities; 1 non-reachable module advisory).

### GOVULNCHECK

**PASS.** `go -C apps/api run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...`: 0 reachable vulnerabilities.

### GOSEC

**PASS AT CI GATE; FINDINGS TRIAGED.** `gosec@v2.29.0 -severity high -confidence high ./...`: 0 issues. Default scan: 40 low/medium reports reviewed in `SECURITY_AUDIT.md`.

### OWASP ZAP BASELINE

**NOT AVAILABLE.** ZAP was not installed/runnable; the supplied instruction made the baseline optional.

### DATABASE INTEGRATION

**PASS.** Docker 29.8.2, Compose 5.5.1, repository MariaDB image `mariadb:11.4` (server 11.4.13), container health `healthy`. `pnpm api:integration` ran `TestDatabaseIntegration` against `bem_fmikom_test` without skip, migrated empty version 0 through migration 7, rolled down and back up, passed 12 named auth/RBAC/domain subtests plus the inline auth/document/CMS regression checks, and cleaned back to migration version 0.

### FULL VERIFY

**PASS.** Final `pnpm verify` passed lint (103 files), format (103 files), TypeScript, frontend tests (8 files / 36 tests), Next production build, Go formatting (63 files), vet, all Go tests including real MariaDB integration, and Go build on Go 1.26.9.

### BROWSER SECURITY SMOKE

**PARTIAL / NOT AVAILABLE FOR CONSOLE.** Host loopback HTTP verified `/admin/login` status 200 and production CSP/security headers. The in-app browser timed out connecting to loopback, so no browser console, login/dashboard role UI, or authenticated routes were asserted. No public bind or tunnel was used.

## Source Integrity

`git diff --quiet -- source_docs design_reference` returned clean. Ignored local `.env` remains untracked and unchanged by this work. `PLANS.md` edits are restricted to verified Phase 10 items.

## Phase Boundary

Phase 11 and Phase 12 checklists remain untouched. No Phase 11 visual/integration work or Phase 12 deployment readiness work was started.

## Phase Result

**PHASE 10: PASS WITH DOCUMENTED NON-BLOCKING FINDINGS**

No confirmed HIGH or CRITICAL application vulnerability remains. Residual evidence/policy limitations are listed in `docs/security/SECURITY_AUDIT.md` and do not block Phase 10 security acceptance.
