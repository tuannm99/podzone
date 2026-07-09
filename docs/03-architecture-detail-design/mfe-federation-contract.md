# MFE Federation Version Contract

Version contract for shared modules in the Podzone Module Federation setup.
The host shell (`frontend/`) is the authoritative provider of all shared
singletons. Remotes must declare these as `peerDependencies` and must **not**
bundle them.

## Shared Modules

| Package | Required version | Singleton | Notes |
|---|---|---|---|
| `solid-js` | `^1.9.14` | yes | One SolidJS runtime across all remotes |
| `@tanstack/solid-router` | `^1.170.17` | yes | One router instance; remotes register routes via `createRouteTree` |
| `@podzone/shared` | workspace | yes | Design system, services, auth context; provided by host alias |

## Rule For New Remotes

When adding a remote app to the federation:

1. **Declare shared packages as `peerDependencies`** in the remote's `package.json` at the versions above — do not add them to `dependencies`.
2. **Do not bundle** `solid-js`, `@tanstack/solid-router`, or `@podzone/shared` — the federation `shared` config in `vite.config.ts` handles deduplication.
3. **Expose only page components** via `exposes` — no contexts, no service modules, no store state.
4. **Add the remote's vite config** under `apps/<name>/vite.config.ts` with `root: __dirname` to prevent the host `index.html` being used as entry.
5. **Register in the host** `frontend/vite.config.ts` `remotes` map and add a corresponding `__MFE_<NAME>__` define constant.
6. **Add TypeScript declarations** for the exposed modules in `frontend/src/federation.d.ts`.
7. **Wrap the exposed page** with `remotePage(importFn, name)` in the remote's `routes.ts` for error isolation.

## Where Versions Come From

Canonical installed versions are in `frontend/package.json`. If either shared
library is upgraded, update this doc and verify all remote builds still pass.
