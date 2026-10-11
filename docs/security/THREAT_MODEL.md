# Threat Model

**Scope:** BEM FMIKOM Portal implementation through Phase 10. This is a lightweight STRIDE-style review of implemented surfaces; deployment infrastructure outside this repository is not assumed.

## Assets and actors

Assets include admin credentials and access JWTs, refresh sessions, student registration PII, aspiration identity, private documents, activity-log metadata, CMS content, MariaDB records, and environment secrets. Actors include unauthenticated/external attackers, authenticated low-privilege admins, compromised KEPALA_BIDANG and BPH accounts, malicious applicants, and malicious document uploaders.

Entry points include public/admin REST APIs, login/refresh/logout, public forms, rich text rendering, externally hosted image URLs, document upload/download, the Next.js frontend, Go API, and MariaDB.

## Material threats

| STRIDE | Asset / entry point | Existing mitigation | Residual risk / Phase 10 action | Status |
|---|---|---|---|---|
| Spoofing | Admin credentials / login | bcrypt hashes, generic invalid-credential result, inactive-account denial, login rate limit, minimum password length and bcrypt byte ceiling | Rate limits are per-process and IP-based; deployment must preserve client-IP integrity | Hardened and regression tested |
| Spoofing / replay | Refresh sessions / refresh endpoint | 256-bit opaque token, SHA-256 storage, rotation, reuse family revocation, HttpOnly scoped cookie, trusted Origin, new refresh rate limit | No distributed limiter; review deployment cookie/origin settings | Hardened and integration tested |
| Tampering / elevation | Admin APIs / role and division IDs | DB-current role/status/division, API role middleware, division-scoped queries, no trusted proxy headers | Continue integration coverage as handlers evolve | Existing matrix retained; integration tested |
| Information disclosure | Registration PII, aspiration identity, activity logs | Role-scoped APIs, anonymous identity stripping, recursive activity metadata redaction, no IP/User-Agent in Activity Log UI/API | DB retention and operational log retention remain deployment/policy decisions | Reviewed; retention deferred |
| Tampering / XSS | CMS rich content and public rendering | Server-side HTML sanitizer allowlist; frontend renders sanitized API content; CSP nonce baseline in production | CSP compatibility needs browser validation for all normal flows | Sanitizer regression present; browser review pending |
| SSRF / disclosure | External image URLs | URL validation restricts schemes; images load in browser with Next optimization disabled; no server-side URL fetch path found | Trusted content editors can still publish third-party URLs and expose visitor IP to that host | Reviewed; residual content/privacy risk |
| DoS / parser abuse | JSON APIs, list queries, uploads | Global bounded JSON/query/pagination guards; dedicated upload size/type checks; bounded in-memory limiter key maps | Rate limits are not distributed and multipart resource limits remain route-specific | Added and unit tested; integration/full verify pending |
| Traversal / disclosure | Private archive upload/download | `os.Root` confinement, opaque storage keys, scoped authorization, safe streaming, size/hash checks; OOXML entries are inspected without extraction and bounded | No antivirus/CDR scanning; infected files may be stored if they pass type checks | Existing controls reviewed; runtime regression pending |
| Repudiation / privacy | Activity logs and operational logs | Request IDs validated; request logs exclude body/query/IP; panic response/log is generic; Activity Log omits raw IP and User-Agent | Authentication/refresh service events may log source IP; access must be restricted and retention defined | Reviewed; privacy finding documented |
| Elevation / secrets | Production config and DB | Production requires a 32-byte diverse/non-placeholder JWT secret, HTTPS trusted origins, Secure refresh cookie, absolute storage path, and non-root DB principal | DB grants, secret manager, TLS termination, backups, and file permissions depend on deployment | Config checks added; deployment evidence pending |

## Trust boundaries

1. Browser to Next.js: CSP and response headers; admin pages are `private, no-store`.
2. Next.js to API: explicit configured API origin and credentialed allow-list; bearer access token plus refresh cookie.
3. API to database: parameterized SQL and application DSN; production must use a dedicated principal.
4. API to private storage: confined opaque object keys; client filenames are metadata only.
5. API to outside image hosts: browser-side loads only; the server does not retrieve supplied image URLs.

## Decisions

- Cookie-authenticated mutation endpoints require an exact trusted `Origin`; CORS is explicit and credentialed, while access-token endpoints additionally require bearer authorization.
- No CSRF token was added: refresh/logout require trusted Origin and use a path-scoped SameSite cookie; browser CORS cannot authorize arbitrary origins. Reassess if cookie-authenticated business endpoints are added.
- Distributed rate limiting, malware scanning, CDR, SIEM, and external secret management remain deployment/product work and were not introduced in this phase.
