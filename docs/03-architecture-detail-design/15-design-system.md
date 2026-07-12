# Podzone Design System

Stack: SolidJS + Tailwind v4 + Flowbite plugin.
Source root: `frontend/packages/shared/ui/` (moved here from
`internal/ui-podzone/src/solid/` during the MFE split).

**Staleness note (2026-07-11):** the component inventory, gap list, and file
names below were written against the pre-MFE-split tree and have not been
re-verified against the current `frontend/packages/shared/ui/` layout. Some
files may have moved, merged, or been renamed. Confirm a component still
exists at the stated path before treating a gap as still open.

Read `agent/SOLID_STYLE_GUIDE.md` for architecture rules.
This document covers **what** exists in `frontend/packages/shared/ui/`, its
API conventions, current gaps, and how to extend it correctly.

---

## Component Families

| Family | Files | Status |
|--------|-------|--------|
| Primitives / Form fields | `Primitives.tsx` | ⚠️ Needs token refactor, label fix |
| Navigation shells | `Navigation.tsx`, `layout/PodzoneNavbar.tsx` | ⚠️ Raw anchors; workspace switcher is one compact row (see Patterns → Workspace Switcher) |
| Overlays | `Overlay.tsx` | ❌ No focus trap, no ARIA dialog |
| Collection controls | `CollectionControls`, `CollectionFilters`, `CollectionToolbar` | ✅ |
| Table | `DataTable.tsx` | ⚠️ No sort headers, no built-in empty/skeleton rows |
| Pagination | `Pagination.tsx`, `pagination/createPaginatedResource`, `createClientPagination` | ✅ |
| Feedback | `Feedback.tsx` | ✅ |
| Typography | `Typography.tsx` | ✅ |
| Display | `display/` (Alert, Avatar, Skeleton, Toast, Stepper, …) | ⚠️ Toast has no manager |
| Tabs | `Tabs.tsx` | ✅ |
| Search-select | `SearchSelectField.tsx` | ✅ |
| Code / rich text | `CodeEditor.tsx`, `RichTextEditor.tsx`, `MarkdownPreview.tsx` | ✅ (third-party wrappers) |
| Datepicker | `Datepicker.tsx` | ✅ |
| Media | `Media.tsx` | ✅ |
| Layout shells | `AppShell.tsx`, `PageShell.tsx`, `PanelHeader.tsx`, `SectionTitle.tsx` | ✅ |
| Forms (logic) | `forms/createFormStore`, `forms/fields`, `forms/validators` | ✅ |
| Workspace context | `workspace/context.tsx` | ✅ |
| **Missing** | Router-link wrapper, focus-trapping dialog, toast manager, confirm dialog | ❌ |

> **IAM-specific builders** (`IamKeyValueBuilder.tsx`, `IamTrustPolicyBuilder.tsx`) currently live here
> but are domain-specific. Move them to `modules/iam/components/` — see Gap G4 below.

---

## Design Tokens

### Current state

`global.css` defines exactly one custom property:

```css
@theme {
  --color-bgContrast: #ffffff;  /* only token */
}
```

Everything else — colors, radius, shadow, typography — is hardcoded Tailwind
utility classes inside each component. Dark mode is declared via
`@custom-variant dark` but almost no component uses `dark:` classes.

### Required tokens

Add these to `global.css` `@theme` block. Use them in component class strings
instead of hardcoded Tailwind utilities.

