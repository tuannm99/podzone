# Podzone Solid Rules

Mandatory for `internal/ui-podzone`. Read in full before touching any `.tsx`/`.ts`
file under `src/`. The concrete examples below are derived from real bugs found
in this codebase — they are not hypothetical.

## Workflow

1. Read the route, feature state, service adapter, and adjacent UI first.
2. Identify remote, form, local, and derived state and their owners.
3. Keep changes inside one feature unless extracting a stable shared primitive.
4. Use Docker hot reload; do not start another dev server.
5. Run all four before finishing:

```bash
cd internal/ui-podzone
npm run format
npm run lint
npm run build
npm run format:check
```

Smoke-test affected desktop/mobile loading, error, empty, success, validation,
permission, and pagination states. Frontend unit tests are optional.

---

## Boundaries

```text
src/modules/<feature>/    route pages, panels, feature state/ViewModels
src/services/<feature>/   HTTP/GraphQL DTOs, queries, commands, mapping
src/solid/                domain-neutral components and primitives
```

- Route pages are **thin composition roots**, not God ViewModels. Put all logic
  in a sibling `x/` feature folder.
- One feature ViewModel (`createXViewModel`) owns its query, mutation, loading,
  error, and actions. Reserve `useX` for context/framework consumers.
- Panels render and invoke feature actions; they **never** import from
  `@/services/` or read storage adapters directly.
- Do not import another feature's internal page/state/component.
- `src/solid/` files must never import from `src/modules/` or `src/services/`.
  If a component needs domain types, it belongs in the module, not `solid/`.
- Share state only at the lowest common owner.

---

## Solid Reactivity

### Components run once

```ts
// ✅ read props as props.value — always reactive
function MyPanel(props: { label: string }) {
  return <span>{props.label}</span>
}

// ❌ destructuring breaks reactivity — value is captured at call time
function MyPanel({ label }: { label: string }) { ... }
```

Use `splitProps` when forwarding props to native elements:

```ts
const [local, rest] = splitProps(props, ['color', 'size', 'class'])
return <button class={className()} {...rest}>{local.children}</button>
```

### Derived values must stay reactive

```ts
// ✅ accessor — re-evaluated when props.color changes
const className = () => classes('base', colorMap[props.color], props.class)

// ❌ string — computed once at mount, never updates
const className = classes('base', colorMap[props.color], props.class)
```

### `createEffect` is NOT a data fetcher

`createEffect` is for **external synchronization** only (DOM focus, third-party
libraries, localStorage write-back). It has no built-in cancellation, no
loading state, and no stale-response protection.

```ts
// ❌ anti-pattern — found in TenantOrdersPageView, createAdminIamViewModel
createEffect(() => {
  if (!workspaceReady()) return
  void loadOrders()          // no stale guard, no cancellation
})

// ✅ use createResource with a reactive source signal
const [orders] = createResource(
  () => workspaceReady()
    ? { tenantId: params().tenantId, storeId: currentStoreId() }
    : null,
  ({ tenantId, storeId }) => fetchOrders(tenantId, storeId)
)
// Solid automatically discards the previous request when the source changes.
```

For paginated collections use `createPaginatedResource` from `solid/pagination/`.
For secondary reads triggered by selection, pass the selection as the resource
source. **Never** use `void asyncFn()` inside `createEffect` for remote data.

### List rendering

```ts
// ✅ <For> is reactive — only re-renders changed items
<For each={props.items}>{(item) => <Row item={item} />}</For>

// ❌ .map() runs once — does not track reactive changes to the array
{props.items.map(item => <Row item={item} />)}
```

Use `<For>`, `<Show>`, and `<Switch/Match>` for all conditional and list
rendering. Avoid nested ternaries in JSX.

---

## Async And Mutations

### Submitting state must use try/finally

If `setSubmitting(true)` is called before an `await`, the matching
`setSubmitting(false)` **must** be in a `finally` block. A thrown transport
error otherwise leaves the button permanently disabled.

