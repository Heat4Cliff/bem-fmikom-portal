# Phase 12 — Deployment Readiness Runbook

This is a preparation guide, not authorization to deploy. The approved Go/API and MySQL/MariaDB hosting providers, DNS, region, RPO/RTO, retention, and operational owner must be selected outside this repository before a production change. Do not invent provider-specific settings from this generic baseline.

## Topology and prerequisites

- Frontend: Vercel project rooted at `apps/web`, using the repository lockfile/workspace or the platform-supported monorepo root/build configuration. Build with `pnpm --filter @bem-fmikom/web build`; launch the generated Next.js app with the Vercel Next.js runtime. Do not use `BEM_DEMO_CONTENT=true` in production.
- API: a persistent Go service from `apps/api`, built with the pinned Go toolchain and `go build ./...`; listen on the host-provided port by setting `APP_PORT`. Expose HTTPS through the approved host's TLS/reverse-proxy layer and health-check `/healthz`.
- Database: managed MySQL/MariaDB compatible with the repository's MySQL driver and checked-in migrations. Use a dedicated non-root app user, TLS where supported, private networking, and backups. Compatibility must be confirmed against the selected engine/version before launch; the local and CI baseline is MariaDB 11.4.
- Storage: durable, private API-local paths for `DOCUMENT_STORAGE_PATH`, `REGISTRATION_STORAGE_PATH`, and `MEDIA_STORAGE_PATH`. Mount them outside web roots. Back them up as a matched recovery set with SQL. The repository has no S3 adapter or object-storage deployment integration.
- Browser origins: the frontend and API must be HTTPS. Configure one exact API `FRONTEND_ORIGINS` origin for credentialed CORS. Refresh-cookie domain/SameSite must match the chosen origin topology; `SameSite=None` requires Secure.

## Build and release sequence

1. Select approved providers and populate the non-secret release ticket with target URLs, owner, region, resource sizing, scaling, health check, restore plan, RPO/RTO, retention, and rollback operator.
2. Review Phase 11 follow-ups and complete the role-based manual UAT checklist appropriate to the release. Verify current `main` commit and clean source evidence.
3. In CI/release candidate, run `pnpm api:integration` against a dedicated disposable `_test` database (not skipped), `pnpm verify`, and `pnpm audit`. Build web/API artifacts from the same commit; capture immutable commit SHA and checksums.
4. Provision production secrets/configuration in the hosting secret facility and private database/storage resources. Never build secret values into a Vercel public variable, image, artifact, or log.
5. Before migrations, take a verified recovery set per `PHASE_12_BACKUP_RESTORE.md` and record its recovery-set ID.
6. Apply pending migrations from the release artifact using `pnpm api:migrate:up` only after the release operator verifies its `DATABASE_URL` targets the intended production database. Since the CLI loads local env files, run in a controlled job/container with only the production secret injected and no repository `.env` mounted.
7. Deploy API and Vercel frontend from the same commit. Set Vercel `NEXT_PUBLIC_API_BASE_URL` to the public API origin plus `/api/v1`, and `NEXT_PUBLIC_SITE_URL` to the approved canonical public origin. Rebuild frontend whenever public build-time values change.
8. Verify API `/healthz`, public endpoints, login/refresh cookie, CORS from the exact frontend origin, core public page routes, admin route protection, private file authorization, and database-backed health before announcing availability.
9. Observe API error/auth logs and provider health/backup status. Keep the previous API artifact and Vercel deployment identifiable until the release is accepted.

## Production environment and secret inventory

Supply secret values only through the host secret manager; this table is names/purpose, not values.

