# C1: System Context

```mermaid
flowchart LR
    Admin["Platform Admin / Operator"]
    Merchant["Merchant / Tenant User"]
    SellerPortal["Backoffice Portal"]
    Storefront["Storefront / Buyer UI"]
    AuthSvc["Auth Service"]
    IAMSvc["IAM Service"]
    CatalogSvc["Catalog Service"]
    PartnerSvc["Partner Service"]
    GatewaySvc["gRPC Gateway"]
    OnboardingSvc["Onboarding Service"]
    ApiGateway["API Gateway / APISIX"]
    Kafka["Kafka"]
    Postgres["Postgres (service-owned DBs)"]
    Redis["Redis"]
    RuntimeKV["Mongo runtime_kv"]

    Admin --> SellerPortal
    Merchant --> SellerPortal
    Merchant --> Storefront

    SellerPortal --> ApiGateway
    Storefront --> ApiGateway

    ApiGateway --> GatewaySvc
    ApiGateway --> OnboardingSvc
    ApiGateway --> SellerPortal

    GatewaySvc --> AuthSvc
    GatewaySvc --> IAMSvc
    GatewaySvc --> CatalogSvc
    GatewaySvc --> PartnerSvc

    AuthSvc --> Redis
    AuthSvc --> Postgres
    AuthSvc --> Kafka

    IAMSvc --> Postgres
    IAMSvc --> Kafka

    CatalogSvc --> Postgres
    PartnerSvc --> Postgres
    OnboardingSvc --> RuntimeKV
    OnboardingSvc --> Postgres

    Kafka --> AuthSvc
    Kafka --> IAMSvc
```

## Notes

- `Auth Service` handles identity, sessions, refresh tokens, and scoped access tokens.
- `IAM Service` owns tenants, memberships, roles, policies, boundaries, organizations, and simulations.
- `Catalog Service` and `Partner Service` are gRPC-facing business services behind the gateway.
- `Backoffice Portal` is the main operator-facing entry point.
- `Storefront` is the buyer-facing web surface.
- `gRPC Gateway` is the HTTP facade for gRPC admin-style services.
- `Kafka` is the event backbone for service-to-service asynchronous propagation.
