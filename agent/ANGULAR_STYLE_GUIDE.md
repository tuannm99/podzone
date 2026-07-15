# Podzone Angular Rules

Mandatory for `frontend-v2/`. Read in full before touching any `.ts`/`.html`
file there. `frontend/` (SolidJS) is untouched by this guide — see
`agent/SOLID_STYLE_GUIDE.md` for that tree, and
`docs/09-pzep/PZEP-0004-solidjs-to-angular-frontend-migration.md` for why
both exist and how the cutover is sequenced.

This guide translates the architectural boundaries and failure modes
already learned the hard way in `SOLID_STYLE_GUIDE.md` (and
`docs/03-architecture-detail-design/13-frontend-solid-audit.md`) into
Angular idioms. Angular 22 in this project is **zoneless by default**
(confirmed: no `zone.js` dependency, no `provideZoneChangeDetection` call
in `frontend-v2/src/app/app.config.ts`) and uses Signals
(`signal`/`computed`/`effect`/`resource`) as its reactivity primitive —
this maps closely to Solid's `createSignal`/`createMemo`/`createEffect`/
`createResource`, so most rules below carry the same *reason*, translated
to a different API surface. Where an example is marked "anticipated" it is
a known Angular/zoneless/federation pitfall documented ahead of time, not
(yet) an observed bug in this codebase — `frontend-v2` is new.

## Workflow

1. Read the route, feature state, service, and adjacent UI first.
2. Identify remote, form, local, and derived state and their owners.
3. Keep changes inside one feature unless extracting a stable shared primitive.
4. Run all before finishing:

```bash
cd frontend-v2
npm run format    # add to package.json if not present yet — see below
ng lint            # add @angular-eslint if not present yet — see below
ng build
ng test
```

`frontend-v2`'s scaffold does not yet have `format`/`lint` scripts wired —
add them (Prettier + `@angular-eslint`) before or alongside the first real
feature PZEP, matching `frontend/`'s existing four-command pattern from
`CLAUDE.md`'s Frontend Commands section.

Smoke-test affected desktop/mobile loading, error, empty, success,
validation, permission, and pagination states. Frontend unit tests are
optional, same policy as `frontend/`.

---

## Boundaries

```text
src/app/features/<feature>/   route components, panels, feature state/services
src/app/core/                 app-wide singletons: auth, HTTP interceptors, guards
src/app/shared/               domain-neutral, reusable UI components and primitives
```

- Route components are **thin composition roots**, not God components. Put
  logic in a sibling feature service, not the route component's class body.
- One injectable feature service (`@Injectable()`, `providedIn` scoped to
  the feature's route, not `'root'` unless genuinely app-wide) owns its
  data (as `resource`/`rxResource` or explicit signals), loading, error,
  and actions — the direct Angular equivalent of Solid's `createXViewModel`.
- Presentational components render `input()`s and emit `output()`s; they
  **never** inject `HttpClient` or a storage service directly.
- Do not import another feature's internal component/service/state.
- `src/app/shared/` components must never import from
  `src/app/features/` or `src/app/core/`. If a component needs domain
  types, it belongs in the feature, not `shared/`.
- Share state only at the lowest common owner — inject the feature service
  at the route level (`providers: [FeatureService]` on the route config or
  a parent component), not at `providedIn: 'root'`, unless it is genuinely
  cross-feature/app-wide (auth, toaster).

---

## UI / Design System

`frontend-v2` uses **Angular Material + SCSS** exclusively — no Tailwind CSS,
no Flowbite, no `tailwind-merge` (see
`docs/08-adr/ADR-0006-angular-material-replaces-ported-tailwind-design-system.md`).
This is a permanent, deliberate divergence from `frontend/`'s Tailwind/
Flowbite-derived design system, not a transitional state.

- Each component gets its own `*.component.scss` (`styleUrl`, not inline
  `styles`) rather than utility classes in the template.
- Use Material's CSS custom properties (`var(--mat-sys-primary)`,
  `var(--mat-sys-on-surface-variant)`, `var(--mat-sys-outline-variant)`,
  typography tokens like `var(--mat-sys-body-medium)`, etc.) for color and
  type instead of hardcoded hex values or Tailwind-style utility classes —
  this is what makes components correctly follow the Material 3 theme
  (`src/styles.scss`) automatically, including any future light/dark switch.
