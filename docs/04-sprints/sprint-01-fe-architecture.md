# Sprint 1: Frontend Architecture — MFE Readiness

**Status:** complete  
**Duration:** 2–3 sprints (tracked as sprint 1a / 1b / 1c below)  
**Goal:** Make `frontend/` structurally ready for Module Federation extraction
without breaking the current monolith build or the recovery backbone flow.

**Completed:** All slices shipped across 8 commits:
- 1a/1b/1c (AuthContext, cross-module decoupling, monorepo scaffold) → `33c8ac`
- Sprint 2 (Module Federation infrastructure) → `7cfbe40`
- Sprint 3 (backoffice remote + Docker service) → `0885de4`
- Sprint 4 (move backoffice pages, fix remote build) → `aa05a79`
- Sprint 5 (IAM + onboarding remotes) → `e59f60a`
- Error boundary per remote (`remotePage`) → `60516c3`
- Docker Compose profiles per bounded context → `ff9c949`
- `packages/shared` source migration → `bf98ac8`

Each sprint 1x is independently shippable. Stop after 1a or 1b if backbone
stabilization takes priority again.

**Post-completion note (2026-07-11):** the "Target stack" below names pnpm
workspaces, and slice 1c.3 added `frontend/pnpm-workspace.yaml`. The project
settled on **npm + Vite `resolve.alias`** instead — no `pnpm-lock.yaml` or
npm `workspaces` field was ever added, and `@podzone/shared` resolution has
always gone through the alias in `frontend/vite.config.ts`, not a
package-manager workspace protocol. `pnpm-workspace.yaml` was dead weight
and has been deleted. Do not re-add it without an explicit decision to
migrate to pnpm.

---

## Context

Current state (audited 2026-07-09):

- 4 module boundaries exist: `shell`, `onboarding`, `iam`, `backoffice`
- Each module has its own `routes.ts` — route ownership is clear
- **Blockers for MFE extraction:**
  1. `services/` (tokenStorage, tenantStorage, storeStorage, auth) imported
     directly by all modules — no shell-provided contract
  2. `backoffice` imports `shell/workspace/context` and `iam/components` at
     runtime — cross-module coupling
  3. `app-router.tsx` is a single router that knows all module routes
  4. No shared package boundary — design system and service clients live inside
     the same Vite project

Target stack when ready: **pnpm workspaces + Vite Module Federation**
(`@originjs/vite-plugin-federation`).

---

## Sprint 1a — Auth Contract + Shell Context Provider

**Goal:** Replace direct `services/` imports in modules with a shell-provided
context so modules no longer read browser storage themselves.

### Why first

Every module touches `tokenStorage`, `tenantStorage`, `storeStorage`. This is
the highest-coupling surface. Fixing it unblocks the cross-module import audit
in 1b and reduces localStorage coupling to a single boundary.

### Slice 1a.1 — Define `AuthContext` contract

Agent role: Architect

Goal:

- Write a TypeScript interface `AuthContext` that a shell provider exposes and
  modules consume. No implementation yet.

Interface must cover:

```ts
interface AuthContext {
  token: () => string | null
  activeTenantId: () => string
  activeStoreId: () => string
  switchTenant: (id: string) => Promise<void>
  logout: () => void
}
```

Allowed files:

- `frontend/src/modules/shell/auth-context.ts` (new)

Out of scope:

- implementation
- changing any existing code

Acceptance criteria:

- interface is exported with named type
- no `import` of localStorage or tokenStorage inside this file
- TypeScript compiles: `cd frontend && npm run build`

---

### Slice 1a.2 — Implement `AuthContextProvider` in shell

Agent role: Frontend

Goal:

- Create a SolidJS context provider in `shell` that reads from
  `tokenStorage/tenantStorage/storeStorage` and exposes the `AuthContext`
  interface defined in 1a.1.

Allowed files:

- `frontend/src/modules/shell/AuthContextProvider.tsx` (new)
- `frontend/src/modules/shell/auth-context.ts`

Out of scope:

- changing consumer modules yet
- adding to router

Acceptance criteria:

- provider uses `createContext` + typed accessor hook `useAuthContext()`
- hook throws if called outside provider
- build passes

---

### Slice 1a.3 — Mount provider in Root, migrate `shell` module

Agent role: Frontend

Goal:

- Wrap `<Root>` in `<AuthContextProvider>`. Migrate all `shell` module files
  from direct `tokenStorage/tenantStorage/storeStorage` imports to
  `useAuthContext()`.

Allowed files:

