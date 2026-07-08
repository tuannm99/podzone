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

1. [C4 Architecture Index](./c4.md)
2. [System Context](./system-context.md)
3. [Containers](./containers.md)
4. [Sequences](./sequences.md)
5. [Architecture Detail Design](../03-architecture-detail-design/README.md)

## Topic Index

Core boundaries:

- [Data Ownership](./data-ownership.md)
- [Bounded Contexts](../03-architecture-detail-design/bounded-contexts.md)
- [DDD And Clean Architecture](../03-architecture-detail-design/ddd-clean-architecture.md)
- [Code Map](../03-architecture-detail-design/code-map.md)

Runtime and transport:

- [Transport and Contracts](../03-architecture-detail-design/transport-contracts.md)
- [Collection API Contract](../03-architecture-detail-design/collection-api-contract.md)
- [Async Messaging](../03-architecture-detail-design/async-messaging.md)
- [Platform Runtime](../03-architecture-detail-design/platform-runtime.md)
- [Gateway Bootstrap](../03-architecture-detail-design/gateway-bootstrap.md)
- [Deployment](../03-architecture-detail-design/deployment.md)

Product surfaces:

- [IAM Platform](../03-architecture-detail-design/iam-platform.md)
- [Frontend and Edge](../03-architecture-detail-design/frontend-edge.md)
- [Frontend Solid Audit](../03-architecture-detail-design/frontend-solid-audit.md)
- [Design System](../03-architecture-detail-design/design-system.md)
- [Agent Onboarding](../03-architecture-detail-design/agent-onboarding.md)

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
