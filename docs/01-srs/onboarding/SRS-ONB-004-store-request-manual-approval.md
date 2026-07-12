# SRS-ONB-004 Store Request Manual Approval

Parent: [Podzone SRS](../podzone-srs.md) · [Traceability Matrix](../traceability-matrix.md)

The system shall let an authorized platform operator (System Admin / Platform
Operator actor, see `../../00-project-vision/02-actors-and-business-flows.md`)
inspect, approve, reject, or cancel a pending store onboarding request that
requires manual approval, per `10-store-onboarding-pipeline.md`'s request
lifecycle and its "operators must be able to inspect, retry, cancel, or
manually approve blocked requests without direct database edits" requirement.
This approval surface is distinct from the store owner's own request-status
view ([SRS-ONB-001](./SRS-ONB-001-workspace-and-store-entry.md)) — a store
owner may see their own request's progress but shall not see or act on other
tenants' requests.

**Correction (2026-07-12):** an earlier pass of this requirement claimed no
approve/reject capability existed at all — that was wrong. Re-verified
against `internal/onboarding/controller/httphandler/store/controller.go:44-45`
and `internal/onboarding/domain/store/interactor.go:194-206`:
`POST /requests/:id/approve` and `POST /requests/:id/reject` are real,
already wired to `StoreInteractor.ApproveStoreRequest`/`RejectStoreRequest`,
both gated by `authorizeApproval` → `AuthorizeStoreApproval` →
`CheckPlatformPermission(..., "store:approve")`. The actual remaining gaps
are narrower than originally stated:

1. No frontend code calls either endpoint —
   `frontend/packages/shared/services/onboarding.ts` exports
   `createStoreRequest`/`retryStoreRequest`/`listStoreRequests` only.
2. `ListStoreRequests` (`interactor.go:251`) is workspace-scoped
   (`s.repo.ListPage(ctx, workspaceID, ...)`) — there is no cross-tenant
   "every pending request, any workspace" query yet, which is what a
   platform-wide approval queue needs. An operator today would have to
   already know each workspace ID and query them one at a time.

Status: planned, not yet scheduled — this expands feature breadth beyond the
current backbone flow (see `../../06-recovery/recovery-plan.md` Phase R5
"Stabilize Runtime"); do not implement until the backbone flow
([SRS-ONB-001](./SRS-ONB-001-workspace-and-store-entry.md),
[SRS-ONB-002](./SRS-ONB-002-store-provisioning-workflow.md),
[SRS-ONB-003](./SRS-ONB-003-placement-source-of-truth.md),
[SRS-BO-001](../backoffice/SRS-BO-001-store-scoped-backoffice.md)) works end
to end in Docker dev.

Linked docs:

- `../../00-project-vision/02-actors-and-business-flows.md`
- `../../00-project-vision/10-store-onboarding-pipeline.md`
