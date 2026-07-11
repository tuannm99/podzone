# 02 Architecture Overall

This section documents the current Podzone architecture using a lightweight C4
structure and Mermaid diagrams.

Parent index: [Podzone Documentation Index](../README.md).

## Scope

- `C1` System context: external actors and major systems
- `C2` Container view: deployable services and shared infrastructure
- `C3` Component/module view: main modules inside `auth`, `iam`, and `backoffice`
- `C4` Detail design: sequence diagrams only

## Read In This Order

1. [C4 Architecture Index](./01-c4.md)
2. [System Context](./02-system-context.md)
3. [Containers](./03-containers.md)
4. [Data Ownership](./04-data-ownership.md)
5. [Sequences](./05-sequences.md)
6. [Architecture Detail Design](../03-architecture-detail-design/README.md)

## Topic Index

Core boundaries:

- [Data Ownership](./04-data-ownership.md)
- [Bounded Contexts](../03-architecture-detail-design/02-bounded-contexts.md)
- [DDD And Clean Architecture](../03-architecture-detail-design/03-ddd-clean-architecture.md)
- [Code Map](../03-architecture-detail-design/04-code-map.md)

Runtime and transport:

- [Transport and Contracts](../03-architecture-detail-design/05-transport-contracts.md)
- [Collection API Contract](../03-architecture-detail-design/06-collection-api-contract.md)
- [Async Messaging](../03-architecture-detail-design/07-async-messaging.md)
- [Platform Runtime](../03-architecture-detail-design/08-platform-runtime.md)
- [Gateway Bootstrap](../03-architecture-detail-design/09-gateway-bootstrap.md)
- [Deployment](../03-architecture-detail-design/10-deployment.md)

Product surfaces:

- [IAM Platform](../03-architecture-detail-design/11-iam-platform.md)
- [Frontend and Edge](../03-architecture-detail-design/12-frontend-edge.md)
- [Frontend Solid Audit](../03-architecture-detail-design/13-frontend-solid-audit.md)
- [MFE Federation Contract](../03-architecture-detail-design/14-mfe-federation-contract.md)
- [Design System](../03-architecture-detail-design/15-design-system.md)
- [Agent Onboarding](../03-architecture-detail-design/16-agent-onboarding.md)

## Current Direction

- `auth`, `iam`, `backoffice`, `partner`, and `onboarding` are treated as separate services.
- Synchronous cross-service reads go through `gRPC`.
- Asynchronous propagation goes through `Kafka`.
- Transactional domain events use outbox/CDC; best-effort operational jobs use direct pub/sub.
- `iam` owns the authorization source of truth and publishes commit-coupled domain events through an outbox.
- `auth` owns identity, sessions, and token issuance, and now maintains a small IAM projection for local read paths.

## Links Back To Delivery

- [SRS baseline](../01-srs/podzone-srs.md)
- [Recovery plan](../06-recovery/recovery-plan.md)
- [Sprint 0](../04-sprints/sprint-00-foundation.md)
