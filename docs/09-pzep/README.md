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
| [PZEP-0002](./PZEP-0002-platform-admin-bootstrap-and-recovery.md) | Platform admin bootstrap and recovery | Draft | 2026-07-12 |
| [PZEP-0003](./PZEP-0003-iam-decoupling-and-sdk-phase-1.md) | Backoffice/IAM decoupling and IAM SDK Phase 1 | Draft | 2026-07-12 |
| [PZEP-0004](./PZEP-0004-solidjs-to-angular-frontend-migration.md) | SolidJS → Angular frontend migration plan | Draft | 2026-07-13 |
| [PZEP-0005](./PZEP-0005-host-federation-migration-to-mf2.md) | Host federation migration to Module Federation 2 (`@module-federation/vite`) | Done (uncommitted) | 2026-07-13 |
| [PZEP-0006](./PZEP-0006-angular-base-mfe-scaffold.md) | Angular base MFE scaffold — shell, onboarding feature area, design system port | Done (uncommitted) | 2026-07-13 |
| [PZEP-0007](./PZEP-0007-angular-design-system-port.md) | Angular design-system port — full `packages/shared/ui` component library | Done (uncommitted) | 2026-07-14 |
| [PZEP-0008](./PZEP-0008-angular-onboarding-backbone-integration.md) | Angular onboarding backbone integration — core infra + workspace/store/handoff | Approved | 2026-07-14 |

## Links Back To Delivery

- [SRS baseline](../01-srs/podzone-srs.md)
- [Traceability matrix](../01-srs/traceability-matrix.md)
- [Recovery plan](../06-recovery/recovery-plan.md)