```ts
// ❌ current bug in useOrderActions, createBootstrapLoader, and others
form.setSubmitting(true)
const result = await mutate(...)    // network throw → setSubmitting never clears
form.setSubmitting(false)

// ✅ required pattern
form.setSubmitting(true)
try {
  const result = await mutate(...)
  if (!result.success) {
    setError(result.message)
    return
  }
  form.reset()
  setMessage('Saved.')
  await resource.reload()
} finally {
  form.setSubmitting(false)
}
```

Same rule applies to any `setSaving`, `setLoading`, `setDeleting`, or similar
in-progress state.

### Stale response prevention

When a `createEffect` is the only option (e.g. selection-triggered secondary
load that cannot be modelled as a resource source), use a version counter:

```ts
let requestVersion = 0

async function loadSelectedPolicy(id: string) {
  const thisVersion = ++requestVersion
  const result = await fetchPolicy(id)
  if (thisVersion !== requestVersion) return   // superseded — discard
  setPolicy(result.data)
}
```

Prefer `createResource` over this pattern — it handles cancellation automatically.

### Fire-and-forget mutation is a bug

```ts
// ❌ double-click sends two concurrent requests; errors are silently lost
<Button onClick={() => void remove(id)}>Remove</Button>

// ✅ guard with saving state; expose it to disable the button
const remove = async (id: string) => {
  setSaving(id)
  try {
    await doRemove(id)
    await list.reload()
  } finally {
    setSaving('')
  }
}
<Button disabled={saving() === id} onClick={() => void remove(id)}>Remove</Button>
```

### Expected transport failures stay in feature state

Service calls return `{ success: boolean; data: T; message: string }`. A
`success: false` response is a **handled error** — set it in feature error
state, do not rethrow. Only unexpected throws (network down, bug) should
propagate.

---

## Forms

- One cohesive `createFormStore<T>` per form. No per-field signals.
- Keep form types, defaults, and validators next to the feature ViewModel.
- Validate before mutation; separate field errors from server errors.
- Reset only after success; `form.reset()` is already in the `FormStore` API.
- `setSubmitting` must be in `try/finally` — see Async And Mutations above.
- Map server validation errors back to fields via `form.setFieldError(field, msg)`.

### Field accessibility

Every field primitive must thread a stable `id` through:

```ts
import { createUniqueId } from 'solid-js'

export function InputField(props: InputFieldProps) {
  const id = props.id ?? createUniqueId()
  const errorId = `${id}-error`
  return (
    <div class="space-y-1.5">
      <label for={id} class="...">
        {props.label}
      </label>
      <input
        id={id}
        aria-invalid={props.error || undefined}
        aria-describedby={props.errorText ? errorId : undefined}
        ...
      />
      <Show when={props.errorText}>
        <span id={errorId} class="text-xs text-danger">{props.errorText}</span>
      </Show>
    </div>
  )
}
```

For radio/checkbox groups: use `<fieldset>` + `<legend>`, not a wrapping
`<label>` — clicking the group label does nothing useful otherwise.

---

## Collections

- Operational collections require search, filters, sort, loading, error, empty,
  table/list, and pagination states. Use `createPaginatedResource`.
- Unbounded collections use server pagination. Client-side `slice()` or
  `filter()` on a fully-fetched array is a bug for operational data.
- Page/search/filter/sort changes reset page to `1`; page changes preserve the
  selected page. `updateQuery` in `createPaginatedResource` does this automatically.
- `pageInfo` from the backend is authoritative. Keep previous items visible while
  loading (`.latest` on the resource).

### Collection state belongs in URL search params

Page, search, filter, and sort must be in URL search parameters so the state
survives Back navigation and can be bookmarked or shared.

```ts
// ✅ bind createPaginatedResource to useSearchParams
import { useSearchParams } from '@solidjs/router'

function createPartnersViewModel() {
  const [search, setSearch] = useSearchParams()

  const resource = createPaginatedResource(
    {
      page:     Number(search.page ?? 1),
      pageSize: 20,
      search:   search.q ?? '',
    },
    fetchPartners
  )

  const updateSearch = (q: string) => setSearch({ q, page: '1' })
  const goToPage     = (p: number) => setSearch({ page: String(p) })

  return { resource, updateSearch, goToPage }
}
```

