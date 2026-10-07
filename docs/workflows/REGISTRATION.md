# Workflow — Dynamic Registration

1. Authorized admin creates registration form in DRAFT.
2. Admin configures ordered fields.
3. Admin opens registration; status becomes OPEN if date/state rules pass.
4. Public detail/form endpoint exposes only safe/public definitions.
5. Public submission payload is validated against active field definition server-side.
6. Duplicate rules are checked when enabled.
7. Submission stored with form reference and timestamp.
8. Authorized admin lists/filters submissions.
9. Export endpoint generates XLSX in backend; action is logged as EXPORT.
10. Close action sets CLOSED; public submission then rejected.

Do not implement advanced branching/scoring/file upload by applicants in v1.4.
