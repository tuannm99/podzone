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
        AUTH["Auth gRPC :50051"]
        CATALOG["Catalog gRPC :50052"]
        IAM["IAM gRPC :50053"]
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
    AUTH --> KAFKA
    IAM --> PG
    IAM --> KAFKA
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
    UI["Seller Portal"]
    Gateway["gRPC Gateway"]
    Auth["Auth Service"]
    IAM["IAM Service"]
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
    Ingress --> Backoffice
    Ingress --> Onboarding

    Gateway --> Auth
    Gateway --> IAM
    Gateway --> Partner
    Gateway --> Catalog

    Backoffice --> IAM
    Backoffice --> Partner

    Auth --> SQL
    Auth --> Redis
    Auth --> Kafka
    IAM --> SQL
    IAM --> Kafka
    Partner --> SQL
    Partner --> Kafka
    Catalog --> SQL
    Backoffice --> SQL
    Backoffice --> Kafka
    Onboarding --> Mongo
    Onboarding --> Consul
```

## Notes

- Local runtime mirrors the target service split with simplified single-node infrastructure.
- `APISIX` fronts both GraphQL and HTTP/gRPC-gateway surfaces.
- Kafka is the single async backbone target for service integration.