```css
@theme {
  /* ── surface ── */
  --color-surface:         #ffffff;
  --color-surface-page:    #f9fafb;    /* body background */
  --color-surface-overlay: rgba(0,0,0,0.5);

  /* ── primary action ── */
  --color-primary:         #030712;    /* gray-950 — current "blue" button */
  --color-primary-hover:   #1f2937;    /* gray-800 */
  --color-primary-ring:    #d1d5db;    /* gray-300 focus ring */

  /* ── semantic ── */
  --color-danger:          #b91c1c;    /* red-700 */
  --color-danger-hover:    #991b1b;
  --color-danger-ring:     #fca5a5;
  --color-success:         #15803d;    /* green-700 */
  --color-success-hover:   #166534;
  --color-success-ring:    #86efac;
  --color-warning:         #92400e;
  --color-warning-surface: #fffbeb;

  /* ── text ── */
  --color-text:            #111827;    /* gray-900 */
  --color-text-muted:      #6b7280;    /* gray-500 */
  --color-text-subtle:     #9ca3af;    /* gray-400 */
  --color-text-inverse:    #ffffff;

  /* ── border ── */
  --color-border:          #e5e7eb;    /* gray-200 */
  --color-border-strong:   #030712;    /* focus border */
  --color-focus-ring:      #f3f4f6;    /* gray-100 inner ring */

  /* ── radius ── */
  --radius-sm:   0.375rem;   /* buttons, inputs — rounded-md */
  --radius-md:   0.5rem;     /* cards, panels — rounded-lg */
  --radius-full: 9999px;     /* pills */

  /* ── shadow ── */
  --shadow-card: 0 1px 3px 0 rgba(0,0,0,.10), 0 1px 2px -1px rgba(0,0,0,.10);
  --shadow-overlay: 0 10px 15px -3px rgba(0,0,0,.10), 0 4px 6px -4px rgba(0,0,0,.10);
}

@layer theme {
  .dark {
    --color-surface:        #111827;
    --color-surface-page:   #030712;
    --color-text:           #f9fafb;
    --color-text-muted:     #9ca3af;
    --color-border:         #374151;
    --color-focus-ring:     #1f2937;
    --color-bgContrast:     #111827;
  }
}
```

Map tokens to Tailwind utilities in `@theme` using `--color-*` naming so
`text-text`, `bg-surface`, `border-border` work in class strings.

---

## Component API Conventions

### Props

```tsx
// ✅ Export the props type so callers can extend
export type ButtonProps = ParentProps<{
  color?: ButtonVariant   // 'primary' | 'secondary' | 'danger' | 'ghost'
  size?: ButtonSize       // 'xs' | 'sm' | 'md'
  href?: string
  type?: 'button' | 'submit' | 'reset'
  loading?: boolean
  disabled?: boolean
  class?: string
  onClick?: JSX.EventHandlerUnion<...>
}>

// ✅ splitProps before using props.x
export function Button(props: ButtonProps) {
  const [local, rest] = splitProps(props, ['color', 'size', ...])
  ...
  return <button {...rest}>...</button>
}

// ✅ Derive class names reactively
const className = () => classes('base', variantClass(), local.class)

// ❌ Do not do this — runs once, loses reactivity
const className = classes('base', local.class)
```

### Form fields

Every field primitive must:
1. Accept and forward an `id` prop (or generate one with `createUniqueId`).
2. Set `for={id}` on the `<label>` / `id={id}` on the native control.
3. Expose `aria-invalid` and `aria-describedby` when `error` is truthy.
4. For radio/checkbox groups: use `<fieldset>` + `<legend>` instead of
   wrapping `<label>`.

```tsx
// ✅ Correct field pattern
import { createUniqueId } from 'solid-js'

export function InputField(props: InputFieldProps) {
  const id = props.id ?? createUniqueId()
  const errorId = () => (props.errorText ? `${id}-error` : undefined)
  return (
    <div class="space-y-1.5">
      <label for={id} class="block text-xs font-semibold uppercase text-text-muted">
        {props.label}
      </label>
      <input
        id={id}
        aria-invalid={props.error || undefined}
        aria-describedby={errorId()}
        ...
      />
      <Show when={props.errorText}>
        <span id={errorId()} class="text-xs text-danger">{props.errorText}</span>
      </Show>
    </div>
  )
}
```

### Event handler conventions

Use `onInput` for text/number/textarea value changes.
Use `onChange` for select, checkbox, and radio changes.
Do not mix the two for the same semantic.

### Internal links

**Never** use a raw `<a href>` for internal SPA routes. Use the router-aware
link wrapper (see Gap G1 below). External URLs may use `<a target="_blank">`.

