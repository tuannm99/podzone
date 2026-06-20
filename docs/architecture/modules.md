# C3: Module View

## Auth Service

```mermaid
flowchart LR
    AuthServer["controller/grpchandler"]
    AuthDomain["domain"]
    IAMClient["infrastructure/iamclient"]
    IAMService["IAMService gRPC"]
    AuthRepo["infrastructure/repository"]
    ProjectionHandler["controller/eventhandler/iamprojection"]
    ProjectionRuntime["infrastructure/messaging/iamprojection"]
    AuthAPI["cmd/auth"]
    AuthWorker["cmd/auth-worker"]
    AuthDB["auth DB"]
    Redis["redis"]
    Kafka["kafka"]

    AuthAPI --> AuthServer
    AuthServer --> AuthDomain
    AuthDomain --> AuthRepo
    AuthDomain --> IAMClient
    IAMClient -->|"gRPC"| IAMService
    AuthRepo --> AuthDB
    AuthRepo --> Redis
    AuthWorker --> ProjectionRuntime
    ProjectionRuntime --> Kafka
    ProjectionRuntime --> ProjectionHandler
    ProjectionHandler --> AuthDB
```

### Main modules

- `domain`: login, register, refresh token, switch tenant, session policy, assume-role session state
- `infrastructure/iamclient`: synchronous calls to `IAMService`
- `controller/eventhandler/iamprojection`: inbound Kafka event handler for IAM-derived projection updates
- `infrastructure/messaging/iamprojection`: consumer runtime, inbox/idempotency wiring, and worker lifecycle
- `cmd/auth`: Auth API runtime
- `cmd/auth-worker`: projection-only runtime

## IAM Service

```mermaid
flowchart LR
    IAMServer["controller/grpchandler"]
    IAMCommandHandler["controller/grpchandler command handler"]
    IAMQueryHandler["controller/grpchandler query handler"]
    IAMCommand["domain/inputport command usecases"]
    IAMQuery["domain/inputport query usecases"]
    IAMInteractor["domain/interactor"]
    IAMWriteRepo["infrastructure/repository write model"]
    IAMReadRepo["infrastructure/repository read/evaluation model"]
    EventRelay["outbox CDC / fallback relay"]
    IAMAPI["cmd/iam"]
    IAMWorker["cmd/iam-worker"]
    IAMDB["iam DB"]
    Kafka["kafka"]

    IAMAPI --> IAMServer
    IAMServer --> IAMCommandHandler
    IAMServer --> IAMQueryHandler
    IAMCommandHandler --> IAMCommand
    IAMCommandHandler --> IAMQuery
    IAMQueryHandler --> IAMQuery
    IAMCommand --> IAMInteractor
    IAMQuery --> IAMInteractor
    IAMInteractor --> IAMWriteRepo
    IAMInteractor --> IAMReadRepo
    IAMWriteRepo --> IAMDB
    IAMReadRepo --> IAMDB
    IAMWorker --> EventRelay
    IAMInteractor -->|"append outbox"| IAMDB
    IAMDB -->|"CDC preferred; bounded polling fallback"| EventRelay
    EventRelay -->|"publish"| Kafka
```

### Main modules

- `entity`: authorization core types, policy statements, trust policies, memberships
- `inputport`: IAM command and query usecase contracts; `IAMUsecase` remains a compatibility facade during migration
- `outputport`: command repository, query repository, and outbox contracts
- `controller/grpchandler`: gRPC facade delegates to separate command and query handlers while preserving the public IAM service contract
- `api/proto/iam/v1/iam_service.proto`: exposes `IAMCommandService` and `IAMQueryService` for CQRS gRPC clients; `IAMService` remains the REST/gateway compatibility contract
- IAM currently runs as one API binary, but the module exposes separate command/query server registrations so it can become `cmd/iam-command` and `cmd/iam-query` later without changing proto contracts
- Fx wiring mirrors that boundary: `internal/iam.CommandModule` and `internal/iam.QueryModule` wire domain dependencies separately, while `internal/iam/server.Module` keeps the current all-in-one `cmd/iam` runtime
- `interactor`: command handling, policy lifecycle, authz evaluation, groups, tenants, org/SCP, assume-role
- command side owns tenant, policy, group, membership, org, and boundary mutations
- query side owns policy reads, membership reads, permission checks, simulations, and read-model access
- `cmd/iam`: IAM API runtime
- `cmd/iam-worker`: transactional event publisher runtime; polling relay is fallback until CDC is wired

## Backoffice Service

```mermaid
flowchart LR
    GraphQL["controller/graphql"]
    Catalog["domain/catalog"]
    Routing["domain/routing"]
    Store["domain/store"]
    Repo["infrastructure/repository"]
    PartnerDir["infrastructure/partnerdirectory"]
    BOAuthz["tenant middleware/authz"]
    BODB["backoffice DB"]

    GraphQL --> BOAuthz
    GraphQL --> Catalog
    GraphQL --> Routing
    GraphQL --> Store
    Catalog --> Repo
    Routing --> Repo
    Store --> Repo
    Routing --> PartnerDir
    Repo --> BODB
```

