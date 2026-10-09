# Phase 5 — Organization & Period Management

**Status: PASS WITH MANUAL VERIFICATION REQUIRED** — Phase 5 organization and period behavior is implemented and automated/database verification passed. Exact mobile-device viewport visual verification remains required. Phase 6 has not started.

## Scope delivered

- Reversible migration `00003_organization.sql` adds division display order and period-scoped organization members with foreign keys, checks, indexes, and an active-head unique constraint.
- Period list/create/update/activate/archive; activation transactionally archives the old active period and preserves historical data.
- Division list/create/update/delete with stable display order and safe conflict on referenced division deletion.
- Member list/create/update/soft-deactivate, optional NIM and URL-backed photo, six PRD positions, filters, division/period validation, and sort order.
- Admin screens: `/admin/periode`, `/admin/divisi`, `/admin/pengurus`, member create/edit routes, and `/admin/struktur`.
- Public endpoints for active period, divisions, and minimal active-period organization hierarchy; public landing `#struktur` consumes that hierarchy with a no-data state.
- Activity rows for organization mutations. SUPER_ADMIN writes periods; SUPER_ADMIN/BPH manage divisions and members; KEPALA_BIDANG reads only its database-assigned division and cannot mutate organization data.
- Admin pages provide responsive table/card treatments. Exact 360/390/412 viewport emulation could not be set through the available in-app browser interface; no mobile pixel-parity claim is made.

## Source and contract audit

- Re-read the active phase instructions, PRD v1.4, API/Security Contracts, data model notes, prior reports, decisions/gaps/matrix, route matrix, workflows, Stitch screen index, and visual checklist before implementation.
- The source inventory remains 37 Stitch `screen.png`/`code.html` pairs, with 8 explicitly labeled mobile references. Phase 5 visual target references were opened for periods, divisions, member management/form, organization structure, and the public mobile landing pattern.
- `screen.png` is the visual target. `code.html` informed structure/style only and was not copied into production. Production remains reusable Next.js App Router, TypeScript, and Tailwind components.
- PRD v1.4 fixes role behavior more precisely than the earlier organization route summary: SUPER_ADMIN alone mutates periods; SUPER_ADMIN/BPH manage divisions/members; KEPALA_BIDANG has read-only, current-division scope. This policy is now recorded in the API Contract and enforced in API middleware/service behavior.
- The public API summary did not fully specify active-period empty-state payloads. The implementation fixes these responses explicitly: active-period endpoint returns 404 when none exists; organization returns `period: null` and empty lists; public divisions returns an empty list.
- Some Stitch member form extras conflict with the phase schema: this implementation keeps photo URL and NIM optional, as specified for Phase 5. It does not add file upload, academic program/year, contact/social fields, print, or export behavior based solely on reference-only controls.
- There is no separate public structure route in the route matrix; public hierarchy appears in landing section `#struktur`. The admin route is `/admin/struktur`.

## Fixed decisions

1. Preserve the foundation/content migrations; add Phase 5 only in migration 00003.
2. Periods are archived, never deleted. Only one can be ACTIVE, including under concurrent activation.
3. A division cannot be deleted while referenced. No cascading historical deletion.
4. Member DELETE changes status to INACTIVE and retains the record/history.
5. KEPALA_BIDANG and ANGGOTA require a division from the selected period; other leadership positions may have no division.
6. Database principal division is the authority. KEPALA_BIDANG cross-division detail is hidden as 404; write attempts are denied.
7. Public hierarchy excludes NIM, status, ordering metadata, and audit fields. Active divisions/members are sorted deterministically.
8. At most one ACTIVE KEPALA_BIDANG exists in one division/period, enforced both by validation and MariaDB unique constraint.

## Verification evidence