---

## Patterns

### Form

```
createFormStore<T>(options)    ← one store per form; validators; touched; reset
  └─ FormStore<T>.values       ← reactive store (read as store.values.field)
  └─ FormStore<T>.errors       ← per-field error strings
  └─ FormStore<T>.touched      ← per-field touched flags
  └─ FormStore<T>.isSubmitting ← signal
  └─ FormStore<T>.setSubmitting← setter (always wrap in try/finally)
  └─ FormStore<T>.validate()   ← touches all fields, returns boolean
  └─ FormStore<T>.reset()      ← clears values, errors, touched, submitting
```

**Required mutation lifecycle:**

```ts
form.setSubmitting(true)
try {
  const result = await mutate(...)
  if (!result.success) { setError(result.message); return }
  form.reset()
  setMessage('Saved.')
  await resource.reload()
} finally {
  form.setSubmitting(false)   // ← always runs, even on network throw
}
```

### Collection

```
createPaginatedResource(initialQuery, fetcher, options?)
  └─ .items()       ← current page items (stable via .latest)
  └─ .pageInfo()    ← { page, pageSize, total, totalPages }
  └─ .query         ← reactive store: { page, pageSize, search, sort, filters }
  └─ .updateQuery() ← patches query; resets page to 1 unless patch includes page
  └─ .loading()     ← boolean
  └─ .error()       ← string
  └─ .reload()      ← refetch current query
```

Page/search/filter/sort state must be in URL search parameters for all
operational tables. Bind `updateQuery` to `useSearchParams` setter.

### Overlay

Wrap all modal and drawer open/close with:
- `role="dialog"` / `aria-modal="true"` / `aria-labelledby`
- Focus moved to first interactive element on open
- Focus trapped inside while open (Tab / Shift-Tab cycle)
- Focus restored to trigger on close
- `Escape` key closes

Until the shared `Modal`/`Drawer` components are fixed (Gap G2), feature teams
must add these attributes and focus management at the call site.

### Workspace Switcher

The sidebar's workspace control (`PodzoneNavbar.tsx`) is **one row, fixed
height (`h-12`), regardless of state** — an icon/initials chip, store name +
tenant id (or "Choose a store" when empty), and a chevron, wrapped in one
clickable `Link` to `/admin`. There is no separate uppercase section label
above it ("WORKSPACE") and no dedicated "Change store" button below it —
those doubled the block's height across two states (empty: label + button;
selected: label + info box + subtext + button) for a control that is just a
link to the store chooser.

```tsx
<Link to="/admin" class="flex h-12 items-center gap-2.5 rounded-md border border-gray-200 px-3 hover:bg-gray-50">
  <span class="flex size-7 shrink-0 items-center justify-center rounded-md bg-gray-100 text-[10px] font-bold text-gray-500">
    {hasTenant() ? initials(storeName()) : 'PZ'}
  </span>
  <span class="min-w-0 flex-1 text-left">
    <span class="block truncate text-sm font-medium text-gray-950">{storeName()}</span>
    <Show when={hasTenant()}>
      <span class="block truncate text-[11px] text-gray-500">{activeTenantId().trim()}</span>
    </Show>
  </span>
  {/* chevron svg */}
</Link>
```

Reuse this one-row pattern for any other "current X, click to switch" control
(e.g. a future organization switcher) instead of a label+card+button stack.

---

## Known Gaps

### G1 — No router-link wrapper (P1)

All `href` navigation in `Button`, `NavAction`, `SpeedDial`, and `ListGroup`
uses raw `<a>`. Every internal SPA link triggers a full page reload.

**Fix:** Add `src/solid/components/common/Link.tsx`:

```tsx
import { A, type AnchorProps } from '@solidjs/router'
import { splitProps } from 'solid-js'

export function Link(props: AnchorProps) {
  return <A {...props} />
}
```

In `Button`, replace:
```tsx
// Before
return local.href ? <a href={local.href} ...> : <button ...>

// After
import { isExternalUrl } from '../../shared/utils'
return local.href
  ? isExternalUrl(local.href)
    ? <a href={local.href} target="_blank" rel="noopener" ...>
    : <A href={local.href} ...>
  : <button ...>
```

