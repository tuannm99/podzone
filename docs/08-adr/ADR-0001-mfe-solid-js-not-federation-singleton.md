# ADR-0001: MFE remotes do not share `solid-js` as a federation singleton

## Status
Accepted

## Date
2026-07-10

## Related Commit
`7f333ab` fix(mfe): fix tab active state not updating on click + hot reload
(preceded by `97051fa` which set up the singleton config this ADR reverses
for `solid-js` specifically)

## Context

Podzone's frontend uses `@originjs/vite-plugin-federation` to compose a HOST
shell (`frontend/src`) with three remotes (`frontend/apps/{iam,backoffice,
onboarding}`). The natural instinct — shared by most Module Federation
setups — is to declare every shared library as a `singleton: true` federation
share, including the UI framework itself (`solid-js`), so there is exactly
one framework instance across HOST and remotes.

Doing this broke reactivity across the board: clicking a tab in IAM,
Settings, or Provisioning updated the URL/state but the active-tab visual
never changed. Initial render was always correct; only click-triggered
updates failed.

Root cause: `@originjs/vite-plugin-federation` registers shared modules in
`globalThis.__federation_shared__` even in `vite dev` mode. With `solid-js`
declared as a singleton, a remote's `importShared("solid-js")` call
resolved to the **HOST's** solid-js module (loaded from a different URL than
the remote's own bundle). But the remote's compiled JSX still used **static
imports** of `createRenderEffect`/`createComponent` from its own bundle —
Vite's JSX transform emits these as static imports regardless of the
federation `shared` config. Result: `createSignal` (from the HOST's module)
and `createRenderEffect` (from the remote's own module) belonged to two
different ES module instances with two different `Listener` globals.
Setting a signal notified subscribers registered in one system; the render
effect was tracked in the other. The setter fired, nothing re-rendered.

Full incident detail and detection method: `agent/SOLID_STYLE_GUIDE.md`,
section "MFE / Vite-Plugin-Federation — SolidJS Reactive System Split".

## Decision

**Do not declare `solid-js` as a federation singleton in any remote's
`vite.config.ts`.** Each remote bundles its own `solid-js` copy. Only
`@tanstack/solid-router` and `@podzone/shared` remain federation singletons.

Consequence: HOST and remotes run in **separate SolidJS reactive scopes**.
`useContext` cannot cross that boundary. Auth context is bridged via a
`window.__pz_auth_value__` global instead of Solid context for any consumer
inside a remote (`packages/shared/auth/auth-context.ts`,
`src/modules/shell/AuthContextProvider.tsx`).

## Alternatives Considered

### Option A: Keep `solid-js` singleton, fix the reactive split some other way
Pros:
- Single framework instance is the theoretically "correct" Module
  Federation pattern; smaller total JS payload (no duplicate solid-js
  per remote).

Cons:
- No fix was found that doesn't require controlling Vite's JSX compiler
  output per-module (make it always route through `importShared` instead
  of static import) — not supported by `vite-plugin-solid` today.
- Any future remote added without deep awareness of this issue would
  silently reintroduce the bug.

### Option B (chosen): Un-share `solid-js`, bridge cross-boundary state manually
Pros:
- Structurally impossible to reintroduce the reactive-split bug — there
  is no shared instance to split.
- Small, well-understood cost: one `window` global for auth, extra
  solid-js bytes per remote bundle (acceptable at current remote count).

Cons:
- Slightly larger total bundle size (each remote ships its own solid-js).
- Anything else besides auth that needs to cross the HOST/remote boundary
  must use the same global-bridge pattern, not Solid context.

## Consequences

- Any new global state that must be shared HOST↔remote follows the
  `window.__pz_*__` bridge pattern, not `createContext`/`useContext`.
- `docs/03-architecture-detail-design/14-mfe-federation-contract.md` is the
  living contract; it documents this decision operationally. This ADR is
  the historical record of why.
- Adding a new remote must not add `solid-js` back to that remote's
  `shared` config — see the Rule of Thumb below.

## Rule Of Thumb

A package is safe as an MFE federation singleton only if **both** its
reactive primitives (`createSignal`, `createEffect`, ...) and its DOM
runtime primitives (`createRenderEffect`, `createComponent`) are served
through `importShared` — never split between a static import and
`importShared`. `solid-js` fails this test because JSX compilation always
emits static imports for the DOM runtime primitives.
