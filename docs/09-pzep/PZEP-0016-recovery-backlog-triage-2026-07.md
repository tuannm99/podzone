# PZEP-0016: Recovery backlog triage — remaining `docs/07-problems` findings

## Status
Draft — Waves 1, 3, 4, 5, and 6 all done (see Rollout Log). Wave 2 removed
(already resolved). Waves 4, 5, and 6 were implemented via their own
follow-up PZEPs
([PZEP-0017](./PZEP-0017-multitenancy-correctness-observability.md),
[PZEP-0018](./PZEP-0018-store-authorization-iam-single-authority.md),
[PZEP-0019](./PZEP-0019-staging-k8s-survivability-track-a.md), all Done).
Only Wave 7 remains — not yet scoped into a PZEP.

## Date
2026-07-25 (Wave 1 executed 2026-07-26; Wave 4's MT6 piece, PZEP-0015, and
Wave 3 also executed 2026-07-26; PZEP-0017 and PZEP-0018 written and
implemented 2026-07-26)

## Related Commit
(none yet — implemented in the working tree this session; not committed as
of this update.)

## Requirement Sources
- Business: recovery track, `docs/06-recovery/recovery-plan.md` — the
  backbone flow must stabilize before feature breadth expands; this backlog
  is the residue of the 2026-07-07/07-23/07-24 problem sweeps once the
  backbone-blocking items (MT1 Phase 1, MT2, MT3, MT8, MT17, RQ1-RQ4, K1, K2,
  K6, P1) were already fixed.
- Feature: none new — this PZEP triages and sequences existing findings, it
  does not design new behavior.
- Use Cases: none new.
- Functional/Non-functional Requirements: n/a — see each wave's own
  follow-up PZEP once written.
- Acceptance Criteria: "every item below is either scheduled into a wave with
  an owner PZEP, or explicitly deferred with a reason" — that is this
  document's own acceptance bar, not a code acceptance bar.
- UI Specs: none.

## Summary

`docs/07-problems/` accumulated four dated problem notes (2026-07-07 ×2,
2026-07-23 ×2, 2026-07-24 ×1) totaling around 50 findings. When this PZEP was
first drafted (2026-07-25), nine items/groups were believed resolved (K1, K2,
K6, P1, MT2, MT3, MT8, MT17, RQ1-RQ4 as one closed group) and MT1 had a
Phase 1 fix live with a Phase 2 proposal in
[PZEP-0015](./PZEP-0015-backoffice-per-tenant-db-role-phase-2.md).

**Verifying Wave 1 against current code (2026-07-26) turned up a second,
larger stale-docs problem**: `pkg-2026-07-07.md`'s own status footer only
credited K1/K2/K6 as resolved, but reading the actual code showed K3-K11
were **all** already fixed too — by commit `1a51d79`, the very commit that
introduced that problem note in the first place. Same story for P5 in
`onboarding-provisioning-2026-07-07.md`. In other words: "identified on date
X" and "fixed on date X" were both true, and nobody updated the status line
after. See the Rollout Log below and each note's own corrected status
footer — this is now fixed in the docs, not just noted here.

Net effect on this PZEP's scope: what looked like ~38 open items is actually
~27 (K3-K11, P5, and P8/MT15 — the latter two fixed in this PZEP's own
Wave 1 — remove 11 items from the original count). The three stale frontend
notes (`fe-2026-07-08.md`, `fe-2026-07-09.md`, `fe-2026-07-09b.md`) reference
a SolidJS path (`internal/ui-podzone/src`) that predates even the frozen
`frontend/` tree's current layout — per `docs/08-adr/ADR-0007`, SolidJS is
frozen with no new work, so these are **not triaged into a wave**; re-scope
only if `frontend/` is ever revived.

This PZEP does not design fixes for every remaining item — that would
violate the same "no PZEP, no cross-component implementation" rule this
document itself is written under, just at bulk scale. Instead it groups them
into waves by theme and blast radius, states dependency order between waves,
and marks which waves can proceed as direct scoped fixes under this PZEP's
own authorization (single-component, no contract change) versus which need
their own follow-up PZEP before any code is written (cross-component, or
changes a contract/permission model).