| Gate | Result | Evidence |
|---|---|---|
| Docker preflight | PASS | Docker 29.8.2; Compose 5.5.1; Docker Desktop Engine 29.8.2, Linux x86_64. |
| Existing database configuration | PASS | Used only the repository's `infra/docker-compose.dev.yml` through `pnpm db:up`; `infra-mariadb-1` was healthy. MariaDB 11.4.13. No new Compose configuration was created. |
| Migration and real integration | PASS | `pnpm api:integration` ran `TestDatabaseIntegration` against `bem_fmikom_test`, not skipped. Goose applied migrations 00001–00003 from schema version 0, ran migration 00003 DOWN/UP, and test cleanup returned the schema to version 0. |
| Organization database/API flow | PASS | Integration created/activated periods and divisions, created members, rejected invalid date/period/division and missing-head division data, rejected a second active head, enforced division delete dependency, verified public minimal DTO/order/inactive exclusion, soft-deactivated a member, switched public organization between periods, retained historical member data, exercised no-active-period response, and raced two activations. |
| Role/division security | PASS | Existing auth integration role matrix passed: SUPER_ADMIN cross-division, BPH cross-division, KEPALA_BIDANG own division, cross-division denials, and KB without division. Organization integration confirmed KB cannot read out-of-scope details or mutate members, can read only its own division, BPH can manage organization records, and BPH cannot mutate periods. |
| Frontend automated tests | PASS | 5 files; 22 tests passed. |
| Full verification | PASS | `pnpm verify`: Biome lint and format, TypeScript typecheck, 22 frontend tests, Next.js 16.4.0 webpack production build, Go format (41 files), `go vet`, Go tests, and Go build all passed. The first sandbox build attempt was blocked by EPERM creating `.next/server`; the authorized local rerun passed. |
| Browser public UAT | PASS | Public home rendered with no organization before test data; after period/divisions/members were created, it rendered the active period, leadership, division, head, and member from the API. Screenshot/AX inspection showed the public `#struktur` hierarchy and empty behavior. |
| Browser admin UAT | PASS | SUPER_ADMIN logged in, created/activated a period, created two divisions, created a leader/head/member, and saw the admin hierarchy update. Reload preserved the authenticated session; logout returned to login; direct navigation to a protected period route after logout remained at login. |
| Browser role UI UAT | PARTIAL | SUPER_ADMIN admin UI was manually exercised. BPH/KEPALA_BIDANG organization authorization and scope were exercised against real MariaDB through integration, but separate role logins were not manually exercised in the browser. |
| Responsive visual review | PARTIAL | Admin and public screens were visually inspected in the available browser. The current CUA in-app tab API exposes no exact viewport override; 360/390/412 visual UAT and mobile screenshot comparison were not claimed. Responsive classes and mobile card variants were reviewed in source, and public mobile Stitch evidence was reviewed. |
| Immutable source evidence | PASS | `git diff --quiet -- source_docs design_reference` returned clean. Source and design reference trees were not modified. |
| Phase boundary | PASS | No Phase 6 work or checklist changes were made. |

The temporary browser UAT records and account were created only in the dedicated `bem_fmikom_test` schema. The final integration rerun cleans this schema back to migration version 0.

## Gaps and risks

- Exact-size mobile browser visual review (360/390/412) remains an explicit manual follow-up because the available browser-control surface does not permit viewport emulation. Responsive code paths are present, but this report does not upgrade code inspection to mobile UAT.
- BPH and KEPALA_BIDANG role-specific organization visibility was verified by real-database API integration; separate manual browser sessions for those two roles remain unverified.
- The source reference has extra member form fields/actions that are outside the approved Phase 5 schema/scope. They remain documented and unimplemented; confirm any future additions against PRD and contract first.
- Organization activity events are persisted, but the activity-log dashboard belongs to a later phase.
- No upload/storage service, approval workflow, or public detail route for organization hierarchy was introduced.

## Recommendation for the next authorized phase

Phase 6 is not started. Before release-wide visual sign-off, run mobile browser UAT at 360/390/412 and manual browser sessions for BPH and KEPALA_BIDANG. Proceed to Phase 6 only after a separate user instruction.

**PHASE 5: PASS WITH MANUAL VERIFICATION REQUIRED**
