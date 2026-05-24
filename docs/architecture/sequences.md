# C4: Detail Design Sequences

## 1. Switch Active Tenant

```mermaid
sequenceDiagram
    participant UI as UI
    participant Auth as Auth Service
    participant LocalProj as Auth IAM Projection
    participant IAM as IAM Service
    participant AuthDB as Auth DB

    UI->>Auth: SwitchActiveTenant(userID, tenantID, accessToken)
    Auth->>LocalProj: lookup membership(tenantID, userID)
    alt projection hit and active
        LocalProj-->>Auth: active membership
    else projection miss
        Auth->>IAM: GetTenantMembership
        IAM-->>Auth: membership
    end
    Auth->>AuthDB: update auth_sessions.active_tenant_id
    Auth->>Auth: issue JWT for updated session
    Auth-->>UI: new access token
```

## 2. Create Tenant with Outbox

```mermaid
sequenceDiagram
    participant Admin as Admin / Caller
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant Outbox as IAM Outbox Worker
    participant Kafka as Kafka
    participant AuthProj as Auth IAM Projection Worker
    participant AuthDB as Auth DB

    Admin->>IAM: CreateTenant(name, slug, owner)
    IAM->>IAMDB: insert tenant
    IAM->>IAMDB: upsert owner membership
    IAM->>IAMDB: append message_outbox(tenant.created)
    IAM-->>Admin: tenant created

    Outbox->>IAMDB: poll pending outbox
    Outbox->>Kafka: publish tenant.created
    Outbox->>IAMDB: mark published

    AuthProj->>Kafka: consume tenant.created
    AuthProj->>AuthDB: upsert iam_tenants_projection
```

## 3. Add Tenant Member

```mermaid
sequenceDiagram
    participant Operator as Operator
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant Outbox as IAM Outbox Worker
    participant Kafka as Kafka
    participant AuthProj as Auth IAM Projection Worker
    participant AuthDB as Auth DB

    Operator->>IAM: AddMember(tenantID, userID, role)
    IAM->>IAMDB: upsert tenant_membership
    IAM->>IAMDB: append message_outbox(tenant.member.added)
    IAM-->>Operator: ok

    Outbox->>Kafka: publish tenant.member.added
    AuthProj->>Kafka: consume tenant.member.added
    AuthProj->>AuthDB: upsert iam_tenant_memberships_projection
```

## 4. Create Routed Order Recommendation

```mermaid
sequenceDiagram
    participant UI as Backoffice UI
    participant BO as Backoffice Service
    participant Partner as Partner Service
    participant IAM as IAM Service
    participant BODB as Backoffice DB

    UI->>BO: routedOrderRecommendation(input)
    BO->>IAM: check tenant permission
    IAM-->>BO: allow
    BO->>Partner: list/find partner capabilities
    Partner-->>BO: partner capabilities
    BO->>BO: evaluate eligibility, priority, SLA, margin snapshot
    BO-->>UI: recommendation

    UI->>BO: createRoutedOrder(selected recommendation)
    BO->>BODB: persist routed order + activity
    BO-->>UI: routed order created
```

## 5. IAM Event Projected into Auth

```mermaid
sequenceDiagram
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant Relay as IAM Outbox Worker
    participant Kafka as Kafka
    participant AuthWorker as Auth IAM Projection Worker
    participant AuthDB as Auth DB

    IAM->>IAMDB: append message_outbox(policy.attached / tenant.member.added)
    Relay->>IAMDB: poll pending outbox
    Relay->>Kafka: publish podzone.iam.events
    Relay->>IAMDB: mark published

    AuthWorker->>Kafka: consume podzone.iam.events
    alt tenant.created
        AuthWorker->>AuthDB: upsert iam_tenants_projection
    else tenant.member.added
        AuthWorker->>AuthDB: upsert iam_tenant_memberships_projection
    else other event
        AuthWorker->>AuthWorker: ignore
    end
```

## 6. Login and Tenant Workspace Bootstrap

```mermaid
sequenceDiagram
    participant User as User
    participant UI as Seller Portal
    participant Auth as Auth Service
    participant Redis as Redis
    participant AuthDB as Auth DB

    User->>UI: submit login form
    UI->>Auth: Login(username/email, password)
    Auth->>AuthDB: load user and memberships
    Auth->>Redis: persist refresh token/session cache
    Auth->>AuthDB: create auth_session
    Auth-->>UI: access token + refresh token + session
    UI->>UI: store token and route to /admin or tenant workspace
```

## 7. Admin IAM Policy Attachment

```mermaid
sequenceDiagram
    participant UI as Admin IAM UI
    participant GW as gRPC Gateway
    participant IAM as IAM Service
    participant IAMDB as IAM DB
    participant Outbox as IAM Outbox Worker
    participant Kafka as Kafka

    UI->>GW: AttachTenantUserPolicy
    GW->>IAM: AttachTenantUserPolicy
    IAM->>IAMDB: write attachment
    IAM->>IAMDB: append message_outbox(policy.attached)
    IAM-->>GW: ok
    GW-->>UI: success

    Outbox->>IAMDB: poll pending outbox
    Outbox->>Kafka: publish policy.attached
    Outbox->>IAMDB: mark published
```

## 8. Onboarding Connection Publish to Consul

```mermaid
sequenceDiagram
    participant Client as Operator / Bootstrap Script
    participant Onboarding as Onboarding Service
    participant Mongo as Mongo Event Store
    participant Worker as Onboarding Outbox Worker
    participant Consul as Consul

    Client->>Onboarding: create/update connection
    Onboarding->>Mongo: persist connection event + outbox item
    Onboarding-->>Client: accepted

    Worker->>Mongo: poll pending outbox
    Worker->>Consul: publish connection snapshot and placement
    Worker->>Mongo: mark published
```