- Reach for a real Material component (`MatButtonModule`, `MatCardModule`,
  `MatFormFieldModule`, `MatListModule`, `MatIconModule`, etc.) before
  hand-rolling markup. Only build a custom `shared/ui` primitive when
  Material has no equivalent (e.g. `SectionLead`, `NavLink`).
- Do not add a new Tailwind/Flowbite-style component to `shared/ui` and do
  not reintroduce Tailwind CSS, PostCSS-Tailwind config, or `tailwind-merge`
  as a dependency.

---

## Angular Reactivity

### Signal inputs, not `@Input()` decorators

```ts
// ✅ input() — a Signal, always current, composes with computed()/effect()
export class StatusBadge {
  status = input.required<OrderStatus>()
  label = computed(() => statusLabel(this.status()))
}

// ❌ @Input() decorator — works, but doesn't compose with computed()/
// effect() the way a Signal does, and mixes two reactivity models in one
// codebase for no benefit in a zoneless app
export class StatusBadge {
  @Input() status!: OrderStatus
}
```

Use `input.required<T>()` for a required prop (the Angular analogue of a
required, non-optional Solid prop) and `input<T>(defaultValue)` for one
with a default. Never read `.status` as a bare property expecting Solid-style
"props read once" semantics — a Signal is always current *because* you call
it as a function (`this.status()`), same discipline as Solid's `props.value`.

### Derived values must stay signals

```ts
// ✅ computed() — re-evaluated only when its signal dependencies change
statusColor = computed(() => statusColorMap[this.status()])

// ❌ plain method — recomputed every change-detection pass, and (in a
// zoneless app) may not even be recomputed reliably when nothing else
// triggers CD — don't rely on implicit recomputation for derived state
getStatusColor() {
  return statusColorMap[this.status()]
}
```

### `effect()` is NOT a data fetcher

Same rule as Solid's `createEffect`: `effect()` is for **external
synchronization** only (DOM focus, a third-party non-Angular library,
`localStorage` write-back). It has no built-in cancellation and no loading
state.

```ts
// ❌ anti-pattern (anticipated — mirrors the exact bug class found in
// SOLID_STYLE_GUIDE.md's createEffect section)
constructor() {
  effect(() => {
    if (!this.workspaceReady()) return
    this.orderService.loadOrders(this.tenantId())   // no stale guard, no cancellation
  })
}

// ✅ resource() — params signal in, loader out, cancellation and
// loading/error state included. The option is `params`, not `request` —
// confirmed against the actual installed @angular/core type declarations
// (frontend-v2/node_modules/@angular/core/types/_api-chunk.d.ts); an
// earlier draft of this guide had this wrong and it broke a real build.
orders = resource({
  params: () => (this.workspaceReady() ? { tenantId: this.tenantId(), storeId: this.storeId() } : undefined),
  loader: ({ params }) => this.orderService.fetchOrders(params.tenantId, params.storeId),
})
// orders.value(), orders.isLoading(), orders.error(), orders.reload()
// Angular cancels the in-flight loader via AbortSignal when `params`
// changes or the resource is destroyed — same guarantee Solid's
// createResource gives when its source signal changes.
```

Use `resource()` for reads driven by a `fetch`/promise-based loader, and
`rxResource()` when the underlying call is naturally an `Observable`
(e.g. wrapping an existing RxJS-based HTTP call). **Never** call an async
method with no cancellation/staleness handling from inside `effect()` for
remote data — same failure mode Solid's audit found repeatedly
(`createEffect` + fire-and-forget fetch).

