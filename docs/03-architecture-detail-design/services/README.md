# Per-Service Detail Design

Parent index: [03 Architecture Detail Design](../README.md).

The 16 numbered docs in the parent folder are cross-cutting (modules,
contracts, messaging, platform, deployment, design system). This folder
holds **per-service** detail design: one component spec (BE + FE where a
frontend surface exists) and one DB design/ERD per service, following
`docs/05-process/component-spec-template.md` and
`docs/05-process/db-contract-template.md`.

## Services (Stabilize — full docs)

| Service | DB | Tables/collections | FE remote | README | DB Design |
|---|---|---|---|---|---|
| auth | Postgres, platform-scoped | 7 | none (UI lives in HOST shell `frontend/src/modules/shell/pages/auth/`) | [README](./auth/README.md) | [DB Design](./auth/db-design.md) |
| iam | Postgres, platform-scoped | 32 | `frontend/apps/iam` | [README](./iam/README.md) | [DB Design](./iam/db-design.md) |
| onboarding | MongoDB | 10 collections | `frontend/apps/onboarding` | [README](./onboarding/README.md) | [DB Design](./onboarding/db-design.md) |
| backoffice | Postgres only, tenant-routed via `pdtenantdb` (no Mongo — corrected 2026-07-11) | 6 | `frontend/apps/backoffice` | [README](./backoffice/README.md) | [DB Design](./backoffice/db-design.md) |
| partner | Postgres, single shared DB | 1 | none | [README](./partner/README.md) | [DB Design](./partner/db-design.md) |

All schemas above were built by reading actual migrations/entity structs
per service, not inferred — each `db-design.md` states its own
verification method. Cross-table relationships without a real `REFERENCES`
FK constraint are called out explicitly as logical references, not drawn
as enforced foreign keys.

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
