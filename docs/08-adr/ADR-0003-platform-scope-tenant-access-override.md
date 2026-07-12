# ADR-0003: Platform-scope roles do not implicitly bypass tenant-scoped access checks

## Status
Proposed

## Date
2026-07-12

## Related Commit
(none yet — proposal only, nothing implemented)

## Context

Investigating a 403 on the onboarding "Resource inventory" screen
(`platform:read_infrastructure` denied) led to reading
`internal/iam/domain/interactor/authz.go`'s `CheckPermissionForResource`
end to end. Finding: **no platform-scope role or policy is ever consulted
when evaluating a tenant-scoped permission check.**

`CheckPermissionForResource` only evaluates, in order: an assumed role (if
present), the caller's `tenant_memberships` row for that exact tenant,
tenant-user / tenant-group policy statements, the tenant role's statements,
tenant-user permission boundaries, and the tenant's organization SCP. There
is exactly one cross-scope fallback (`membershipForAuthorization`): if the
caller has no tenant membership but holds `organization:manage_iam` on the
tenant's owning org, a synthetic `tenant_owner` membership is granted — but
only for tenants under that specific org.

Consequence, confirmed against the local dev database: `devowner@podzone.dev`
holds `platform_owner` (global platform role, including
`platform:read_infrastructure` and `platform:manage_infrastructure` per
migration `0026_add_infrastructure_permissions.sql`) and can therefore read
platform-wide infrastructure inventory. But that same account has **no**
implicit access to any tenant's store data — `store:read`/`store:create`
checks go through the tenant-scoped path above, and `platform_owner` is
never on that path. To open a specific tenant's store, `platform_owner`
would need an explicit `tenant_memberships` row (or organization-level
`manage_iam`) for that tenant, same as anyone else.

There is currently no "root admin" identity in Podzone that can open any
store across any tenant based on a platform-level permission grant. This
was raised as a product question: should such a role exist, and if so, how
does it interact with the tenant-isolation boundary that
`docs/00-project-vision/02-actors-and-business-flows.md` already documents
("organization membership ... never grants platform permissions")?

## Decision

**Proposed, not yet accepted:** add one additional, explicitly-scoped
fallback branch to `CheckPermissionForResource`, evaluated only after the
existing tenant-membership and org-fallback paths find nothing:

- If the caller holds a platform role or platform-scope policy whose
  statements match the requested `(action, resource)` pattern (e.g. an
  action pattern like `store:*` with resource pattern
  `tenant/*/stores/*`), and that grant passes the caller's platform
  permission boundary, treat the request as allowed.
- This path must be a separate, clearly-named branch (not a silent
  fallthrough) so it is easy to find, disable, or audit independently of
  the tenant-membership evaluation.
- Every allow decision taken through this path must be audit-logged
  (`iam_audit_logs`) with a distinct `resource_type`/`action` marker (e.g.
  `platform_override`) so cross-tenant access via this path is always
  traceable — this is the one path in the system that crosses the tenant
  isolation boundary by design, so it needs stronger observability than a
  normal same-tenant grant, not less.
- No existing role gets this behavior automatically. It only activates for
  a principal that holds a platform policy statement naming the tenant
  resource pattern explicitly — `platform_owner`/`platform_admin` having
  `platform:manage_infrastructure` today does **not** imply store access;
  a separate statement (e.g. `store:*` on `tenant/*/stores/*`) would need
  to be attached.

This ADR does not decide whether such a policy statement should actually be
seeded for `platform_owner` by default, nor which permission name to use —
that is a product/security call for a follow-up PZEP once this boundary
mechanism itself is accepted.

## Alternatives Considered

### Option A: Explicit per-tenant membership only (status quo, no change)
Pros:
- Simplest model, already implemented, already audited by existing
  tenant-membership tooling.
- No new code path across the tenant-isolation boundary — smallest attack
  surface.

Cons:
- Operational break-glass (support debugging a specific tenant's store) has
  no path today except manually inserting a `tenant_memberships` row per
  incident, which is itself an unaudited, ad-hoc bypass of the same
  boundary this ADR is trying to protect.
- "Root admin" as a product concept (mentioned by the user) has no
  representation in the permission model at all.

### Option B: Platform-scope override branch in `CheckPermissionForResource` (this proposal)
Pros:
- Root-admin capability becomes a normal permission grant (a policy
  statement), reusable across services, auditable, revocable, and subject
  to permission boundaries like every other grant in the system.
- Keeps tenant-scoped evaluation as the primary, default path — the
  override only fires when nothing else matched.

Cons:
- Adds a second place cross-tenant access can originate (in addition to
  the existing org-manage_iam fallback), increasing the paths a reviewer
  must reason about when auditing tenant isolation.
- Requires new audit-log discipline (see Decision) to stay safe — without
  it, this becomes an easy-to-miss god-mode hole.

### Option C: Separate "impersonation"/assume-role flow instead of a permission-check branch
Pros:
- Reuses the existing `AssumeRole`/trust-statement machinery
  (`canAssumeRole`, `evaluateAssumedRolePermission`) which already has
  explicit, auditable session boundaries and is designed for exactly this
  kind of temporary elevated access.
- No change needed to `CheckPermissionForResource`'s default path at all.

Cons:
- Requires the caller to explicitly assume a tenant-scoped role per target
  tenant, which is a worse UX for "browse any store" style admin tooling
  than a standing platform grant, and does not fit a use case where the
  admin doesn't know the target tenant ID in advance (e.g. a cross-tenant
  onboarding dashboard).

## Consequences

If Option B is accepted: `internal/iam/domain/interactor/authz.go` gains a
new branch and a new audit event type; every service's `AccessAuthorizer`
(`onboarding`, `backoffice`, `partner`) keeps calling the same
`CheckPermission`/`CheckPermissionForResource` RPCs unchanged — the change
is contained to IAM's decision logic. A follow-up PZEP is required before
implementation (per `docs/05-process/pzep-template.md`'s rule: changes
touching permissions and multiple components need a PZEP), covering which
permission name to introduce, whether `platform_owner` gets it by default,
and the audit log schema addition.

If rejected (Option A retained): document that cross-tenant "root admin"
access is intentionally not supported, and that support/ops access to a
specific tenant's store must go through an explicit, per-tenant
`tenant_memberships` grant (existing mechanism) — update
`docs/00-project-vision/02-actors-and-business-flows.md` to state this
explicitly so the question doesn't resurface as an assumed gap later.

## Rule Of Thumb

A platform-scope role (`platform_owner`, `platform_admin`, or any future
platform role) grants **no** tenant-scoped access by default. Tenant/store
access always requires either an explicit `tenant_memberships` row, an
org-level `manage_iam` fallback scoped to that tenant's org, or — if this
ADR is accepted — an explicit platform-policy statement matching the
tenant resource pattern, evaluated through the audited override branch.
Never assume "owner"-named roles imply cross-tenant reach.
