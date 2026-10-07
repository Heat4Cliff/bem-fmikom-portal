# Workflow — Authentication & RBAC

## Login
`email + password -> verify active user -> issue access token -> set refresh cookie -> log LOGIN -> return principal`.

## Refresh
`refresh cookie -> validate/rotate according to implementation -> issue new access token`.

## Protected request
`verify JWT -> load principal/account status -> resolve role -> apply module policy -> apply division scope -> execute`.

## Scope examples
- KEPALA_BIDANG article list: repository query must include authenticated `division_id`.
- KEPALA_BIDANG export: form must belong to own division.
- BPH: no division restriction for operational modules permitted by matrix.
- SUPER_ADMIN: no division restriction.

## Logout
Invalidate/clear refresh mechanism, clear cookie, log LOGOUT.