Add `isExternalUrl(url: string) = /^https?:\/\//i.test(url)` to `shared/utils`.

### G2 — Modal and Drawer have no focus trap or ARIA dialog semantics (P1)

`Overlay.tsx` `Modal` and `Drawer` components are missing:
- `role="dialog"` on the panel element
- `aria-modal="true"`
- `aria-labelledby` pointing to the heading
- Focus trap (Tab / Shift-Tab stay inside while open)
- Auto-focus on open, focus restore on close

**Fix:** Add a `useFocusTrap` primitive to `solid/shared/` and apply it in
`Modal` and `Drawer`. Minimal implementation:

```ts
export function useFocusTrap(container: Accessor<HTMLElement | undefined>) {
  createEffect(() => {
    const el = container()
    if (!el) return
    const focusable = el.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    first?.focus()
    const trap = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return
      if (e.shiftKey ? document.activeElement === first : document.activeElement === last) {
        e.preventDefault()
        ;(e.shiftKey ? last : first)?.focus()
      }
    }
    el.addEventListener('keydown', trap)
    onCleanup(() => el.removeEventListener('keydown', trap))
  })
}
```

### G3 — No toast notification manager (P2)

`display/Toast.tsx` is a static positioned element. Every feature manages its
own `message` / `error` signal visible only while the panel is mounted.

**Fix:** Add `solid/toaster/createToaster.ts` providing a singleton queue:

```ts
// Global singleton
const [toasts, setToasts] = createSignal<Toast[]>([])
export const toaster = {
  success: (msg: string) => push({ type: 'success', msg }),
  error: (msg: string) => push({ type: 'error', msg }),
  info: (msg: string) => push({ type: 'info', msg }),
}
export function ToasterProvider() { /* renders fixed toast stack */ }
```

Mount `<ToasterProvider />` once in `AppShell`. Features call
`toaster.success(...)` instead of managing local message state.

### G4 — Domain-specific IAM components in shared library (P2)

`IamKeyValueBuilder.tsx` and `IamTrustPolicyBuilder.tsx` implement IAM-specific
data shapes and live in `solid/components/common/`. Shared code must be
domain-neutral.

**Fix:** Move to `modules/iam/components/` (create directory).
Update the two import sites in `modules/iam/pages/admin-iam/`.

### G5 — `FieldLabel` breaks label-input association for all field components (P1)

`FieldLabel` wraps children in `<label>` without a `for` attribute, and no
field passes an `id` to its control. Implicit label wrapping works for single
controls but fails for `RadioGroupField` (multiple controls under one label).
Screen readers, click-to-focus, and form autofill may not work correctly.

See Component API Conventions → Form fields above for the fix pattern.
Apply across all 9 field components in `Primitives.tsx`.

### G6 — `ButtonColor` vocabulary is misleading (P2)

`'blue'` and `'dark'` both map to `bg-gray-950 text-white` (identical classes).
The default button color is named "blue" but renders black.

**Fix:** Rename variant map to `ButtonVariant` with values:
`'primary' | 'secondary' | 'danger' | 'success' | 'ghost'`. Derive colors from
CSS custom properties so a brand change touches only `global.css`, not
every call site. Run a codemod to update existing `color="blue"` → `color="primary"`.

### G7 — `PageShell` has no max-width (P3, not yet prioritized)

`PageShell.tsx` is `<div class="space-y-4 lg:space-y-5">{children}</div>` —
no width cap at all. On a wide viewport, a low-content page (e.g. a metrics
row of 2-3 cards) stretches full width and reads as sparse/empty — observed
while mocking up the SRS-IAM-003 platform-overview placeholder screen. Not
fixed yet: current low-content pages are placeholders (SRS-IAM-003 planned,
not built), so it's unclear whether this is really a gap or will resolve
itself once those pages have real content. Revisit once a real page hits
this — don't add a max-width speculatively, since some pages (wide tables,
`DataTable`) may rely on full width.