### Main modules

- `store`: store aggregate, store-scoped access, and workspace-facing store metadata
- `catalog`: product setup draft/candidate flow inside the active store context
- `routing`: routed order aggregate, recommendation, shipment, settlement, and audit feed
- Backoffice DDD boundary: GraphQL maps transport DTOs to context usecases; repositories stay behind each context output port.
- Cross-context calls should use domain ports or external adapters, not direct repository imports.

## Partner Service

```mermaid
flowchart LR
    PartnerServer["controller/grpchandler"]
    PartnerDomain["domain"]
    PartnerRepo["infrastructure/repository"]
    PartnerDB["partner DB"]

    PartnerServer --> PartnerDomain
    PartnerDomain --> PartnerRepo
    PartnerRepo --> PartnerDB
```

### Main modules

- `domain`: partner profile, capabilities, cost rules, operational settings
- `controller/grpchandler`: gRPC transport surface for partner management
- `infrastructure/repository`: SQL persistence

## Onboarding Service

```mermaid
flowchart LR
    HTTP["controller/http"]
    InfraInteractor["domain/infrasmanager/interactor"]
    StoreInteractor["domain/store/interactor"]
    InfraPorts["domain/infrasmanager/outputport"]
    StorePorts["domain/store/outputport"]
    MongoStore["infrastructure/repository/infrasmanager"]
    StoreRepo["infrastructure/repository/store"]
    Provider["infrastructure/provisioning/provider"]
    ConsulBridge["controller/eventhandler/consulbridge"]
    Messaging["infrastructure/messaging"]
    Mongo["Mongo onboarding DB"]
    Consul["Consul router projection"]

    HTTP --> InfraInteractor
    HTTP --> StoreInteractor
    StoreInteractor --> InfraInteractor
    InfraInteractor --> InfraPorts
    StoreInteractor --> StorePorts
    InfraPorts --> MongoStore
    InfraPorts --> Provider
    StorePorts --> StoreRepo
    MongoStore --> Mongo
    StoreRepo --> Mongo
    MongoStore -->|"placement/outbox records"| Messaging
    Messaging --> ConsulBridge
    ConsulBridge --> Consul
```

### Main modules

- `domain/infrasmanager`: placement allocation, connection publication, and infrastructure manager usecases
- `domain/store`: store request lifecycle, approval state, and provisioning orchestration
- `infrastructure/repository`: Mongo-backed repositories for store requests, connection events, and placement allocations
- `infrastructure/provisioning/provider`: runtime-specific placement provider for local Docker, Kubernetes, and future Terraform/cloud
- `controller/eventhandler/consulbridge`: router projection publisher; Consul is rebuilt from onboarding allocation state
- `infrastructure/messaging`: CDC/fallback publisher and background worker wiring

Local Docker and Kubernetes schema-mode placement use the `podzone_tenants` Postgres database for tenant schemas.
The `postgres` database remains the admin/default connection database and must not host service-owned public tables.

## Catalog Service

```mermaid
flowchart LR
    CatalogServer["controller/grpchandler"]
    CatalogDomain["domain"]
    CatalogInfra["infrastructure"]
    CatalogDB["catalog DB"]

    CatalogServer --> CatalogDomain
    CatalogDomain --> CatalogInfra
    CatalogInfra --> CatalogDB
```

### Main modules

- current repo shape keeps `catalog` lighter than `backoffice/catalog`
- it mainly exposes gRPC APIs and persistence for catalog-facing workflows

## Gateway and gRPC Gateway

```mermaid
flowchart LR
    APISIX["internal/gateway (APISIX config)"]
    GatewayRegistrar["internal/grpcgateway"]
    Proto["pkg/api/proto"]
    Services["Auth / IAM / Catalog / Partner gRPC"]
    UI["internal/ui-podzone"]

    UI --> APISIX
    GatewayRegistrar --> Proto
    GatewayRegistrar --> Services
    APISIX --> GatewayRegistrar
```

### Main modules

- `internal/gateway`: APISIX runtime config
- `deployments/docker/apisix-init`: local seed for APISIX services, routes, and sample JWT edge plugin
- `internal/grpcgateway`: service registration and HTTP translation
- `pkg/api/proto`: generated contracts shared by transport layers

## Seller Portal UI

```mermaid
flowchart LR
    Router["solid/app-router.tsx"]
    Pages["pages/podzone/*"]
    Components["components/common/*"]
    Services["src/services/*"]
    Gateway["HTTP + GraphQL endpoints"]

    Router --> Pages
    Pages --> Components
    Pages --> Services
    Services --> Gateway
```

### Main modules

- `solid/app-router.tsx`: auth/admin/tenant route ownership
- `pages/podzone/*`: page-level application flows
- `services/*`: API adapters for Auth, IAM, Partner, GraphQL Backoffice
