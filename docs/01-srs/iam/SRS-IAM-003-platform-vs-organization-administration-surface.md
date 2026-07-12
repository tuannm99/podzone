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

Status: **unblocked 2026-07-12** — explicit product decision to proceed
after the backbone flow
([SRS-ONB-001](../onboarding/SRS-ONB-001-workspace-and-store-entry.md),
[SRS-ONB-002](../onboarding/SRS-ONB-002-store-provisioning-workflow.md),
[SRS-ONB-003](../onboarding/SRS-ONB-003-placement-source-of-truth.md),
[SRS-BO-001](../backoffice/SRS-BO-001-store-scoped-backoffice.md)) was
verified at the API level (not full browser UI — see
`../../06-recovery/backbone-flow-refactor.md`). First slice implemented
same day: sidebar nav gating for Provisioning/IAM, see
`../../04-sprints/tasks/gate-platform-nav-by-role.md`. Splitting `/admin/iam`
into two distinct consoles (vs. today's one console that internally adapts
via `canManagePlatform`) is not yet done — the nav-visibility slice was
judged sufficient for now.

Linked docs:

- `../../00-project-vision/02-actors-and-business-flows.md`
- `../../03-architecture-detail-design/11-iam-platform.md`
