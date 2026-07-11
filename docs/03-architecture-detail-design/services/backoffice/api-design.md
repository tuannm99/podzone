# Backoffice Service — API Design

Parent: [Services Index](../README.md) · [Backoffice README](./README.md) · [DB Design](./db-design.md)

## C3: Component View

```mermaid
flowchart TB
    Schema["controller/graphql/schema/*.graphqls"]
    Resolver["controller/graphql/resolver/*.resolvers.go"]
    Mid["tenant_middleware.go (TenantMiddleware)"]

    Catalog["domain/catalog (ProductSetupDraftAggregate)"]
    Store["domain/store"]
    Order["domain/order"]
    Routing["domain/routing (OrderRoutingRepository)"]
    Fulfillment["domain/fulfillment"]
    Settlement["domain/settlement (SettlementRecord aggregate)"]
    Exception["domain/exception"]
    Activity["domain/activity"]
    Shared["domain/shared"]

    RepoRouting["infrastructure/repository/routing"]
    RepoCatalog["infrastructure/repository/catalog"]
    RepoStore["infrastructure/repository (store)"]
    PartnerDir["infrastructure/partnerdirectory"]
    Tenancy["runtime/tenancy"]

    Schema --> Resolver
    Resolver --> Mid
    Resolver --> Catalog
    Resolver --> Store
    Resolver --> Order
    Resolver --> Routing
    Resolver --> Fulfillment
    Resolver --> Settlement
    Resolver --> Exception
    Resolver --> Activity

    Catalog --> RepoCatalog
    Store --> RepoStore
    Routing --> RepoRouting
    Settlement --> RepoRouting
    Exception --> RepoRouting
    Activity --> RepoRouting

    Mid --> Tenancy
    Resolver --> PartnerDir
```

Domain subdomains confirmed from `internal/backoffice/domain/`: `activity`,
`catalog`, `exception`, `fulfillment`, `order`, `routing`, `settlement`,
`store` (plus `shared`, common code, not a subdomain). `catalog` and
`settlement` use explicit DDD aggregate roots with domain events
(`ProductSetupDraftAggregate.PromoteCandidate`, `SettlementRecord.
UpdateSettlement`/`UpdateIssueHandling`) — the others are simpler
CRUD-shaped usecases over the `routing` repository, which backs
`routed_orders`/`customer_orders`/`routed_order_activities` for all of
routing, fulfillment, settlement, exception, and activity concerns (see
[DB Design](./db-design.md) — one wide table set, several domain packages
reading/writing it).

## GraphQL API Surface