- `frontend/src/solid/root.tsx`
- `frontend/src/modules/shell/**`

Out of scope:

- migrating `onboarding`, `iam`, `backoffice` yet

Acceptance criteria:

- `grep -r "tokenStorage\|tenantStorage\|storeStorage" frontend/src/modules/shell` → 0 hits
- build passes
- `npm run lint` passes

---

### Slice 1a.4 — Migrate remaining modules to `useAuthContext()`

Agent role: Frontend

Goal:

- Replace all `tokenStorage/tenantStorage/storeStorage` direct imports in
  `onboarding`, `iam`, and `backoffice` with `useAuthContext()`.

Allowed files:

- `frontend/src/modules/onboarding/**`
- `frontend/src/modules/iam/**`
- `frontend/src/modules/backoffice/**`

Acceptance criteria:

- `grep -r "tokenStorage\|tenantStorage\|storeStorage" frontend/src/modules/{onboarding,iam,backoffice}` → 0 hits
- build passes, lint passes
- backbone flow (sign in → backoffice) works in Docker dev

---

### Sprint 1a Exit Criteria

- No module imports browser-storage adapters directly
- `AuthContext` is the single storage boundary
- `tokenStorage/storeStorage/tenantStorage` files still exist — only their
  import surface has moved to shell
- Build and lint pass
- Backbone flow unbroken in Docker dev

---

## Sprint 1b — Cross-Module Coupling Removal

**Goal:** Remove the two cross-module runtime imports so `backoffice` has zero
compile-time dependency on `shell` internals or `iam` components.

Current violations:

```
backoffice → shell/workspace/context   (useTenantWorkspace)
backoffice → iam/components            (IamKeyValueBuilder, IamTrustPolicyBuilder)
```

### Slice 1b.1 — Promote `workspace/context` to shell public API

Agent role: Frontend

`shell/workspace/context.tsx` is already in `modules/shell/workspace/`. The
problem is that `backoffice` imports it directly, creating a compile-time
dependency on the shell module's internal path.

Goal:

- Export `useTenantWorkspace` from the `AuthContext` provider or a separate
  `WorkspaceContext` that the shell mounts alongside `AuthContextProvider`.
- This makes the dependency go through the context layer (runtime), not an
  import (compile-time).

Allowed files:

- `frontend/src/modules/shell/workspace/context.tsx`
- `frontend/src/modules/shell/AuthContextProvider.tsx`
- `frontend/src/modules/backoffice/**` (update import path)

Acceptance criteria:

- `grep -r "from '@/modules/shell" frontend/src/modules/backoffice` → 0 hits
- build passes

---

### Slice 1b.2 — Move IAM builder components to `solid/components`

Agent role: Frontend

`IamKeyValueBuilder` and `IamTrustPolicyBuilder` are generic policy editors,
not IAM-module-specific. `backoffice` imports them as cross-module dependency.

Goal:

- Move both components from `modules/iam/components/` to
  `solid/components/policy/` (new subfolder).
- Update all imports.

Allowed files:

- `frontend/src/solid/components/policy/` (new)
- `frontend/src/modules/iam/components/` (move source + update barrel)
- `frontend/src/modules/backoffice/**` (update imports)

Acceptance criteria:

- `grep -r "from '@/modules/iam" frontend/src/modules/backoffice` → 0 hits
- `grep -r "from '@/modules" frontend/src/solid` → 0 hits
- build passes, lint passes

---

### Slice 1b.3 — Module isolation audit

Agent role: Reviewer

Goal:

- Run a full cross-module import audit and confirm the 3 rules hold:
  1. modules do not import from other modules
  2. modules do not import from `services/` storage adapters directly
  3. `solid/` does not import from `modules/`

Verification:

```bash
# Rule 1 — cross-module imports (allowed: shell workspace context via useAuthContext only)
grep -rn "from '@/modules/" frontend/src/modules --include="*.ts" --include="*.tsx"

# Rule 2 — storage adapter direct imports in modules
grep -rn "tokenStorage\|tenantStorage\|storeStorage" frontend/src/modules --include="*.ts" --include="*.tsx"

# Rule 3 — solid importing modules
grep -rn "from '@/modules" frontend/src/solid --include="*.ts" --include="*.tsx"
```

All three must return 0 hits after 1b.1 and 1b.2.

---

### Sprint 1b Exit Criteria

- Zero cross-module imports
- Zero direct storage adapter imports in non-shell modules
- `solid/` layer is import-clean
- Module boundaries match what `SOLID_STYLE_GUIDE.md` requires

