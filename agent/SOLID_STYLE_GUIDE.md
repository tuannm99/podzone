# Podzone Solid Style Guide

Use this guide for every change under `internal/ui-podzone`.

## Required Workflow

1. Read the feature, route, service adapter, and adjacent checks before editing.
2. Identify remote state, local UI state, form state, and derived state.
3. Keep the change inside one feature boundary unless extracting stable shared UI.
4. Run `npm run format`, `npm run lint`, `npm run build`, and
   `git diff --check`.
5. Verify affected desktop and mobile workflows through the Docker hot-reload runtime.
6. Do not start a second dev server when Docker already serves the UI.

## Project Boundaries

Use this shape:

```text
src/
  modules/<feature>/
    pages/
    components/
    state/
    routes.ts
  services/<feature>/
    types.ts
    queries.ts
    commands.ts
  solid/
    components/common/
    forms/
    pagination/
    workspace/
```

- Keep route pages as composition roots.
- Keep business-facing UI and state inside its feature module.
- Keep HTTP and GraphQL transport code in `services`.
- Keep generic, domain-neutral UI in `solid/components/common`.
- Do not import another feature's page, state, or internal component.
- Keep compatibility facades such as `services/iam.ts` thin.
- Keep files below 500 lines. Prefer pages below 250 lines and components below
  200 lines.

## Solid Reactivity

Solid components run once. Reactive getters update their consumers directly.
Do not reason in React re-render cycles.

### Props

- Read reactive props as `props.value`.
- Use `splitProps` when forwarding or separating props.
- Do not destructure reactive props into local values.
- Wrap a prop read in an accessor when a child API requires `Accessor<T>`.

```tsx
// Good
const selected = () => props.selected
return <Show when={props.order}>{props.order?.id}</Show>

// Bad: values can become stale
const { order, selected } = props
```