## Problem

The backlog is real but undifferentiated — `docs/07-problems/README.md`
lists five notes with no priority or sequencing across them, and two of the
notes (`multitenancy-2026-07-23.md`, `platform-scaling-2026-07-23.md`)
already contain their own internal "Suggested Order Of Work" sections that
don't account for each other or for the `pkg-2026-07-07.md` /
`onboarding-provisioning-2026-07-07.md` items at all. Without a merged
sequencing, the risk is picking items ad hoc by whichever doc was read most
recently, redoing ordering analysis each time, or — worse — implementing a
cross-component item (e.g. MT7, PS4) directly from a problem note, which
`docs/00-governance/agent-working-rule.md` already prohibits.

## Goals

- One merged, de-duplicated view of every open item across all four active
  problem notes.
- A wave assignment for every item, with an explicit reason.
- Correct cross-note dependency ordering preserved (e.g. PS7 before PS6, MT1
  before PS6 — both already called out in `platform-scaling-2026-07-23.md`).
- A clear split between "small enough to fix directly once this PZEP is
  approved" (single file/component, no contract change) and "needs its own
  PZEP" (cross-component, or changes a permission/contract boundary) — so
  approving this PZEP doesn't silently authorize the large items too.

## Non-Goals

- Designing the actual fix for any cross-component item — that is each
  wave's own follow-up PZEP's job.
- The three stale SolidJS frontend notes (see Summary).
- Re-litigating MT1 (covered by PZEP-0009/PZEP-0015) or the already-resolved
  items (K1/K2/K6, P1, MT2/MT3/MT8/MT17, RQ1-RQ4).
- A committed calendar/timeline — "wave order" here means dependency and risk
  order, not a schedule; the user picks which wave to actually start.

## Proposed Solution

### Wave 1 — Direct scoped fixes (single component, no contract change)

**Done as of 2026-07-26 — see Rollout Log.** Originally scoped as six items;
verifying each against current code before touching it found that K9, K10,
K11, and P5 were already fixed (by commit `1a51d79`, predating this PZEP —
see Summary). Only P8 and MT15 needed real work:

