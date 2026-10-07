# Phase 0 — Decisions

## Fixed by user or authoritative source

| Area | Decision | Authority |
|---|---|---|
| Repository | Separate project directory named bem-fmikom-portal; keep preparation source separate. | User request |
| Source evidence | Treat source_docs/ and design_reference/ as read-only; copied unchanged into the new repo for a self-contained project. | User request |
| Scope order | PRD v1.4 controls behavior and scope; API/Security Contract controls wire/auth behavior subject to v1.4 reconciliation. | User request and v1.4 contract rule |
| Visual authority | Stitch screen.png is the target; code.html only informs layout/style; Stitch HTML is not production architecture. | User request and AGENTS.md |
| Design tokens | Use Stitch modern_academic_horizon/DESIGN.md as the token reference; screen images win visual conflicts. | User request and source design note |
| Frontend | Next.js App Router, TypeScript strict, Tailwind CSS, reusable component implementation. | User request / AGENTS.md |
| Backend | Go, Gin, REST API. | User request / AGENTS.md |
| Database | MySQL or MariaDB with versioned migrations in later phase. | User request / AGENTS.md |
| Authentication | JWT access token plus Secure HttpOnly refresh cookie on HTTPS. | User request / PRD v1.4 |
| Authorization | Backend is final authority; role and division scope must be enforced server-side. | AGENTS.md, Security Contract, PRD v1.4 |
| Period invariant | Exactly one active period; transitions preserve history and are atomic. | PRD v1.4 |
| Documents | Private/non-public storage and authorized streaming; no public direct file path. | PRD v1.4 / Security Contract |
| Phase boundary | Phase 0 only; no app scaffold, feature, migration, or Phase 1 checklist changes in this run. | User request |

## Audit resolutions (not product scope changes)

- Where summary endpoint inventory differs from the full v1.4 endpoint table, the full v1.4 contract is the authority; additions still need explicit reconciliation before implementation.
- The 37 Stitch files are evidence, not a requirement to reproduce raw HTML. Implement screens as reusable Next.js/TypeScript/Tailwind components in a later authorized phase.
- Explicitly mobile-named references are evidence for 8 screens. Other routes remain responsive requirements; do not imply that missing mobile captures exist.

## Not decided here

See GAPS_AND_QUESTIONS.md. Phase 0 does not choose new endpoint behavior, role access, analytics provider, retention policy, rate limit thresholds, or unprovided visual design.

## Phase 1 — Foundation choices

| Area | Decision | Reason / boundary |
|---|---|---|
| Workspace | pnpm 11.25.0 workspace with `apps/*` and `packages/*`; exact direct dependency versions are recorded in manifests and `pnpm-lock.yaml`. | Frozen lockfile install passed. A later offline supply-chain metadata check lacked a cached `@babel/helpers` record, so the final install check used the configured registry. |
| Frontend | Next.js 16.4.0 App Router, React 19.2.8, TypeScript 6.0.3 strict, Tailwind CSS 4.3.3, Biome 2.5.15 for lint/format. | ESLint 9 is out of support and ESLint 10 is outside the peer ranges of the Next React/import/a11y plugins. Biome keeps lint and format on a supported, peer-clean toolchain. TypeScript 7 was rejected after typescript-eslint reported its compiler API was unsupported. |
| Fonts | Inter and Plus Jakarta Sans are bundled from Fontsource 5.3.0 and loaded with `next/font/local`. | Keeps the specified typefaces while allowing offline builds without fetching Google Fonts. |
| Build bundler | Production build script uses `next build --webpack`; `next dev` keeps the Next.js default. | Turbopack's Windows worker failed to spawn under the restricted local build environment; Webpack production build passed. |
| Backend | Go 1.26.1 module with Gin 1.12.0, `slog` JSON logging, request ID, recovery, common JSON envelope, JSON 404, and operational `GET /healthz`. | Operational foundation only; `/healthz` is not a product API contract route. |
| Runtime pin | `.node-version` is 24.19.0 and `.go-version` is 1.26.1. | CI installs the declared versions. This workstation executed Node commands on Node 25.2.1; record the mismatch when interpreting local verification. |

These choices do not resolve Phase 0 product/API/security gaps or introduce authentication, authorization, database schema, migrations, or business endpoints.
