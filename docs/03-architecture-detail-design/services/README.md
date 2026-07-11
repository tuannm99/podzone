# Per-Service Detail Design

Parent index: [03 Architecture Detail Design](../README.md).

The 16 numbered docs in the parent folder are cross-cutting (modules,
contracts, messaging, platform, deployment, design system). This folder
holds **per-service** detail design — everything specific to one service,
BE + FE where a frontend surface exists:

- `README.md` — component spec (Purpose, Responsibilities, Owned Data,
  Dependencies, Failure Modes, Security, Frontend Surface), per
  `docs/05-process/component-spec-template.md`.
- `db-design.md` — ERD and full schema, per
  `docs/05-process/db-contract-template.md`.
- `api-design.md` — C3 component view, full API surface (gRPC/HTTP/GraphQL
  contract), and one C4 sequence diagram per usecase. Sequence diagrams
  live here, next to the API/usecase they describe — not in a standalone
  cross-cutting file (see `docs/02-architecture-overall/01-c4.md`).

## Services (Stabilize — full docs)

| Service | DB | Tables/collections | FE remote | README | DB Design | API Design |
|---|---|---|---|---|---|---|
| auth | Postgres, platform-scoped | 7 | none (UI lives in HOST shell `frontend/src/modules/shell/pages/auth/`) | [README](./auth/README.md) | [DB Design](./auth/db-design.md) | [API Design](./auth/api-design.md) |
| iam | Postgres, platform-scoped | 24 | `frontend/apps/iam` | [README](./iam/README.md) | [DB Design](./iam/db-design.md) | [API Design](./iam/api-design.md) |
| onboarding | MongoDB | 10 collections | `frontend/apps/onboarding` | [README](./onboarding/README.md) | [DB Design](./onboarding/db-design.md) | [API Design](./onboarding/api-design.md) |
| backoffice | Postgres only, tenant-routed via `pdtenantdb` (no Mongo — corrected 2026-07-11) | 6 | `frontend/apps/backoffice` | [README](./backoffice/README.md) | [DB Design](./backoffice/db-design.md) | [API Design](./backoffice/api-design.md) |
| partner | Postgres, single shared DB | 1 | none | [README](./partner/README.md) | [DB Design](./partner/db-design.md) | [API Design](./partner/api-design.md) |

All schemas above were built by reading actual migrations/entity structs
per service, not inferred — each `db-design.md` states its own
verification method. Cross-table relationships without a real `REFERENCES`
FK constraint are called out explicitly as logical references, not drawn
as enforced foreign keys. Same rule for `api-design.md`: sequence diagrams
were reproduced from actual code, not guessed, and known drift between an
old diagram and current code is called out inline rather than silently
corrected.

## Services (Later — stub, not documented in detail)

Per `docs/06-recovery/legacy-inventory.md`, these are not needed for the
current backbone recovery slice. Documented only when work on them starts.

| Service | Status | Notes |
|---|---|---|
| catalog | Later | `internal/catalog` — 1 file, not yet implemented beyond scaffold. Mongo per legacy-inventory. |
| storefront | Later | `internal/storefront` — 1 file, not yet implemented. |
| gateway | Infra, not a domain service | APISIX config generation, no DB. |
| grpcgateway | Infra, not a domain service | HTTP→gRPC adapter, no DB. |

## Links Back To Delivery

- [SRS baseline](../../01-srs/podzone-srs.md)
- [Legacy Inventory](../../06-recovery/legacy-inventory.md)
- [Recovery plan](../../06-recovery/recovery-plan.md)