Do **not** use `createSignal` for page, search, filter, or sort in collection
ViewModels — these are URL state.

### Pagination buttons

Pagination `<button>` elements must:
- Have `type="button"` (prevent accidental form submit)
- Expose `aria-current="page"` on the active page button
- Not trigger document scroll (scroll only on pathname change)

---

## Routing, UI, And Accessibility

### Internal navigation — use the Link wrapper

**Never** use a raw `<a href>` or `window.location.href =` for internal SPA routes.
Both trigger a full document reload, destroying all reactive state.

```ts
// ❌ full page reload — found in Button, NavAction, AdminHomePage
<a href="/admin/iam">IAM Console</a>
window.location.href = `/t/${tenantId}`

// ✅ use the Link primitive (solid/components/common/Link.tsx)
import { Link } from '@/solid/components/common/Link'
<Link href="/admin/iam">IAM Console</Link>

// Button accepts href — it routes via Link internally
<Button href="/admin/iam">IAM Console</Button>
```

Use `navigate()` from `useNavigate()` for programmatic navigation in
ViewModel actions.

### Tab / section routing

Active tab and section state must be in URL search params — not in
`window.location.hash`, `window.history.replaceState`, or local signals.

```ts
// ❌ hash — found in AdminSettingsPage, AdminIamView, createProvisioningShellViewModel
window.location.hash = 'sessions'
const tab = window.location.hash.slice(1)

// ✅ search param — survives Back, is shareable
const [params, setParams] = useSearchParams()
const activeTab = () => params.tab ?? 'sessions'
const setTab = (tab: string) => setParams({ tab })
```

### Overlays — focus trap and ARIA

Every Modal and Drawer must have:
- `role="dialog"` and `aria-modal="true"` on the panel element
- `aria-labelledby` pointing to the heading id
- Focus moved to the first focusable element on open
- Tab / Shift-Tab trapped inside while open
- Focus restored to the trigger element on close
- `Escape` key closes the overlay

Apply `useFocusTrap` from `solid/shared/` — see
`docs/03-architecture-detail-design/15-design-system.md`.

### Global notifications

Do not manage success/error messages as local signals inside features — they are
lost when the user navigates away. Use the global `ToasterContext` provided by
`AppShell`:

```ts
// ✅ notifications survive navigation
import { useToaster } from '@/solid/toaster'

function createOrdersViewModel() {
  const toast = useToaster()
  const save = async () => {
    try { ... toast.success('Order created.') }
    catch { toast.error('Failed to create order.') }
  }
}
```

### Semantic elements and ARIA

- Use `<table>` for operational row data; `<ul>`/`<li>` for lists.
- Every icon-only button needs `aria-label`.
- Interactive elements need visible focus rings (already provided by Tailwind
  focus utilities; do not override with `outline-none` without a custom ring).
- Status changes beyond color — always pair color with text or icon with label.

---

## TypeScript And Services

- No `any` types at ViewModel or service boundaries.
- No `as unknown as T` casts at runtime boundaries — validate instead.
- Keep DTOs (service layer), form values (form layer), and presentation models
  (ViewModel/panel layer) separate. Do not pass a raw DTO to a form field.
- Split large service adapters by feature and query/command responsibility.
- Services return `{ success: boolean; data: T; message: string }` — handle both
  branches; do not assume success.

---

## Reject

The following patterns are bugs, not style preferences:

| Pattern | Why |
|---------|-----|
| Reactive prop destructuring | Props read once — reactivity lost |
| `.map()` in JSX | Not tracked — use `<For>` |
| `createEffect` + `void asyncFn()` for remote reads | No stale guard, no cancellation |
| `setSubmitting(true)` without `try/finally` | Permanent stuck state on network error |
| `window.location.href =` for internal routes | Full page reload, SPA state lost |
| `window.location.hash` / `window.history` for tab state | Not URL-search-param, not shareable |
| Fire-and-forget mutation (`void remove(id)` with no saving guard) | Double-submit sends duplicate requests |
| Per-field `createSignal` in a form | Use `createFormStore` |
| Client `slice()` / `filter()` on unbounded operational data | Use server pagination |
| Service call or storage read inside a Panel | Panels receive actions from the ViewModel |
| Domain-specific component inside `src/solid/` | Belongs in the owning module |
| `<a href>` for internal SPA routes | Full page reload — use `Link` wrapper |
| Raw `window.confirm()` | Use the shared confirm dialog |
| Local `message`/`error` signals for notifications | Lost on navigate — use `ToasterContext` |
| `solid-js` as a federation singleton in remote vite configs | Cross reactive-system: signals from HOST's solid-js, `createRenderEffect` from remote bundle — subscriptions silently fail (see MFE section below) |

