# Campus Deployment Handoff v1.0.1 — preparation evidence

Date: 2026-10-16. The owner accepted the latest branding, theme, Hero media, and About/Tentang presentation before this release preparation. No production deployment was performed.

## Source and security audit

- Reviewed the accepted working tree. The deliverable source includes the multi-theme system, public/admin branding, landing settings, cabinet identity, social links, Hero documentation media, and `/tentang` layout. Temporary browser scripts, screenshots, build output, local uploads, environment files, database files, dependency directories, and generated ZIPs are excluded from Git history.
- `source_docs/` and `design_reference/` are unchanged. `git diff --check` passed.
- Scans of candidate source and extracted ZIP contents found no high-confidence private keys, API/JWT tokens, credential-bearing DSNs, local absolute machine paths, or actual `.env` files. Each package contains only its safe `.env.example` template.
- A historical UAT report was edited to remove a local absolute download path. The UAT content itself was not changed.

## Automated regression

- `pnpm api:integration`: PASS against dedicated local MariaDB `*_test` database. `TestDatabaseIntegration` ran rather than skipped, applied migrations 00001–00013, exercised migration reversibility, auth and RBAC/division cases, branding settings, and media upload behavior.
- `pnpm verify`: PASS. Biome lint and format checked 142 source files; strict TypeScript passed; frontend suite passed 27 files/125 tests; Next.js production build passed; gofmt checked 71 files; Go vet, tests, and build passed.
- `pnpm audit`: PASS, no known vulnerabilities at audit time.

## Browser smoke and visual classification

- Chrome headless checked `/`, `/tentang`, and `/admin/login` at 360×800, 390×844, 412×915, 768×900, 1280×800, and 1440×900. No horizontal overflow or visible clipped actions were detected. The cabinet logo loaded and remained horizontally centered at all sizes.
- Chrome selected Matcha from the theme switcher and confirmed that the selection remained after reload.
- Admin header, sidebar, and mobile drawer were checked at the same six sizes with **simulated auth responses** solely for layout review. No overflow or visible clipped actions were detected. This is visual smoke evidence, not a real login or RBAC UAT result. The owner acceptance and integration tests remain the evidence for the respective acceptance/security behavior.

## Standalone archive validation

Both archives were regenerated from the current source, extracted into clean directories, and every extracted source file matched its package source by SHA-256. The extracted archive trees contained no `node_modules`, `.next`, `tsbuildinfo`, actual `.env`, runtime storage, database dump, or private upload.

| Archive | Files | Bytes | SHA-256 | Independent checks |
|---|---:|---:|---|---|
| `bem-fmikom-frontend.zip` | 156 | 337738 | `505FCA8FEBAED8620C8C4C0B3D06ED30994EFAAA671606C38F2CEAC5C017FE53` | Frozen install, format, lint, typecheck, 125 tests, production build: PASS |
| `bem-fmikom-backend.zip` | 90 | 187027 | `9AE3FFACE37DAF2F0EB1442CF0F0442284496CA4E9B33B9561F705D88119678D` | Go mod download, test, vet, build: PASS |

The frontend export adjusts its font paths and standalone Biome configuration; the backend export adjusts its dotenv loader to read only its own `.env`. These export-only adaptations do not alter the monorepo application source.

## Operational boundary

Frontend and backend remain independently deployable. Public assets are controlled by application settings. Database migrations 00012–00013 add branding and social presentation fields; do not assume the database contract is unchanged from v1.0.0. Production database, secrets, DNS/TLS, durable storage, backups, monitoring, rollback, and post-deployment UAT belong to the campus deployment operator. No production credential or provider configuration is included.