---

## Control Rules

### File sizes (enforced via ESLint `max-lines`)

| File pattern | Limit |
|---|---|
| `solid/components/**/*.tsx` | 200 |
| `modules/**/pages/**/*Page.tsx` | 250 |
| `modules/**/pages/**/*View.tsx` | 300 |
| All other `.tsx` | 500 |

Add to `.eslintrc` or `eslint.config.js`:

```js
{ files: ['src/solid/components/**/*.tsx'],
  rules: { 'max-lines': ['warn', { max: 200 }] } },
{ files: ['src/modules/**/pages/**/*Page.tsx'],
  rules: { 'max-lines': ['warn', { max: 250 }] } },
```

### Reactivity (eslint-plugin-solid)

The project already has `eslint-plugin-solid`. Enable:

```js
'solid/no-destructure': 'error',
'solid/reactivity': 'warn',
'solid/prefer-for': 'warn',      // <For> instead of .map()
'solid/no-react-deps': 'error',  // no useEffect/useCallback patterns
```

### No raw internal anchors

Add a custom rule or use `jsx-a11y/anchor-is-valid` with config to require the
`Link` component for paths starting with `/`. Alternatively, lint for
`<a href="/"` and `<a href="/` patterns in non-`Link.tsx` files.

### Module boundary

No file in `src/solid/` may import from `src/modules/`.
No file in `src/modules/<A>/` may import from `src/modules/<B>/` internal paths.

Enforce with `eslint-plugin-import` boundaries:

```js
'import/no-restricted-paths': ['error', {
  zones: [
    { target: 'src/solid', from: 'src/modules' },
    { target: 'src/modules/iam', from: 'src/modules/backoffice' },
    { target: 'src/modules/backoffice', from: 'src/modules/iam' },
  ]
}]
```

### Pre-commit gate (already exists)

```bash
npm run format && npm run lint && npm run build && npm run format:check
```

All four must pass before merge. Build failure = hard block.

---

## MVVM Compliance by Module

| Module | Route pages thin? | ViewModels exist? | Panels clean? |
|--------|-------------------|-------------------|---------------|
| **shell** | ❌ auth pages own all logic | ❌ none | — |
| **onboarding** | ⚠️ `AdminHomePage` (379 ln) — ViewModel inline | ✅ well-named per feature | ⚠️ `InvitesPanel` reads storage directly |
| **iam** | ✅ | ✅ feature-split | ✅ |
| **backoffice** | ❌ `TenantOrdersPageView` (349 ln) IS the ViewModel | ⚠️ missing for `orders` | ✅ |

### Priority fixes

1. **Backoffice orders** — extract `createOrdersViewModel.ts`; reduce `TenantOrdersPageView` to composition root.
2. **Shell auth** — extract `createLoginViewModel` and `createRegisterViewModel`; replace per-field signals with `createFormStore`.
3. **Onboarding home** — move `createAdminHomeViewModel` to `admin-home/createAdminHomeViewModel.ts`; page file becomes ≤20 lines.
4. **Onboarding invites panel** — expose `currentUserEmail()` from `createInvitesViewModel`; remove `tokenStorage` import from the panel.
5. **Backoffice home** — replace 16 individual order-metric signals with `createStore<OrderMetrics>`.

---

## Recommended Migration Order

1. Add design tokens to `global.css` (no breaking changes to existing code).
2. Fix `FieldLabel` / add `id` threading to all field components.
3. Add `Link` wrapper; update `Button`, `NavAction`, `ListGroup` to use it.
4. Fix `Modal` and `Drawer` focus trap and ARIA attributes.
5. Extract `createOrdersViewModel` (backoffice orders).
6. Extract shell auth ViewModels.
7. Move `AdminHomePage` ViewModel; shrink page file.
8. Move `IamKeyValueBuilder` / `IamTrustPolicyBuilder` to IAM module.
9. Add toast manager and confirm dialog.
10. Refactor `ButtonColor` → `ButtonVariant` with CSS token backing.
