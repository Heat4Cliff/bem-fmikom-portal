# Acceptance Criteria

- **AC-01** Public pages responsive desktop/mobile and match reviewed Stitch references.
- **AC-02** In-scope API endpoints implemented or explicitly excluded before code freeze.
- **AC-03** Protected unauthenticated request -> 401; forbidden role/scope -> 403 without leaking data.
- **AC-04** KEPALA_BIDANG cannot read/write another division through URL/body/query manipulation.
- **AC-05** Exactly one active period; previous history preserved.
- **AC-06** Dynamic registration works end-to-end: builder -> public render -> submit -> dashboard -> XLSX export.
- **AC-07** Aspirations route to division and status can be managed by permitted roles.
- **AC-08** Private documents are not directly publicly accessible; download requires authorization.
- **AC-09** Activity log captures login/logout, key CRUD, publish, upload/download, export.
- **AC-10** Public input endpoints have server validation and rate limiting.
- **AC-11** Repository contains no secret/credential.
- **AC-12** Backup, deployment and basic rollback are documented before production.

## Additional implementation DoD
A feature is not done until:
- UI happy/loading/empty/error/forbidden states are handled where relevant.
- Backend validation and authorization are tested.
- API response follows contract.
- relevant automated/integration tests pass.
- visual screen is reviewed against Stitch when a matching reference exists.
- phase report records evidence and unresolved manual checks.
