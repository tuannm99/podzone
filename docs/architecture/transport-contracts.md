# Transport and Contracts

## API Surface Ownership

```mermaid
flowchart TB
    subgraph Proto["api/proto"]
        AuthProto["auth/v1"]
        CatalogProto["catalog/v1"]
        PartnerProto["partner/v1"]
        CommonProto["common/v1"]
    end

    subgraph Transport
        AuthGRPC["Auth gRPC handlers"]
        IAMGRPC["IAM gRPC handlers"]
        PartnerGRPC["Partner gRPC handlers"]
        CatalogGRPC["Catalog gRPC handlers"]
        Gateway["gRPC Gateway registrars"]
        BackofficeGraphQL["Backoffice GraphQL schema/resolvers"]
    end

    AuthProto --> AuthGRPC
    AuthProto --> IAMGRPC
    CatalogProto --> CatalogGRPC
    PartnerProto --> PartnerGRPC
    CommonProto --> Gateway

    AuthProto --> Gateway
    CatalogProto --> Gateway
    PartnerProto --> Gateway
```

## Auth and IAM Proto Split

```mermaid
flowchart LR
    AuthSvc["auth_service.proto"]
    IAMSvc["iam_service.proto"]
    AuthTypes["auth.proto + auth_session.proto"]
    IAMTypes["iam_tenant.proto + iam_policy.proto + iam_simulation.proto"]

    AuthTypes --> AuthSvc
    IAMTypes --> IAMSvc
    AuthTypes --> IAMSvc
```

## Backoffice Transport Split

```mermaid
flowchart LR
    Schema["schema/*.graphqls"]
    Resolver["resolver/*.resolvers.go"]
    Domain["domain/catalog|store|routing"]
    Repo["infrastructure/repository/*"]
    PartnerDir["infrastructure/partnerdirectory"]

    Schema --> Resolver
    Resolver --> Domain
    Domain --> Repo
    Domain --> PartnerDir
```

## Notes

- `AuthService` owns identity, session, and token lifecycle.
- `IAMService` owns authorization, tenant, membership, policy, boundary, simulation, and assume-role decisioning.
- `Backoffice` is GraphQL-owned and maps into context-local domain packages.
- `grpcgateway` is only a transport adapter, not business logic.
