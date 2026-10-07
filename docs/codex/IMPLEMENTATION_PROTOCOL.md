# Codex Implementation Protocol

## Before each phase
- Read `AGENTS.md`, active phase in `PLANS.md`, relevant PRD/workflow/API/design files.
- Inspect existing code before proposing changes.
- Create a small phase execution note and identify tests.

## During implementation
- Keep commits/changes scoped to the active phase.
- Prefer simple maintainable architecture over premature abstractions.
- Any discovered contract/design mismatch goes to phase report; do not silently improvise.

## Verification taxonomy
### STATIC VERIFICATION
Formatting, lint, typecheck, compile/build.

### AUTOMATED TEST
Unit/component/service tests that do not require full external integration.

### INTEGRATION TEST
API + DB/storage/auth interactions, preferably against deterministic local test environment.

### VISUAL REVIEW
Compare rendered route against Stitch `screen.png`; document deviations.

### MANUAL UAT
Human interaction validation in browser/device. Never claim this from static checks.

## Defect record
For meaningful defects record:
- precondition
- reproduction
- expected
- actual
- evidence
- status
- root cause
- files changed
- regression risk
- prevention/test added

## Phase exit
- Run required commands.
- Write phase report from template.
- Update only completed checkboxes.
- State `PASS`, `PASS WITH MANUAL VERIFICATION REQUIRED`, or `BLOCKED`.
- Stop and wait for user before starting next phase unless instructed otherwise.
