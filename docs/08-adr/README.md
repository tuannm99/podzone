# 08 Architecture Decision Records

Parent index: [Podzone Documentation Index](../README.md).

An ADR records one architecture decision that is expensive to reverse:
service boundaries, data ownership, cross-service communication,
dependency direction, or a technology/pattern choice. Use
[`docs/05-process/adr-template.md`](../05-process/adr-template.md) to write
one.

Per `docs/00-governance/traceability-rule.md`: **no ADR, no architecture
boundary change.**

## Index

| ID | Title | Status | Date |
|---|---|---|---|
| [ADR-0001](./ADR-0001-mfe-solid-js-not-federation-singleton.md) | MFE remotes do not share `solid-js` as a federation singleton | Accepted | 2026-07-10 |
| [ADR-0002](./ADR-0002-npm-vite-alias-not-pnpm-workspaces.md) | Monorepo dependency resolution uses npm + Vite `resolve.alias`, not pnpm workspaces | Accepted | 2026-07-11 |
| [ADR-0003](./ADR-0003-platform-scope-tenant-access-override.md) | Platform-scope roles do not implicitly bypass tenant-scoped access checks | Proposed | 2026-07-12 |
| [ADR-0004](./ADR-0004-backoffice-static-rbac-not-iam-decision-api.md) | Backoffice enforces authorization with a static role→action table, not IAM's decision API | Proposed | 2026-07-12 |
| [ADR-0005](./ADR-0005-frontend-framework-angular-replaces-solidjs.md) | Frontend framework — Angular replaces SolidJS as the target, migrated incrementally | Proposed | 2026-07-13 |
| [ADR-0006](./ADR-0006-angular-material-replaces-ported-tailwind-design-system.md) | Angular Material replaces the ported Tailwind/Flowbite design system for `frontend-v2` | Accepted | 2026-07-14 |

## Links Back To Delivery

- [SRS baseline](../01-srs/podzone-srs.md)
- [Architecture detail design](../03-architecture-detail-design/README.md)
- [Recovery plan](../06-recovery/recovery-plan.md)
