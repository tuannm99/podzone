# 07 Problems And Recovery Evidence

Parent index: [Podzone Documentation Index](../README.md).

This folder stores dated problem notes used as evidence for recovery work.
These are not task specs by themselves. Link them from SRS, recovery, sprint,
or agent task docs before assigning implementation.

## Problem Notes

- [Frontend issues, 2026-07-08](./fe-2026-07-08.md) — scoped to `internal/ui-podzone/src`, a path that no longer exists (frontend moved to `frontend/apps/*/src`). Re-check each item's file path before treating it as current.
- [Frontend issues, 2026-07-09](./fe-2026-07-09.md) — same stale-path caveat as above.
- [Frontend issues, 2026-07-09b](./fe-2026-07-09b.md) — uses `src/modules/backoffice/...` paths; current path is `frontend/apps/backoffice/src/...`. Re-check before treating as current.
- [Onboarding provisioning, 2026-07-07](./onboarding-provisioning-2026-07-07.md) — P1 resolved, P2-P8 unverified.
- [Package issues, 2026-07-07](./pkg-2026-07-07.md) — K1, K2, K6 resolved, K3-K5/K7-K11 unverified.
- [Multi-tenancy issues, 2026-07-23](./multitenancy-2026-07-23.md) — `pkg/pdtenantdb` + `internal/backoffice` tenant isolation, placement routing, pooling and migration path. MT1-MT16, all open, none reproduced against a running environment.
- [Platform scaling and runtime topology, 2026-07-23](./platform-scaling-2026-07-23.md) — DB-cluster placement/capacity admission, PgBouncer, and the Kubernetes staging topology. PS1-PS12, all open, derived from code and manifests only — no load test run, no cluster inspected.
- [Repository and query layer, 2026-07-24](./repository-query-layer-2026-07-24.md) — SQL construction maintainability in the backoffice repositories, plus one real defect (RQ1: `contains` on timestamp columns). RQ1-RQ4 open. Also records a measured benchmark showing the query builder is **not** a performance bottleneck — read that before proposing to replace Squirrel.
