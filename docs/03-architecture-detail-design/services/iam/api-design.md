# IAM Service — API Design

Parent: [Services Index](../README.md) · [IAM README](./README.md) · [DB Design](./db-design.md)

gRPC only, no HTTP. Full request/response message shapes live in
`api/proto/iam/v1/{iam_service,iam_tenant,iam_policy,iam_simulation}.proto`
— not reproduced field-by-field here to avoid drift; this doc lists method
names, groups, and callers.

## C3: Component View

```mermaid
flowchart TB
    subgraph Controller["controller/grpchandler"]
        CmdSrv["IAMCommandServer\n(command_server.go)"]
        QrySrv["IAMQueryServer\n(query_server.go)"]
        Methods["*_methods.go\n(tenant, group, role, policy,\nprincipal, organization, permission,\ndirectory)"]
    end

    subgraph Domain["domain"]
        Interactor["interactor\n(CheckPermission, AssumeRole,\ntenant/policy/group usecases)"]
        InputPort["inputport\n(usecase_authz.go, ...)"]
        OutputPort["outputport"]
        Entity["entity"]
    end

    subgraph Infra["infrastructure"]
        Repo["repository\n(*_repo_impl.go, Postgres)"]
        AuthClient["authclient\n(user directory lookups)"]
        Outbox["messaging/outbox\n(outbox_worker.go)"]
    end

    Controller --> Domain
    Domain --> InputPort
    Domain --> Entity
    Infra --> OutputPort
    Repo --> OutputPort
    AuthClient --> OutputPort
    Outbox --> Repo
    Outbox -->|"publish"| Kafka["Kafka\n(pkg/messaging, pkg/pdkafka)"]
```

`IAMServer` (`server.go`) embeds both `IAMCommandServer` and
`IAMQueryServer` to satisfy the legacy unified `IAMServiceServer`
interface — new callers should target the split
`IAMCommandService`/`IAMQueryService` for CQRS routing.

## gRPC API Surface

`IAMService` (unified, 78 RPCs) is the canonical proto definition;
`IAMCommandService` (47 RPCs) and `IAMQueryService` (32 RPCs, overlapping
on read-only methods like `CheckPermission`) are the CQRS-split surfaces
actually intended for new callers. Grouped by domain area:

| Domain area | Representative RPCs | Caller |
|---|---|---|
| Tenant/org lifecycle | `CreateTenant`, `CreateOrganization`, `EnsureRootOrganization`, `AttachTenantToOrganization`, `DetachTenantFromOrganization` | `onboarding` (tenant creation), platform admin UI |
| Organization membership | `AddOrganizationMember`, `RemoveOrganizationMember`, `ListOrganizationMembers` | `frontend/apps/iam` organizations section |
| Tenant membership | `AddTenantMember`, `AddTenantMemberByIdentity`, `RemoveTenantMember`, `ListTenantMembers`, `GetTenantMembership` | `onboarding`, `frontend/apps/iam` |
| Tenant invites | `CreateTenantInvite`, `RevokeTenantInvite`, `AcceptTenantInvite`, `ListTenantInvites` | `frontend/apps/shell` invite-accept flow, `frontend/apps/iam` |
| Permission decision | `CheckPermission`, `CheckPermissionForResource`, `CheckPlatformPermission`, `SimulateAccess` | Every backend service's inbound guard (`backoffice`, `partner`, `onboarding`), `frontend/apps/iam` trust-simulation page |
| Managed policy engine | `CreatePolicy`, `CreatePolicyVersion`, `SetDefaultPolicyVersion`, `DeletePolicyVersion`, `DeletePolicy`, `GetPolicy`, `ListPolicies`, `ListPolicyVersions`, `ListPolicyAttachments` | `frontend/apps/iam` policies section |
| Policy attachment (by principal) | `AttachTenantUserPolicy`/`Detach...`, `AttachPlatformUserPolicy`/`Detach...`, `AttachGroupPolicy`/`Detach...` | `frontend/apps/iam` |
| Inline policies (4 parallel families: platform user, tenant user, group, and role-adjacent) | `Put/Get/List/Delete{Platform,TenantUser,Group}InlinePolicy` | `frontend/apps/iam` |
| Permission boundaries | `Put/Get/Delete{Platform,TenantUser,Role}PermissionBoundary` | `frontend/apps/iam` |
| Groups | `CreateGroup`, `DeleteGroup`, `ListGroups`, `AddGroupMember`/`Remove...`, `ListGroupMembers` | `frontend/apps/iam` groups section |
| Role / role trust | `AddPlatformRole`, `RemovePlatformRole`, `ListPlatformRoles`, `PutRoleTrustPolicy`, `GetRoleTrustPolicy`, `DeleteRoleTrustPolicy`, `AssumeRole` | Platform admin, cross-account-style role assumption |
| Service control policies (org-level) | `AttachServiceControlPolicy`, `DetachServiceControlPolicy`, `ListServiceControlPolicies` | Platform admin |
| Directory | `ListDirectoryUsers`, `ListUserTenants`, `ListPermissions` | `frontend/apps/iam` directory section |

