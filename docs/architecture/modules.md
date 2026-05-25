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
    AuthDB["auth DB"]
    Redis["redis"]
    Kafka["kafka"]

    AuthServer --> AuthDomain
    AuthDomain --> AuthRepo
    AuthDomain --> IAMClient
    IAMClient -->|"gRPC"| IAMService
    AuthRepo --> AuthDB
    AuthRepo --> Redis
    ProjectionRuntime --> Kafka
    ProjectionRuntime --> ProjectionHandler
    ProjectionHandler --> AuthDB
```

### Main modules

- `domain`: login, register, refresh token, switch tenant, session policy, assume-role session state
- `infrastructure/iamclient`: synchronous calls to `IAMService`
- `controller/eventhandler/iamprojection`: inbound Kafka event handler for IAM-derived projection updates
- `infrastructure/messaging/iamprojection`: consumer runtime, inbox/idempotency wiring, and worker lifecycle

## IAM Service

```mermaid
flowchart LR
    IAMServer["controller/grpchandler"]
    IAMInteractor["interactor"]
    IAMRepo["infrastructure/repository"]
    OutboxWorker["worker/outbox"]
    IAMDB["iam DB"]
    Kafka["kafka"]

    IAMServer --> IAMInteractor
    IAMInteractor --> IAMRepo
    IAMRepo --> IAMDB
    IAMInteractor -->|"append outbox"| IAMDB
    OutboxWorker -->|"poll outbox"| IAMDB
    OutboxWorker -->|"publish"| Kafka
```

### Main modules

- `entity`: authorization core types, policy statements, trust policies, memberships
- `inputport`: IAM usecase contracts
- `outputport`: repository and outbox contracts
- `interactor`: policy lifecycle, authz evaluation, groups, tenants, org/SCP, assume-role

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

- `catalog`: product setup draft/candidate flow
- `routing`: routed orders, recommendation, shipment, settlement, audit feed
- `store`: tenant store metadata and store-owned operations

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
    InfraManager["infrasmanager/core"]
    StoreSvc["store"]
    MongoStore["mongo store / eventstore"]
    ConsulPub["consul publisher"]
    OutboxWorker["outbox worker"]

    HTTP --> InfraManager
    HTTP --> StoreSvc
    InfraManager --> MongoStore
    InfraManager --> ConsulPub
    OutboxWorker --> MongoStore
    OutboxWorker --> ConsulPub
```

### Main modules

- `infrasmanager/core`: connection lifecycle and outbox/event storage
- `store`: onboarding-facing store CRUD
- `core/worker`: outbox publisher to Consul

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
