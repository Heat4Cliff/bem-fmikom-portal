# Security Baseline

This document records the application controls in the repository, not proof of production deployment configuration.

## Authentication and sessions

- Passwords use bcrypt (default cost 12); values above bcrypt's 72-byte input ceiling are rejected rather than silently truncated.
- Invalid credentials share a generic response; inactive accounts cannot authenticate. The database remains authoritative for current account status, role, and division.
- Access JWTs use HS256 only, a configured issuer, required expiration and issued-at claims, positive numeric subject, and JTI.
- Refresh credentials are 32 random bytes encoded as base64url; only SHA-256 hashes are stored. Rotation/reuse revocation and logout are covered by MariaDB integration tests.
- Refresh cookie is HttpOnly, scoped to `/api/v1/auth`, SameSite-configurable, and Secure by default in production. Production rejects Secure=false and requires Secure for SameSite=None.

## Authorization and CSRF

- API middleware enforces role and division checks from database-current identity. KEPALA_BIDANG is division-scoped; SUPER_ADMIN and BPH receive only their documented global operational access.
- Credentialed CORS uses an explicit origin allow-list. Auth lifecycle POSTs require an allow-listed Origin. Public write APIs reject a supplied untrusted browser Origin.
- No synchronizer token is used because cookie-authenticated actions are limited to refresh/logout and protected by exact Origin checks plus SameSite; do not add new cookie-authenticated mutation routes without reassessing.

## Web/API headers and CSP

- Next.js adds `nosniff`, frame denial, strict referrer policy, restricted Permissions Policy, and same-origin opener policy. Production routes use a nonce CSP and HTTPS-only HSTS when the configured site URL is HTTPS. Admin paths are `private, no-store`.
- Go API responses set corresponding headers and a restrictive API CSP (`default-src 'none'; frame-ancestors 'none'`). Local HTTP does not set HSTS.
- CSP behavior must be checked in a real browser after changes to scripts, styles, fonts, image hosts, or API origin.

## Input, content, and storage

- JSON bodies are capped at 2 MiB; document multipart requests use the document service's configured size cap. Query length, duplicate values, search size, pagination size, and maximum offset are bounded.
- SQL uses prepared/bound values; variable sort/filter fragments are allow-listed. CMS rich text is sanitized on the server before public rendering. Plain text is rendered through React escaping.
- URL validation restricts accepted schemes. External images are browser-loaded with Next image optimization disabled; the API does not fetch user-supplied URLs.
- Private documents use confined `os.Root` paths, opaque object keys, authorization before streaming, size/type/signature checks, content hashes, and bounded OOXML inspection without archive extraction.
- Malware scanning and CDR are not implemented. Do not describe accepted files as malware-free.

## Abuse, logs, and operations

- Login, refresh, aspirations, and registration use bounded per-process IP rate limiters. Gin trusted proxies are disabled; forwarded headers are not a trusted rate-limit identity.
- Panic responses and logs do not include panic contents. Request logs use route templates and omit request bodies, query values, IP, and User-Agent. Request IDs are validated before reuse.
- Activity Log API/UI omit raw IP/User-Agent and recursively redact sensitive metadata keys. Operational auth/security events may include source IP; protect log access and set retention at deployment.
- Production config rejects weak/default JWT secrets, non-HTTPS/local trusted origins, Secure=false refresh cookies, root DB principal, and relative private document storage path. Provision a least-privilege DB account and external secret handling in deployment.
- Configure HTTPS, secure backups, restricted document storage permissions, log access/retention, patch cadence, and recovery testing before production.

## Known limitations

- Rate limiting is process-local and can be bypassed by distributing requests across instances; configure an edge limit if horizontally scaled.
- No antivirus/CDR scanner, SIEM, WAF, or cloud secret manager exists.
- DB activity and authentication events contain some client network metadata; policy retention and access controls remain deployment decisions.
- A dependency audit reports only advisories known to the registry snapshot at scan time.
