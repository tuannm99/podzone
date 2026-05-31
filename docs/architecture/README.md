# Architecture

This section documents the current Podzone architecture using a lightweight C4 structure and Mermaid diagrams.

## Scope

- `C1` System context: external actors and major systems
- `C2` Container view: deployable services and shared infrastructure
- `C3` Component/module view: main modules inside `auth`, `iam`, and `backoffice`
- `C4` Detail design: sequence diagrams only

## Documents

- [System Context](./system-context.md)
- [Containers](./containers.md)
- [Modules](./modules.md)
- [Frontend and Edge](./frontend-edge.md)
- [Transport and Contracts](./transport-contracts.md)
- [Data Ownership](./data-ownership.md)
- [Code Map](./code-map.md)
- [Deployment](./deployment.md)
- [Async Messaging](./async-messaging.md)
- [Gateway Bootstrap](./gateway-bootstrap.md)
- [Bounded Contexts](./bounded-contexts.md)
- [Platform Runtime](./platform-runtime.md)
- [Sequences](./sequences.md)

## Current Direction

- `auth`, `iam`, `backoffice`, `partner`, and `onboarding` are treated as separate services.
- Synchronous cross-service reads go through `gRPC`.
- Asynchronous propagation goes through `Kafka`.
- Transactional domain events use outbox/CDC; best-effort operational jobs use direct pub/sub.
- `iam` owns the authorization source of truth and publishes commit-coupled domain events through an outbox.
- `auth` owns identity, sessions, and token issuance, and now maintains a small IAM projection for local read paths.
