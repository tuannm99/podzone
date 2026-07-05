# C2: Container View

```mermaid
flowchart TB
    subgraph Edge
        UI["UI / Seller Portal"]
        STOREFRONT["Storefront App"]
        APISIX["APISIX / Gateway"]
        GRPCGW["gRPC Gateway"]
    end

    subgraph Services
        AUTH["Auth Service"]
        IAM["IAM Service"]
        BACKOFFICE["Backoffice Service"]
        CATALOG["Catalog Service"]
        PARTNER["Partner Service"]
        ONBOARDING["Onboarding Service"]
    end

    subgraph Provisioning
        PROVWORKER["Onboarding Provisioning Worker"]
        PROVIDER["Placement Provider\nDocker / Kubernetes / Terraform"]
    end

    subgraph Infra
        AUTHDB["Postgres: auth"]
        IAMDB["Postgres: iam"]
        BACKOFFICEDB["Postgres: tenant data pools"]
        CATALOGDB["Postgres: catalog"]
        PARTNERDB["Postgres: partner"]
        ONBOARDINGDB["Mongo: onboarding"]
        REDIS["Redis"]
        KAFKA["Kafka"]
        RUNTIME_KV["Mongo runtime_kv"]
    end

    UI --> APISIX
    STOREFRONT --> APISIX
    APISIX --> GRPCGW
    APISIX --> BACKOFFICE
    APISIX --> ONBOARDING

    GRPCGW --> AUTH
    GRPCGW --> IAM
    GRPCGW --> CATALOG
    GRPCGW --> PARTNER

    BACKOFFICE --> AUTH
    BACKOFFICE --> IAM
    BACKOFFICE --> PARTNER

    AUTH --> AUTHDB
    AUTH --> REDIS
    AUTH --> KAFKA

    IAM --> IAMDB
    IAM --> KAFKA

    BACKOFFICE --> BACKOFFICEDB
    BACKOFFICE --> KAFKA

    CATALOG --> CATALOGDB
    PARTNER --> PARTNERDB
    PARTNER --> KAFKA
    ONBOARDING --> ONBOARDINGDB
    ONBOARDING --> KAFKA
    ONBOARDING --> PROVWORKER
    PROVWORKER --> PROVIDER
    PROVIDER --> BACKOFFICEDB
    PROVWORKER --> ONBOARDINGDB
    PROVWORKER --> RUNTIME_KV
    BACKOFFICE --> RUNTIME_KV
```

## Boundaries

- `auth` and `iam` are separated by API and database ownership.
- `auth` uses `IAMService` over `gRPC` for synchronous authorization-sensitive operations.
- `iam` publishes commit-coupled domain events through Kafka outbox/CDC.
- `auth` consumes a subset of IAM events into a local projection.
- `grpcgateway` remains the HTTP translation layer for gRPC services.
- `backoffice` is GraphQL-first and talks to service APIs plus its own database.
- `onboarding` owns store provisioning requests and the placement allocation source of truth.
- `runtime_kv` is a Mongo-backed router projection consumed by `pdtenantdb`, not the source of truth for placement.
- placement providers create or bind the tenant storage target for Docker, Kubernetes, or future Terraform/cloud runtimes.