### List rendering

```html
<!-- ✅ @for with a track expression — Angular's control-flow syntax,
     equivalent to Solid's <For>, requires a track key -->
@for (order of orders.value(); track order.id) {
  <app-order-row [order]="order" />
}

<!-- ❌ *ngFor without OnPush/track discipline invites the same
     "re-renders everything" class of bug -->
```

Use `@for`/`@if`/`@switch` (the modern control-flow syntax, not the legacy
`*ngFor`/`*ngIf` structural directives) for all conditional and list
rendering. Always provide `track` — an untracked `@for` degrades to
index-based diffing, the Angular analogue of Solid's `.map()` anti-pattern.

---

## Async And Mutations

### Submitting state must use try/finally

Same rule, same reason as `SOLID_STYLE_GUIDE.md`: if `submitting.set(true)`
runs before an `await`, the matching `submitting.set(false)` **must** be in
a `finally` block.

```ts
// ❌ anticipated anti-pattern
async save() {
  this.submitting.set(true)
  const result = await this.orderService.update(this.orderId(), this.form.value())
  this.submitting.set(false)   // never runs if update() throws
}

// ✅ required pattern
async save() {
  this.submitting.set(true)
  try {
    const result = await this.orderService.update(this.orderId(), this.form.value())
    if (!result.success) {
      this.error.set(result.message)
      return
    }
    this.toaster.success('Saved.')
    this.orders.reload()
  } finally {
    this.submitting.set(false)
  }
}
```

Same rule applies to any `saving`/`loading`/`deleting` signal driving a
mutation's in-progress UI state.

### Fire-and-forget mutation is a bug

```html
<!-- ❌ double-click sends two concurrent requests -->
<button (click)="remove(id)">Remove</button>

<!-- ✅ guard with a saving signal; expose it to disable the button -->
<button [disabled]="saving() === id" (click)="remove(id)">Remove</button>
```

```ts
async remove(id: string) {
  this.saving.set(id)
  try {
    await this.orderService.remove(id)
    this.orders.reload()
  } finally {
    this.saving.set('')
  }
}
```

### Expected transport failures stay in feature state

Same contract as `frontend/`: services return
`{ success: boolean; data: T; message: string }`. A `success: false`
response is a **handled error** — set it in feature error state, do not
throw. Only unexpected throws (network down, bug) should propagate to
Angular's error handling (`provideBrowserGlobalErrorListeners`, already in
the scaffold's `app.config.ts`).

---

## Forms

Angular 22 ships two form APIs. **Prefer Signal Forms** (`form()`, `Field`,
`schema()` from `@angular/forms`, confirmed present and not marked
experimental in this project's installed version) — it is the direct
Angular analogue of `createFormStore<T>`: one reactive tree built from a
model signal, not per-field signals wired by hand.

```ts
import { form, Field, schema, required, submit } from '@angular/forms'

// ✅ one cohesive form, matching createFormStore<T>'s "no per-field signals" rule
storeName = signal({ name: '', region: 'us-east' })
storeForm = form(this.storeName, schema((path) => {
  required(path.name, { message: 'Store name is required.' })
}))

async onSubmit() {
  await submit(this.storeForm, async (value) => {
    const result = await this.storeService.create(value)
    if (!result.success) return { kind: 'server', message: result.message }
    this.toaster.success('Store created.')
    return
  })
}
```

If a form's needs outgrow Signal Forms (verify current maturity before
assuming — check `@angular/forms`'s changelog for the installed version
before depending on an edge-case feature), fall back to typed Reactive
Forms (`FormGroup<T>`/`FormBuilder`) — never per-field `signal()`s wired by
hand for a multi-field form, same reasoning as the Solid guide's "no
per-field `createSignal`" rule.

