# 09 Podzone Enhancement Proposals

Parent index: [Podzone Documentation Index](../README.md).

A PZEP (Podzone Enhancement Proposal) describes the feature-level solution
for a change that touches multiple components, contracts, permissions,
database schema, events, workers, or external integrations. Use
[`docs/05-process/pzep-template.md`](../05-process/pzep-template.md) to
write one.

Per `docs/00-governance/agent-working-rule.md`: **no PZEP, no
cross-component implementation.**

PZEP is not a business requirement (that's `docs/00-project-vision`) and
does not replace an ADR for pure architecture-boundary decisions (see
`docs/08-adr/`).

## Index

| ID | Title | Status | Date |
|---|---|---|---|
| [PZEP-0001](./PZEP-0001-onboarding-store-readiness-endpoint.md) | Onboarding store readiness endpoint | Done | 2026-07-10 |
| [PZEP-0002](./PZEP-0002-platform-admin-bootstrap-and-recovery.md) | Platform admin bootstrap and recovery (`cmd/iam-bootstrap`) | Done | 2026-07-12 |
| [PZEP-0003](./PZEP-0003-iam-decoupling-and-sdk-phase-1.md) | Backoffice/IAM decoupling and IAM SDK Phase 1 | Draft | 2026-07-12 |
| [PZEP-0004](./PZEP-0004-solidjs-to-angular-frontend-migration.md) | SolidJS → Angular frontend migration plan | Superseded by ADR-0007 | 2026-07-13 |
| [PZEP-0005](./PZEP-0005-host-federation-migration-to-mf2.md) | Host federation migration to Module Federation 2 (`@module-federation/vite`) | Done (uncommitted) | 2026-07-13 |
| [PZEP-0006](./PZEP-0006-angular-base-mfe-scaffold.md) | Angular base MFE scaffold — shell, onboarding feature area, design system port | Done (uncommitted) | 2026-07-13 |
| [PZEP-0007](./PZEP-0007-angular-design-system-port.md) | Angular design-system port — full `packages/shared/ui` component library | Done (uncommitted) | 2026-07-14 |
| [PZEP-0008](./PZEP-0008-angular-onboarding-backbone-integration.md) | Angular onboarding backbone integration — core infra + workspace/store/handoff | Approved | 2026-07-14 |
| [PZEP-0009](./PZEP-0009-backoffice-least-privilege-app-db-role.md) | Backoffice tenant DB — separate least-privilege app role from the admin/bootstrap role (Phase 1 of MT1) | Done (rolled out to docker dev) | 2026-07-25 |
| [PZEP-0010](./PZEP-0010-backoffice-declarative-collection-spec.md) | Backoffice repository layer — declarative collection spec (closes RQ1) | Done | 2026-07-25 |
| [PZEP-0011](./PZEP-0011-angular-backoffice-minimal-shell.md) | Angular Backoffice minimal shell — store-scoped home page + one protected API call | Done | 2026-07-25 |
| [PZEP-0012](./PZEP-0012-angular-store-collection-screen.md) | Angular store collection screen — first ADR-0008 declarative-spec screen | Done | 2026-07-25 |
| [PZEP-0013](./PZEP-0013-angular-shell-wraps-backoffice-routes.md) | Angular `Shell` wraps Backoffice routes; Stores nav link | Done | 2026-07-25 |
| [PZEP-0014](./PZEP-0014-angular-routed-orders-collection-screen.md) | Angular routed-orders collection screen — second ADR-0008 declarative-spec screen | Done | 2026-07-25 |
| [PZEP-0015](./PZEP-0015-backoffice-per-tenant-db-role-phase-2.md) | Backoffice tenant DB — per-tenant Postgres roles via `SET LOCAL ROLE` (Phase 2 of MT1) | Done (code; not yet rolled out) | 2026-07-25 |
| [PZEP-0016](./PZEP-0016-recovery-backlog-triage-2026-07.md) | Recovery backlog triage — remaining `docs/07-problems` findings | Draft (Waves 1/3/4/5 done, Waves 6/7 not started) | 2026-07-25 |
| [PZEP-0017](./PZEP-0017-multitenancy-correctness-observability.md) | `pkg/pdtenantdb` correctness & observability batch (MT4, MT9, MT10, MT11, MT16) | Done | 2026-07-26 |
| [PZEP-0018](./PZEP-0018-store-authorization-iam-single-authority.md) | Backoffice store authorization — IAM as the single authority (MT5, MT7) | Done | 2026-07-26 |
| [PZEP-0019](./PZEP-0019-staging-k8s-survivability-track-a.md) | Staging Kubernetes survivability, Track A (PS8, PS9, PS10, PS11) | Done | 2026-07-27 |
| [PZEP-0020](./PZEP-0020-frontend-v2-admin-console.md) | frontend-v2 admin console — cross-tenant onboarding review, readiness debug, IAM section | Done (tenant directory dropped — no backend RPC) | 2026-07-27 |

## Links Back To Delivery

- [SRS baseline](../01-srs/podzone-srs.md)
- [Traceability matrix](../01-srs/traceability-matrix.md)
- [Recovery plan](../06-recovery/recovery-plan.md)
