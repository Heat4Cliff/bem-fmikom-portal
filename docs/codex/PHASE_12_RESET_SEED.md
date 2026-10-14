# Phase 12 — Isolated UAT Reset and Seed Procedure

## Purpose and safety boundary

This procedure prepares a repeatable, disposable local UAT environment. It never drops, truncates, or resets the shared development database. It does not create demo accounts or passwords in source control. The repository currently has no general application fixture seeder: `api:create-admin` bootstraps only the first `SUPER_ADMIN`; the remaining role accounts and domain fixtures are created through the authenticated UI/API.

Use a newly provisioned empty MariaDB schema named `bem_fmikom_uat` and a dedicated local database principal with access only to that schema. Provisioning requires a local DBA/operator. Do not point these steps at a production host or at the existing `bem_fmikom_dev` or `bem_fmikom_test` schema. The root Compose file remains the sole local MariaDB configuration.

## Prepare a fresh isolated schema

1. Start the existing local MariaDB service with `pnpm db:up`; wait for its healthcheck and confirm the intended container before proceeding.
2. Using a local database administrator through an interactive SQL client, create an empty `bem_fmikom_uat` schema (`utf8mb4`, `utf8mb4_unicode_ci`) and a dedicated `bem_uat` user. Grant only the privileges needed to run the repository's migrations and the API's normal CRUD against this schema. Do not put a password in a command argument or this document.
3. Confirm that the schema contains no tables and that the `bem_uat` principal cannot access other schemas.
4. In a PowerShell session, set `DATABASE_URL` from a secure local secret prompt or a local ignored environment file; preserve Go MySQL DSN syntax and `parseTime=true`. Keep this shell scoped to UAT and clear the variable at the end.
5. Apply the checked-in schema and bootstrap the first admin:

```powershell
pnpm api:migrate:up
pnpm api:create-admin -name "UAT SUPER_ADMIN" -email uat-super-admin@example.test
```

The bootstrap CLI prompts twice without echo. Use a fresh, randomly generated password meeting the 12-character minimum; share it through an approved password manager, never chat, terminal history, a fixture file, or a report.

## Repeatable fixture recipe

After the first SUPER_ADMIN login, create these records in order through the current admin screens. Use the exact UAT markers below so each fresh schema gets the same logical fixture set. Use clearly synthetic `example.test` email addresses and non-real names/NIMs only.

1. Create active period `UAT-PHASE12-CURRENT` and draft period `UAT-PHASE12-NEXT`; keep exactly the current period active.
2. Create active divisions `UAT-DIV-A` and `UAT-DIV-B` in the current period. Set their public-structure flags explicitly.
3. Keep the bootstrapped `uat-super-admin@example.test`; create active `BPH` account `uat-bph@example.test` without division, `KEPALA_BIDANG` account `uat-kb-a@example.test` assigned to DIV-A, and `KEPALA_BIDANG` account `uat-kb-b@example.test` assigned to DIV-B. Use independent random passwords and deliver them through a password manager.
4. Add synthetic heads `UAT HEAD DIV A` / `UAT HEAD DIV B` and members `UAT MEMBER DIV A` / `UAT MEMBER DIV B`, in that display order where supported.
5. Add article titles `UAT PHASE12 PUBLISHED` (published) and `UAT PHASE12 DRAFT` (draft); announcements `UAT PHASE12 ACTIVE` and `UAT PHASE12 INACTIVE`; gallery category/item `UAT PHASE12 GALLERY`; open registration form `UAT PHASE12 REGISTRATION` with synthetic NIM `UAT0000001`; and pending aspiration `UAT PHASE12 ASPIRATION` addressed to DIV-A.
6. Upload only synthetic files to the private document archive, public media fields, and registration IMAGE field. Do not use real member, applicant, or confidential documents. Keep file hashes/fixture names in the operator's local UAT notes, not in the repository.
7. Record generated resource IDs and expected visibility in a local UAT worksheet. IDs may differ across fresh schemas; fixture marker strings are the stable identity. Create each fixture once per fresh schema and do not use the recipe against a populated shared database.

## Repeat / cleanup

For a fresh clean run, provision another uniquely named empty UAT schema and dedicated user, then repeat from the migration step. Do not use `api:migrate:down` as a reset: it rolls back only the latest migration and can destroy records. `pnpm api:integration` is not a seed command; it resets and cleans the dedicated `_test` database, so never point it at UAT, development, or production. `pnpm db:down` stops the local container and keeps its named volume.

If an operator needs to retire an existing UAT schema, first verify its exact database host/name, take and verify a backup if any records matter, obtain the data owner's approval, and perform removal through the DBA's controlled database tooling. This repository supplies no destructive reset command.

## Completion evidence

Record the schema name, migration version, fixture marker counts, and the fact that passwords were entered through secure prompts. Never record credentials, tokens, PII, or private file contents. Verify login for all four roles and the role/scope checks in `PHASE_12_UAT_CHECKLIST.md` before calling the environment seeded.
