# Security Checklist

## Before development release

- [ ] Keep local secrets in ignored `.env`; use non-production credentials and a dedicated `_test` database.
- [ ] Run `pnpm verify` and `pnpm api:integration`; confirm the integration test executes against MariaDB and does not skip.
- [ ] Review role and division scope for every new handler and UI action.
- [ ] Validate inputs, pagination, file type/size, rich text, and URL schemes on the server.
- [ ] Keep private documents outside public/static directories and retain authorized streaming downloads.
- [ ] Review dependency audit output and static-analysis findings.

## Before staging

- [ ] Set explicit trusted HTTPS frontend origins and `APP_ENV=production`-equivalent settings.
- [ ] Provision a dedicated least-privilege database account; prohibit root application credentials.
- [ ] Generate a unique random JWT secret with at least 32 bytes; store it in the deployment secret facility.
- [ ] Confirm Secure HttpOnly refresh cookie, SameSite policy, cookie domain/path, TLS termination, and trusted proxy topology.
- [ ] Use an absolute private document storage path; verify filesystem ownership, permissions, backup, and restore.
- [ ] Review CSP against deployed frontend/API origins and check browser console for violations.
- [ ] Decide log and activity retention/access policy; treat client IP as personal/operational data.

## Before production

- [ ] Enforce HTTPS end-to-end or at an approved TLS-terminating edge; validate HSTS scope before enabling.
- [ ] Verify database grants, network exposure, backups, restore drill, migration rollback plan, and monitoring.
- [ ] Confirm rate limits at the edge or shared store when running multiple API instances.
- [ ] Decide whether malware scanning/CDR is required for uploaded documents; current application does not scan malware.
- [ ] Run current dependency audit, Go vulnerability scan, Go static security scan, API integration, full verification, and browser security smoke.
- [ ] Complete role-based manual UAT and release approval; do not infer manual approval from automated tests.

## Recurring maintenance

- [ ] Review dependency/security advisories and update supported dependencies.
- [ ] Re-run verification and database integration after auth, API, storage, dependency, or deployment changes.
- [ ] Review logs for access, retention, redaction, and unexpected PII.
- [ ] Test backup restoration and private-document integrity.
- [ ] Revisit rate limits, CSP, trusted origins, database privileges, and incident contacts as infrastructure changes.