Errors: standard gRPC status codes: `NotFound` (unknown tenant/policy/role
ID), `InvalidArgument` (malformed input, e.g. `AssumeRole` with tenant ID
on a platform-scope role), `PermissionDenied` (`AssumeRole` trust
statement rejects, `CheckPermission`-style denials return `false` in the
response body, not a gRPC error — the RPC itself succeeds).

## C4: Sequences Per Usecase

### Create Tenant with Outbox

```mermaid
sequenceDiagram
    participant Admin as Admin / Caller
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant EventRelay as IAM CDC / Fallback Relay
    participant Kafka as Kafka
    participant AuthProj as Auth IAM Projection Worker
    participant AuthDB as Auth DB

    Admin->>IAM: CreateTenant(name, slug, owner)
    IAM->>IAMDB: insert tenant
    IAM->>IAMDB: upsert owner membership
    IAM->>IAMDB: append message_outbox(tenant.created)
    IAM-->>Admin: tenant created

    IAMDB->>EventRelay: stream outbox change or bounded fallback read
    EventRelay->>Kafka: publish tenant.created
    EventRelay->>IAMDB: mark published when using fallback relay

    AuthProj->>Kafka: consume tenant.created
    AuthProj->>AuthDB: upsert iam_tenants_projection
```

### Add Tenant Member

```mermaid
sequenceDiagram
    participant Operator as Operator
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant EventRelay as IAM CDC / Fallback Relay
    participant Kafka as Kafka
    participant AuthProj as Auth IAM Projection Worker
    participant AuthDB as Auth DB

    Operator->>IAM: AddMember(tenantID, userID, role)
    IAM->>IAMDB: upsert tenant_membership
    IAM->>IAMDB: append message_outbox(tenant.member.added)
    IAM-->>Operator: ok

    IAMDB->>EventRelay: stream outbox change or bounded fallback read
    EventRelay->>Kafka: publish tenant.member.added
    AuthProj->>Kafka: consume tenant.member.added
    AuthProj->>AuthDB: upsert iam_tenant_memberships_projection
```

### IAM Event Projected into Auth

```mermaid
sequenceDiagram
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant Relay as IAM CDC / Fallback Relay
    participant Kafka as Kafka
    participant AuthWorker as Auth IAM Projection Worker
    participant AuthDB as Auth DB

    IAM->>IAMDB: append message_outbox(policy.attached / tenant.member.added)
    IAMDB->>Relay: stream outbox change or bounded fallback read
    Relay->>Kafka: publish podzone.iam.events
    Relay->>IAMDB: mark published when using fallback relay

    AuthWorker->>Kafka: consume podzone.iam.events
    alt tenant.created
        AuthWorker->>AuthDB: upsert iam_tenants_projection
    else tenant.member.added
        AuthWorker->>AuthDB: upsert iam_tenant_memberships_projection
    else other event
        AuthWorker->>AuthWorker: ignore
    end
```

### Admin IAM Policy Attachment

