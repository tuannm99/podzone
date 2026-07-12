# ADR-0004: Backoffice enforces authorization with a static role→action table, not IAM's decision API

## Status
Proposed

## Date
2026-07-12

## Related Commit
(none yet — proposal only, nothing implemented)

## Context

Podzone's stated direction (see `docs/03-architecture-detail-design/11-iam-platform.md`)
is to extract IAM into a standalone, open-source, identity-provider-neutral
authorization platform, reusable by unrelated products via an SDK. Before
starting that work, a live inventory of every current cross-service call
into IAM's decision API (`CheckPermission`/`CheckPlatformPermission`) was
taken:

| Caller | Call sites | Resource dependency |
|---|---|---|
| onboarding | 4 (`AuthorizeStoreRequest`, `AuthorizeStoreRead`, `AuthorizeStoreApproval`, `AuthorizeInfrastructureRead`/`Manage`) | Tenant/platform-scoped, no per-instance resource resolution |
| backoffice | 1 (`RequirePermission`, called from `tenant_middleware.go`'s `InterceptField` for every GraphQL resolver) | **Per-instance**: resource string `podzone:tenant/{tenantId}/store/{storeId}`, `storeId` resolved from GraphQL request context at resolve time, not a static route parameter |
| partner | 1 (`AuthorizeTenant`) | Tenant-scoped, resource defaults to `*` |

Backoffice is the only caller whose resource identifier cannot be known
before the request body/GraphQL selection is parsed — this is exactly the
case `11-iam-platform.md`'s "Enforcement Placement" table calls out as
requiring a service-handler PEP, not a gateway PEP: *"GraphQL field or
mutation permission... Request-body-dependent... decision: No [gateway],
Yes [service handler]"*.

Separately: Backoffice's actual permission model, read from the real
migration data (`internal/iam/migrations/sql/0002_create_iam_core.sql`),
is four roles with a strictly nested permission set — `tenant_viewer` ⊂
`tenant_editor` ⊂ `tenant_admin` ⊂ `tenant_owner` — and no per-store ACL is
used anywhere in the system today (confirmed: no policy statement or
inline policy scopes a grant to one store ID out of several in the same
tenant). The dynamic parts of IAM's policy engine (explicit deny,
permission boundaries, organization SCPs, versioned policies, trust
statements) exist to support IAM's own admin console and onboarding's
platform-level checks — Backoffice has never used any of them.

## Decision

Backoffice's authorization stops calling IAM's decision API at request
time. Instead:

1. **Auth embeds the caller's tenant role in the JWT.** At login, tenant
   switch, and token refresh, `auth` reads `role_name` from its own
   `iam_tenant_memberships_projection` (already locally present — no new
   cross-service call) and adds an `active_tenant_role` claim to the
   issued JWT, alongside the existing `active_tenant_id`.
2. **Backoffice enforces with a static, compiled-in role→action table**
   mirroring the real seeded permission set above, checked purely
   in-memory against the JWT's `active_tenant_role` claim — no network
   call, no database read, no per-store distinction.
3. **IAM stops being a backoffice runtime dependency.** Backoffice's
   `TenantAuthorizer`/`RequirePermission` port is replaced with a local
   implementation; the gRPC client to IAM is removed from backoffice's
   wiring entirely.

This does not change IAM itself, onboarding, or partner — their calls
into IAM's decision API continue and are the actual target of the
SDK/gateway work now being scoped separately (PZEP-0003).

## Alternatives Considered

### Option A: Keep calling IAM's `CheckPermission` from backoffice (status quo)
Pros:
- No new code path; permission changes are centrally managed and
  immediately effective (no JWT staleness).
- Room to add real per-store ACLs later without re-architecting.

Cons:
- Every GraphQL resolver call pays a synchronous cross-service gRPC round
  trip for a decision that, empirically, only ever depends on a role that
  doesn't change mid-session for 99% of requests.
- Keeps backoffice as the one caller that structurally cannot move to
  gateway-level enforcement, which weighs down IAM's extraction plan with
  a resource-resolution problem it doesn't actually need to solve for its
  two other real callers.
- IAM's action catalog stays polluted with Backoffice-specific strings
  (`store:*`, `store_config:*`) baked into IAM's own migrations — exactly
  what `11-iam-platform.md`'s "Product Independence Invariants" says a
  reusable IAM must not have.

### Option B: Static role table in Backoffice, role from JWT (this proposal)
Pros:
- Removes the one caller whose resource-dependent checks made
  gateway-level enforcement impossible for IAM as a whole — directly
  unblocks IAM's extraction plan for its remaining callers.
- Zero network calls for the hottest authorization path in the product
  (every Backoffice GraphQL request).
- Backoffice's actual current usage (role-gated CRUD, no per-store ACL)
  is honestly represented in code instead of routed through a dynamic
  engine that's never asked to do anything dynamic.

Cons:
- Role changes don't take effect until the JWT is refreshed — a
  downgraded or revoked tenant member keeps their old permissions until
  their token expires or they re-authenticate. Accepted explicitly: no
  incident or requirement currently demands faster propagation.
- No per-store ACL path exists anymore without a larger follow-up design.
  Accepted explicitly: not used today, not a near-term requirement.
- Backoffice's permission table and IAM's seeded role-permission rows
  (migration `0002_create_iam_core.sql`) can drift apart over time since
  nothing keeps them in sync automatically — must be called out in code
  comments/docs at the point of definition.

### Option C: Backoffice keeps calling IAM, but only for role lookup (not `CheckPermission`), and does its own role→action mapping
Pros:
- IAM stays the single source of truth for "who has what role," no JWT
  claim needed, no staleness window.

Cons:
- Still a network call on the hot path, just a cheaper one — doesn't
  remove IAM from backoffice's critical path, only shrinks the RPC.
- Doesn't solve the "IAM must be reachable for Backoffice to serve any
  request" availability coupling that Option B removes entirely.

## Consequences

- Backoffice becomes independently deployable/testable without a running
  IAM instance — meaningful for local dev, CI, and the eventual
  open-source IAM's "must run standalone" invariant (IAM no longer needs
  to model Backoffice's action catalog just to exist).
- `store:approve` (migration `0024_add_store_approval_permission.sql`,
  granted only to `platform_owner`/`platform_admin`) is unaffected — that
  check lives in **onboarding** (`AuthorizeStoreApproval`), not backoffice,
  and is out of scope for this decision.
- Future requirement for per-store ACLs or faster permission propagation
  would require revisiting this decision, not a small patch — flagged
  explicitly in Consequences so it isn't rediscovered as a surprise.
- `internal/iam/migrations/sql/0002_create_iam_core.sql`'s `store:*`/
  `store_config:*` permission rows and their role assignments become the
  historical source Backoffice's static table was seeded from; IAM keeps
  serving them for its own admin console (an operator can still see/edit
  these role-permission rows there) but Backoffice no longer reads them at
  request time. Docs must say this explicitly to avoid a future reader
  assuming the two are still connected.

## Rule Of Thumb

If a service's authorization need is "which static role does this
principal have, and does that role permit this action" with no
per-resource-instance policy, dynamic deny, or cross-tenant boundary —
don't route it through IAM's decision API. Embed the role in the session
token and check it locally. Reserve IAM's decision API for callers that
actually need policy dynamism (organization guardrails, permission
boundaries, trust policies, explicit deny) — today that's onboarding's
platform-scope checks and IAM's own admin console, not Backoffice.
