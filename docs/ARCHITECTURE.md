# Architecture Baseline

## System topology
```text
Browser
  |
  | HTTPS
  v
Next.js Web (Vercel)
  |
  | REST /api/v1
  v
Go API (Gin)
  |---- MySQL/MariaDB
  |---- Private document storage
  `---- Google Photos external URLs (gallery only)
```

## Monorepo rationale
Keep frontend/backend/contracts/docs in one repository for Codex consistency and atomic contract changes while retaining decoupled deployments.

```text
apps/web            Next.js public + admin UI
apps/api            Go REST API
packages/contracts  OpenAPI + generated types/client (optional generation)
design_reference    Stitch read-only evidence
docs                architecture/workflows/reports
scripts             dev/test/seed/verification
infra               deploy templates, reverse-proxy notes, DB backup scripts
```

## Frontend architecture
- Next.js App Router.
- Public and Admin route groups.
- Server Components by default; Client Components only for interactive widgets/forms/editors.
- Typed REST client with centralized auth refresh/error handling.
- Design tokens mapped from Stitch, not default library theme.
- Route guards are UX convenience only; backend authorizes every protected operation.

Suggested layering:
```text
apps/web/src/
  app/
  components/
    ui/
    public/
    admin/
  features/
  lib/
    api/
    auth/
    validation/
  types/
```

## Backend architecture
```text
apps/api/
  cmd/server/
  internal/
    config/
    http/
      middleware/
      handler/
    service/
    repository/
    domain/
    auth/
    storage/
    audit/
  migrations/
```

Request flow:
`router -> auth/rate-limit middleware -> handler -> service/policy -> repository -> DB/storage`.

## Cross-cutting
- Structured errors and request IDs.
- Central RBAC/division policy helpers.
- Pagination/search/filter conventions.
- Activity log on sensitive actions.
- Time handling stored UTC, formatted to local context at presentation boundary.
- Database indexes match query patterns from PRD.