```mermaid
sequenceDiagram
    participant UI as Admin IAM UI
    participant GW as gRPC Gateway
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant EventRelay as IAM CDC / Fallback Relay
    participant Kafka as Kafka

    UI->>GW: AttachTenantUserPolicy
    GW->>IAM: AttachTenantUserPolicy
    IAM->>IAMDB: write attachment
    IAM->>IAMDB: append message_outbox(policy.attached)
    IAM-->>GW: ok
    GW-->>UI: success

    IAMDB->>EventRelay: stream outbox change or bounded fallback read
    EventRelay->>Kafka: publish policy.attached
    EventRelay->>IAMDB: mark published when using fallback relay
```

### CheckPermission Evaluation (Normal Membership)

Not an assumed-role call. Precedence, from `domain/interactor/authz.go`:
identity + group policy statements are evaluated first; an **explicit
deny** short-circuits to `false` immediately. Otherwise the request must
also clear the user's permission boundary and the tenant's organization
service-control-policy (SCP) ceiling, then optionally get scoped down
further by any session policy statements attached to the caller's JWT.

```mermaid
sequenceDiagram
    participant Caller as Backend service
    participant IAM as IAM.CheckPermission
    participant DB as Postgres

    Caller->>IAM: CheckPermission(tenantID, userID, permission, resource)
    IAM->>DB: GetTenant, membership lookup (must be active)
    IAM->>DB: ListTenantUserStatements + ListTenantGroupStatements
    alt explicit deny matched
        IAM-->>Caller: false
    else statements allow
        IAM->>DB: evaluate permission boundary
        IAM->>DB: evaluate organization SCP
        alt session policy present on JWT
            IAM->>IAM: scope down by session statements
        end
        IAM-->>Caller: true/false
    else no matching statement (fall through to role evaluation)
        IAM->>DB: evaluate tenant role-attached policies
        IAM-->>Caller: true/false
    end
```

### AssumeRole

```mermaid
sequenceDiagram
    participant Caller as Caller
    participant IAM as IAM.AssumeRole
    participant DB as Postgres

    Caller->>IAM: AssumeRole(roleName, tenantID?, externalID?, servicePrincipal?, durationSeconds)
    IAM->>DB: GetRoleByName
    IAM->>IAM: validate scope: tenant role requires tenantID, platform role forbids it
    IAM->>DB: GetTrustPolicy(roleID)
    IAM->>IAM: canAssumeRole(userID, tenantID, externalID, servicePrincipal, trustStatements)
    alt trust denied
        IAM-->>Caller: ErrAssumeRoleDenied
    else trust allowed
        IAM->>IAM: clamp duration to (0, 12h], default 1h
        IAM-->>Caller: AssumedRole{roleID, roleScope, tenantID, expiresAt, ...}
    end
```

Later `CheckPermission` calls in an assumed-role context branch
differently (see the assumed-role check at the top of
`CheckPermissionForResource`): a tenant-scoped assumed role is rejected
outright if its `TenantID` doesn't match the request's tenant, otherwise
evaluation runs against the assumed role's own policy instead of the
caller's direct membership.

## Cross-Service Dependencies

Inbound (who calls IAM): `auth` (`SwitchActiveTenant` → `GetTenantMembership`
on projection miss), `onboarding` (tenant/store provisioning membership
checks), `backoffice` (`CheckPermission` on every GraphQL resolver via
`tenant_middleware.go`), `partner` (`CheckPermission` via its
`TenantAuthorizer`), `frontend/apps/iam` (all admin console operations).

Outbound (who IAM calls): `auth` gRPC, read-only, for user directory
lookups (`ListDirectoryUsers`) — IAM never writes to auth's tables. Kafka
publish via the outbox — see "IAM Event Projected into Auth" above; the
only consumer today is `internal/auth/controller/eventhandler/iamprojection/`.

For the target-state gap between this implemented API and where IAM is
meant to go (product-independent action catalog, identity-provider
neutrality), see [11-iam-platform.md](../../11-iam-platform.md) — not
re-described here.
