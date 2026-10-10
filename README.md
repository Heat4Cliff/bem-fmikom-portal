# BEM FMIKOM UNUGHA Portal

Monorepo for the BEM FMIKOM UNUGHA portal. Phases 0–6 are implemented. Phase 7 adds public dynamic registrations, an admin form builder, submissions, and privacy-conscious XLSX export. See `docs/codex/PHASE_7_REPORT.md` for scope and current evidence.

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

The public landing page reads no content API yet. Dynamic teaser sections show an empty state by default in production. For local visual review only, set `BEM_DEMO_CONTENT=true` in the web process; all fixture content and sample statistics are visibly marked as examples. Do not enable this flag in a production runtime. The brand currently uses a text mark because no approved logo asset or public contact/social details were found in the source package. Set `NEXT_PUBLIC_SITE_URL` to the approved public origin before relying on absolute canonical, sitemap, or robots sitemap URLs.

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

See `docs/codex/PHASE_2_REPORT.md` through `docs/codex/PHASE_6_REPORT.md` for phase scope and verification evidence. The landing page reads public news, announcements, gallery, and active-period organization data. `/aspirasi` accepts identified or anonymous submissions; `/admin/aspirasi` provides division-scoped review and `/admin/aspirasi/[id]` shows detail and status history.

## Content CMS (Phase 4)

Phase 4 adds:

- Public pages: `/berita`, `/berita/[slug]`, `/pengumuman`, `/pengumuman/[slug]`, `/galeri`, and `/galeri/[slug]`.
- Admin screens: `/admin/berita`, `/admin/berita/new`, `/admin/berita/[id]/edit`, `/admin/kategori-berita`, `/admin/galeri`, and `/admin/pengumuman`.
- API-backed article/category, gallery category/item, and announcement CRUD/publication workflows.
- Reversible migration `00002_content_cms.sql`; `pnpm api:integration` resets only the dedicated database named by `TEST_DATABASE_URL` and ending in `_test`.
- Rich text sanitized on the server; image/preview fields are HTTPS URLs and album URLs point to Google Photos. File uploads and private media storage are not part of this phase.
- Role policy: SUPER_ADMIN/BPH manage shared content across divisions; KEPALA_BIDANG content requests are scoped to the current database division. Article taxonomy and announcements are read-only for KEPALA_BIDANG.
- Public CMS Server Components use `cache: "no-store"` so publication and schedule state is read from the API for each request.

Public announcement archive visibility for scheduled/expired entries still needs a product decision; current list/detail behavior and its impact are recorded in `docs/codex/GAPS_AND_QUESTIONS.md`.

## Organization and periods (Phase 5)

- Admin routes: `/admin/periode`, `/admin/divisi`, `/admin/pengurus`, member create/edit routes, and `/admin/struktur`.
- Public home organization section reads `/api/v1/public/organization`; active-period metadata and division listings also have public API endpoints.
- Reversible migration `00003_organization.sql` adds division ordering and period-scoped organization members. Period transitions preserve archived history; member deactivation is soft (`INACTIVE`).
- SUPER_ADMIN manages periods; SUPER_ADMIN/BPH manage divisions and members; KEPALA_BIDANG reads only its current division and cannot mutate organization records. The API enforces this policy.
- Phase 5 verification: `pnpm verify` and `pnpm api:integration`; see `docs/codex/PHASE_5_REPORT.md` for exact evidence, including browser UAT and viewport limitations.

## Aspiration management (Phase 6)

- Public submission uses `POST /api/v1/public/aspirations`; destination must be an ACTIVE division in the ACTIVE period. The form loads choices from `GET /api/v1/public/divisions`.
- Anonymous submissions persist `name`, `nim`, and `email` as SQL `NULL`. Identified name/NIM are optional; optional email is validated when supplied. Subject/message are plain text (200/5,000 character limits).
- Reversible migration `00004_aspirations.sql` adds the domain. Status progresses only `PENDING → READ → RESOLVED`; updates are transactional and record first `read_at`/`resolved_at` timestamps in UTC.
- SUPER_ADMIN and BPH review across divisions; KEPALA_BIDANG is scoped by its current database division. Scoped detail reads return 404; cross-division status writes return 403.
- Public submissions use a separate configurable in-memory IP limiter (`PUBLIC_ASPIRATION_RATE_LIMIT=5`, `PUBLIC_ASPIRATION_RATE_WINDOW=15m`). Multi-instance deployments need a shared limiter.
- Status activity rows store only aspiration ID, old/new status, and division ID. Request logs contain route/status/timing only. Attachment, taxonomy, disposition notes, SLA, export, and case workflow controls shown in the Stitch mock are outside the Phase 6 PRD/API scope.

## Dynamic registration management (Phase 7)

- Public routes: `/pendaftaran`, `/pendaftaran/[slug]`, and `/pendaftaran/[slug]/form`. Admin routes: `/admin/pendaftaran`, `/admin/pendaftaran/[id]/builder`, `/admin/pendaftaran/[id]/submissions`, and the submission detail route.
- The API supports DRAFT → OPEN → CLOSED forms, public UPCOMING/OPEN/CLOSED status, and eight field types: TEXT, TEXTAREA, EMAIL, NUMBER, SELECT, RADIO, CHECKBOX, and DATE. Forms start with the required core identity fields; a schema becomes immutable on its first submission while historical answers retain metadata snapshots.
- NIM is normalized and unique within one form. MariaDB enforces the duplicate rule, including concurrent submissions. Registration writes have a separate configurable per-process rate limit (`PUBLIC_REGISTRATION_RATE_LIMIT=5`, `PUBLIC_REGISTRATION_RATE_WINDOW=30m`).
- SUPER_ADMIN and BPH can manage forms across divisions; KEPALA_BIDANG is limited to its current database-assigned division. Admin submission reads and exports require the same form scope. XLSX export is capped at 1,000 rows, uses literal text cells for applicant content, and returns `private, no-store`.
- Migration `00005_registration.sql` is reversible. Run `pnpm api:integration` against `TEST_DATABASE_URL` ending in `_test` for real MariaDB coverage. For feature decisions, limitations, and browser/visual verification evidence, see `docs/codex/PHASE_7_REPORT.md`.

## Private document archive (Phase 8)

- Admin archive: `/admin/arsip-dokumen`; API list/upload/detail/metadata/delete is under `/api/v1/admin/documents`, with authorized streaming at `/api/v1/admin/documents/{id}/download`.
- Files are stored in a private local directory configured by `DOCUMENT_STORAGE_PATH`; configure an absolute path outside the web root in production. POSIX storage objects use private `0600` permissions; Windows inherits the directory ACL. The default upload maximum is 10 MiB and can be set with `DOCUMENT_MAX_UPLOAD_BYTES`.
- Allowed formats are PDF, DOC, DOCX, XLS, and XLSX. Upload checks extension, declared MIME type, and file signature/container. SHA-256 and opaque storage keys are persisted; no public/static path is exposed.
- SUPER_ADMIN/BPH can access cross-division archives. KEPALA_BIDANG is restricted by the current database division and period. Back up database metadata and private storage together to retain archive consistency.
- File validation is not antivirus scanning or content disarm. Production should pair the private storage backup with the database backup and verify restore as one recovery set.
- Run `pnpm db:up`, ensure `TEST_DATABASE_URL` names a dedicated database ending `_test`, then use `pnpm api:integration` and `pnpm verify`. See `docs/codex/PHASE_8_REPORT.md` for executed evidence and remaining manual review.
