# Deployment

## Local Docker Runtime

```mermaid
flowchart TB
    subgraph Browser
        UI["Seller Portal UI :3000"]
    end

    subgraph Edge
        APISIX["APISIX :9080"]
        GRPCGW["gRPC Gateway :8080"]
        BO["Backoffice :8000"]
        ONB["Onboarding HTTP :8800"]
    end

    subgraph Services
        AUTH["Auth API gRPC :50051"]
        AUTHW["Auth Worker"]
        CATALOG["Catalog gRPC :50052"]
        IAM["IAM API gRPC :50053"]
        IAMW["IAM Worker"]
        PARTNER["Partner gRPC :50054"]
    end

    subgraph Infra
        PG["Postgres / PgBouncer :5432/:6432"]
        REDIS["Redis :6379"]
        MONGO["Mongo :27017"]
        CONSUL["Consul :8500"]
        KAFKA["Kafka KRaft :9092"]
        ETCD["etcd :2379"]
    end

    UI --> APISIX
    APISIX --> GRPCGW
    APISIX --> BO
    APISIX --> ONB

    GRPCGW --> AUTH
    GRPCGW --> CATALOG
    GRPCGW --> IAM
    GRPCGW --> PARTNER

    BO --> IAM
    BO --> PARTNER
    BO --> PG
    BO --> KAFKA

    AUTH --> PG
    AUTH --> REDIS
    AUTHW --> PG
    AUTHW --> KAFKA
    IAM --> PG
    IAM --> AUTH
    IAMW --> PG
    IAMW --> KAFKA
    CATALOG --> PG
    PARTNER --> PG
    PARTNER --> KAFKA
    ONB --> MONGO
    ONB --> CONSUL

    APISIX --> ETCD
```

## Production-Target Runtime Shape

```mermaid
flowchart LR
    Ingress["Ingress / API Gateway"]
    Router["Tenant Routing Layer"]
    BOPoolA["Backoffice Pool A"]
    BOPoolB["Backoffice Pool B"]
    UI["Seller Portal"]
    Gateway["gRPC Gateway"]
    Auth["Auth Service"]
    AuthWorker["Auth Projection Worker"]
    IAM["IAM Service"]
    IAMWorker["IAM Outbox Worker"]
    Backoffice["Backoffice Service"]
    Partner["Partner Service"]
    Catalog["Catalog Service"]
    Onboarding["Onboarding Service"]
    Kafka["Kafka Cluster"]
    Redis["Redis"]
    SQL["Service-owned Postgres DBs"]
    Consul["Consul"]
    Mongo["Mongo"]

    UI --> Ingress
    Ingress --> Gateway
    Ingress --> Router
    Router --> BOPoolA
    Router --> BOPoolB
    Ingress --> Onboarding

    Gateway --> Auth
    Gateway --> IAM
    Gateway --> Partner
    Gateway --> Catalog

    BOPoolA --> Backoffice
    BOPoolB --> Backoffice

    Backoffice --> IAM
    Backoffice --> Partner

    Auth --> SQL
    Auth --> Redis
    AuthWorker --> SQL
    AuthWorker --> Kafka
    IAM --> SQL
    IAM --> Auth
    IAMWorker --> SQL
    IAMWorker --> Kafka
    Partner --> SQL
    Partner --> Kafka
    Catalog --> SQL
    Backoffice --> SQL
    Backoffice --> Kafka
    Onboarding --> Mongo
    Onboarding --> Consul
```

## Backoffice Tenant Routing Model

```mermaid
flowchart TB
    Browser["Browser / Seller UI"]
    Edge["Ingress / Gateway / LB"]
    PoolPlacement["Tenant -> Backoffice Pool Placement"]
    BOPool["Backoffice Runtime Pool"]
    TenantRuntime["Backoffice Tenancy Runtime"]
    DBPlacement["Tenant -> DB Placement"]
    StoreScope["Store-scoped Operation"]
    TenantDB["Tenant DB / Schema"]

    Browser --> Edge
    Edge --> PoolPlacement
    PoolPlacement --> BOPool
    BOPool --> TenantRuntime
    TenantRuntime --> DBPlacement
    DBPlacement --> TenantDB
    TenantRuntime --> StoreScope
    StoreScope --> TenantDB
```

## Notes

- Local runtime mirrors the target service split with simplified single-node infrastructure.
- `APISIX` fronts both GraphQL and HTTP/gRPC-gateway surfaces.
- Local Docker now includes a one-shot `apisix-init` seed step for routes, services, shared plugins, and a JWT edge probe route.
- Kafka is the single async backbone target for service integration.
- Production follow-up should move APISIX bootstrap into:
  - Kubernetes Job / Helm hook
  - Terraform-managed Admin API resources
  - environment-specific route/plugin manifests
- `auth` and `iam` now run separate API and worker binaries in local Docker.
- Projection and outbox workers no longer share the gRPC server runtime by default.
- `Backoffice` should be treated as stateless API capacity inside tenant-assigned runtime pools.
- Tenant-to-runtime-pool routing belongs to edge/runtime placement, not to business handlers.
- Tenant-to-database placement belongs to `pdtenantdb` and application runtime placement resolution.

## Kubernetes Direction

- Each service keeps its own Deployment and DB binding.
- `auth-api`, `auth-worker`, `iam-api`, and `iam-worker` should be modeled as separate Deployments or ECS services.
- Shared code can still live in one repo and one bounded context, but runtime scaling is now independent.
- `backoffice` should support sharded runtime pools where:
  - ingress routes tenant traffic into the correct pool
  - multiple pods can exist within one pool
  - pod count can scale independently from tenant placement metadata
- tenant placement metadata should remain externalized so scale-in and rescheduling do not require application rebinding.

## Terraform / AWS Future

- Keep service runtime bootstrap declarative enough to migrate later into:
  - MSK topics and ACLs
  - RDS per-service databases
  - ECS/EKS service modules
  - APISIX or alternative gateway Admin API seed resources
- The current Docker seed scripts should be treated as the local equivalent of future infra bootstrap modules, not as the final production mechanism.
