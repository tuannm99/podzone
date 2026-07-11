# Partner Service — API Design

Parent: [Services Index](../README.md) · [Partner README](./README.md) · [DB Design](./db-design.md)

## C3: Component View

```mermaid
flowchart TB
    subgraph Controller["controller/grpchandler"]
        PartnerServer["partner_server.go\n(PartnerServer)"]
        Authz["authz.go\n(authTenantAuthorizer)"]
    end

    subgraph Domain["domain"]
        Interactor["partner interactor\n(CRUD, normalization)"]
    end

    subgraph Infra["infrastructure"]
        Repo["repository\n(Postgres, pkg/pdsql + sqlx)"]
    end

    PartnerServer --> Authz
    Authz -->|"gRPC"| AuthSvc["Auth Service"]
    Authz -->|"gRPC"| IAMSvc["IAM Service"]
    PartnerServer --> Interactor
    Interactor --> Repo
    Repo --> DB["Postgres: partners"]
```

Two controller files: `partner_server.go` maps transport for all 5 RPCs,
`authz.go` is a separate `TenantAuthorizer` component every RPC calls first
(see [README](./README.md) Runtime Flows). No separate domain package split
by subdomain — this service is small enough that CRUD + normalization logic
lives directly behind the one interactor.

## gRPC API Surface

Full request/response field detail already documented in
[README's Interfaces table](./README.md#interfaces) — summarized here for
the C3/C4 pairing:

| RPC | Permission | Notes |
|---|---|---|
| `CreatePartner` | `partner:manage` | |
| `GetPartner` | `partner:read` | Fetches record first, authorizes against its actual `tenant_id`. |
| `ListPartners` | `partner:read` | Paginated (`pkg/collection`). Backoffice calls this in a loop to build its partner directory — see Cross-Service Dependencies. |
| `UpdatePartner` | `partner:manage` | Fetch-then-authorize, same as `GetPartner`. |
| `UpdatePartnerStatus` | `partner:manage` | Activate/deactivate. |

No separate "list active partners" or "query capabilities" RPC exists —
confirmed against `internal/partner/controller/grpchandler/partner_server.go`
(exactly these 5 methods) and
`internal/backoffice/infrastructure/partnerdirectory/adapter.go`
(`ListActivePartners`, which is a **backoffice-side** domain method, not a
partner-service RPC — it pages through `ListPartners` and filters
client-side).

## C4: Sequences Per Usecase

### Tenant Authorization (every RPC)

```mermaid
sequenceDiagram
    participant Caller as Caller (backoffice / gateway)
    participant Partner as Partner Service
    participant Auth as Auth Service
    participant IAM as IAM Service

    Caller->>Partner: gRPC call + Bearer JWT (metadata)
    Partner->>Partner: parse JWT (HS256, partner.auth.jwt_secret)
    Partner->>Partner: validate claims.Key == partner.auth.jwt_key
    Partner->>Partner: require active_tenant_id, session_id, user_id present
    Partner->>Auth: GetSession(sessionId) [forwards original auth header]
    Auth-->>Partner: session (status, userId, tenantId)
    Partner->>Partner: reject if session inactive or user/tenant mismatch
    Partner->>IAM: GetTenantMembership(tenantId, userId)
    IAM-->>Partner: membership (status)
    Partner->>Partner: reject if membership not active
    Partner->>IAM: CheckPermission(tenantId, userId, permission)
    IAM-->>Partner: allowed (bool)
    alt allowed
        Partner->>Partner: proceed to usecase
    else denied
        Partner-->>Caller: PermissionDenied
    end
```

`GetPartner`/`UpdatePartner`/`UpdatePartnerStatus` fetch the record by ID
*before* running this flow, then authorize against the record's actual
`tenant_id` — the caller cannot claim a different tenant to bypass the
check (`partner_server.go`).

This is the only genuinely distinct C4 flow in this service — every RPC
funnels through it, and the 5 RPCs themselves are straightforward CRUD with
no other branching business logic worth a separate diagram.

## Cross-Service Dependencies

**Outbound** (partner → other services, always for authorization, never for
data):

- `auth` gRPC — `GetSession`, on every call.
- `iam` gRPC — `GetTenantMembership` + `CheckPermission`, on every call.

**Inbound** (who calls partner):

- `backoffice` via `internal/backoffice/infrastructure/partnerdirectory/adapter.go`
  (`routingctx.PartnerDirectory` port) — pages through `ListPartners` to
  build a partner routing directory during order routing recommendation.
  See
  [`../backoffice/api-design.md`](../backoffice/api-design.md) "Create
  Routed Order Recommendation" for the consuming sequence.
- API gateway (`internal/grpcgateway`) forwards all 5 RPCs for
  HTTP-facing callers — no gateway-side logic beyond transcoding.
