# Bounded Contexts

## Context Landscape

```mermaid
flowchart TB
    Auth["Auth"]
    IAM["IAM"]
    BackofficeCatalog["Backoffice Catalog"]
    BackofficeRouting["Backoffice Routing"]
    BackofficeStore["Backoffice Store"]
    Partner["Partner"]
    Onboarding["Onboarding"]
    CatalogSvc["Catalog Service"]
    Storefront["Storefront"]

    Auth --> IAM
    BackofficeStore --> IAM
    BackofficeRouting --> IAM
    BackofficeRouting --> Partner
    BackofficeCatalog --> CatalogSvc
    Onboarding --> BackofficeStore
    Storefront --> CatalogSvc
```

## Auth and IAM Boundary

```mermaid
flowchart LR
    AuthIdentity["Auth: identity, sessions, tokens"]
    AuthProjection["Auth: IAM local projection"]
    IAMCore["IAM: tenants, memberships, policies, groups, orgs"]
    IAMEvents["IAM Kafka outbox events"]

    IAMCore --> IAMEvents
    IAMEvents --> AuthProjection
    AuthIdentity --> IAMCore
```

## Backoffice Boundary

```mermaid
flowchart LR
    Shell["Workspace shell"]
    Tenancy["Backoffice tenancy runtime"]
    Store["Store context"]
    Catalog["Catalog context"]
    Routing["Routing context"]
    PartnerDir["Partner directory adapter"]
    IAMAuthz["Tenant/store authz"]
    Placement["Tenant placement runtime"]

    Shell --> Tenancy
    Tenancy --> Placement
    Tenancy --> Store
    Store --> Catalog
    Store --> Routing
    Routing --> PartnerDir
    Routing --> IAMAuthz
    Catalog --> IAMAuthz
```

## Partner and Routing Boundary

```mermaid
flowchart LR
    PartnerModel["Partner capabilities + cost rules"]
    Recommendation["Routing recommendation"]
    RoutedOrders["Routed order lifecycle"]

    PartnerModel --> Recommendation
    Recommendation --> RoutedOrders
```

## Notes

- `Auth` and `IAM` are now separate contexts with gRPC and Kafka integration points.
- `Backoffice` is still one deployable service, but internally split into `store`, `catalog`, and `routing`.
- `Backoffice` also needs a tenancy runtime layer that separates:
  - edge/runtime tenant routing
  - application placement resolution
  - store-scoped business execution
- `Partner` influences routing decisions through capability and cost metadata.
- `Onboarding` owns connection/placement publication and is distinct from operator workflow surfaces.
