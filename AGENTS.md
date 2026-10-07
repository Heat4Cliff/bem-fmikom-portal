# AGENTS.md — Codex Rules for BEM FMIKOM UNUGHA Portal

## Mission
Implementasikan Sistem Informasi & Company Profile BEM FMIKOM UNUGHA sebagai aplikasi web production-oriented yang terdiri dari Public Website + Admin Management Dashboard berbasis role.

## Non-negotiable rules
- Jangan redesign. Stitch adalah visual reference yang sudah final/audited.
- Jangan menambah fitur di luar PRD v1.4.
- Jangan menghapus requirement karena implementasinya sulit; laporkan blocker.
- Jangan mengganti stack utama tanpa persetujuan.
- Backend adalah enforcement final untuk authorization; UI hiding bukan security control.
- Jangan menyimpan secret, token, password, atau credential di repository.
- Jangan menyimpan private document pada public web root bila bisa dihindari.
- Jangan membuat direct public URL untuk dokumen internal.
- Jangan menggunakan raw Stitch HTML sebagai production architecture. Rekonstruksi sebagai komponen React/Next.js yang maintainable.
- Jangan menyatakan sebuah phase PASS sebelum quality gate phase tersebut dijalankan dan buktinya dicatat.
- Jika verifikasi butuh browser/manual visual inspection, tulis `MANUAL VERIFICATION REQUIRED`, jangan mengklaim PASS dari static analysis saja.

## Required working style
1. Baca source terkait sebelum mengubah code.
2. Tulis rencana kecil untuk phase aktif.
3. Implementasikan hanya scope phase aktif.
4. Jalankan format/lint/typecheck/test/integration sesuai phase.
5. Bandingkan UI dengan Stitch reference untuk screen yang disentuh.
6. Catat gap, keputusan, command, result, dan bukti di phase report.
7. Update `PLANS.md` hanya untuk item yang benar-benar selesai dan terverifikasi.
8. Stop sebelum phase berikutnya kecuali prompt pengguna secara eksplisit meminta lanjut.

## Target stack
### Frontend
- Next.js App Router
- TypeScript strict
- Tailwind CSS
- Reusable component layer; Radix/shadcn primitives boleh dipakai bila membantu accessibility, tetapi styling akhir harus mengikuti Stitch.
- REST client typed dari OpenAPI/contract.

### Backend
- Go
- Gin sebagai baseline HTTP framework kecuali repository existing sudah menetapkan alternatif.
- MySQL/MariaDB.
- Database migrations versioned.
- JWT short-lived access token + refresh token via Secure HttpOnly cookie pada HTTPS.
- Server-side validation, rate limiting endpoint publik, RBAC + division scope.
- Excel export menggunakan `excelize` atau library setara yang terawat.

## Repository architecture target
```text
bem-fmikom-portal/
├─ apps/
│  ├─ web/                  # Next.js
│  └─ api/                  # Go API
├─ packages/
│  └─ contracts/            # OpenAPI, generated TS types/client if used
├─ docs/
├─ design_reference/        # immutable reference from Stitch
├─ scripts/
├─ infra/
├─ AGENTS.md
├─ PLANS.md
└─ README.md
```

## Code quality
- TypeScript: no casual `any`; use typed API models and runtime validation at trust boundaries.
- Go: handlers thin, business logic in service layer, persistence isolated in repository/data layer.
- Avoid duplicated permission logic: centralize policy/scope helpers.
- Use transactions for multi-step writes that must be atomic.
- Database constraints complement application validation.
- Use soft delete/archive for historical entities where PRD requires preservation.

## Visual fidelity rules
For every implemented screen:
- Open matching `screen.png`.
- Read matching `code.html` only as layout/style reference.
- Match typography, spacing, color, radius, density, responsive behavior, empty/loading/error states.
- Do not silently substitute generic dashboard styling.

## Security baseline
- Password hash: bcrypt or Argon2id; follow selected implementation consistently.
- JWT secret/keys from environment only.
- Refresh cookie: `HttpOnly`, `Secure` in production, appropriate `SameSite`.
- CORS allow-list from env; never wildcard with credentials.
- Rate limit public write endpoints.
- File upload: allow-list MIME/extension, size limit, generated server filename, path traversal prevention.
- Rich text: sanitize before rendering.
- Protect against IDOR by applying role + division scope at query/service layer.
- Log sensitive operations but never log password/token/file contents/private form payload unnecessarily.

## Testing classification
Every report must distinguish:
- `STATIC VERIFICATION`
- `AUTOMATED TEST`
- `INTEGRATION TEST`
- `VISUAL REVIEW`
- `MANUAL UAT`

Do not collapse these into one “tests passed” statement.
