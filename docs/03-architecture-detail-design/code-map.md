# Code Map

## Top-Level Code Structure

```mermaid
flowchart TB
    Cmd["cmd/*"]
    Internal["internal/*"]
    Pkg["pkg/*"]
    Proto["api/proto/*"]
    UI["internal/ui-podzone"]
    Docs["docs/02-architecture-overall"]
    Deploy["deployments/docker"]

    Cmd --> Internal
    Cmd --> Pkg
    Internal --> Proto
    UI --> Proto
    Docs --> Internal
    Deploy --> Cmd
```

## Internal Service Map

```mermaid
flowchart TB
    Auth["internal/auth"]
    IAM["internal/iam"]
    Backoffice["internal/backoffice"]
    Partner["internal/partner"]
    Catalog["internal/catalog"]
    Onboarding["internal/onboarding"]
    Gateway["internal/grpcgateway + internal/gateway"]
    Storefront["internal/storefront"]

    Auth --> IAM
    Backoffice --> Auth
    Backoffice --> IAM
    Backoffice --> Partner
    Gateway --> Auth
    Gateway --> IAM
    Gateway --> Partner
    Gateway --> Catalog
```

## Shared Package Map

```mermaid
flowchart LR
    PDConfig["pdconfig"]
    PDLog["pdlog"]
    PDSQL["pdsql"]
    PDRedis["pdredis"]
    PDKafka["pdkafka"]
    Messaging["messaging"]
    PDGRPC["pdgrpc"]
    PDGRPCClient["pdgrpcclient"]
    PDGraphQL["pdgraphql"]
    PDTenantDB["pdtenantdb"]
    PDWorker["pdworker"]
    Testkit["testkit"]

    PDConfig --> PDSQL
    PDConfig --> PDRedis
    PDConfig --> PDKafka
    PDKafka --> Messaging
    PDGRPCClient --> PDGRPC
    PDWorker --> Messaging
    Testkit --> PDSQL
```

## Notes

- `cmd/*` wires service processes and Fx modules.
- `internal/*` contains service-owned code and adapters.
- `pkg/*` contains reusable platform/runtime packages.
- `api/proto/*` is the source of gRPC/gateway contracts.
- `docs/02-architecture-overall` and `docs/03-architecture-detail-design` are the current architecture description set.