---

## Sprint 1c — Router Delegation + Monorepo Structure

**Goal:** Restructure the router so each module owns its own route tree, and
prepare the directory layout for future pnpm workspace extraction.

### Slice 1c.1 — Module-scoped route trees

Agent role: Frontend

Currently `app-router.tsx` imports `backofficeRouteComponents`,
`iamRouteComponents`, etc. and wires them into a single TanStack Router tree.

Goal:

- Each module exports a `createRouteTree(parentRoute)` function that returns
  its subtree.
- `app-router.tsx` becomes a shell file that creates the root route and mounts
  each module's subtree.

Allowed files:

- `frontend/src/modules/*/routes.ts` (add `createRouteTree` export)
- `frontend/src/app-router.tsx` (refactor to compose subtrees)

Out of scope:

- splitting into separate Vite builds
- changing route paths

Acceptance criteria:

- router still works end-to-end (backbone flow)
- `app-router.tsx` contains no route `path` strings other than `/` and `/auth/*`
- build passes

---

### Slice 1c.2 — `packages/shared` directory scaffold

Agent role: Architect + Frontend

Goal:

- Create `frontend/packages/shared/` with:
  - `auth/index.ts` — re-export `AuthContext` interface
  - `services/index.ts` — re-export typed API client functions (no axios
    instance, just typed wrappers)
  - `ui/index.ts` — re-export design system components from `solid/components`
- Do **not** move files yet — only re-export. This establishes the public
  contract surface that future workspace packages will expose.

Allowed files:

- `frontend/packages/shared/` (new)

Out of scope:

- pnpm workspace setup
- changing existing `solid/components` or `services/` internals

Acceptance criteria:

- `frontend/packages/shared/` exists with the three index files
- imports from `packages/shared/auth` resolve and build correctly from one
  test consumer (e.g. a backoffice ViewModel)

---

### Slice 1c.3 — `pnpm-workspace.yaml` + package.json scaffold

Agent role: Architect

Goal:

- Add `pnpm-workspace.yaml` at `frontend/` root declaring:
  ```yaml
  packages:
    - 'apps/*'
    - 'packages/*'
  ```
- Add `frontend/apps/backoffice/package.json` (stub, no code yet)
- Add `frontend/packages/shared/package.json`

This does not change the current build — the monolith Vite build in
`frontend/` root still works. This is directory scaffolding only.

Allowed files:

- `frontend/pnpm-workspace.yaml` (new)
- `frontend/apps/backoffice/package.json` (new, stub)
- `frontend/packages/shared/package.json` (new)

Out of scope:

- moving source files into `apps/`
- changing `frontend/package.json` or Vite config
- CI changes

Acceptance criteria:

- `cd frontend && npm run build` still passes (root Vite build unchanged)
- `frontend/pnpm-workspace.yaml` is valid YAML

---

### Sprint 1c Exit Criteria

- Each module exports its own route subtree
- `packages/shared/` exists with typed re-export surfaces
- `pnpm-workspace.yaml` scaffold is in place
- Root Vite monolith build still works — no regression

---

## Overall Exit Criteria (after 1a + 1b + 1c)

| Criterion | Verification |
|---|---|
| No cross-module imports | `grep -rn "from '@/modules/" frontend/src/modules` → shell workspace only via context |
| No direct storage imports in modules | `grep -rn "tokenStorage\|tenantStorage\|storeStorage" frontend/src/modules` → 0 |
| `solid/` clean | `grep -rn "from '@/modules" frontend/src/solid` → 0 |
| Each module owns its route subtree | `app-router.tsx` composes subtrees |
| `packages/shared/` surface exists | auth, services, ui index files |
| pnpm workspace scaffold | `pnpm-workspace.yaml` present |
| Build green | `cd frontend && npm run build` |
| Lint green | `cd frontend && npm run lint` |
| Backbone flow | sign in → backoffice → API call works in Docker dev |

## What This Does NOT Cover (next sprint)

- Splitting into separate Vite builds (Module Federation config)
- Moving `apps/backoffice` source files into `frontend/apps/backoffice/`
- Publishing `packages/shared` as a private npm package
- CI changes for parallel MFE builds
- Remote entry URL config per environment

## Source Docs

- Style guide: `agent/SOLID_STYLE_GUIDE.md`
- Design system: `docs/03-architecture-detail-design/design-system.md`
- FE audit: `docs/03-architecture-detail-design/frontend-solid-audit.md`
- Sprint template: `docs/04-sprints/sprint-template.md`
