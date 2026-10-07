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
