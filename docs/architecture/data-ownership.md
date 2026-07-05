# Data Ownership

## Service-Owned Datastores

```mermaid
flowchart TB
    Auth["Auth Service"]
    IAM["IAM Service"]
    Backoffice["Backoffice Service"]
    Partner["Partner Service"]
    Catalog["Catalog Service"]
    Onboarding["Onboarding Service"]

    AuthDB["auth DB"]
    IAMDB["iam DB"]
    BackofficeDB["backoffice DB"]
    PartnerDB["partner DB"]
    CatalogDB["catalog DB"]
    Mongo["Mongo"]
    RuntimeKV["Mongo runtime_kv"]
    Redis["Redis"]

    Auth --> AuthDB
    Auth --> Redis
    IAM --> IAMDB
    Backoffice --> BackofficeDB
    Partner --> PartnerDB
    Catalog --> CatalogDB
    Onboarding --> Mongo
    Onboarding --> RuntimeKV
```

## Auth Data Ownership

```mermaid
flowchart LR
    Users["users"]
    Sessions["auth_sessions"]
    Refresh["refresh_tokens"]
    Audit["auth_audit_logs"]
    IAMProj["iam_*_projection tables"]

    Users --> Sessions
    Users --> Refresh
    Users --> Audit
    IAMProj --> Sessions
```

## IAM Data Ownership

```mermaid
flowchart LR
    Tenants["tenants"]
    Memberships["tenant_memberships"]
    Platform["platform_memberships"]
    Groups["iam_groups + members"]
    Policies["iam_policies + versions + statements"]
    Boundaries["permission boundaries"]
    Orgs["organizations + SCP attachments"]
    Outbox["message_outbox"]

    Tenants --> Memberships
    Policies --> Boundaries
    Policies --> Groups
    Policies --> Orgs
    Memberships --> Outbox
    Policies --> Outbox
    Tenants --> Outbox
```

## Backoffice Data Ownership

```mermaid
flowchart LR
    Stores["stores"]
    ProductSetup["product_setup_*"]
    RoutedOrders["routed_orders"]
    Activities["routed_order_activities"]

    Stores --> ProductSetup
    Stores --> RoutedOrders
    RoutedOrders --> Activities
```

## Notes

- The target architecture is service-owned persistence with no shared write access.
- `auth` keeps only a small IAM projection for hot read paths.
- Kafka events plus projections replace direct cross-service table reads.
