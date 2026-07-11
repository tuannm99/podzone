# Shared Runtime and Platform Modules

## Runtime Foundation

```mermaid
flowchart LR
    PDConfig["pkg/pdconfig"]
    PDLog["pkg/pdlog"]
    PDSQL["pkg/pdsql"]
    PDRedis["pkg/pdredis"]
    PDKafka["pkg/pdkafka"]
    PDGRPC["pkg/pdgrpc"]
    PDGRPCClient["pkg/pdgrpcclient"]
    PDGraphQL["pkg/pdgraphql"]
    PDTenantDB["pkg/pdtenantdb"]
    PDMessaging["pkg/messaging"]
    PDWorker["pkg/pdworker"]

    PDConfig --> PDSQL
    PDConfig --> PDRedis
    PDConfig --> PDKafka
    PDConfig --> PDGRPC
    PDLog --> PDSQL
    PDLog --> PDRedis
    PDLog --> PDKafka
    PDKafka --> PDMessaging
    PDWorker --> PDMessaging
    PDGRPCClient --> PDGRPC
```

## Package Intent

- `pkg/pdconfig`: Koanf-backed configuration bootstrap
- `pkg/pdlog`: logging abstraction and Fx module
- `pkg/pdsql`: named SQL connection modules per service
- `pkg/pdredis`: named Redis connection modules
- `pkg/pdkafka`: Kafka client/admin/consumer-group wiring with Sarama
- `pkg/messaging`: envelope, outbox, publisher/consumer contracts, Kafka adapters
- `pkg/pdgrpc` and `pkg/pdgrpcclient`: gRPC server/client lifecycle
- `pkg/pdgraphql`: GraphQL runtime wiring
- `pkg/pdtenantdb`: tenant placement and multi-tenant DB resolution
- `pkg/pdworker`: long-running worker lifecycle abstraction

## Tenant Runtime Routing

```mermaid
sequenceDiagram
    participant Edge as Edge / LB
    participant Pool as Backoffice Pool
    participant RT as Backoffice Tenancy Runtime
    participant Place as pdtenantdb PlacementResolver
    participant Repo as Repository
    participant DB as Tenant DB / Schema

    Edge->>Pool: route request using tenant placement
    Pool->>RT: resolve tenant + store runtime scope
    RT->>Place: resolve tenant placement
    Place-->>RT: cluster/db/schema
    RT-->>Pool: runtime scope ready
    Pool->>Repo: execute store-scoped operation
    Repo->>Place: resolve tenant placement (via manager)
    Place-->>Repo: cluster/db/schema
    Repo->>DB: query in tenant placement
```

Notes:

- edge/runtime routing decides which backoffice runtime pool receives tenant traffic
- application runtime placement decides which DB/schema receives tenant traffic
- store scope lives inside the resolved tenant placement

## Data and Event Backbone

```mermaid
flowchart TB
    ServiceDB["Service DB"]
    Outbox["message_outbox"]
    Relay["CDC Connector / Fallback Relay"]
    Kafka["Kafka Topic"]
    Projection["Consumer / Projection Worker"]
    LocalReadModel["Local Projection Tables"]

    ServiceDB --> Outbox
    Outbox -->|"CDC stream preferred"| Relay
    Relay --> Kafka
    Kafka --> Projection
    Projection --> LocalReadModel
```

This pattern is for transactional integration events, not every async job:

- write business state and the outbox record in the same service-owned transaction
- publish outbox records through CDC where possible; bounded polling is only a fallback
- downstream services materialize local read models when needed

Use direct pub/sub through `messaging.Publisher` for best-effort operational jobs that do not need
atomic commit with service state. Examples: cache refresh, search indexing hints, warmups,
non-critical notifications, telemetry, and UI task hints.

The CDC runtime boundary is Kafka Connect/Debezium:

- Postgres service databases publish changes through logical replication.
- The local Docker component is `cdc-connect`.
- MongoDB CDC requires Mongo change streams, so local Mongo must become a replica set before Mongo
  collections such as onboarding `connection_outbox` can use CDC reliably.
