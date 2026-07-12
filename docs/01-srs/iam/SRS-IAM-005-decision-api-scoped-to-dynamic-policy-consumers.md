# SRS-IAM-005 IAM Decision API Scoped To Dynamic-Policy Consumers

Parent: [Podzone SRS](../podzone-srs.md) · [Traceability Matrix](../traceability-matrix.md)

The system shall route a service's authorization check through IAM's
decision API (`CheckPermission`/`CheckPlatformPermission`) only when that
check genuinely needs dynamic policy evaluation — per-resource-instance
grants, explicit deny, permission boundaries, organization guardrails, or
trust policies. A service whose authorization need is fully described by
"which static role does the principal hold, and does that role permit this
action" — with no per-instance resource scoping — shall enforce it locally
using a role carried in the caller's session token, not a network call to
IAM. This narrows, but does not repeal,
[SRS-IAM-001](./SRS-IAM-001-centralized-authorization.md): centralized
*policy* remains IAM's responsibility for callers that need it; centralized
*network calls for every check regardless of whether the caller needs
policy dynamism* is not required and is expensive when the answer never
varies per resource instance.

Confirmed via a live inventory (2026-07-12) of every cross-service call
into IAM's decision API: 6 call sites total (onboarding 4, backoffice 1,
partner 1). Backoffice's one call site is the only one whose resource
identifier is resolved from GraphQL request context at resolve time
(`podzone:tenant/{tenantId}/store/{storeId}`), and its actual permission
model (`internal/iam/migrations/sql/0002_create_iam_core.sql`) is four
roles with a strictly nested, static permission set — no per-store ACL,
explicit deny, boundary, or SCP is used by Backoffice today.

Status: Draft, ready to spec into implementation tasks — see
[ADR-0004](../../08-adr/ADR-0004-backoffice-static-rbac-not-iam-decision-api.md)
(the boundary decision) and
[PZEP-0003](../../09-pzep/PZEP-0003-iam-decoupling-and-sdk-phase-1.md) (the
implementation plan: JWT role claim, Backoffice static table, IAM Go SDK
for the remaining callers). Not gated behind backbone stabilization
(`../../06-recovery/recovery-plan.md`) — this is an internal authorization
refactor with no user-facing behavior change, not new feature breadth —
but touches Backoffice and Auth's core request path, so implementation
must ship with the test coverage in PZEP-0003's Test Plan before merging.

Linked docs:

- `../../03-architecture-detail-design/11-iam-platform.md`
- `../../08-adr/ADR-0004-backoffice-static-rbac-not-iam-decision-api.md`
