# Task: Onboarding Store Readiness Endpoint

**Sprint:** 0, Slice 0.4  
**SRS:** SRS-ONB-002, SRS-ONB-003  
**Recovery source:** `docs/06-recovery/backbone-flow-refactor.md` — Gap 1 (no combined readiness endpoint)  
**Contract:** `docs/03-architecture-detail-design/05-transport-contracts.md` — "Slice 0.3: Store Readiness Contract"  
**Status:** Backend implemented and tested as of 2026-07-11 —
`internal/onboarding/controller/httphandler/store/controller.go:145`
(`Controller.GetStoreReadiness`) and
`internal/onboarding/domain/store/interactor.go:619`
(`StoreInteractor.GetStoreReadiness`), covered by
`interactor_test.go` (Ready/Blocked/Provisioning/Failed/NotFound/
WorkspaceMismatch/NoAuth cases) and `controller_test.go`. Remaining work:
frontend integration (no FE code calls this endpoint yet) and end-to-end
verification in Docker dev. See `docs/06-recovery/backbone-flow-refactor.md`
"First Agent-Ready Task Candidate" for the next slice.

---

```text
You are working in the Podzone monorepo.

Task:
Add a GET /requests/:id/readiness endpoint to the onboarding service that
returns a combined ui_state (pending | provisioning | blocked | failed | ready)
for a store request, by combining the store request status with the placement
allocation and route readiness from infrasmanager.

Goal:
The FE store-chooser can call one endpoint to decide whether to show "Setting
up", "Provisioning", "Blocked", "Failed", or "Open Backoffice" for a store.
No more guessing from store status alone.

References:
- CLAUDE.md
- docs/00-governance/agent-working-rule.md
- agent/SKILL.md
- docs/03-architecture-detail-design/05-transport-contracts.md  (Slice 0.3 section)
- docs/06-recovery/backbone-flow-refactor.md
- internal/onboarding/README.md
- internal/onboarding/domain/store/inputport/store.go
- internal/onboarding/domain/infrasmanager/inputport/  (PlacementStatus type)
- internal/onboarding/controller/httphandler/store/controller.go
- internal/onboarding/controller/httphandler/authentication.go

Scope:
You may modify:
- internal/onboarding/domain/store/inputport/store.go  (add ReadinessResponse type + Usecase method)
- internal/onboarding/domain/store/interactor.go        (implement ReadinessQuery)
- internal/onboarding/domain/store/interactor_test.go   (add readiness tests)
- internal/onboarding/controller/httphandler/store/controller.go  (add GET /:id/readiness route)
- internal/onboarding/controller/httphandler/store/controller_test.go

Out of scope:
- Do not change Mongo collection names or schema.
- Do not modify infrasmanager domain — only call its existing Usecase port.
- Do not change proto files.
- Do not add new packages.
- Do not modify authentication middleware.
- Do not change frontend UI code.
- Do not change the existing GET /:id handler.
- Do not change pdtenantdb.

Architecture rules:
- The HTTP controller calls the store Usecase only.
- The store Usecase calls the infrasmanager Usecase via its input port (injected).
- Domain must not import infrastructure.
- Usecases depend on ports, not adapters.
- The controller extracts workspace_id from JWT context using toolkit.GetTenantID.
- Permission: workspace ownership check only (request.WorkspaceID == workspaceID from JWT).

Acceptance criteria:
- GET /requests/:id/readiness returns 200 with { request_id, request_status, readiness: { store_ready, placement_allocation_ready, route_ready }, failure_reason, ui_state }.
- ui_state = "ready" only when request_status == ready AND placement allocation + route are both ready.
- ui_state = "blocked" when request_status == ready but placement is not ready.
- ui_state maps pending statuses (requested/planning/planned/pending_approval) to "pending".
- ui_state maps active statuses (queued/provisioning/pending_platform_setup) to "provisioning".
- ui_state maps terminal failure statuses (failed/failed_non_retryable/rejected/cancelled/suspended) to "failed".
- ui_state maps failed_retryable to "blocked".
- Returns 404 if request not found or workspace does not match.
- Returns 401 if no valid JWT.
- Tests cover: ready+placement-ready, ready+placement-not-ready (blocked), provisioning state, failed state, 404, 401.

Validation:
- GOCACHE=/tmp/podzone-gocache go test ./internal/onboarding/...
- GOCACHE=/tmp/podzone-gocache GOLANGCI_LINT_CACHE=/tmp/podzone-golangci-cache go tool golangci-lint run --timeout=5m ./internal/onboarding/...
- git diff --check

Handoff:
- Changed files
- Behavior changed
- Verification results (test output)
- Known gaps
- Suggested commit message: feat(onboarding): add store readiness endpoint GET /requests/:id/readiness
```
