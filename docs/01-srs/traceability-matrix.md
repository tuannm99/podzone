# Traceability Matrix

Use this matrix to connect requirements to architecture, contracts, sprints,
and agent tasks.

| SRS ID | Requirement doc | Design/UI spec | Architecture/C4 | Contract/DB | Sprint | Agent task | Verification |
| --- | --- | --- | --- | --- | --- | --- | --- |
| SRS-AUTH-001 | `../00-project-vision/iam-mvp.md` | `../06-recovery/backbone-flow-refactor.md` (Required Screens: Login) | `../02-architecture-overall/03-containers.md` | TBD — login/session request-response contract not locked | Sprint 0 | TBD | `go test ./internal/auth/...` (existing, unverified end-to-end in Docker) |
| SRS-IAM-001 | `../00-project-vision/13-iam-permission-authoring.md` | N/A — backend-only authorization, no direct UI surface | `../03-architecture-detail-design/11-iam-platform.md` | TBD — IAM decision API request/response not locked (`transport-contracts.md` "Authorization Boundary" is a flow diagram, not a contract) | Sprint 0 | TBD | `go test ./internal/iam/...` (existing, unverified end-to-end in Docker) |
| SRS-IAM-002 | `../00-project-vision/13-iam-permission-authoring.md` | TBD | `../03-architecture-detail-design/11-iam-platform.md` | TBD | TBD | TBD | TBD |
| SRS-ONB-001 | `../00-project-vision/09-backoffice-multitenancy.md` | `../06-recovery/backbone-flow-refactor.md` (Required Screens: Workspace/store chooser) | `../02-architecture-overall/05-sequences.md` | TBD — workspace/store selection read contract not locked (`backbone-flow-refactor.md` "Minimum Contracts To Lock" #1) | Sprint 0 | TBD | TBD |
| SRS-ONB-002 | `../00-project-vision/10-store-onboarding-pipeline.md` | `../06-recovery/backbone-flow-refactor.md` (Required Screens: Provisioning status) | `../02-architecture-overall/05-sequences.md` | `../03-architecture-detail-design/05-transport-contracts.md` (Slice 0.3: Store Readiness Contract) | Sprint 0 | `../04-sprints/tasks/onboarding-readiness-api.md` (backend implemented + tested; FE integration pending) | `go test ./internal/onboarding/domain/store/... ./internal/onboarding/controller/httphandler/store/...` |
| SRS-ONB-003 | `../../internal/onboarding/README.md` | N/A — placement source-of-truth rule, no direct UI | `../02-architecture-overall/04-data-ownership.md` | `../03-architecture-detail-design/05-transport-contracts.md` (Slice 0.3: Store Readiness Contract) | Sprint 0 | `../04-sprints/tasks/onboarding-readiness-api.md` (backend implemented + tested; FE integration pending) | `go test ./internal/onboarding/domain/store/... ./internal/onboarding/controller/httphandler/store/...` |
| SRS-BO-001 | `../00-project-vision/09-backoffice-multitenancy.md` | `../06-recovery/backbone-flow-refactor.md` (Required Screens: Backoffice home) | `../02-architecture-overall/03-containers.md` | TBD — one protected Backoffice read contract not locked (`backbone-flow-refactor.md` "Minimum Contracts To Lock" #4) | Sprint 1 | TBD | `go test ./internal/backoffice/...` (existing, unverified end-to-end in Docker) |
| SRS-CAT-001 | `../00-project-vision/05-epics-and-user-stories.md` | TBD | `../03-architecture-detail-design/01-modules.md` | TBD | TBD | TBD | TBD |
| SRS-PARTNER-001 | `../00-project-vision/07-partner-mvp.md` | TBD | `../03-architecture-detail-design/01-modules.md` | TBD | TBD | TBD | TBD |
| SRS-ORDER-001 | `../00-project-vision/11-backoffice-event-storming.md` | TBD | `../03-architecture-detail-design/01-modules.md` | TBD | TBD | TBD | TBD |
| SRS-SETTLEMENT-001 | `../00-project-vision/05-epics-and-user-stories.md` | TBD | `../03-architecture-detail-design/01-modules.md` | TBD | TBD | TBD | TBD |
