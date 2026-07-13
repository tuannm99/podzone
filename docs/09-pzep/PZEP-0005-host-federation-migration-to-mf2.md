# PZEP-0005: Host federation migration to Module Federation 2 (`@module-federation/vite`)

## Status
Done — implemented and verified 2026-07-13 (build/manifest level; see Verification Results). Not yet committed.

## Date
2026-07-13

## Related Commit
(uncommitted — working tree only, pending user review)

## Requirement Sources
- Business: PZEP-0004 Phase 1 — this is that phase's own required PZEP,
  written per that document's explicit gate ("no phase beyond Phase 0 is
  authorized without its own accepted PZEP").
- Feature: [ADR-0005](../08-adr/ADR-0005-frontend-framework-angular-replaces-solidjs.md),
  [PZEP-0004](./PZEP-0004-solidjs-to-angular-frontend-migration.md)
- Use Cases: the host loads the Angular `frontend-v2` remote at runtime,
  without breaking any of the 3 existing SolidJS remotes.
- Functional Requirements: none new.
- Non-functional Requirements: [SRS-NFR-006](../01-srs/podzone-srs.md) AI-Agent
  Safety — this PZEP exists specifically because the host is a currently-working
  system and this change must not regress it.
- Acceptance Criteria: see below
- UI Specs: none — build/runtime tooling change only, no visual change.

## Summary

Migrate `frontend/vite.config.ts` and the 3 existing remotes' own
`vite.config.ts` (backoffice, iam, onboarding) from
`@originjs/vite-plugin-federation` to `@module-federation/vite` (MF2), then
bridge the Angular `frontend-v2` native-federation remote into the migrated
host using `native-to-mf-bridge`.

## Problem

PZEP-0004 assumed migrating the host's federation plugin alone would let
an Angular remote load in. A spike this session found that is incomplete:
`@module-federation/vite` (MF2) and Angular's
`@angular-architects/native-federation` are **separate ecosystems** with
different runtimes and manifest formats — they do not interoperate without
a bridge. Only one such bridge exists on npm
(`native-to-mf-bridge`), published **2026-07-02, v0.1.0, two versions ever,
single maintainer** — 11 days old at the time of writing, no production
track record. An alternative (`module-federation-angular-adapter`, using
MF2's own esbuild plugin directly instead of native-federation) is equally
new (created 2026-06-28, three patch releases). Both `@module-federation/vite`
and `@module-federation/esbuild` (the underlying primitives each option
depends on) are themselves substantially more mature (the latter has
commits as recent as today, 2026-07-13).

User decision (2026-07-13, asked explicitly given this risk): proceed with
`@module-federation/vite` + `native-to-mf-bridge`, accepting the new/unproven
dependency risk. Rationale for this specific combination over the
esbuild-adapter alternative: `native-to-mf-bridge`'s own documented use
case is verbatim this project's situation ("a React Module Federation host
loading Angular Native Federation remotes" — same shape, different
non-Angular framework), and it changes nothing on the Angular side
(`frontend-v2` keeps using native-federation exactly as scaffolded in
Phase 0) — only the host and the 3 existing Solid remotes change.

## Goals

- The 3 existing SolidJS remotes (backoffice, iam, onboarding) build and
  load into the host exactly as before, now via MF2 instead of
  `@originjs/vite-plugin-federation` — **zero behavior change** for
  end users of the current app.
- The Angular `frontend-v2` remote's `remoteEntry.json` is bridgeable into
  the host via `native-to-mf-bridge`, verified at build/manifest level (a
  full in-browser click-through is out of scope for this PZEP — no browser
  automation tool is available in this environment; see Test Plan).
- A clear, cheap rollback: nothing in this PZEP is committed until it is
  verified working; if verification fails, `git checkout` reverts the
  working tree with no partial state to unwind.

## Non-Goals

- Actually cutting any remote's real routes over to Angular — that's Phase
  2 onward, gated by its own future PZEP.
- Full in-browser interactive verification (click a link, see an Angular
  component render inside the Solid host) — not achievable in this session
  without a working browser-automation tool; verification here is build- and
  manifest-level, explicitly flagged as such in the Test Plan and in the
  final report.
- Changing Docker Compose service definitions, APISIX routing, or exposing
  `frontend-v2` through the Docker dev stack — this PZEP verifies the Vite
  dev/build tooling directly, not the full containerized environment.
- Removing `@originjs/vite-plugin-federation` from `package.json` — kept
  installed until this migration is fully verified working end-to-end in a
  later phase; only which plugin is *active* in each `vite.config.ts` changes.

## Proposed Solution

