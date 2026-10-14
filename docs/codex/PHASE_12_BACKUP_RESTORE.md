# Phase 12 — Backup and Restore Runbook

## Recovery set

Back up these components together from one documented recovery point:

| Component | Source | Reason |
|---|---|---|
| MariaDB schema/data | Production `bem_fmikom` database | Domain records, metadata, migrations, users, sessions, and activity logs |
| Private documents | `DOCUMENT_STORAGE_PATH` | Private archive bytes referenced by document rows |
| Registration attachments | `REGISTRATION_STORAGE_PATH` | Private applicant IMAGE bytes referenced by submission file rows |
| Public media | `MEDIA_STORAGE_PATH` | Immutable uploaded images referenced by media metadata/URLs |
| Deployment configuration metadata | Secret manager/config inventory only | Restore the configuration names/versions; never put secret values in the backup repository |

Treat DB and all three storage roots as a single release/recovery set. An isolated database-only restore is not a complete portal restore.

## Before first production use

1. The hosting operator must choose an approved MariaDB/MySQL provider, region, encryption, retention, RPO/RTO, and backup ownership. Those values are not fixed by this repository; record them in the release change ticket.
2. Provision durable private storage for all three configured paths, outside every web/static root. Restrict access to the API runtime and backup operator; preserve ownership, permissions, timestamps, and opaque names.
3. Select a point-in-time consistency method. Prefer a database snapshot plus storage snapshots from the same quiesced write window. If the provider cannot coordinate snapshots, temporarily stop application writes before taking the recovery set. `--single-transaction` provides a consistent InnoDB DB dump but cannot make filesystem snapshots atomic with SQL.
4. Encrypt backups at rest and in transit; store them outside the application host and test decryption/access under a second operator account.

## Logical MariaDB backup

Run from a trusted operator host with a TLS-validated connection. Use a protected client option file (`chmod 600`) or an interactive password prompt; never pass passwords on the command line or print a DSN. Use a dedicated backup principal with read/lock privileges required by the selected dump mode.

```sh
umask 077
mariadb-dump --defaults-extra-file=/secure/path/bem-backup.cnf \
  --single-transaction --routines --triggers --events --hex-blob \
  --default-character-set=utf8mb4 bem_fmikom \
  | gzip -c > bem_fmikom-UTC_TIMESTAMP.sql.gz
gzip -t bem_fmikom-UTC_TIMESTAMP.sql.gz
sha256sum bem_fmikom-UTC_TIMESTAMP.sql.gz > bem_fmikom-UTC_TIMESTAMP.sql.gz.sha256
```

Replace the illustrative file names with a UTC timestamp before running. Ensure the pipeline propagates a failed `mariadb-dump` exit code (`set -o pipefail` in Bash) and upload encrypted artifacts to the approved backup location. A checksum detects accidental corruption; it is not encryption or authenticity protection.

## Storage backup

1. Stop writes or take coordinated filesystem snapshots for the three configured storage roots.
2. Archive each root preserving metadata and ACLs using the platform-approved backup tool. Do not follow symlinks outside the configured root; fail the backup if an unexpected symlink exists.
3. Encrypt each archive and create a SHA-256 checksum. Record only the storage-root label, UTC snapshot time, archive identifier, and checksum in the change ticket.
4. Keep DB dump and three storage artifacts under one recovery-set ID and retention schedule. Never copy uploaded user content into Git or CI artifacts.

## Restore drill (isolated target only)

1. Select a fresh isolated restore database and fresh private storage directories. Confirm host, schema, and paths twice; never restore into production or the live UAT database.
2. Retrieve one recovery-set ID and verify encryption, authorization, timestamps, archive safety, and every checksum before extraction/import.
3. Restore storage archives to empty paths owned by the API runtime with restrictive permissions. Do not extract untrusted paths without the archive tool's traversal protections.
4. Create an empty schema and import the dump using a restore-only principal. Example:

```sh
gzip -dc bem_fmikom-UTC_TIMESTAMP.sql.gz | mariadb --defaults-extra-file=/secure/path/bem-restore.cnf bem_fmikom_restore
```

5. Configure a non-production API instance to use only the restore DB and restored paths. Do not expose it publicly. Verify migration version, representative row counts, document downloads, registration attachment access, public media response, auth/RBAC, and logs for absent secret/file content.
6. Record elapsed restore time, achieved recovery point, mismatches, and operator sign-off. Delete the isolated restored data through approved retention procedures when the drill is complete.

Do not run `api:migrate:down` as a recovery mechanism. Restore a matched database/storage set or apply a reviewed forward fix. A storage/DB mismatch can leave missing assets or orphan bytes; repair before reopening writes.
