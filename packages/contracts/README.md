# API contracts package

This directory reserves the shared contract package for the portal. The authoritative endpoint and security contracts are currently maintained in `docs/backend/API_CONTRACT.md` and `docs/backend/SECURITY_CONTRACT.md`.

Phase 1 does not generate client types or publish OpenAPI because several PRD-to-contract gaps remain open. Phase 2 must resolve those gaps and then add versioned OpenAPI schemas and generated or validated client types here. Do not infer endpoints from Stitch `code.html`.
