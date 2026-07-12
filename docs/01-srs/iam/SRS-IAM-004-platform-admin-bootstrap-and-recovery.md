# SRS-IAM-004 Platform Admin Bootstrap And Recovery

Parent: [Podzone SRS](../podzone-srs.md) · [Traceability Matrix](../traceability-matrix.md)

The system shall provide an explicit, audited, operator-only procedure to
grant the first platform-level role (`platform_owner`) when no platform
admin exists yet, and to recover platform admin access if every existing
platform admin becomes unreachable (break-glass). This procedure must not
be reachable through the public HTTP/gRPC API surface — normal
[SRS-IAM-001](./SRS-IAM-001-centralized-authorization.md) permission checks
cannot apply to it, since by definition no authorized platform actor may
exist yet — so it must require direct operational access (deploy-level
database credentials), equivalent in trust level to running a migration,
not a product feature reachable by any signed-in user. Every use of this
procedure shall write an `iam_audit_logs` entry.

Confirmed gap (as of 2026-07-12): no such procedure exists today.
`internal/iam/migrations/sql/` creates the `platform_owner`/`platform_admin`
role definitions and their permission sets, but no migration assigns either
role to a real user. `AddPlatformRole`
(`internal/iam/controller/grpchandler/role_methods.go`) requires the caller
to already hold `platform:manage_roles` — it cannot bootstrap the first
grant. In the local dev database, the first platform admin exists only
because of a manual, unaudited SQL insert into `user_platform_roles`, which
is exactly the situation this requirement exists to replace.

Status: planned, not yet scheduled — this expands feature breadth beyond
the current backbone flow (see `../../06-recovery/recovery-plan.md` Phase R5
"Stabilize Runtime"); do not implement until the backbone flow
([SRS-ONB-001](../onboarding/SRS-ONB-001-workspace-and-store-entry.md),
[SRS-ONB-002](../onboarding/SRS-ONB-002-store-provisioning-workflow.md),
[SRS-ONB-003](../onboarding/SRS-ONB-003-placement-source-of-truth.md),
[SRS-BO-001](../backoffice/SRS-BO-001-store-scoped-backoffice.md)) works end
to end in Docker dev. See
[PZEP-0002](../../09-pzep/PZEP-0002-platform-admin-bootstrap-and-recovery.md)
for the proposed solution.

Linked docs:

- `../../00-project-vision/02-actors-and-business-flows.md`
- `../../03-architecture-detail-design/11-iam-platform.md`
