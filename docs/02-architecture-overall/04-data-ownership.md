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

    AuthDB["auth DB (Postgres)"]
    IAMDB["iam DB (Postgres)"]
    BackofficeDB["backoffice DB (Postgres, tenant-routed)"]
    PartnerDB["partner DB (Postgres)"]
    CatalogDB["catalog DB (not yet implemented)"]
    Mongo["Mongo (onboarding, 10 collections)"]
    Redis["Redis / Valkey"]

    Auth --> AuthDB
    Auth --> Redis
    IAM --> IAMDB
    Backoffice --> BackofficeDB
    Partner --> PartnerDB
    Catalog --> CatalogDB
    Onboarding --> Mongo
    Onboarding -->|"publishes route projection"| Redis
    Backoffice -->|"reads route projection via pdtenantdb"| Redis
```

Corrected 2026-07-11: the KV route projection published by onboarding and
read by `pkg/pdtenantdb` is **Redis/Valkey, not Mongo** — a prior version
of this diagram mislabeled it as "Mongo runtime_kv". See
[`03-architecture-detail-design/services/onboarding/db-design.md`](../03-architecture-detail-design/services/onboarding/db-design.md)
"Not A Database Table: KV Route Projection".

## Auth Data Ownership

Full schema, columns, and indexes:
[`03-architecture-detail-design/services/auth/db-design.md`](../03-architecture-detail-design/services/auth/db-design.md).

```mermaid
flowchart LR
    Users["users"]
    Sessions["auth_sessions"]
    Refresh["auth_refresh_tokens"]
    Audit["auth_audit_logs"]
    IAMProj["iam_tenants_projection, iam_tenant_memberships_projection"]
    Inbox["message_inbox (Kafka idempotency ledger)"]

    Users --> Sessions
    Users --> Refresh
    Users --> Audit
    IAMProj --> Sessions
```

## IAM Data Ownership

Full schema (32 tables), columns, and indexes:
[`03-architecture-detail-design/services/iam/db-design.md`](../03-architecture-detail-design/services/iam/db-design.md).

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

Full schema, columns, and indexes:
[`03-architecture-detail-design/services/backoffice/db-design.md`](../03-architecture-detail-design/services/backoffice/db-design.md).
Postgres only — no Mongo (corrected 2026-07-11; see linked doc for the
grep evidence).

```mermaid
flowchart LR
    Stores["stores"]
    ProductSetup["product_setup_*"]
    RoutedOrders["routed_orders (legacy, still written)"]
    CustomerOrders["customer_orders (current aggregate)"]
    Activities["routed_order_activities"]

    Stores --> ProductSetup
    Stores --> RoutedOrders
    Stores --> CustomerOrders
    RoutedOrders --> Activities
    CustomerOrders --> Activities
```

## Notes

- The target architecture is service-owned persistence with no shared write access.
- `auth` keeps only a small IAM projection for hot read paths.
- Kafka events plus projections replace direct cross-service table reads.
