# SRS-IAM-003 Platform Vs Organization Administration Surface

Parent: [Podzone SRS](../podzone-srs.md) · [Traceability Matrix](../traceability-matrix.md)

The system shall present distinct administration surfaces for platform-level
administration (System Admin / Platform Operator actor) and
organization-level administration (Organization Root / Organization Admin
actor), matching the actor model in `../../00-project-vision/02-actors-and-business-flows.md`:
platform administration manages organizations, platform-wide roles, and
cross-organization policy; organization administration manages only members,
roles, and policy inside the signed-in actor's own organization. Navigation
and console entry points shall reflect only the actions the signed-in
actor's role can plausibly use. This is a usability requirement, not a
security boundary — actual authorization is enforced per
[SRS-IAM-001](./SRS-IAM-001-centralized-authorization.md) regardless of what
the UI shows or hides.

Status: planned, not yet scheduled — this expands feature breadth beyond the
current backbone flow (see `../../06-recovery/recovery-plan.md` Phase R5
"Stabilize Runtime"); do not implement until the backbone flow
([SRS-ONB-001](../onboarding/SRS-ONB-001-workspace-and-store-entry.md),
[SRS-ONB-002](../onboarding/SRS-ONB-002-store-provisioning-workflow.md),
[SRS-ONB-003](../onboarding/SRS-ONB-003-placement-source-of-truth.md),
[SRS-BO-001](../backoffice/SRS-BO-001-store-scoped-backoffice.md)) works end
to end in Docker dev.

Linked docs:

- `../../00-project-vision/02-actors-and-business-flows.md`
- `../../03-architecture-detail-design/11-iam-platform.md`
