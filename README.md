# BEM FMIKOM UNUGHA Portal

Repository foundation for the BEM FMIKOM UNUGHA portal. This repository currently contains the Phase 0 audit and Phase 1 technical foundation. Business features, database migrations, authentication, authorization, and contract APIs have not started.

## Source of truth

- `docs/PRD_SUMMARY.md` defines behavior and scope.
- `docs/backend/API_CONTRACT.md` and `docs/backend/SECURITY_CONTRACT.md` define API and authorization contracts.
- `design_reference/` contains the read-only Stitch material. `screen.png` is the visual target; `DESIGN.md` defines design tokens; `code.html` is a layout/style reference only and is not production architecture.
- `docs/codex/` contains the audit, decisions, gaps, implementation matrix, and phase reports.

Do not edit `source_docs/` or `design_reference/` as part of implementation.

## Repository architecture

```text
apps/web/                 Next.js public and admin shells, reusable UI, tokens
apps/api/                 Gin server foundation, typed config, HTTP middleware
packages/contracts/       Reserved for resolved OpenAPI/shared contracts
scripts/                  Cross-platform repository verification helpers
infra/                    Reserved for later local infrastructure decisions
docs/codex/                Audit, implementation matrix, gaps, and phase reports
source_docs/              Read-only copies of source documents
design_reference/         Read-only Stitch screenshots and references
```

## Stack and prerequisites

- Node.js 24.19.0 and pnpm 11.25.0
- Go 1.26.1
- Next.js App Router, TypeScript strict, Tailwind CSS v4
- Go, Gin, REST foundation
- MySQL/MariaDB is a later phase; no database is required to run the Phase 1 health server.

Install dependencies and copy environment templates as needed:

```powershell
pnpm install
Copy-Item apps/web/.env.example apps/web/.env.local
Copy-Item apps/api/.env.example apps/api/.env
```

The templates contain placeholders only. Do not commit local `.env` files or real secrets.

## Run locally

```powershell
pnpm dev
```

The public shell runs at `http://localhost:3000`. To run the API foundation separately:

```powershell
pnpm api:dev
```

The API listens on port 8080 by default and exposes only `GET /healthz` at this stage.

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
pnpm verify
```

CI runs the same aggregate verification command after a frozen pnpm install.

## Phase status

See `docs/codex/PHASE_1_REPORT.md` for the verified scope and any manual visual-review limitations. Follow `PLANS.md` one phase at a time. Phase 2 has not begun.
