# MFE Federation Version Contract

Version contract for shared modules in the Podzone Module Federation setup.
Decision record: [ADR-0001](../08-adr/ADR-0001-mfe-solid-js-not-federation-singleton.md).

## Shared Modules

| Package | Singleton in remotes? | Notes |
|---|---|---|
| `solid-js` | **No — deliberately excluded** | See "Why solid-js is not a singleton" below. Each remote bundles its own copy. |
| `@tanstack/solid-router` | yes | One router instance; remotes register routes via `createRouteTree` |
| `@podzone/shared` | yes (config only, see note) | Design system, services, auth context |

## Why `solid-js` Is Not A Singleton

`@originjs/vite-plugin-federation` registers shared modules in
`globalThis.__federation_shared__` even in `vite dev` mode. If a remote
declares `solid-js` as a federation singleton, `importShared("solid-js")`
resolves to the **host's** solid-js module, while the remote's own JSX
templates use **static imports** of `createRenderEffect`/`createComponent`
from the remote's own bundle. These are two different ES module instances —
`createSignal` subscribes listeners in one, `createRenderEffect` tracks in
the other, so UI does not update on click (it can still render correctly on
first paint, since effects run once synchronously regardless of
subscription).

This bug and its fix are documented in full in
`agent/SOLID_STYLE_GUIDE.md` ("MFE / Vite-Plugin-Federation — SolidJS
Reactive System Split"). Do not re-add `solid-js` to a remote's federation
`shared` config without reading that section first.

Because `solid-js` is not shared, `useContext` cannot cross the host/remote
boundary either (each side has its own reactive scope). Auth context is
bridged via a `window.__pz_auth_value__` global instead of Solid context for
consumers inside remotes — see `packages/shared/auth/auth-context.ts`.

## `@podzone/shared` Is Not A Real Federation Singleton

Vite's `resolve.alias` (`@podzone/shared -> ../../packages/shared`) resolves
the import path **before** the federation plugin sees it, so the package
name never matches the `shared` config key. `@podzone/shared` is always
bundled locally into each remote's own output — the `singleton: true` entry
in `vite.config.ts` has no effect. This is harmless (the source is identical
across bundles), but do not rely on it for true singleton semantics (e.g. a
module-level cache or event bus inside `@podzone/shared` will NOT be shared
across host/remote).

## Rule For New Remotes

When adding a remote app to the federation:

1. **Do not add `solid-js` to the remote's federation `shared` config.** Only
   `@tanstack/solid-router` and `@podzone/shared` go there.
2. **Expose only page components** via `exposes` — no contexts, no service
   modules, no store state.
3. **Add the remote's vite config** under `apps/<name>/vite.config.ts` with
   `root: __dirname` to prevent the host `index.html` being used as entry.
4. **Register in the host** `frontend/vite.config.ts` `remotes` map and add a
   corresponding `__MFE_<NAME>__` define constant.
5. **Add TypeScript declarations** for the exposed modules in
   `frontend/src/federation.d.ts`.
6. **Wrap the exposed page** with `remotePage(importFn, name)` from
   `@podzone/shared/ui/remotePage` in the remote's `routes.ts` for error
   isolation.
7. **Use `npm run dev:<name>`** (`vite build --watch` + `vite preview`) for
   Docker hot reload — `vite dev` mode does not rebuild remote bundles the
   host consumes. See `deployments/docker/services.yml`.

## Where Versions Come From

Canonical installed versions are in `frontend/package.json`. If a shared
library is upgraded, verify all remote builds still pass — federation
version mismatches fail silently at runtime, not at build time.