See the official [Solid props guide](https://docs.solidjs.com/concepts/components/props).

### Primitive Selection

- Use `createSignal` for independent scalar UI state.
- Use `createStore` for cohesive object or form state.
- Use `createMemo` for derived values that are read in multiple places or are
  expensive.
- Use `createResource`, router queries, or `createAsync` for remote reads.
- Use `createEffect` only to synchronize with an external side effect such as
  browser storage, focus, scrolling, subscriptions, or imperative libraries.
- Do not use effects to derive one signal from another.
- Do not manually enumerate signal reads merely to make an effect subscribe.
- Register timers and event listeners with `onCleanup`.

See [fine-grained reactivity](https://docs.solidjs.com/advanced-concepts/fine-grained-reactivity).

### Control Flow

- Use `For` for reactive collections.
- Use `Show` for one condition.
- Use `Switch` and `Match` for mutually exclusive states.
- Do not use nested ternaries for application states.
- Render explicit loading, error, empty, and success states.

## Async Data

- Prefer route-level query/preload for primary page data.
- Prefer a resource/query primitive for dependency-driven reads.
- Keep mutation functions explicit and invalidate or refetch affected queries.
- Prevent stale responses from overwriting newer state.
- Support request cancellation where the transport permits it.
- Do not duplicate remote data into another signal unless creating an edit draft.
- Keep transport result mapping in the service adapter.
- Never decide authorization only in the UI; render backend authorization results.

For route data, follow Solid Router
[preloading](https://docs.solidjs.com/solid-router/data-fetching/how-to/preload-data).

## State Ownership

- Place state at the lowest common owner.
- Use context only for a real feature subtree.
- Type every context value; `any` is forbidden.
- Split command actions, query state, and presentation helpers when the model grows.
- Name owner-bound custom reactive primitives `createX`.
- Reserve `useX` for context consumers and framework primitives.
- Return accessors for reactive reads instead of snapshots.
- Do not build one god view-model for an entire admin application.

## Routing

- Lazy-load route components.
- Use router links and navigation APIs for internal navigation.
- Do not use raw anchors for internal routes.
- Store shareable state in typed search parameters:
  - filters
  - sort
  - page or cursor
  - selected resource
  - active workspace section
- Use route guards for authentication and tenant access.
- Keep preload functions pure and deduplicated.

See Solid Router [lazy loading](https://docs.solidjs.com/solid-router/advanced-concepts/lazy-loading).

## Forms

- Use one typed form store for one cohesive form.
- Keep initial values, value types, and validators in a nearby `forms.ts`.
- Do not create one signal per field.
- Validate before mutation.
- Set submitting state immediately before mutation and clear it in `finally`.
- Reset only after a successful mutation.
- Map backend field errors to their fields when available.
- Keep server errors separate from field validation errors.
- Use native control semantics and labels.

## Collection And Admin UX

- Use a table for operational collections.
- Use cards only for repeated visual objects that require card presentation.
- Provide search, filters, sort, loading, error, empty, and pagination states.
- Use server-side cursor pagination for unbounded collections.
- Treat client pagination as a temporary optimization for bounded data only.
- Persist filter, sort, page, and selection in URL state when shareable.
- Open create and edit workflows in a drawer, modal, or dedicated route.
- Open one detail panel or route at a time.
- Do not mount full detail editors for every list row.
- Keep bulk actions in a contextual toolbar after selection.
- Ensure tables scroll horizontally on narrow screens.

## Component Design

- Prefer explicit typed props over a generic model object.
- Keep components focused on one workflow or presentation responsibility.
- Extract repeated behavior into a custom primitive, not a base component class.
- Do not copy React patterns such as `useCallback`, blanket memoization, or
  immutable state spreading for every update.
- Avoid wrapper components that only rename an HTML element.
- Use shared primitives consistently for fields, feedback, tables, pagination,
  dialogs, and buttons.
- Do not nest decorative cards.

## Services

- Split large adapters by feature and query/command responsibility.
- Export transport-independent DTO types from `types.ts`.
- Keep endpoint paths, GraphQL documents, and response mapping out of pages.
- Return a consistent typed result or throw a typed application error; do not mix
  both styles inside one feature.
- Add cursor/page parameters to collection contracts before data becomes unbounded.
- Remove unused client dependencies instead of keeping parallel data stacks.

## Browser Storage

- Access local and session storage through typed adapters.
- Namespace keys by application, tenant, and store where relevant.
- Handle malformed JSON and unavailable storage.
- Do not use browser storage as the source of truth for authorization.
- Keep tokens out of logs and UI errors.

## Accessibility

- Use semantic buttons, links, forms, headings, tables, and navigation.
- Give icon-only actions an accessible name and tooltip.
- Preserve keyboard navigation and visible focus.
- Associate every input with a label and error message.
- Use `aria-current`, `aria-expanded`, and dialog attributes where applicable.
- Do not encode status using color alone.

## TypeScript

- Enable and preserve strict type checking.
- Do not introduce `any` or suppress `no-explicit-any`.
- Model finite UI states with unions.
- Model IDs consistently as strings at the UI boundary unless arithmetic is needed.
- Keep API DTOs separate from form values and presentation models.
- Avoid broad `Record<string, unknown>` when a route search schema or DTO is known.

## Verification

- Frontend unit tests are not required in this repository.
- Use lint and production build as mandatory static gates.
- Use E2E or browser smoke checks for:
  - login and registration
  - workspace and store selection
  - onboarding progress
  - IAM permission denial and management
  - partner create/edit/filter
  - order creation, routing, shipment, issue, and settlement
- Verify loading, error, empty, success, validation, and permission states in the
  affected workflow.

## Formatting And Verification

- Keep one Prettier configuration for the frontend.
- Format touched TypeScript and TSX consistently.
- Run:

```bash
cd internal/ui-podzone
npm run format
npm run lint
npm run build
git diff --check
```

- Verify no touched file exceeds the size limits.
- Verify no horizontal page overflow at mobile and desktop widths.
- Verify Docker hot reload instead of starting duplicate Vite processes.

## Prohibited Patterns

- Destructuring reactive props.
- Fetching remote data through unguarded `createEffect`.
- Context or view models typed as `any`.
- One signal per form field.
- Rendering an unbounded collection without server pagination.
- Rendering every row's full editor or detail controls.
- Raw internal `<a href>` navigation.
- Authorization based only on hidden buttons or frontend checks.
- Fire-and-forget mutation without loading and error handling.
- Duplicating API types across pages.
- Adding shared abstractions before two real consumers exist.

## Review Checklist

- Props stay reactive.
- Remote reads have ownership, cancellation, and stale-response behavior.
- Derived values use accessors or memos.
- Contexts are typed and feature-scoped.
- Internal navigation uses the router.
- Forms are typed and reset only after success.
- Collections use the correct pagination strategy.
- Permission failures expose actionable backend error details.
- Loading, error, empty, and success states exist.
- Keyboard, focus, labels, and responsive overflow are verified.
- Lint, build, diff checks, and affected workflow smoke checks pass.
