# ADR-0002: Monorepo dependency resolution uses npm + Vite `resolve.alias`, not pnpm workspaces

## Status
Accepted

## Date
2026-07-11

## Related Commit
`b0f9254` chore(frontend): remove unused pnpm-workspace.yaml

## Context

`docs/04-sprints/sprint-01-fe-architecture.md` (slice 1c.3, commit `33c8a5f`)
originally scaffolded `frontend/pnpm-workspace.yaml`, naming "pnpm workspaces
+ Vite Module Federation" as the target stack for the `frontend/` monorepo
(HOST `frontend/src` + remotes `frontend/apps/*` + shared lib
`frontend/packages/shared`).

That migration was never completed. No `pnpm-lock.yaml` was ever committed,
`frontend/package.json` never declared an npm `"workspaces"` field, and
`@podzone/shared` resolution has worked from day one through a plain Vite
`resolve.alias` entry (`'@podzone/shared': path.resolve(__dirname,
'./packages/shared')`) in `frontend/vite.config.ts`, not through any
package-manager workspace protocol. `frontend/package-lock.json` (npm) has
been the only lockfile in the repo the entire time. `pnpm-workspace.yaml`
sat unused for a full sprint cycle before being noticed and removed.

## Decision

**The `frontend/` monorepo resolves cross-package imports (`@podzone/shared`,
`@iam`, `@onboarding`, `@backoffice`) through Vite's `resolve.alias`, using
plain npm for package management.** No pnpm, no npm workspaces protocol.

## Alternatives Considered

### Option A: Finish the pnpm workspaces migration
Pros:
- pnpm workspaces would give real dependency isolation per package
  (`apps/iam` and `apps/onboarding` could have divergent dependency
  versions if ever needed) and faster installs via pnpm's content-
  addressable store.
- Matches the sprint doc's originally stated target.

Cons:
- Nobody had used it in practice for an entire sprint — the FE build,
  Docker dev, and CI all worked correctly on npm + Vite alias the whole
  time, meaning the actual pain point workspaces would solve was never
  felt.
- Migration cost (converting `package-lock.json` to `pnpm-lock.yaml`,
  updating all Docker build/dev scripts, updating CI) with no concrete
  problem it fixes today.

### Option B (chosen): Keep Vite `resolve.alias`, drop the pnpm scaffold
Pros:
- Zero migration cost — this is what was already running.
- One lockfile, one package manager, fewer moving parts for a project
  already in `docs/06-recovery/recovery-plan.md` stabilization mode,
  where the explicit rule is "do not expand feature breadth" — a package
  manager migration is exactly that kind of scope creep.

Cons:
- All `apps/*` and the HOST share one flat `node_modules` and one
  dependency version per package — acceptable at current scale (3
  remotes + 1 host + 1 shared lib), would need revisiting if remotes
  ever need genuinely divergent dependency versions.

## Consequences

- Do not re-add `pnpm-workspace.yaml` without a new ADR that names a
  concrete problem it solves.
- New shared packages under `frontend/packages/` get a new
  `resolve.alias` entry in `frontend/vite.config.ts`, not a workspace
  `package.json` dependency declaration.
- `docs/04-sprints/sprint-01-fe-architecture.md` carries a post-completion
  note pointing here; this ADR is the decision record, that note is the
  pointer.

## Rule Of Thumb

If a future FE dependency-isolation problem shows up (version conflicts
between remotes, slow installs, need for private/local packages with real
semver), that is the trigger to open a new ADR proposing pnpm/npm
workspaces — not something to scaffold speculatively ahead of the need.