Schema files: `internal/backoffice/controller/graphql/schema/
{store,catalog,routing,common}.graphqls`. Full operation list already
enumerated in [README.md](./README.md#interfaces) — summarized here by
subdomain:

| Subdomain | Queries | Mutations |
|---|---|---|
| store | `stores`, `store` | `createStore`, `activateStore`, `deactivateStore` |
| catalog (product setup) | `productSetupSnapshot` | `createProductSetupDraft`, `promoteProductSetupCandidate`, `updateProductSetupCandidateStatus` |
| routing / order | `routedOrders`, `routedOrderRecommendation` | `createRoutedOrder`, `forceRerouteBlockedOrder`, `advanceRoutedOrder`, `bulkUpdateRoutedOrders` |
| exception | — | `openOrderException`, `updateOrderExceptionStatus` |
| fulfillment | — | `updateOrderShipment` |
| settlement | — | `updateOrderSettlement`, `updateOrderIssueHandling` |
| activity | `routedOrderActivities` | — |
| operator queue | — | `updateOrderQueueControl` |

All mutations/queries go through `TenantMiddleware` (`InterceptOperation` +
per-field `InterceptField`) before reaching a resolver — see Runtime Flows
in [README.md](./README.md#runtime-flows) for that request path, not
repeated here.

## C4: Sequences Per Usecase

### Create Routed Order Recommendation → Create Routed Order

Reused from the former `docs/02-architecture-overall/05-sequences.md`,
corrected: `Create` in `infrastructure/repository/routing/repository.go`
inserts into **both** `routed_orders` and `customer_orders` in the same
call (dual-write, not a `routed_orders`-only legacy path) — the given
diagram's single "persist routed order + activity" step undersold this.

```mermaid
sequenceDiagram
    participant UI as Backoffice UI
    participant BO as Backoffice Service
    participant Partner as Partner Service
    participant IAM as IAM Service
    participant BODB as Backoffice DB (Postgres)

    UI->>BO: routedOrderRecommendation(input)
    BO->>IAM: check tenant permission
    IAM-->>BO: allow
    BO->>Partner: list/find partner capabilities
    Partner-->>BO: partner capabilities
    BO->>BO: evaluate eligibility, priority, SLA, margin snapshot
    BO-->>UI: recommendation

    UI->>BO: createRoutedOrder(selected recommendation)
    BO->>BODB: INSERT routed_orders (legacy table, still written)
    BO->>BODB: INSERT customer_orders (current aggregate, aggregate_version set)
    BO->>BODB: INSERT routed_order_activities (audit entry)
    BO-->>UI: routed order created
```

### Product Setup Draft → Candidate Promotion

```mermaid
sequenceDiagram
    participant UI as Backoffice UI
    participant Resolver as catalog.resolvers.go
    participant Agg as ProductSetupDraftAggregate
    participant RepoCatalog as catalog repository (Postgres)

    UI->>Resolver: promoteProductSetupCandidate(input)
    Resolver->>RepoCatalog: load draft aggregate by ID
    RepoCatalog-->>Resolver: ProductSetupDraftAggregate
    Resolver->>Agg: PromoteCandidate(cmd)
    Agg->>Agg: validate draft status, build candidate, raise ProductSetupCandidatePromoted event
    Agg-->>Resolver: candidate + events
    Resolver->>RepoCatalog: persist candidate (product_setup_candidates), mark draft promoted
    RepoCatalog-->>Resolver: ok
    Resolver-->>UI: candidate
```

### Open Order Exception → Update Exception Status

```mermaid
sequenceDiagram
    participant UI as Backoffice UI
    participant Resolver as routing.resolvers.go
    participant RepoRouting as routing repository (Postgres)

    UI->>Resolver: openOrderException(input)
    Resolver->>RepoRouting: load order (customer_orders)
    RepoRouting-->>Resolver: order
    Resolver->>Resolver: set exception_type, exception_status = open
    Resolver->>RepoRouting: UPDATE customer_orders + routed_orders, INSERT activity entry
    RepoRouting-->>Resolver: ok
    Resolver-->>UI: updated order

    UI->>Resolver: updateOrderExceptionStatus(input)
    Resolver->>RepoRouting: UPDATE exception_status, append activity entry
    Resolver-->>UI: updated order
```

### Settlement Update (DDD Aggregate)

```mermaid
sequenceDiagram
    participant UI as Backoffice UI
    participant Resolver as routing.resolvers.go
    participant Agg as SettlementRecord aggregate
    participant RepoRouting as routing repository (Postgres)

    UI->>Resolver: updateOrderSettlement(input)
    Resolver->>RepoRouting: rehydrate SettlementRecord from order row (RehydrateSettlementRecord)
    RepoRouting-->>Resolver: SettlementRecord snapshot
    Resolver->>Agg: UpdateSettlement(cmd)
    Agg->>Agg: compute realized_margin, raise SettlementUpdated event, append timeline entry
    Agg-->>Resolver: updated snapshot + events
    Resolver->>RepoRouting: persist snapshot (settlement_status, settlement_notes, realized_margin)
    Resolver-->>UI: updated order
```

`updateOrderIssueHandling` follows the same aggregate-rehydrate-persist
shape via `SettlementRecord.UpdateIssueHandling` — not diagrammed
separately, same pattern.

## Cross-Service Dependencies

| Direction | Target/Caller | Protocol | Purpose |
|---|---|---|---|
| Outbound | `auth` service | gRPC (`AuthServiceClient.GetSession`) | Session validation in `TenantMiddleware` |
| Outbound | `iam` service | gRPC (`GetTenantMembership`, `CheckPermission`) | Tenant membership + per-field permission checks |
| Outbound | `partner` service | gRPC (`infrastructure/partnerdirectory/adapter.go`) | Partner capability/directory read-through |
| Outbound | Postgres (tenant-routed) | `pkg/pdtenantdb` | All domain reads/writes — DB route resolved from the KV projection onboarding publishes |
| Inbound | Frontend (`frontend/apps/backoffice`) | GraphQL over HTTPS, via APISIX `/backoffice/graphql` → rewritten to `/query` | Only inbound caller — see
[knowledge base: backoffice GraphQL 404](../../../10-knowledge-base/local-dev/2026-07-11-backoffice-graphql-404.md)
for why that route exists instead of hitting the service's own port directly |

No inbound gRPC/REST — this service has no server-to-server callers besides
the frontend through the gateway.
