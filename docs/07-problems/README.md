# 07 Problems And Recovery Evidence

Parent index: [Podzone Documentation Index](../README.md).

This folder stores dated problem notes used as evidence for recovery work.
These are not task specs by themselves. Link them from SRS, recovery, sprint,
or agent task docs before assigning implementation.

## Problem Notes

- [Frontend issues, 2026-07-08](./fe-2026-07-08.md) — scoped to `internal/ui-podzone/src`, a path that no longer exists (frontend moved to `frontend/apps/*/src`). Re-check each item's file path before treating it as current.
- [Frontend issues, 2026-07-09](./fe-2026-07-09.md) — same stale-path caveat as above.
- [Frontend issues, 2026-07-09b](./fe-2026-07-09b.md) — uses `src/modules/backoffice/...` paths; current path is `frontend/apps/backoffice/src/...`. Re-check before treating as current.
- [Onboarding provisioning, 2026-07-07](./onboarding-provisioning-2026-07-07.md) — P1-P8 all resolved (P2-P7 turned out already fixed by earlier commits with a stale status footer; P8 and one new residual gap in P7's fix were fixed 2026-07-26, PZEP-0016 Wave 3).
- [Package issues, 2026-07-07](./pkg-2026-07-07.md) — K1-K11 all resolved (K3-K5/K7-K11 turned out already fixed by the same commit that introduced this note — the status footer was just stale until 2026-07-26).
- [Multi-tenancy issues, 2026-07-23](./multitenancy-2026-07-23.md) — `pkg/pdtenantdb` + `internal/backoffice` tenant isolation, placement routing, pooling and migration path. MT2/MT3/MT4/MT6/MT8/MT9/MT10/MT11/MT15/MT17 fully resolved; MT16 mostly resolved (real metrics still pending an observability-stack ADR); MT1 Phase 1 live, Phase 2 implemented but not rolled out ([PZEP-0009](../09-pzep/PZEP-0009-backoffice-least-privilege-app-db-role.md), [PZEP-0015](../09-pzep/PZEP-0015-backoffice-per-tenant-db-role-phase-2.md)); MT5 fully resolved and MT7 partially resolved (`FindPage` still blocked on a missing IAM capability) via [PZEP-0018](../09-pzep/PZEP-0018-store-authorization-iam-single-authority.md). Every item in this note has been addressed in some form.
- [Platform scaling and runtime topology, 2026-07-23](./platform-scaling-2026-07-23.md) — DB-cluster placement/capacity admission, PgBouncer, and the Kubernetes staging topology. PS8-PS11 (Track A / Wave 6) fully resolved via [PZEP-0019](../09-pzep/PZEP-0019-staging-k8s-survivability-track-a.md) (found a new `grpcgateway` port-mismatch bug this note never described along the way). PS1-PS7/PS12 (Track B / Wave 7) still open, not yet scoped into a PZEP.
- [Repository and query layer, 2026-07-24](./repository-query-layer-2026-07-24.md) — SQL construction maintainability in the backoffice repositories, plus one real defect (RQ1: `contains` on timestamp columns). RQ1-RQ4 resolved via [PZEP-0010](../09-pzep/PZEP-0010-backoffice-declarative-collection-spec.md). Also records a measured benchmark showing the query builder is **not** a performance bottleneck — read that before proposing to replace Squirrel.