- Keep form types, defaults, and validators next to the feature service.
- Validate before mutation; separate field errors from server errors.
- Reset only after success.
- `submitting`/`saving` state around the submit call must be in
  `try/finally` — see Async And Mutations above (Signal Forms' `submit()`
  helper handles this internally for the schema-validation path, but a
  service-layer network failure outside `submit()`'s own try/catch still
  needs the same discipline if you're not routing every mutation through
  `submit()`).
- Map server validation errors back to fields via Signal Forms' server-error
  reporting (`submit()`'s callback return value) or, on the Reactive Forms
  fallback, `control.setErrors(...)`.

### Field accessibility

Every field primitive must thread a stable `id` through, same requirement
as `SOLID_STYLE_GUIDE.md`. Prefer `mat-form-field` + `matInput` + `mat-label`
over hand-rolled `<label>`/`<input>` markup — it wires label/id association
and error styling (via `[color]="hasError() ? 'warn' : 'primary'"`, since
`mat-error`'s automatic error-state detection requires a bound
`FormControl`/`NgModel` that a plain `value`/`valueChange` input pair does
not have) for you:

```html
<mat-form-field appearance="outline" [color]="hasError() ? 'warn' : 'primary'">
  <mat-label>{{ label() }}</mat-label>
  <input matInput [id]="id()" [attr.aria-invalid]="hasError() || null" [attr.aria-describedby]="hasError() ? errorId() : null" />
  @if (hasError()) {
    <mat-hint [attr.id]="errorId()">{{ errorText() }}</mat-hint>
  }
</mat-form-field>
```

```ts
id = input<string>(createUniqueId()) // shared/utils.ts — Angular has no public equivalent of Solid's createUniqueId()
```

Generate a stable id once per component instance (e.g. in a field
constant or via Angular's own unique-id utilities), not recomputed on
every change-detection pass. For radio/checkbox groups: use
`<fieldset>` + `<legend>`, not a wrapping `<label>`.

---

## Collections

- Operational collections require search, filters, sort, loading, error,
  empty, table/list, and pagination states — same list as `frontend/`.
- Unbounded collections use server pagination. Client-side `.slice()` or
  `.filter()` on a fully-fetched array is a bug for operational data, same
  as `SOLID_STYLE_GUIDE.md`.
- Page/search/filter/sort changes reset page to `1`; page changes preserve
  the selected page.
- `pageInfo` from the backend is authoritative. Keep previous items visible
  while loading — `resource()`'s previous `.value()` is retained by default
  until the new load resolves (unlike a naive re-fetch that clears state
  first); do not manually clear `.value()` before a reload.

### Collection state belongs in URL query params

Page, search, filter, and sort must be in the URL so state survives Back
navigation and is bookmarkable/shareable — same rule, same reason as
`SOLID_STYLE_GUIDE.md`.

```ts
// ✅ bind resource() to ActivatedRoute's queryParams (as a signal via toSignal)
import { toSignal } from '@angular/core/rxjs-interop'

export class PartnersListComponent {
  private route = inject(ActivatedRoute)
  private router = inject(Router)

  private queryParams = toSignal(this.route.queryParamMap, { requireSync: true })
  page = computed(() => Number(this.queryParams().get('page') ?? '1'))
  search = computed(() => this.queryParams().get('q') ?? '')

  partners = resource({
    params: () => ({ page: this.page(), search: this.search() }),
    loader: ({ params }) => this.partnerService.list(params),
  })

  updateSearch(q: string) {
    this.router.navigate([], { queryParams: { q, page: 1 }, queryParamsHandling: 'merge' })
  }
}
```

Do **not** use a plain `signal()` for page, search, filter, or sort in a
collection's feature service — these are URL state, same rule as Solid's
"not `createSignal`" prohibition.

### Pagination buttons

Pagination `<button>` elements must:
- Have `type="button"` (prevent accidental form submit)
- Expose `[attr.aria-current]="isActive() ? 'page' : null"` on the active
  page button
- Not trigger document scroll (scroll only on route/pathname change)

---

## Routing, UI, And Accessibility

### Internal navigation — use `routerLink`, never a raw anchor

**Never** use a raw `<a href>` with a literal string or set
`window.location.href` for internal SPA routes — both trigger a full
document reload, destroying all component/signal state. Same rule as
`SOLID_STYLE_GUIDE.md`.

```html
<!-- ❌ full page reload -->
<a href="/admin/iam">IAM Console</a>

<!-- ✅ routerLink — Angular's Link equivalent -->
<a routerLink="/admin/iam">IAM Console</a>
```

Use `Router.navigate()`/`Router.navigateByUrl()` (injected via `inject(Router)`)
for programmatic navigation in feature service actions.

### Route guards are functions, not classes

```ts
// ✅ functional guard (CanActivateFn), current idiom
export const authGuard: CanActivateFn = (route, state) => {
  const auth = inject(AuthService)
  return auth.isAuthenticated() || inject(Router).parseUrl('/login')
}
```

Avoid the older class-based `CanActivate` interface for new guards — no
functional benefit lost, and it matches the DI-via-`inject()` idiom used
everywhere else in this guide.

### Tab / section routing

Active tab/section state must be a route query param, not a signal or
`window.location.hash` — same rule as `SOLID_STYLE_GUIDE.md`'s hash
prohibition, translated to Angular's `ActivatedRoute`/`Router` APIs shown
above under Collections.

### Overlays — focus trap and ARIA

Every modal/drawer-equivalent component must have the same guarantees as
`SOLID_STYLE_GUIDE.md`'s Overlays section: `role="dialog"`,
`aria-modal="true"`, `aria-labelledby`, focus moved in on open and restored
on close, Tab/Shift-Tab trapped, `Escape` closes. Do not hand-roll this.
`@angular/cdk` is already a dependency (pulled in by `@angular/material`,
see ADR-0006) — prefer `MatDialog` (built on CDK's overlay + focus trap)
for a real modal, or `@angular/cdk/a11y`'s `FocusTrap`/`FocusMonitor`
directly for a custom overlay that isn't a good fit for `MatDialog`. Never
reimplement focus management from scratch.

### Global notifications

Do not manage success/error messages as local component signals — they're
lost on navigation. Use a single app-wide, `providedIn: 'root'` toaster
service (the direct analogue of Solid's `ToasterContext`/`useToaster`):

```ts
@Injectable({ providedIn: 'root' })
export class ToasterService {
  private readonly _toasts = signal<Toast[]>([])
  readonly toasts = this._toasts.asReadonly()

  success(message: string) { this.push('success', message) }
  error(message: string) { this.push('error', message) }
  private push(type: ToastType, message: string) {
    const id = crypto.randomUUID()
    this._toasts.update((list) => [...list, { id, type, message }])
    setTimeout(() => this.dismiss(id), 4000)
  }
  dismiss(id: string) { this._toasts.update((list) => list.filter((t) => t.id !== id)) }
}
```

### Semantic elements and ARIA

Same requirements as `SOLID_STYLE_GUIDE.md`: `<table>` for operational row
data, `<ul>`/`<li>` for lists, `aria-label` on every icon-only button,
visible focus rings preserved (Material components ship a visible focus
indicator by default — do not override it away in component SCSS), status
changes always paired with text/icon, not color alone.

---

## TypeScript And Services

- No `any` types at a feature service or HTTP boundary.
- No `as unknown as T` casts at runtime boundaries — validate instead.
- Keep DTOs (HTTP-layer), form models (`Field`/`FormGroup` layer), and
  presentation models (component-input layer) separate. Do not pass a raw
  DTO to a form.
- Inject `HttpClient` (from `@angular/common/http`, `provideHttpClient()`
  registered in `app.config.ts`) only inside feature services — never
  directly inside a component, same boundary as Solid's "Panels never
  import from `@/services/`" rule.
- Services return `{ success: boolean; data: T; message: string }` — same
  contract as `frontend/`'s services, handle both branches; do not assume
  success.
- Use `inject()` at field-initializer position for DI, not constructor
  parameter injection — matches Angular's current idiom and composes
  better with standalone components (no constructor boilerplate to keep in
  sync with a base class).

---

## Reject

The following patterns are bugs, not style preferences:

| Pattern | Why |
|---------|-----|
| `@Input()` decorator on new components | Use `input()`/`input.required()` — composes with `computed()`/`effect()` |
| `*ngFor`/`*ngIf` in new templates | Use `@for`/`@if` — modern control flow, requires explicit `track` |
| `effect()` + fire-and-forget async call for remote reads | No stale guard, no cancellation — use `resource()`/`rxResource()` |
| `submitting.set(true)` without `try/finally` | Permanent stuck state on network error |
| `<a href="...">` or `window.location.href =` for internal routes | Full page reload, all component/signal state lost — use `routerLink` |
| `window.location.hash` for tab/section state | Not a route query param, not shareable |
| Fire-and-forget mutation (`(click)="remove(id)"` with no saving guard) | Double-submit sends duplicate requests |
| Per-field `signal()` wiring for a multi-field form | Use Signal Forms (`form()`/`schema()`) or typed `FormGroup` |
| Client `.slice()`/`.filter()` on unbounded operational data | Use server pagination |
| `HttpClient` injected directly into a component | Components receive data/actions from a feature service |
| Domain-specific component inside `src/app/shared/` | Belongs in the owning feature |
| Class-based `CanActivate` for new guards | Use `CanActivateFn` |
| Local component signal for success/error toast messages | Lost on navigate — use the app-wide `ToasterService` |
| Hand-rolled focus-trap/ARIA for overlays | Use `@angular/cdk/a11y` |
| A shared library declared a federation singleton without confirming BOTH its reactive primitives and its compiled output resolve through the shared runtime | See MFE section below — same class of bug as Solid's `solid-js` singleton footgun, different mechanism |

---

## MFE / Native-Federation

### Confirmed today

`frontend-v2` is configured as a native-federation **remote** on port
`3004` (`federation.config.mjs`), verified live this session: real dev
server, real `remoteEntry.json` (manifest format `$version: "v4"`),
`@angular/core` shared as a strict-version singleton (the schematic's
default `shareAll(...)` plus an explicit override forcing
`includeSecondaries: { keepAll: true }` on `@angular/core` specifically —
do not remove that override without understanding why it's there; secondary
entry points of `@angular/core` must stay bundled with the primary to avoid
a split-module class of bug analogous to Solid's federation reactive-split).

### Not yet true — do not assume it

The host (`frontend/vite.config.ts`) migrated to `@module-federation/vite`
(MF2) in PZEP-0005, bridged to native-federation remotes via
`native-to-mf-bridge` — the bridge mechanism was verified live (real
import map fetched from this app's dev server). **`frontend-v2` is still
not mounted as an actual route in the host's route tree** — the bridge was
only proven with a throwaway one-off import, not a real page. Do not write
code in either `frontend-v2` or `frontend/` that assumes a user can reach
`frontend-v2` through the host today — that mount is a separate,
not-yet-scheduled step (PZEP-0004 Phase 2+), not a config tweak. Until
then, `frontend-v2` is developed and gate-reviewed standalone via
`ng serve` (port 3004) against the real backend, with its own `/login`
route (see PZEP-0008).

### Rule of thumb (anticipated — apply once Phase 1 is real)

Same shape of rule as `SOLID_STYLE_GUIDE.md`'s Solid federation section,
translated: a package is safe as a federation singleton only if every
consumer across the federation boundary resolves it through the **same**
shared-runtime path — not partly through the shared manifest and partly
through a locally-bundled copy. Verify this explicitly (inspect the built
`remoteEntry.json`'s `shared` array and the actual bundle output) before
trusting a `singleton: true` declaration alone; a declaration that isn't
backed by consistent resolution is exactly what caused Solid's version of
this bug.
