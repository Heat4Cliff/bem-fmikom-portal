# BEM FMIKOM UNUGHA — Campus Deployment Handoff v1.0.1

The frontend and backend are standalone packages that can be installed and deployed independently. This handoff contains no production deployment, production credentials, database, or runtime uploads.

## Packages

- `frontend/`: Next.js App Router, tests, local fonts, public assets, and current multi-theme/branding presentation.
- `backend/`: Go/Gin API, tests, embedded migrations, and configuration template.
- `bem-fmikom-frontend.zip` and `bem-fmikom-backend.zip`: release assets containing those respective directories.

The application settings manage public logo assets, cabinet identity, landing Hero media, and social links. Theme choice is persisted in the browser. The API/database contract includes migrations through 00013; migrations 00012 and 00013 add landing brand identity and presentation controls.

## Independent frontend

From the extracted `frontend/` directory, use Node.js 24.19.0 or newer and pnpm 11.25.0:

```powershell
pnpm install --frozen-lockfile
Copy-Item .env.example .env.local
# Set the public HTTPS API/site origins in .env.local before building.
pnpm lint
pnpm format
pnpm typecheck
pnpm test
pnpm build
pnpm start
```

`NEXT_PUBLIC_API_BASE_URL` and `NEXT_PUBLIC_SITE_URL` are public build-time settings. Never put secrets in `NEXT_PUBLIC_*`. Disable `BEM_DEMO_CONTENT` in production.

## Independent backend

From the extracted `backend/` directory, use the Go toolchain declared in `go.mod`:

```powershell
go mod download
go test ./...
go vet ./...
go build ./...
Copy-Item .env.example .env
# Replace placeholders with secrets from the hosting secret manager.
go run ./cmd/migrate up
go run ./cmd/server
```

Provision MySQL/MariaDB and a least-privilege account before migrations. Set a production `JWT_SECRET`, exact HTTPS `FRONTEND_ORIGINS`, and Secure HttpOnly refresh cookies. Keep private document and registration attachment storage outside public/static web roots. Public media storage must be durable and backed up together with its database metadata.

### Database upgrade for v1.0.1

This release includes migrations 00012–00013 for branding and social presentation fields. Before activating the new backend, the campus deployment operator must:

1. Create and verify a restorable backup of the production database.
2. Verify that `DATABASE_URL` points to the intended production database and uses the authorized application account.
3. From the extracted `backend/` directory, run the existing forward migration command: `go run ./cmd/migrate up`.
4. Confirm that the migration succeeds, then deploy or activate the new backend version.

Do not use `go run ./cmd/migrate down` as a routine deployment step. Plan any rollback separately against a verified backup.

## Initial SUPER_ADMIN Bootstrap

After the production database is available, `DATABASE_URL` is verified, a backup/recovery checkpoint is ready, and production migrations have succeeded, the campus operator must create the first `SUPER_ADMIN` with the backend bootstrap CLI. The CLI creates an active `SUPER_ADMIN` account with no division, normalizes and validates the email, hashes the password, and rejects an email that already exists. It does not seed an account automatically.

From the extracted `backend/` directory, run:

```powershell
go run ./cmd/create-admin -name "Super Admin BEM FMIKOM" -email "<production-admin-email>"
```

The command requires an interactive terminal and prompts for `Password (minimum 12 characters)` and `Confirm password` without echoing either value. Passwords must be at least 12 characters and no more than 72 bytes. Do not pass a password as a command argument. Choose a unique production-only password, preferably 16–24 characters or longer, and store it in the campus operator's password manager.

Recommended deployment order:

1. Provision the production database.
2. Verify `DATABASE_URL` points to the intended production database and least-privilege account.
3. Establish a verified backup/recovery checkpoint.
4. Run `go run ./cmd/migrate up` from `backend/` and verify success. This release includes migrations through 00013; migrations 00012–00013 add branding and social presentation fields.
5. Create the initial `SUPER_ADMIN` with the CLI above.
6. Start or activate the backend.
7. Deploy or activate the frontend.
8. Sign in at `/admin/login` and perform the production smoke test.

Do not use `go run ./cmd/migrate down` as a routine deployment or rollback procedure. Plan rollback separately against a verified recovery point.

The first production `SUPER_ADMIN` should be created directly by the campus operator or IT team so the project developer does not know the production password. Never reuse UAT, development, or testing accounts or passwords in production. Credentials must not be placed in source code, `.env.example`, README examples, Git, GitHub Release notes, command-line arguments, shell history, logs, chat, or browser/network output. The CLI has no production default account or password. After first login, authorized administrators may create additional accounts through **Manajemen Pengguna**, subject to application RBAC and division rules.

First login verification checklist (operator must record actual results; this handoff does not claim production UAT passed):

- [ ] Initial `SUPER_ADMIN` account was created.
- [ ] Login at `/admin/login` succeeded.
- [ ] Reload preserved the session according to the auth contract.
- [ ] `/api/v1/auth/me` reports role `SUPER_ADMIN`.
- [ ] Admin navigation matches the role.
- [ ] **Manajemen Pengguna** opens.
- [ ] Logout succeeded.
- [ ] Credentials do not appear in logs, browser console, or network responses.
- [ ] No UAT credential was used in production.

## Operator responsibilities

The campus deployment operator supplies DNS, TLS, hosting providers and region, persistent storage, database, secret management, backup/restore, monitoring, rollback ownership, and post-deployment UAT. No production credentials are included; `.env.example` files contain placeholders only. This release is a handoff, not a production deployment.