| Variable / setting | Runtime | Required policy |
|---|---|---|
| `APP_ENV` | API | `production` |
| `APP_PORT` | API | Match host-provided listening port |
| `FRONTEND_ORIGINS` | API | Explicit trusted HTTPS frontend origin(s); no wildcard/localhost/private origin |
| `DATABASE_URL` | API / migration job | MySQL DSN with dedicated non-root least-privilege account, TLS/private connection; no URL logging |
| `JWT_SECRET` | API | Unique random secret, at least 32 bytes, high diversity; never reuse local/example value |
| `AUTH_REFRESH_COOKIE_SECURE` | API | `true`; `SameSite=None` also requires Secure |
| `AUTH_REFRESH_COOKIE_SAMESITE` | API | Explicitly choose `lax`, `strict`, or `none` for the origin topology |
| `AUTH_REFRESH_COOKIE_DOMAIN` | API | Leave unset unless the approved shared-domain topology requires it |
| `AUTH_ACCESS_TOKEN_TTL`, `AUTH_REFRESH_TOKEN_TTL`, `AUTH_BCRYPT_COST` | API | Reviewed security defaults; do not weaken without approved decision |
| `AUTH_LOGIN_RATE_LIMIT`, `AUTH_REFRESH_RATE_LIMIT`, `PUBLIC_ASPIRATION_RATE_LIMIT`, `PUBLIC_REGISTRATION_RATE_LIMIT` and `*_WINDOW` | API | Review against expected usage; limits are per process, so use an edge/shared limiter before multiple API instances |
| `DOCUMENT_STORAGE_PATH` | API | Absolute durable private path outside web root |
| `REGISTRATION_STORAGE_PATH` | API | Absolute durable private path outside web root |
| `MEDIA_STORAGE_PATH` | API | Absolute durable API-controlled path outside web/static roots |
| `DOCUMENT_MAX_UPLOAD_BYTES`, `REGISTRATION_IMAGE_MAX_UPLOAD_BYTES`, `MEDIA_MAX_UPLOAD_BYTES` | API | Reviewed bounds; preserve server-side validation |
| `NEXT_PUBLIC_API_BASE_URL` | Vercel build | HTTPS public API base ending `/api/v1`; public at build time, never a secret |
| `NEXT_PUBLIC_SITE_URL` | Vercel build | HTTPS canonical public origin; public at build time |
| `BEM_DEMO_CONTENT` | Vercel build/runtime | Must be unset/false in production |

Also inventory provider-managed TLS/DNS, database CA material if required, backup encryption keys, deployment tokens, and operator access in the secret manager. Track secret owner/rotation date, not secret values, in release records. Rotate/revoke all credentials through the provider if exposed.

## Rollback procedure

1. Declare rollback owner and capture release SHA, migration version, current health, and recovery-set ID. Pause high-risk writes if data consistency is uncertain.
2. Roll back Vercel to the last known-good immutable deployment and API to its matching compatible artifact; verify `/healthz`, CORS, login/refresh, and critical read paths.
3. Determine whether the new binary can read the current schema. Prefer code rollback only when schema changes are backward-compatible.
4. Do not automatically run `api:migrate:down`: downs may destroy or invalidate newly written data. If schema rollback is unavoidable, get database-owner approval, take another backup, assess each down migration, and rehearse on an isolated restore first. Prefer a forward-fix migration for production data.
5. If database or storage corruption/loss occurred, restore one matched DB+documents+registration-attachments+media recovery set to an isolated target first. Switch traffic only after integrity and security checks pass and the incident owner approves.
6. Record trigger, impact window, actions, data-loss estimate, release IDs, and follow-up. Reopen traffic only after smoke, auth/RBAC, storage, and monitoring checks pass.

## Go-live gate

- [ ] Approved provider, engine/version, ownership, region, DNS/TLS, RPO/RTO and retention are recorded.
- [ ] Production env checklist complete; config validation accepts production values.
- [ ] Restore drill for a complete recovery set passed and has operator sign-off.
- [ ] `pnpm api:integration` ran, not skipped; `pnpm verify` and `pnpm audit` passed on the release SHA.
- [ ] Manual UAT results and unresolved Phase 11 follow-ups reviewed by release owner.
- [ ] Rollback target, migration compatibility, health alarms, logs, and on-call operator are ready.
- [ ] No secret, local `.env`, test account password, private upload, DB dump, or build output is in Git/artifacts.

This repository does not perform a production deployment. A release owner must authorize and execute that separately.
