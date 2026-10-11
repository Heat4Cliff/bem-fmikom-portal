# Security Audit — Phase 10

**Audit date:** 2026-10-11
**Scope:** Implemented application through Phase 9 and Phase 10 security changes. Findings below distinguish confirmed defects from static-scanner triage and verification limitations.

## Findings

| ID | Severity | Area | Description and evidence | Fix / verification | Status |
|---|---|---|---|---|---|
| SEC-10-01 | MEDIUM | JWT | JWT library did not require `exp`/`iat`; issuer, HMAC method, subject, and JTI checks existed. | Require expiration and issued-at validation, pin HS256, reject missing/invalid claims; token regression tests and integration PASS. | Fixed |
| SEC-10-02 | MEDIUM | Refresh abuse | Refresh lacked a dedicated request limit; arbitrary strings could reach the token hash/database path. | Add per-process refresh limiter and strict 32-byte base64url token shape validation; limiter/token tests and integration PASS. | Fixed |
| SEC-10-03 | MEDIUM | Request resource limits | Not all JSON mutations and list queries had a shared maximum. | Add 2 MiB JSON cap, query/search/page/limit/offset bounds and supported-media-type checks; middleware tests and full verification PASS. Document multipart is separately bounded by document upload limits. | Fixed |
| SEC-10-04 | MEDIUM | Production configuration | Config could accept unknown environment names and insecure production cookie/origin/database combinations. | Validate APP_ENV, explicit HTTPS non-local origins, non-root DB account, production Secure cookie, `SameSite=None`/Secure pairing, secret strength, and absolute private storage path; config tests PASS. | Fixed |
| SEC-10-05 | LOW | HTTP server | Read/write and header-size protections were incomplete. | Set 5 s header timeout, 3 min read/write timeout (supports bounded document transfers), 60 s idle timeout, 32 KiB max headers; full Go build/test PASS. | Fixed |
| SEC-10-06 | LOW | In-memory rate limiters | Unique-IP limiter keys could grow until stale-entry sweep. | Bound per-limiter entries at 4096 and remove expired keys before admitting new keys; rate-limit tests PASS. | Fixed |
| SEC-10-07 | MEDIUM | Go dependencies | govulncheck identified reachable standard-library, `x/net`, and `quic-go` advisories in initial versions. | Upgrade Go 1.26.1 → 1.26.9, `x/net` 0.58.0 → 0.60.0, `quic-go` 0.59.0 → 0.59.1. Final govulncheck reports zero reachable vulnerabilities. | Fixed |
| SEC-10-08 | INFO | Go dependency graph | `GO-2026-5932` applies to `golang.org/x/crypto/openpgp`, for which the advisory lists no fixed version. This application imports bcrypt, not OpenPGP; final govulncheck reports it only among required modules and reports zero affected code paths. | Keep `x/crypto` for bcrypt. Reassess if OpenPGP packages are introduced or module contents change. | Deferred; non-reachable |
| SEC-10-09 | MEDIUM / LOW | gosec | Full gosec produced 40 reports: 2 G124 cookie-attribute reports, 11 G202 SQL-concatenation reports, and 27 LOW G104 ignored cleanup/close errors. | Manual review found cookie security selected from config and production-validated; all G202 queries concatenate only static clauses and bind dynamic values (generated `?` placeholders for `IN`); G104 findings concern cleanup/close paths. No `nosec` suppressions added. High severity/high confidence gosec gate passes with 0 findings. | Triaged; non-blocking |
| SEC-10-10 | LOW | Log privacy / retention | Auth/security events and database activity rows include source IP (and activity rows may include User-Agent). Activity Log API/UI omit these fields and sensitive metadata is redacted, but the retention period is a deployment/policy decision. | Keep operational log access restricted and set retention before production; no unverified retention period was invented. | Deferred policy decision |
| SEC-10-11 | INFO | Browser evidence | Loopback production login returned 200 with nonce CSP and expected headers. The in-app browser could not reach the local loopback server, so browser console/CSP-violation capture and role-authenticated UI smoke were unavailable. | Production build and direct HTTP header checks pass; follow-up browser review should use a browser environment sharing the host loopback. Server remained loopback-only. | Evidence limitation |

## Scanner and review totals

- `pnpm audit`: 0 info, 0 low, 0 moderate, 0 high, and 0 critical advisories across 309 resolved frontend dependencies (including dev and optional dependencies).
- `govulncheck`: 0 vulnerabilities in reachable application code; 1 module-only OpenPGP advisory described above.
- `gosec`: 0 findings at the CI gate `-severity high -confidence high`; full default scan findings and manual triage are recorded in SEC-10-09.
- Tracked-source secret pattern review: no private-key, cloud-key, GitHub-token, or JWT-shaped matches. The ignored local `.env` was not printed or added to Git; `.env.example` values remain clearly local placeholders.
- SQL review: service queries use bound `?` arguments for request values; assembled predicates/orderings are static and pagination is bounded. No user-controlled SQL identifier or raw sort expression was found.
- URL/SSRF review: external image URLs are HTTPS-validated and loaded by the browser with Next image optimization disabled. No server-side request path for user-supplied URLs or redirect-to-query flow was found.
- Private archive review: object keys are opaque and confined by `os.Root`; upload size/type/signature checks and bounded OOXML inspection are present; ZIP entries are read in place, never extracted. Antivirus/CDR is not implemented.

## Verification limitations

OWASP ZAP is optional in the supplied Phase 10 instructions and was not installed/runnable in the local environment. Browser console inspection was unavailable because the in-app browser could not connect to the loopback-only host, even though direct host HTTP returned the page and headers. Neither limitation is evidence of a confirmed high/critical application vulnerability.
