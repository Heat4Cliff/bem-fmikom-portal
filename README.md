# BEM FMIKOM UNUGHA Portal

Monorepo for the BEM FMIKOM UNUGHA portal. Phase 0 source audit and Phase 1 foundation are complete. Phase 2 adds the database foundation, admin authentication, and RBAC/division-scope primitives. Product modules remain for later phases.

## Source of truth

- `docs/PRD_SUMMARY.md` defines behavior and scope.
- `docs/backend/API_CONTRACT.md` and `docs/backend/SECURITY_CONTRACT.md` define request/response and authorization contracts. Report conflicts rather than silently choosing a source.
- `design_reference/` contains read-only Stitch evidence. `screen.png` is the visual target; `DESIGN.md` defines design tokens; `code.html` is a layout/style reference only.
- `docs/codex/` contains the audit, decisions, gaps, implementation matrix, and phase reports.
- Do not edit `source_docs/` or `design_reference/` as part of implementation.

## Architecture

```text
apps/web/                 Next.js App Router, strict TypeScript, Tailwind
apps/api/                 Go/Gin REST API, services, repositories, auth
apps/api/internal/database/migrate/  Embedded, versioned SQL migrations
packages/contracts/       Reserved for resolved OpenAPI/shared contracts
infra/                    Local MariaDB Compose and initialization SQL
scripts/                  Cross-platform repository verification helpers
docs/codex/                Audit, implementation matrix, and phase reports
source_docs/               Read-only copies of source documents
design_reference/          Read-only Stitch screenshots and references
```

## Prerequisites

- Node.js 24.19.0, pnpm 11.25.0
- Go 1.26.1
- Docker Desktop with Docker Compose
- MariaDB 11.4 for local development and database integration tests

The Phase 2 verification used the project-pinned Node.js 24.19.0 runtime. `.node-version` and CI remain pinned to Node 24.19.0.

## Local setup

Copy the template and replace its local placeholders. The example credentials are for local development only; generate a unique random `JWT_SECRET` with at least 32 bytes before starting the API.

```powershell
Copy-Item .env.example .env
```

The root `.env` supplies Docker Compose and Go commands. Go commands also load `apps/api/.env` if present. Existing process environment variables take precedence. Never commit local `.env` files or production credentials.

Start MariaDB and apply the foundation migration:

```powershell
pnpm db:up
pnpm api:migrate:up
```

Create the first `SUPER_ADMIN`. The password is entered twice without terminal echo and must be at least 12 characters and no more than 72 bytes.

```powershell
pnpm api:create-admin -name "Admin BEM" -email admin@bem.unugha.ac.id
```

Run the API and web app in separate terminals:

```powershell
pnpm api:dev
pnpm dev
```

The public shell is at `http://localhost:3000`, admin login at `http://localhost:3000/admin/login`, API at `http://localhost:8080`, and operational health at `http://localhost:8080/healthz`.

## Database and migrations

The local Compose service binds MariaDB to `127.0.0.1:3306`, persists data in a named Docker volume, and creates the dedicated `bem_fmikom_test` database for integration tests. `pnpm db:down` stops the service but keeps its volume. Migrations are embedded and versioned with Goose.

```powershell
pnpm db:up             # Start MariaDB
pnpm db:down           # Stop MariaDB; keep local data
pnpm api:migrate:up    # Apply pending migrations
pnpm api:migrate:down  # Revert the latest migration (destructive to its tables)
pnpm api:integration   # Reset and exercise TEST_DATABASE_URL; database must end in _test
```

`TEST_DATABASE_URL` must point to a dedicated database whose name ends with `_test`. The integration suite refuses other names, runs migrations from an empty schema, verifies schema/auth/session behavior, and drops the foundation migration again during cleanup. Never point it at a development or production database.

## Authentication and authorization

- Only `/api/v1/auth/login`, `/refresh`, `/logout`, and `/me` are implemented for product authentication.
- Access JWTs default to 15 minutes and remain in frontend memory. The browser does not persist access tokens to web storage.
- Opaque refresh tokens contain 256 bits of random entropy. Only SHA-256 hashes are stored in MariaDB; rotation reuses the session family and detects replay.
- Refresh cookie is HttpOnly, uses `/api/v1/auth` path, defaults to `SameSite=Lax`, and defaults to `Secure=false` in development and `Secure=true` in production. `SameSite=None` requires Secure.
- `FRONTEND_ORIGINS` is a comma-separated explicit origin allow-list. Credentialed CORS never uses `*`, and state-changing auth requests require a trusted Origin.
- Each protected API request verifies the JWT, reloads the active user and current role/division from the database, then applies role and scope policy. JWT role/division claims are not authorization authority.
- Role summary: `SUPER_ADMIN` has full access and no division; `BPH` has operational cross-division access without user/activity-log powers; `KEPALA_BIDANG` must belong to one division and is limited to its scoped content/operations. Backend route authorization remains mandatory even when navigation is hidden.
- Login is rate-limited per IP by a configurable single-instance in-memory limiter. Distributed limiting is a later hardening concern.

Login and refresh return the access token in JSON and set the refresh token only as a cookie. Logout revokes the related refresh-session family when present, clears the cookie, and the frontend clears in-memory auth state.

## Verification

```powershell
pnpm lint
pnpm format
pnpm typecheck
pnpm test
pnpm build
pnpm api:fmt
pnpm api:vet
pnpm api:test
pnpm api:build
pnpm api:integration
pnpm verify
```

CI runs the aggregate static/unit checks and a MariaDB 11.4 service for actual database integration tests. It uses test-only credentials and no production secrets.

## Phase status

See `docs/codex/PHASE_2_REPORT.md` for implementation scope and verification evidence. Phase 2 is limited to schema foundation, authentication, RBAC, and division-scope primitives. Phase 3 and all product feature modules have not started.