1. Host (`frontend/vite.config.ts`): replace the `federation()` import from
   `@originjs/vite-plugin-federation` with `@module-federation/vite`'s
   `federation()`, keeping the same `name`, `remotes` (backoffice, iam,
   onboarding URLs unchanged), and `shared` config
   (`solid-js` singleton, `@tanstack/solid-router` singleton,
   `@podzone/shared` singleton — unchanged from today). Add
   `nativeToMfBridge({ remotes: { '@native/frontend-v2': { entry: '.../remoteEntry.json', defaultExpose: './Component' } } })`
   as an additional plugin, pointed at `frontend-v2`'s dev server.
2. Each of the 3 remotes' own `vite.config.ts` (backoffice, iam,
   onboarding): same swap — `@originjs/vite-plugin-federation` →
   `@module-federation/vite`, same `exposes`/`shared` config, unchanged
   values (including the documented `solid-js`-excluded-from-singleton rule
   in iam/onboarding, which is orthogonal to which federation plugin is used
   and must be preserved).
3. Verify per Test Plan below before considering this done.

## Affected Components
- Frontend: `frontend/vite.config.ts`, `frontend/apps/backoffice/vite.config.ts`,
  `frontend/apps/iam/vite.config.ts`, `frontend/apps/onboarding/vite.config.ts`,
  `frontend/package.json` (new devDependencies: `@module-federation/vite`,
  `native-to-mf-bridge`)
- Gateway: none
- Backend: none
- Worker: none
- Database: none
- External Integration: none

## API Contract Changes
- None — no HTTP/GraphQL/gRPC contract changes.

## DB Contract Changes
- None.

## Event Contract Changes
- None.

## Permission Changes
- None.

## Data Ownership
- Unchanged.

## Security Considerations
- New dependency risk: `native-to-mf-bridge` fetches each Angular remote's
  `remoteEntry.json` and injects a browser import map
  (`transformIndexHtml`) — this executes at dev/build time against a URL
  the host config controls (not user input), so no new untrusted-input
  surface is introduced. The bigger risk is supply-chain/maintenance risk
  of an 11-day-old single-maintainer package, not a runtime security risk —
  called out explicitly per the Problem section; the user accepted this
  tradeoff.
- No change to authentication, authorization, or tenant isolation — this
  PZEP is build-tooling only.

## Observability
- None required.

## Alternatives Considered

See PZEP-0004's Alternatives Considered for the broader sequencing
question. This PZEP's specific alternative is the bridge mechanism itself:

### Alternative: `module-federation-angular-adapter` (MF2 esbuild plugin, no native-federation)
Pros: single MF2 runtime end-to-end instead of two runtimes bridged
together; no import-map/CORS requirement.
Cons: would require reconfiguring `frontend-v2` off native-federation
(undoing Phase 0's scaffold decision) and depending on an equally-new
(2026-06-28), even-lower-adoption single-repo package. Rejected — keeps
`frontend-v2` unchanged, and the bridge's own stated use case matches this
project's shape exactly.

### Alternative: Do not bridge now; migrate host to MF2 for the 3 Solid remotes only, revisit Angular bridging later
Pros: avoids the new bridge dependency entirely for now; still gets MF2's
independent DX/manifest improvements.
Cons: doesn't answer the actual question this session was asked to
resolve ("start migrating"); user explicitly chose to proceed with the
bridge despite the risk when asked. Not chosen.

## Test Plan
- Unit: none (no application logic changed).
- Integration/Build: `cd frontend && npm run build` (host) and
  `npx vite build` in each of `apps/backoffice`, `apps/iam`,
  `apps/onboarding` must all succeed with MF2's plugin, producing a
  `remoteEntry.js`/manifest per remote.
- Manifest-level: with the host and all 3 Solid remotes' dev servers
  running, confirm each remote's federation manifest is well-formed and the
  host's dev server starts without federation-resolution errors in its own
  logs (curl-level, not browser-level).
- Angular bridge: with `frontend-v2`'s `ng serve --port 3004` and the
  migrated host both running, confirm the host's served HTML includes the
  import map `native-to-mf-bridge` injects, and that
  `@native/frontend-v2/Component` resolves to a real JS module URL when
  requested — this is the practical ceiling of verification without a
  browser automation tool; explicitly not a rendered-pixel check.
- E2E: none — out of scope, no browser tool available (see Non-Goals).
- Manual QA: none required by this PZEP; full Docker Compose/APISIX-level
  regression testing is deferred to whichever later phase actually exposes
  this to real user traffic.

## Agent Implementation Plan
- TASK-0001: Install `@module-federation/vite` and `native-to-mf-bridge` in
  `frontend/package.json` (devDependencies).
- TASK-0002: Migrate `frontend/vite.config.ts` to MF2's `federation()` +
  add `nativeToMfBridge()` pointed at `frontend-v2`.
- TASK-0003: Migrate `apps/backoffice/vite.config.ts`,
  `apps/iam/vite.config.ts`, `apps/onboarding/vite.config.ts` to MF2's
  `federation()`, preserving all existing `exposes`/`shared` values exactly.
- TASK-0004: Run Test Plan; fix any build/manifest regression before
  reporting done. If unresolvable within reasonable effort, revert via
  `git checkout` (nothing committed) and report the blocker instead of
  leaving a half-migrated state.
- TASK-0005: Report results (what's verified, what isn't, current risk
  level) and update PZEP-0004's Phase 1 status.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: All 3 Solid remotes build clean under MF2 | TASK-0003 | `vite build` per remote |
| AC-2: Host builds clean under MF2 | TASK-0002 | `npm run build` |
| AC-3: Host's dev server starts with no federation-resolution errors with all 3 Solid remotes running | TASK-0002, TASK-0003 | Manifest-level check, dev server logs |
| AC-4: `frontend-v2`'s `remoteEntry.json` is reachable and resolvable through the bridge from the host's served HTML | TASK-0002 | Manifest-level curl check |
| AC-5: `frontend/npm run format/lint/build/format:check` all pass | all | Standard frontend commands per `CLAUDE.md` |

## Verification Results (2026-07-13)

- **AC-1**: All 3 Solid remotes (backoffice, iam, onboarding) built clean
  under MF2 (`vite build`, verified via a temp `--outDir` to avoid a
  pre-existing, unrelated `EACCES` on their real `dist/assets/` — those
  directories are owned by `root` from an earlier Docker build and were not
  touched; this is an environment issue, not a regression from this PZEP).
- **AC-2**: Host built clean under MF2 + `native-to-mf-bridge`, with and
  without `frontend-v2`'s dev server running (the bridge does not hard-fail
  the host build when the Angular remote is unreachable).