---

## MFE / Vite-Plugin-Federation — SolidJS Reactive System Split

### The Bug

When a remote app (IAM, Onboarding) declares `solid-js` as a federation
singleton, the HOST's `vite dev` server registers its own `solid-js` module
in `globalThis.__federation_shared__`. Remote bundle code that calls
`importShared("solid-js")` receives the **HOST's** solid-js module (loaded
from `localhost:3000/...`).

At the same time, the remote bundle's JSX templates import `createRenderEffect`
and `createComponent` as **static imports** from `solid-wpaq3ZqP.js`
(`localhost:3002/assets/...`). Different URL → different ES module instance →
different global `Listener` variable.

Result: `createSignal` subscribes to HOST's reactive system; `createRenderEffect`
registers tracking in the REMOTE's reactive system. When a signal setter is
called, it notifies HOST's subscribers only. The `createRenderEffect` inside
`Tabs` (or any component) was **never registered** as a HOST subscriber →
UI does not update on click, but renders correctly on initial load (effects
run once synchronously regardless of subscription).

**Symptom:** "clicking tab does not change active visual; reload shows the
correct initial state."

### The Fix

**Do not declare `solid-js` as a singleton in remote vite configs.**

```ts
// apps/iam/vite.config.ts  and  apps/onboarding/vite.config.ts
shared: {
  // solid-js intentionally excluded — see MFE section in SOLID_STYLE_GUIDE.md
  '@tanstack/solid-router': { singleton: true },
  '@podzone/shared': { singleton: true },
},
```

Without `solid-js` in the remote's `shared` map, `importShared("solid-js")`
falls back to the local bundle (`solid-wpaq3ZqP.js`). Both `createSignal` and
`createRenderEffect` resolve to the same module instance → one reactive system
→ subscriptions work.

### Bridging Auth Context

Removing the singleton breaks `useContext(AuthContextToken)` in remotes because
the provider is in the HOST's owner tree and the REMOTE's `useContext` searches
the REMOTE's owner tree (a separate reactive scope).

Fix: store the auth value globally in `AuthContextProvider`, and read it first
in `useAuthContext()`:

```ts
// packages/shared/auth/auth-context.ts
export function useAuthContext(): AuthContext {
  if (window.__pz_auth_value__) return window.__pz_auth_value__
  const ctx = useContext(AuthContextToken)
  if (!ctx) throw new Error('useAuthContext must be used inside AuthContextProvider')
  return ctx
}
```

```tsx
// src/modules/shell/AuthContextProvider.tsx
window.__pz_auth_value__ = ctx  // set before returning the Provider
```

### Why `@podzone/shared` Singleton Does Not Cause This

The `@podzone/shared` singleton config has no effect because Vite's alias
(`@podzone/shared → ../../packages/shared`) is resolved **before** the
federation plugin sees the import. The aliased path no longer matches the
package name key in `shared`, so `@podzone/shared` is always bundled locally
into the remote — `pagination-CH3jtJvn.js` et al. This is why Tabs' compiled
output always uses `solid-wpaq3ZqP.js` directly.

### Rule of Thumb

For MFE remotes built with `vite-plugin-solid` + `@originjs/vite-plugin-federation`:

> A package is safe as a federation singleton **only if** its reactive primitives
> (`createSignal`, `createEffect`, etc.) and its DOM runtime primitives
> (`createRenderEffect`, `createComponent`) are **both** served through
> `importShared` — not split between static import and `importShared`.
> `solid-js` fails this test because JSX compilation emits static imports for
> the DOM runtime primitives regardless of the shared config.
