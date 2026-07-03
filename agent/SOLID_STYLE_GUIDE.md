# Podzone Solid Rules

Mandatory for `internal/ui-podzone`.

## Workflow

1. Read the route, feature state, service adapter, and adjacent UI first.
2. Identify remote, form, local, and derived state and their owners.
3. Keep changes inside one feature unless extracting a stable shared primitive.
4. Use Docker hot reload; do not start another dev server.
5. Run:

```bash
cd internal/ui-podzone
npm run format
npm run lint
npm run build
npm run format:check
git diff --check
```

Smoke-test affected desktop/mobile loading, error, empty, success, validation,
permission, and pagination states. Frontend unit tests are optional.

## Boundaries

```text
src/modules/<feature>/    route pages, panels, feature state/ViewModels
src/services/<feature>/   HTTP/GraphQL DTOs, queries, commands, mapping
src/solid/                 domain-neutral components and primitives
```

- Route pages are composition roots, not God ViewModels.
- Keep route entry files thin. Put `XView`, `createXViewModel`, panels, forms,
  collection state, and presentation helpers in a sibling `x/` feature folder.
- One feature ViewModel owns its query, mutation, loading, error, and actions.
- Context values are typed and namespaced: `state.sessions.items()`, not a flat
  object assembled with spreads.
- Share state only at the lowest common owner. A shared scope may serve sibling
  features when it prevents duplicate remote state.
- Panels render and invoke feature actions; they never call services directly.
- Name owner primitives `createXViewModel`; reserve `useX` for context/framework
  consumers. Return accessors, not reactive snapshots.
- Keep files below 500 lines; prefer pages below 250 and components below 200.
- Do not import another feature's internal page/state/component.

## Solid Reactivity

- Components run once. Read reactive props as `props.value`; never destructure
  them. Use `splitProps` when forwarding.
- Any value derived from reactive props must remain an accessor/memo, including
  class names: `const className = () => classes(props.color)`.
- Use `createSignal` for independent scalar state, `createStore` for cohesive
  state/forms, and `createMemo` for reused derived values.
- Use `createResource`, router queries, or `createAsync` for remote reads.
- Use `createEffect` only for external synchronization, not derived state or
  ordinary data fetching. Clean up listeners/timers with `onCleanup`.
- Use `For`, `Show`, and `Switch/Match`; avoid nested state ternaries.

## Async And Mutations

- Prefer route preload for primary route data and resources for dependency-driven
  reads.
- Keep transport mapping in services; do not duplicate remote data unless making
  an edit draft.
- Paginated resources use `.latest` so refetching preserves rows and DOM height.
- Feature mutations expose explicit `saving`, `error`, and success state, then
  refetch/invalidate affected reads.
- Expected transport failures such as permission denial and validation errors
  stay in feature/resource error state. They must not escape to the route error
  boundary or replace the current screen.
- Set submitting/saving before mutation and clear it in `finally`.
- Prevent stale responses where the transport/query primitive does not.
- Authorization is enforced by the backend; UI only renders backend results.
- FE must never call IAM permission-check endpoints to decide access. Each
  business service must authorize through its IAM gRPC guard, while IAM-owned
  endpoints enforce their own guards. UI capability state may only be derived
  from protected business responses and is never a security boundary.

## Forms

- One cohesive typed form store per form; no signal per field.
- Keep form types/defaults/validators near the feature.
- Validate before mutation; separate field errors from server errors.
- Reset only after success. Use native labels, controls, and submit semantics.

## Collections

- Operational collections require search, filters, sort, loading, error, empty,
  table/list, and pagination states.
- Unbounded collections use server pagination; client pagination is only for
  known bounded data.
- Page size/search/filter/sort changes reset page to `1`; page changes preserve
  the selected page.
- Backend `pageInfo` is authoritative. Keep previous items visible while loading.
- Pagination buttons must be `type="button"`, must not submit/navigate, and must
  expose reactive `aria-current`.
- Pagination must not scroll the document. Scroll to top only when pathname
  actually changes.
- Put shareable page/filter/sort/selection state in typed URL search parameters.
- Use tables for operational rows, horizontal overflow on small screens, one
  detail editor at a time, and contextual bulk actions.

## Routing, UI, And Accessibility

- Lazy-load routes. Use router links/navigation for internal paths, not raw
  anchors or `window.location`.
- Use shared fields, feedback, tables, pagination, dialogs, and buttons; avoid
  wrapper-only components and nested decorative cards.
- Prefer explicit typed props. Do not copy React patterns such as `useCallback`,
  blanket memoization, or immutable spreading for every update.
- Use semantic elements, visible focus, labels/errors, accessible icon names,
  `aria-current`/`aria-expanded`, and status text beyond color.
- Access storage only through typed adapters; never treat it as authorization
  truth or log credentials.

## TypeScript And Services

- Preserve strict typing: no `any`, broad suppression, or duplicated API DTOs.
- Keep DTOs, form values, and presentation models separate.
- Split large service adapters by feature and query/command responsibility.
- Use one consistent typed result/error convention per feature.
- Do not add a shared abstraction until at least two real consumers need it.

## Reject

- Flat global/God ViewModels or feature ViewModels spread into page context.
- Reactive prop destructuring or one-time derived classes.
- Service calls from panels; unguarded effect fetching; fire-and-forget mutation.
- One signal per form field; unbounded lists without server pagination.
- Raw internal anchors; UI-only authorization; full editors mounted per row.
- Missing loading/error/empty/success states or skipped format/lint/build checks.
