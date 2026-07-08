# Backbone Flow Refactor

Status: draft.

This document describes the first runtime flow to stabilize before expanding
Podzone.

## Flow

```text
User
  -> sign in
  -> choose workspace
  -> request or select store
  -> onboarding verifies placement/readiness
  -> Backoffice opens store-scoped workspace
  -> UI calls one protected Backoffice API
```

## SRS Links

- SRS-AUTH-001
- SRS-IAM-001
- SRS-ONB-001
- SRS-ONB-002
- SRS-ONB-003
- SRS-BO-001

## Required Screens

| Screen | Required states |
| --- | --- |
| Login | loading, validation error, auth error, success |
| Workspace/store chooser | loading, empty workspace, pending store, failed store, ready store, permission denied |
| Provisioning status | queued, planning, provisioning, blocked, failed, ready |
| Backoffice home | loading, missing store, placement error, permission denied, success |

## Required Backend Capabilities

| Capability | Owner | Status |
| --- | --- | --- |
| Validate session/token | auth | Existing, verify |
| Resolve workspace membership | IAM/auth projection | Existing, verify |
| List/select stores with readiness state | onboarding/backoffice | Partial |
| Check placement allocation and route projection | onboarding | Partial |
| Resolve tenant DB route | pdtenantdb/onboarding projection | Existing, verify |
| Enforce protected API permission | backoffice -> IAM gRPC | Existing, verify |

## Minimum Contracts To Lock

1. Workspace/store chooser read contract.
2. Store request/provisioning status contract.
3. Placement readiness contract.
4. One protected Backoffice read contract.

## Exit Criteria

- A new user can sign in.
- The first workspace/root ownership is clear.
- Store list distinguishes pending, failed, and ready stores.
- Ready store has resolvable placement.
- Missing placement shows a provisioning/readiness error, not a generic crash.
- Backoffice never opens store-scoped mode without ready placement.
- One protected Backoffice API returns success for authorized user and
  permission detail for unauthorized user.

## First Agent-Ready Task Candidate

Do not implement yet until contracts are written.

Candidate:

```text
Define onboarding placement readiness query contract.
```

Allowed docs:

- `docs/03-architecture-detail-design/transport-contracts.md`
- `docs/03-architecture-detail-design/collection-api-contract.md`
- `docs/01-srs/traceability-matrix.md`
- `docs/04-sprints/sprint-00-foundation.md`

Done:

- request/response/errors/permission/UI behavior documented;
- no code change;
- traceability updated.