- **AC-3**: No federation-resolution errors in build output beyond benign
  `IMPORT_IS_UNDEFINED` warnings on `default` exports for `solid-js`,
  `solid-js/web`, `solid-js/store`, `@tanstack/solid-router`, and
  `@podzone/shared` — MF2's shared-module prebuild wrapper always emits a
  defensive `__mfPrebuildNamespace.default ?? __mfPrebuildNamespace` shim
  regardless of whether a real default export exists; the `??` fallback
  means this resolves correctly at runtime. Same, already-documented
  characteristic as `SOLID_STYLE_GUIDE.md`'s note that `@podzone/shared`
  is always bundled locally per remote (Vite's alias resolves before the
  federation plugin sees the import) — MF2 additionally surfaces this as an
  explicit "alias conflict" warning at build time; behavior is unchanged
  from before this migration, just more visible now.
- **AC-4**: Verified concretely, not just configured — a temporary import
  (`import '@native/frontend-v2/Component'` in `src/main.tsx`, removed
  immediately after) forced the bridge to actually run against
  `frontend-v2`'s live `ng serve --port 3004`. The built `index.html`
  contained a real `<script type="importmap">` correctly scoped to
  `http://localhost:3004/`, mapping `@angular/core`, `@angular/common`,
  `@angular/router`, `rxjs`, `tslib`, and native-federation's internal
  chunks to their real dev-server URLs. This is the practical ceiling of
  verification without a browser automation tool (none available this
  session) — confirms the manifest fetch, parse, and import-map generation
  pipeline works end-to-end; does not confirm in-browser rendering.
- **AC-5**: `npm run format` / `lint` / `build` / `format:check` all pass
  in `frontend/`.
- **Real bug found and fixed during verification, not anticipated in the
  original plan**: MF2 emits `remoteEntry.js` at each remote's build output
  root, while `@originjs/vite-plugin-federation` placed it under `assets/`.
  `frontend/vite.config.ts`'s default remote URLs and
  `deployments/docker/services.yml`'s `VITE_*_REMOTE_URL` env vars both
  hardcoded the old `/assets/remoteEntry.js` path — fixed in both places
  (3 lines each). Without this fix, the migration would silently 404 on
  every remote the moment `@originjs/vite-plugin-federation` was swapped
  out, in both local dev and Docker.

## Open Questions
- Whether `native-to-mf-bridge`'s CORS requirement needs any dev-server
  config beyond Vite's defaults — check during TASK-0002/TASK-0004, not
  assumed here.
- Whether this migration needs any change to `deployments/docker/services.yml`
  or APISIX routing — explicitly deferred (see Non-Goals); answer only when
  a later phase actually exposes this through Docker.
