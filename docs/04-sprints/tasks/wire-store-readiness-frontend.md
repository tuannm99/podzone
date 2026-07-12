# Task: Wire Store Readiness Endpoint Into Frontend

**Sprint:** 0, Slice 0.2 (Backbone Flow) — follow-up to Slice 0.4
**SRS:** SRS-ONB-001, SRS-ONB-002, SRS-ONB-003
**Recovery source:** `docs/06-recovery/backbone-flow-refactor.md` — "First Agent-Ready Task Candidate" and Known Gap #4 (15 statuses not mapped to 3 UI states)
**Contract:** `docs/03-architecture-detail-design/05-transport-contracts.md` — "Slice 0.3: Store Readiness Contract"
**Status:** Implemented 2026-07-12 —
`frontend/packages/shared/services/onboarding.ts` (`getStoreReadiness`),
`frontend/apps/onboarding/src/pages/admin-home/createStoreReadinessViewModel.ts`
(new), `createAdminHomeViewModel.ts` (wired), `presentation.ts`
(`readinessBadgeColor`), `ProvisioningRequestsPanel.tsx` (badge now driven by
`ui_state`). `npm run build` / `build:onboarding` / `lint` / `format:check`
all pass.

**API-level verified 2026-07-12** against the running `full`-profile Docker
dev stack: `curl`'d `GET :8800/onboarding/v1/requests/<id>/readiness` with
the dev-bootstrap JWT for the seeded "Demo Store" request — response was
`{"request_status":"ready","readiness":{"store_ready":true,"placement_allocation_ready":true,"route_ready":true},"ui_state":"ready"}`,
exactly the shape `getStoreReadiness` expects, and `readinessBadgeColor('ready')`
correctly resolves to green. See `docs/06-recovery/backbone-flow-refactor.md`
"Current Status" for the full verification trail.

**Still not done:** (1) browser-rendered confirmation that the badge
actually paints green in the running UI — no headless-browser tool was
available in this environment this pass; (2) the `pending`/`provisioning`/
`blocked`/`failed` branches — only `ready` has seed data to exercise.

**Confirmed bug this fixes:** `ProvisioningRequestsPanel.tsx` currently
derives the status badge color from a client-side prefix check —
`request.status === 'ready' ? 'green' : request.status.startsWith('failed') ? 'red' : 'yellow'`.
Per the readiness contract's status mapping table, this is wrong in two
concrete cases: a `failed_retryable` request should map to `blocked`, not
plain `failed`; and a request with `request_status: 'ready'` but
placement/route **not** ready should map to `blocked`, not `ready` — the
current code would show it green (ready) when it isn't.

---

```text
You are working in the Podzone monorepo.

Task:
Add a getStoreReadiness client to frontend/packages/shared/services/onboarding.ts,
a small createStoreReadinessViewModel that fetches readiness for the visible
page of store requests, and wire it into ProvisioningRequestsPanel's status
badge, replacing the client-side status-prefix heuristic.

Goal:
The provisioning requests table shows the backend's authoritative ui_state
(pending | provisioning | blocked | failed | ready) for every visible
request instead of guessing from request.status alone.

References:
- CLAUDE.md
- docs/00-governance/agent-working-rule.md
- agent/SKILL.md
- agent/SOLID_STYLE_GUIDE.md
- docs/03-architecture-detail-design/05-transport-contracts.md (Slice 0.3 section)
- docs/06-recovery/backbone-flow-refactor.md
- frontend/packages/shared/services/onboarding.ts
- frontend/apps/onboarding/src/pages/admin-home/createStoreCollectionsViewModel.ts (pattern to follow)
- frontend/apps/onboarding/src/pages/admin-home/ProvisioningRequestsPanel.tsx
- frontend/apps/onboarding/src/pages/admin-home/createAdminHomeViewModel.ts

Scope:
You may modify:
- frontend/packages/shared/services/onboarding.ts (add StoreReadiness type + getStoreReadiness function)
- frontend/apps/onboarding/src/pages/admin-home/createStoreReadinessViewModel.ts (new file)
- frontend/apps/onboarding/src/pages/admin-home/createAdminHomeViewModel.ts (wire the new view model in, expose on the returned object)
- frontend/apps/onboarding/src/pages/admin-home/ProvisioningRequestsPanel.tsx (use ui_state for badge color; keep provisioningStatusLabel for the text label)

Out of scope:
- Do not change the retry button's enablement logic (still keyed off raw request.status) — that's a separate, unconfirmed correctness question, not part of this task.
- Do not change StoreChooser.tsx or the openStore flow — that path lists already-provisioned backoffice stores (via listStores), not onboarding requests; readiness does not apply there.
- Do not add a batch/bulk readiness endpoint — the contract only defines a single-id GET; call it once per visible row.
- Do not change backend code.
- Do not change API contracts.
- Do not add new dependencies.

Architecture rules:
- Frontend components must not call services directly — ProvisioningRequestsPanel reads from the ViewModel only, never imports from @podzone/shared/services directly.
- Remote reads use createResource or the shared pagination resource, per agent/SOLID_STYLE_GUIDE.md.
- Frontend must not call IAM permission-check endpoints as authorization probes (not applicable here — this task touches onboarding, not IAM).

Acceptance criteria:
- getStoreReadiness(tenantId, requestId) calls GET {ONBOARDING_API_URL}/onboarding/v1/requests/{id}/readiness with the tenant header, matching the existing service functions' pattern in onboarding.ts.
- createStoreReadinessViewModel fetches readiness for exactly the request IDs currently visible on the page (from storeRequests.items()), re-fetching when the visible set or selected workspace changes.
- ProvisioningRequestsPanel's status Badge color is driven by ui_state: 'ready' -> green, 'failed' -> red, 'pending' | 'provisioning' | 'blocked' -> yellow (matches the panel's existing 3-color vocabulary; do not invent new Badge colors).
- While readiness for a row has not loaded yet, the badge falls back to the existing status-label text with a neutral color, not an error.
- A request with request_status 'failed_retryable' displays as 'blocked' semantics (yellow), not 'failed' (red) — this is the confirmed bug this task fixes.
- npm run build, npm run lint, npm run format:check all pass.

Validation:
- cd frontend && npm run build
- cd frontend && npm run lint
- cd frontend && npm run format:check
- git diff --check

Handoff:
- Changed files
- Behavior changed
- Verification results
- Known gaps (e.g. end-to-end Docker verification still pending — that requires a live backend/DB, out of scope for this pass)
- Suggested commit message: fix(onboarding): drive provisioning status badge from readiness endpoint, not status prefix
```
