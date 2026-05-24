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

    subgraph Infra
        AUTHDB["Postgres: auth"]
        IAMDB["Postgres: iam"]
        BACKOFFICEDB["Postgres: backoffice"]
        CATALOGDB["Postgres: catalog"]
        PARTNERDB["Postgres: partner"]
        REDIS["Redis"]
        KAFKA["Kafka"]
        CONSUL["Consul"]
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
    ONBOARDING --> CONSUL
    ONBOARDING --> BACKOFFICEDB
```

## Boundaries

- `auth` and `iam` are separated by API and database ownership.
- `auth` uses `IAMService` over `gRPC` for synchronous authorization-sensitive operations.
- `iam` publishes domain events through Kafka outbox.
- `auth` consumes a subset of IAM events into a local projection.
- `grpcgateway` remains the HTTP translation layer for gRPC services.
- `backoffice` is GraphQL-first and talks to service APIs plus its own database.