| Item | File | Fix | Outcome |
|---|---|---|---|
| ~~K9~~ | `pkg/pdconfig/config.go:29` | ~~Replace blanket `os.ExpandEnv`...~~ | Already fixed, no change needed |
| ~~K10~~ | `pkg/pdgrpcgateway/lifecycle.go:26-41` | ~~Replace `strings.Contains`...~~ | Already fixed, no change needed |
| ~~K11~~ | `pkg/pdredis/fxmodule.go:79` | ~~Delete the dead line~~ | Already fixed, no change needed |
| ~~P5~~ | `internal/onboarding/.../provider/health.go:53` | ~~Add `AND pid <> pg_backend_pid()`~~ | Already fixed, no change needed |
| P8 | `scripts/dev/reconcile_legacy_tenant.go:178` | `findRequest` now filters server-side via `collection.filters[0].field=subdomain` instead of paginating a fixed 100-row page | Fixed |
| MT15 | `internal/backoffice/migrations/sql/0015_*.sql` → `0017_...` | Renumbered the duplicate to `0017_` (0016 was already taken by PZEP-0009's role migration); updated `backoffice_schema_migrations.version` directly in all five docker dev tenant schemas so the rename doesn't trigger a redundant (harmless but messy) re-apply | Fixed |

### Wave 2 — Messaging reliability (`pkg/messaging`, single component, behavior-changing)

**Removed — already resolved.** K3, K4, K5, K7, K8 were all fixed by the
same commit `1a51d79` that fixed K1/K2/K6/K9/K10/K11 (see Summary). No PZEP
needed; nothing left to sequence here. Verified directly against current
code: `waitUntilReady` reschedules instead of blocking (K3), decode failures
drop+observe instead of killing the consumer session (K4), `InboxStore.Begin`
retries on `ErrNoRows` (K5), `IdempotentConsumer` errors on empty `msg.ID`
(K7), `Relay.RunOnce` batches `MarkPublished` after the publish loop (K8).

### Wave 3 — Onboarding provisioning correctness (`internal/onboarding`)

**Removed — already resolved, plus one residual gap fixed directly.**
Verifying P2/P3/P4/P6/P7 against current code (2026-07-26, same
double-check discipline that found Wave 1/Wave 2's stale status) found all
five already fixed by commits `73be9f5` and `27b26f3` (2026-07-09) — see
`docs/07-problems/onboarding-provisioning-2026-07-07.md`'s corrected status
footer. No PZEP needed for the original five items.

One **new**, genuinely open bug surfaced while re-verifying P7: its fix
(removing `Planning` from `ClaimNextQueued`'s filter, closing the
double-provisioning race) left no recovery path for a request whose worker
crashes mid-`Planning` — `Planning → Queued` was never a valid transition,
so neither the automatic claim path nor manual `RetryStoreRequest` could
recover it; it would stay `Planning` forever past lease expiry. This was
small enough (single component, no contract change, no new API surface —
implemented as an automatic sweep inside the existing worker tick) to fix
directly rather than write a new PZEP for: added
`StoreRepository.ReclaimStalePlanning` (atomic `FindOneAndUpdate`, same
concurrency-safety shape as `ClaimNextQueued`) and
`Usecase.ReclaimNextStalePlanningRequest`, wired into
`StoreProvisioningWorker.tick()` ahead of the existing claim loops. Fixed
2026-07-26 — see Rollout Log.

### Wave 4 — Multi-tenancy correctness & observability (no permission-model change)

MT4, MT6, MT9, MT10, MT11, MT16 (MT15 already in Wave 1). All in
`pkg/pdtenantdb` / `internal/backoffice/runtime`, none change the IAM/
permission contract (that's Wave 5). MT6 (`search_path` drops `public`) is
worth doing **before** PZEP-0015 (Phase 2 roles) ships, since Phase 2's
`SET LOCAL ROLE` test plan already wants a real mutual-invisibility test —
cleaning up `search_path` first removes a variable from that test. MT16
(observability + the boundary test) is partly satisfied by PZEP-0015's own
Test Plan (AC-3) but MT16 also asks for metrics, which PZEP-0015 doesn't
cover — keep MT16 open here for the metrics half even after PZEP-0015 lands.
**Needs its own PZEP** — MT9/MT10/MT11 change caching/lifecycle behavior
multiple call sites rely on.

### Wave 5 — Authorization model changes (permission contract)

MT5, MT7. Both change how store-level access control works — MT7 in
particular flips the authority from `owner_id` to IAM and needs "an IAM
listing capability" that doesn't exist yet per MT7's own fix note. These are
exactly the kind of change `docs/00-governance/agent-working-rule.md` says
must not be coded from a problem note directly. **Needs its own PZEP**,
likely after Wave 4, and depends on IAM's listing-capability gap being
scoped first (a sub-dependency to flag in that PZEP's Open Questions, not
resolve here).

### Wave 6 — Platform survivability, infra-only (`deployments/kubernetes/staging/`)

**Done** — [PZEP-0019](./PZEP-0019-staging-k8s-survivability-track-a.md),
implemented. Re-verifying PS9/PS10 against the actual current
manifests while writing that PZEP found the picture had partially moved
since this note was written: `backoffice`'s CPU limit was already bumped
(not still `50m`), `auth`/`catalog` already have correct `tcpSocket`
probes, and `onboarding` already has a correctly-ported `httpGet` probe
(just pointed at the still-fake `/healthz` handler). It also found a
**new, more severe bug PS10 didn't originally describe**: `grpcgateway`'s
`Service`/`containerPort` declare `8080` but the app actually listens on
`8088` (`GW_HTTP_PORT`) — a connectivity bug, not just a missing health
check. See PZEP-0019's Problem section for the full corrected per-service
table.

### Wave 7 — Platform tenancy scaling (`internal/onboarding` provisioning + `pdtenantdb`)

PS1, PS2, PS3, PS4, PS5, PS6, PS7, PS12 — matches that doc's "Track B",
**depends on MT1** (Phase 1 done; Phase 2 is PZEP-0015) since PS6/PS7 change
the same `clusterRoutePayload` code MT1's fix touches (documented ordering
constraint: "MT1 before PS6... both change how the cluster route payload is
produced; doing them in the other order means writing that code twice").
Internal ordering already specified in `platform-scaling-2026-07-23.md`:
PS7 → PS6 → PS2 → PS1/PS3 → PS12; PS4 and PS5 are each large enough
("placement behaviour... needs a migration plan") to warrant splitting out
of this wave into their own PZEPs rather than bundling with PS1-PS3/PS6/PS7/
PS12. **Needs its own PZEP** (or two, given PS4/PS5's note above) — changes
placement/scheduling behavior end to end.

## Suggested Overall Order

1. ~~Wave 1 (direct fixes)~~ — done, see Rollout Log.
2. ~~Wave 4's MT6 piece~~ — done, see Rollout Log.
3. ~~PZEP-0015 (MT1 Phase 2)~~ — done (code; not yet rolled out).
4. ~~Wave 3~~ — done, see Rollout Log. (Wave 2 removed — already resolved.)
5. ~~Remainder of Wave 4 (MT4, MT9, MT10, MT11, MT16)~~ — done, implemented
   via [PZEP-0017](./PZEP-0017-multitenancy-correctness-observability.md).
6. ~~Wave 5 (MT5, MT7)~~ — done, implemented via
   [PZEP-0018](./PZEP-0018-store-authorization-iam-single-authority.md).
7. ~~Wave 6~~ — done, implemented via
   [PZEP-0019](./PZEP-0019-staging-k8s-survivability-track-a.md).
8. Wave 7 — last, since it depends on MT1 Phase 1 (done) and benefits from
   Phase 2 being live first, plus PS4/PS5 need their own splits. Not yet
   scoped into its own PZEP.

## Rollout Log

**2026-07-26, Wave 1:**
1. Read current code for all six originally-scoped items before touching
   anything (per `docs/00-governance/agent-working-rule.md` — verify before
   acting, not just trust a problem note's status line). Found K9, K10, K11,
   P5 already fixed — see Summary for how that was discovered and
   `docs/07-problems/pkg-2026-07-07.md` /
   `docs/07-problems/onboarding-provisioning-2026-07-07.md` for the
   corrected per-item status notes.
2. P8: changed `scripts/dev/reconcile_legacy_tenant.go`'s `findRequest` to
   filter server-side (`collection.filters[0].field=subdomain`,
   `operator=eq`, `pageSize=1`) instead of a flat 100-row page scan.
   `gofmt`-clean, `go build ./...` and `go vet` pass.
3. MT15: `git mv` the duplicate `0015_create_customer_order_aggregate_store.sql`
   to `0017_create_customer_order_aggregate_store.sql`. Checked docker dev's
   `podzone_tenants` database directly (`docker exec podzone-postgres psql`)
   — found the old version string already recorded in
   `backoffice_schema_migrations` for all five tenant schemas present
   (`t_tenant_dev`, `t_tenant_23d2bab8_…`, `t_tenant_2e0df8f6_…`,
   `t_tenant_456a11db_…`, `t_e1a252ad_…`). Ran a direct `UPDATE ... SET
   version = '0017_...' WHERE version = '0015_...'` against each schema so
   the rename is treated as already-applied rather than silently re-run on
   next bootstrap (the migration's SQL is idempotent, so a re-run would have
   been harmless, just extra/duplicate history). `go build ./...`,
   `go test ./internal/backoffice/migrations/...`,
   `go test ./internal/onboarding/infrastructure/provisioning/provider/...`
   all pass. No `podzone-backoffice-dev` container was running at the time,
   so no service restart was needed to pick up the renamed embedded file.
4. Updated status footers in `pkg-2026-07-07.md`,
   `onboarding-provisioning-2026-07-07.md`, and
   `multitenancy-2026-07-23.md` (MT15) to reflect reality, and updated
   [PZEP-0015](./PZEP-0015-backoffice-per-tenant-db-role-phase-2.md)'s
   proposed migration number from `0017_` to `0018_` since `0017` is now
   consumed by MT15's fix.

Not done: no commit yet (working tree only, per this session's "don't commit
without explicit instruction" convention).

**2026-07-26, Wave 4's MT6 + PZEP-0015 (MT1 Phase 2):** MT6 fixed directly
(dropped `public` from `search_path` in both `pkg/pdtenantdb/manager.go` and
`internal/backoffice/migrations/migrations.go` — no extension needed
`public`, no test depended on it). PZEP-0015 implemented in full (per-tenant
Postgres roles via `SET LOCAL ROLE`, gated behind `TenantRoleEnabled`,
default off) — see that PZEP's own Rollout Log for detail. Both verified
against real Postgres 15 via testcontainers; full `pkg/pdtenantdb` +
`internal/backoffice` + `internal/onboarding` test suites green.

**2026-07-26, Wave 3:** Re-verified P2/P3/P4/P6/P7 against current code
before touching anything (same discipline as Wave 1) and found all five
already resolved by commits `73be9f5`/`27b26f3` (2026-07-09) — see
`docs/07-problems/onboarding-provisioning-2026-07-07.md`'s corrected status.
Found one new, real, still-open bug while re-verifying P7 (see this
document's Wave 3 section) and fixed it directly: added
`outputport.StoreRepository.ReclaimStalePlanning`,
`inputport.Usecase.ReclaimNextStalePlanningRequest`, wired into
`StoreProvisioningWorker.tick()`. Regenerated the two affected mocks
(`go tool mockery`) — confirmed via `git status` that only
`mock_Usecase.go`/`mock_StoreRepository.go` changed, no unrelated mock
churn. Added interactor-level and worker-level unit tests (mocked repo/
usecase, matching this package's existing test-doubles convention — no
Mongo-backed integration test exists for this package to extend, unlike
`pdtenantdb`/`migrations`). `go build ./...`, `go vet ./...`,
`golangci-lint run ./internal/onboarding/...` (0 issues), and the full
`internal/onboarding/...` test suite all pass.

**2026-07-26, Wave 4 (remainder) and Wave 5:** Both written up as their own
PZEPs ([PZEP-0017](./PZEP-0017-multitenancy-correctness-observability.md),
[PZEP-0018](./PZEP-0018-store-authorization-iam-single-authority.md), per
this document's own rule that cross-component/contract-changing waves need
explicit approval before coding) and, once approved, both implemented in
full the same session. See each PZEP's own Rollout Log for detail — in
short: MT4/MT9/MT10/MT11/MT16 (PZEP-0017) and MT5/MT7 (PZEP-0018) are all
now resolved. MT10's fix deviated from the Draft in one material way (the
real SQLSTATE is `42P01`, not the `3F000` originally guessed) after
verifying empirically against real Postgres rather than assuming — see
PZEP-0017's own Rollout Log. Full-repo `go test ./...` passes after both.

**2026-07-27, Wave 6:** Written up and implemented via
[PZEP-0019](./PZEP-0019-staging-k8s-survivability-track-a.md) (PS8-PS11).
Verifying the manifests before writing that PZEP found a new, more severe
`grpcgateway` port-mismatch bug (manifest declared `8080`, app actually
listened on `8088`) beyond what PS10 originally described — see
PZEP-0019's own Rollout Log for full detail.

With this, every wave in this PZEP except Wave 7 is closed.

## Affected Components
Spans every backend component and the Kubernetes staging manifests — see
each wave above. No frontend impact.

## API / DB / Event Contract Changes
- Wave 1, Wave 6: none.
- Wave 2: potential event/DLQ contract addition (K4's fix) — to be specified
  in that wave's own PZEP.
- Wave 3: none (internal worker state machine only).
- Wave 4: none.
- Wave 5: permission/authorization contract change — to be specified in that
  wave's own PZEP.
- Wave 7: placement/provisioning contract changes — to be specified in that
  wave's own PZEP(s).

## Permission Changes
None from this PZEP itself. Wave 5 is the only wave that changes permission
behavior, and it is explicitly deferred to its own PZEP.

## Security Considerations
K10 (Wave 1) is a live CORS bypass and should not wait on the rest of this
sequencing — recommend treating it as a standalone fix ahead of even
starting Wave 1's other items if there's any near-term external exposure.
Everything else in this PZEP is sequencing, not a new security decision.

## Test Plan
Not applicable to this document itself beyond Wave 1 (no code in the other
waves yet). Wave 1: `go build ./...`, `go vet` on the touched packages,
`go test ./internal/backoffice/migrations/...` and
`./internal/onboarding/infrastructure/provisioning/provider/...` — see
Rollout Log. Each remaining wave's own follow-up PZEP carries its own Test
Plan.

## Agent Implementation Plan
- TASK-0001: Wave 1 direct fixes. **Done** — see Rollout Log (P8, MT15
  actually changed; K9/K10/K11/P5 turned out already fixed).
- ~~TASK-0002: write Wave 2's messaging-reliability PZEP.~~ Not needed —
  Wave 2 removed, already resolved.
- TASK-0003: Wave 3. **Done** — see Rollout Log (all five original items
  turned out already fixed; one new residual gap found and fixed directly,
  no separate PZEP needed for it either — small enough for direct scoped
  fix per this document's own Wave 1 precedent).
- TASK-0004: write Wave 4's multi-tenancy-correctness PZEP (MT6 piece
  first, standalone, ahead of PZEP-0015). **Done, and implemented** — MT6
  fixed directly (Rollout Log); remainder (MT4, MT9, MT10, MT11, MT16)
  written up as and implemented via
  [PZEP-0017](./PZEP-0017-multitenancy-correctness-observability.md) (Done).
- TASK-0005: write Wave 5's authorization-model PZEP (after Wave 4).
  **Done, and implemented** — written up as and implemented via
  [PZEP-0018](./PZEP-0018-store-authorization-iam-single-authority.md) (Done).
- TASK-0006: write Wave 6's infra-survivability PZEP. **Done, and
  implemented** —
  [PZEP-0019](./PZEP-0019-staging-k8s-survivability-track-a.md) (Done).
- TASK-0007: write Wave 7's tenancy-scaling PZEP(s), split PS4/PS5 out per
  Proposed Solution. Not started.

TASK-0006 and TASK-0007 are not started — each is "go write that PZEP" when
the user picks it up. PZEP-0015 (Phase 2 of MT1, referenced throughout this
document) is implemented but not yet rolled out to any real environment —
see its own Status/Rollout Log.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: every open item in the four active problem notes is assigned to exactly one wave or explicitly excluded with a reason | (this document) | Manual cross-check against `docs/07-problems/*.md` |
| AC-2: cross-note ordering constraints (PS7→PS6, MT1→PS6) are preserved in the wave sequencing | (this document) | Manual review |
| AC-3: no wave containing a cross-component or contract-changing item is marked as directly implementable without its own PZEP | (this document) | Manual review |

## Open Questions
- Whether Wave 6 (infra-only) really needs its own PZEP versus being treated
  like Wave 1 (direct fixes) — leaning "needs one" here because PS9/PS11
  change real resource/storage allocations that deserve explicit sign-off,
  but this is a judgment call the user may want to overrule.
- Whether Wave 2 and Wave 3 should be split further once actually scoped —
  they're grouped by package here for locality, not because every item
  inside necessarily ships together.
- PS4 and PS5's exact split (one PZEP each, or a shared "tenant lifecycle"
  PZEP covering provisioning-mode support and rebalancing together, since
  both eventually need a placement-versioning concept) — left open for
  whoever starts Wave 7.
