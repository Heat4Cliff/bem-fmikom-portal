# BEM FMIKOM Backend Handoff v1.0.1

Standalone Go/Gin API with source, tests, configuration template, and embedded migrations through 00013. Run `go mod download`, `go test ./...`, `go vet ./...`, and `go build ./...` from this directory. Provision the database and private storage first, then configure secrets from `.env.example` and run `go run ./cmd/migrate up`. The exported dotenv loader reads only this package's `.env`.

## Initial production SUPER_ADMIN

After verifying the production `DATABASE_URL`, completing a database backup/recovery checkpoint, and applying migrations, the campus operator creates the first administrator interactively:

```powershell
go run ./cmd/create-admin -name "Super Admin BEM FMIKOM" -email "<production-admin-email>"
```

Run from this directory in an interactive terminal. The CLI creates an active `SUPER_ADMIN`, validates/normalizes the email, rejects a duplicate email, and prompts twice for a password without terminal echo. The password must be at least 12 characters and no more than 72 bytes; never pass it as an argument. Use a unique production-only password and store it in the campus password manager. Do not reuse UAT/development credentials. This release includes migrations through 00013, with 00012–00013 adding branding/social presentation fields. Run `go run ./cmd/migrate up` before activating the new backend; do not use `go run ./cmd/migrate down` as a routine deployment step. The complete operator sequence and first-login checklist are in `../DEPLOYMENT_HANDOFF.md`.
