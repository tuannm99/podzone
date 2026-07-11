# ui-podzone Remaining Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 5 remaining issues from `docs/07-problems/fe-2026-07-09.md` — spurious API calls, silent UI truncation, misleading Button API, and URL state lost on refresh.

**Architecture:** All fixes are in `internal/ui-podzone/src`. Tasks 1–3 touch ViewModels/router; Tasks 4–5 touch `Primitives.tsx` and `OrdersInsightsPanel.tsx`. Each task is independently committable. Task 3 (URL state migration) is the largest and must be done atomically since it changes the router schema and ViewModel at the same time.

**Tech Stack:** SolidJS, TanStack Router (`@tanstack/solid-router`), TypeScript, Vite, Tailwind CSS.

## Global Constraints

- Run `npm run build` (from `internal/ui-podzone/`) after every task — zero type errors required before commit.
- Run `npm run lint` and `npm run format` before the final commit of each task.
- Do not add new dependencies.
- Do not change public API contracts (service layer, gRPC, REST).
- Follow `agent/SOLID_STYLE_GUIDE.md`: ViewModels own state; panels/views do not call services directly.

---

## File Map

| File | Task | Change |
|------|------|--------|
| `src/modules/backoffice/pages/orders/createTenantOrdersViewModel.ts` | 1, 3 | Fix `recommendationResource` source; migrate 5 signals to URL state |
| `src/solid/app-router.tsx` | 3 | Add `validateSearch` to `tenantOrdersRoute` |
| `src/modules/backoffice/pages/orders/board-context.tsx` | 3 | Update `TenantOrdersBoardContextValue` type for new setter signatures |
| `src/modules/backoffice/pages/orders/OrdersInsightsPanel.tsx` | 4 | Add count badges to truncated arrays |
| `src/solid/components/common/Primitives.tsx` | 5 | Remove `'blue'` from `ButtonColor` |
| `src/modules/backoffice/pages/orders/CreateRoutedOrderPanel.tsx` | 5 | `color="blue"` → `color="primary"` |
| `src/modules/backoffice/pages/orders/order-card/ShipmentPanel.tsx` | 5 | `color="blue"` → `color="primary"` |
| `src/modules/backoffice/pages/home/TenantHomeSections.tsx` | 5 | `color="blue"` → `color="primary"` (3 sites) |

---

## Task 1: Fix `recommendationResource` unstable object source

**Finding:** NEW1 in `fe-2026-07-09.md`.  
`createResource` compares sources with `===`. An object literal always produces a new reference → every reactive evaluation triggers a spurious refetch even when field values are unchanged.

**Files:**
- Modify: `src/modules/backoffice/pages/orders/createTenantOrdersViewModel.ts:164–176`

**Interfaces:**
- Produces: no interface change — `recommendationResource` signature unchanged.

- [ ] **Step 1: Open the file and locate the source**

Read `src/modules/backoffice/pages/orders/createTenantOrdersViewModel.ts` lines 164–176.
Current code:
```ts
const [recommendationResource] = createResource(
    () => {
        if (!workspaceReady()) return undefined
        const candidateID = orderForm.values.selectedCandidateId.trim()
        if (!candidateID) return undefined
        return {
            candidateId: candidateID,
            productType: orderForm.values.selectedProductType,
            shipRegion: orderForm.values.selectedShipRegion,
            preferredPartner: orderForm.values.preferredPartner.trim() || undefined,
        }
    },
    async (source) => getRoutedOrderRecommendation(source)
)
```

- [ ] **Step 2: Replace with stable string key**

Replace the entire `recommendationResource` block with:
```ts
const [recommendationResource] = createResource(
    () => {
        if (!workspaceReady()) return undefined
        const candidateID = orderForm.values.selectedCandidateId.trim()
        if (!candidateID) return undefined
        return [
            candidateID,
            orderForm.values.selectedProductType,
            orderForm.values.selectedShipRegion,
            orderForm.values.preferredPartner.trim(),
        ].join('|')
    },
    async () =>
        getRoutedOrderRecommendation({
            candidateId: orderForm.values.selectedCandidateId.trim(),
            productType: orderForm.values.selectedProductType,
            shipRegion: orderForm.values.selectedShipRegion,
            preferredPartner: orderForm.values.preferredPartner.trim() || undefined,
        })
)
```

The fetcher reads `orderForm.values.*` directly — this is safe and untracked in the fetcher context (same pattern as the `queuePageResource` fix in Part 6).

- [ ] **Step 3: Build**

```bash
cd internal/ui-podzone && npm run build
```
Expected: `✓ built in ~2s`, zero errors.

- [ ] **Step 4: Commit**

```bash
git add internal/ui-podzone/src/modules/backoffice/pages/orders/createTenantOrdersViewModel.ts
git commit -m "fix(vibe): ui-podzone - stabilize recommendationResource source key"
```

---

## Task 2: Add `validateSearch` to orders route (prerequisite for Task 3)

**Finding:** NEW2 in `fe-2026-07-09.md`.  
`tenantOrdersRoute` has no `validateSearch` schema. TanStack Router discards unrecognised params → the one-way sync `createEffect` at line 314 reads `unknown` types and silently no-ops. This must be done before Task 3 so that `useSearch({ from: '/t/$tenantId/orders' })` returns typed values.

**Files:**
- Modify: `src/solid/app-router.tsx:121–126`

**Interfaces:**
- Produces: `search().queueView: QueueView | 'all'`, `search().queueSort: QueueSort | 'priority'`, `search().operatorLens: string`, `search().queuePage: number`, `search().appliedQueueSearch: string` — used in Task 3.

- [ ] **Step 1: Check what type guards exist**

```bash
grep -n "isQueueView\|isQueueSort\|QueueView\|QueueSort" internal/ui-podzone/src/modules/backoffice/pages/orders/createTenantOrdersViewModel.ts | head -10
```

Note the import paths for `isQueueView` and `isQueueSort`. If they are not exported, you'll need to export them or inline the guard.

- [ ] **Step 2: Add `validateSearch` to `tenantOrdersRoute`**

In `src/solid/app-router.tsx`, locate `tenantOrdersRoute` (around line 121). Replace:
```ts
const tenantOrdersRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/orders',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    component: lazyRouteComponent(backofficeRouteComponents.tenantOrders),
})
```
With:
```ts
const tenantOrdersRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/orders',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    validateSearch: (search: Record<string, unknown>) => ({
        queueView: typeof search.queueView === 'string' ? search.queueView : 'all',
        queueSort: typeof search.queueSort === 'string' ? search.queueSort : 'priority',
        operatorLens: typeof search.operatorLens === 'string' ? search.operatorLens : '',
        queuePage: typeof search.queuePage === 'number' ? search.queuePage : 1,
        appliedQueueSearch: typeof search.appliedQueueSearch === 'string' ? search.appliedQueueSearch : '',
    }),
    component: lazyRouteComponent(backofficeRouteComponents.tenantOrders),
})
```

Note: `appliedQueueSearch` replaces both `queueSearch` (live input, stays local) and `appliedQueueSearch` (committed, goes to URL).

- [ ] **Step 3: Build**

```bash
cd internal/ui-podzone && npm run build
```
Expected: `✓ built in ~2s`, zero errors. (No functional change yet — the ViewModel still reads `useSearch({ strict: false })` which ignores the schema.)

- [ ] **Step 4: Commit**

```bash
git add internal/ui-podzone/src/solid/app-router.tsx
git commit -m "fix(vibe): ui-podzone - add validateSearch to tenantOrdersRoute"
```

---

## Task 3: Migrate orders ViewModel signals to URL state

**Finding:** NEW2 / Part 8 in `fe-2026-07-09.md`.  
`activeQueueView`, `activeQueueSort`, `operatorLens`, `queuePage`, `appliedQueueSearch` are `createSignal` — their values are lost on page refresh. They must be read from / written to URL search params. Requires Task 2 to have run first (so `validateSearch` types are available).

**Rule:** `agent/SOLID_STYLE_GUIDE.md` lines 283–284 — collection page/filter/sort must be URL state, not `createSignal`.

**Files:**
- Modify: `src/modules/backoffice/pages/orders/createTenantOrdersViewModel.ts`
- Modify: `src/modules/backoffice/pages/orders/board-context.tsx` (if setter types change)

**Interfaces:**
- The ViewModel's `boardContextValue` still exposes the same keys (`activeQueueView`, `setActiveQueueView`, etc.) — only their implementations change. The View does not need to change.

- [ ] **Step 1: Switch to typed `useSearch`**

At the top of `createTenantOrdersViewModel` (around line 28), replace:
```ts
const search = useSearch({ strict: false }) as () => Record<string, unknown>
```
With:
```ts
const search = useSearch({ from: '/t/$tenantId/orders' })
```
Add `useNavigate` to the import if not already present:
```ts
import { useNavigate, useParams, useSearch } from '@tanstack/solid-router'
```

- [ ] **Step 2: Replace the 5 `createSignal` declarations with URL-derived accessors**

Find these lines (around 34–51) and remove them:
```ts
const [queuePage, setQueuePage] = createSignal(1)
const [queueSearch, setQueueSearch] = createSignal('')
const [appliedQueueSearch, setAppliedQueueSearch] = createSignal('')
const [activeQueueView, setActiveQueueView] = createSignal<QueueView>('all')
const [activeQueueSort, setActiveQueueSort] = createSignal<QueueSort>('priority')
const [operatorLens, setOperatorLens] = createSignal('')
```

Add `const navigate = useNavigate()` right after `const params = useParams(...)`.

Add these URL-backed accessors in their place:
```ts
const navigate = useNavigate()

const queuePage = () => search().queuePage
const setQueuePage = (page: number) =>
    void navigate({ search: (prev) => ({ ...prev, queuePage: page }) })

// queueSearch stays local (live input value — ephemeral, no need to survive refresh)
const [queueSearch, setQueueSearch] = createSignal('')

const appliedQueueSearch = () => search().appliedQueueSearch
const setAppliedQueueSearch = (value: string) =>
    void navigate({ search: (prev) => ({ ...prev, appliedQueueSearch: value }) })

const activeQueueView = () => {
    const v = search().queueView
    return isQueueView(v) ? v : 'all'
}
const setActiveQueueView = (v: QueueView) =>
    void navigate({ search: (prev) => ({ ...prev, queueView: v }) })

const activeQueueSort = () => {
    const s = search().queueSort
    return isQueueSort(s) ? s : 'priority'
}
const setActiveQueueSort = (s: QueueSort) =>
    void navigate({ search: (prev) => ({ ...prev, queueSort: s }) })

const operatorLens = () => search().operatorLens
const setOperatorLens = (lens: string) =>
    void navigate({ search: (prev) => ({ ...prev, operatorLens: lens }) })
```

- [ ] **Step 3: Update `applyQueueSearch` in `boardContextValue`**

The current implementation calls `setQueuePage(1)` then `setAppliedQueueSearch(...)`.
With URL state, both must happen in a single `navigate` call to avoid two history entries:

```ts
applyQueueSearch: () => {
    void navigate({
        search: (prev) => ({
            ...prev,
            appliedQueueSearch: queueSearch().trim(),
            queuePage: 1,
        }),
    })
},
```

- [ ] **Step 4: Remove the one-way sync `createEffect`**

Find and delete the block around line 314 (now a few lines lower):
```ts
createEffect(() => {
    const current = search()
    const queueView = String(current.queueView || '')
    const queueSort = String(current.queueSort || '')
    const lens = String(current.operatorLens || '')

    if (isQueueView(queueView)) {
        setActiveQueueView(queueView)
    }
    if (isQueueSort(queueSort)) {
        setActiveQueueSort(queueSort)
    }
    if (lens) {
        setOperatorLens(lens)
    }
})
```

This effect no longer has a purpose — the accessors read directly from `search()`.

- [ ] **Step 5: Update `queuePageResource` source**

The current stable-key source (from Part 6 fix) reads `queuePage()`, `appliedQueueSearch()`, `activeQueueView()`, `activeQueueSort()`, `operatorLens()`. These are now URL-backed accessors so they still work as-is. No change needed here.

- [ ] **Step 6: Build and check for type errors**

```bash
cd internal/ui-podzone && npm run build
```

Expected: `✓ built in ~2s`. If there are type errors about `setActiveQueueView` etc., check `board-context.tsx` — the setter type may be `(v: QueueView) => void` (a synchronous setter), but the new implementation calls `navigate` which returns `void` synchronously too. Should be compatible.

If `board-context.tsx` has explicit setter type annotations, update them to match (the signature `(v: QueueView) => void` is unchanged).

- [ ] **Step 7: Lint and format**

```bash
cd internal/ui-podzone && npm run lint && npm run format
```

- [ ] **Step 8: Commit**

```bash
git add \
  internal/ui-podzone/src/modules/backoffice/pages/orders/createTenantOrdersViewModel.ts \
  internal/ui-podzone/src/modules/backoffice/pages/orders/board-context.tsx
git commit -m "fix(vibe): ui-podzone - migrate orders queue state to URL search params (Part 8)"
```

---

## Task 4: Add count indicators to truncated insight arrays

**Finding:** FE10 in `fe-2026-07-09.md`.  
Three `<For each={items().slice(0, 6)}>` render at most 6 rows silently. A store with 20 entries shows 6 with no hint that more exist.

**Files:**
- Modify: `src/modules/backoffice/pages/orders/OrdersInsightsPanel.tsx:43, 64, 110`

**Interfaces:**
- No ViewModel change needed — `forcedRerouteSummary()`, `reconciliationOrders()`, `partnerFinanceSummary()` already return full arrays. The slice is only in the view.

- [ ] **Step 1: Read the three affected sections**

Read `OrdersInsightsPanel.tsx` around lines 40–120 to understand the surrounding JSX structure.

- [ ] **Step 2: Add count badge pattern to each truncated section**

For each of the three sites, before the `<For each={insights.xyzArray().slice(0, 6)}>`, add:
```tsx
<Show when={insights.xyzArray().length > 6}>
    <p class="text-xs text-gray-400">
        Showing 6 of {insights.xyzArray().length}
    </p>
</Show>
```

Example for `forcedRerouteSummary` (line ~43):
```tsx
<Show when={insights.forcedRerouteSummary().length > 6}>
    <p class="mb-1 text-xs text-gray-400">
        Showing 6 of {insights.forcedRerouteSummary().length}
    </p>
</Show>
<For each={insights.forcedRerouteSummary().slice(0, 6)}>
    ...
</For>
```

Apply the same pattern to `reconciliationOrders` and `partnerFinanceSummary`.

- [ ] **Step 3: Build**

```bash
cd internal/ui-podzone && npm run build
```
Expected: `✓ built in ~2s`, zero errors.

- [ ] **Step 4: Commit**

```bash
git add internal/ui-podzone/src/modules/backoffice/pages/orders/OrdersInsightsPanel.tsx
git commit -m "fix(vibe): ui-podzone - show count when insight arrays are truncated"
```

---

## Task 5: Remove misleading `'blue'` from `ButtonColor`

**Finding:** FE14 in `fe-2026-07-09.md`.  
`ButtonColor` has `'blue'` and `'dark'` both mapping to identical `bg-gray-950 text-white` classes — same as `'primary'`. A caller passing `color="blue"` expecting a blue button gets a black one. `'blue'` is the misleading duplicate; remove it and rename call sites to `'primary'`.

**Files:**
- Modify: `src/solid/components/common/Primitives.tsx:5,9–11`
- Modify: `src/modules/backoffice/pages/orders/CreateRoutedOrderPanel.tsx:323`
- Modify: `src/modules/backoffice/pages/orders/order-card/ShipmentPanel.tsx:71`
- Modify: `src/modules/backoffice/pages/home/TenantHomeSections.tsx:169, 181, 271`

**Interfaces:**
- `ButtonColor` exported type changes: `'blue'` removed. Any downstream TypeScript file using `color="blue"` on a `Button` will get a compile error until renamed.
- `Badge` has its own `BadgeColor` type — `color="blue"` on `<Badge>` is unaffected.

- [ ] **Step 1: Remove `'blue'` from `ButtonColor` and its class entry**

In `src/solid/components/common/Primitives.tsx`:

Change line 5:
```ts
// Before
type ButtonColor = 'primary' | 'blue' | 'alternative' | 'light' | 'dark' | 'green' | 'red'
// After
type ButtonColor = 'primary' | 'alternative' | 'light' | 'dark' | 'green' | 'red'
```

Remove the `blue` entry from `buttonColorClasses` (line 11):
```ts
// Before
const buttonColorClasses: Record<ButtonColor, string> = {
    primary: 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300',
    blue: 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300',
    ...
}
// After — remove the blue line entirely
const buttonColorClasses: Record<ButtonColor, string> = {
    primary: 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300',
    ...
}
```

- [ ] **Step 2: Build to surface all `color="blue"` call sites**

```bash
cd internal/ui-podzone && npm run build 2>&1 | grep "blue"
```
Expected: type errors listing every `<Button ... color="blue" ...>` file and line.

- [ ] **Step 3: Rename call sites**

In each file surfaced by the build errors, change `color="blue"` → `color="primary"` on `Button` components only. Do not touch `Badge color="blue"` — that is `BadgeColor`, a different type.

Files to update (from `grep` result earlier):
- `src/modules/backoffice/pages/orders/CreateRoutedOrderPanel.tsx:323`
- `src/modules/backoffice/pages/orders/order-card/ShipmentPanel.tsx:71`
- `src/modules/backoffice/pages/home/TenantHomeSections.tsx:169, 181, 271`

- [ ] **Step 4: Build clean**

```bash
cd internal/ui-podzone && npm run build
```
Expected: `✓ built in ~2s`, zero errors.

- [ ] **Step 5: Lint and format**

```bash
cd internal/ui-podzone && npm run lint && npm run format
```

- [ ] **Step 6: Commit**

```bash
git add \
  internal/ui-podzone/src/solid/components/common/Primitives.tsx \
  internal/ui-podzone/src/modules/backoffice/pages/orders/CreateRoutedOrderPanel.tsx \
  internal/ui-podzone/src/modules/backoffice/pages/orders/order-card/ShipmentPanel.tsx \
  internal/ui-podzone/src/modules/backoffice/pages/home/TenantHomeSections.tsx
git commit -m "fix(vibe): ui-podzone - remove misleading ButtonColor 'blue' alias (rename to primary)"
```

---

## Self-Review

**Spec coverage:**

| Finding | Task | Covered? |
|---------|------|----------|
| NEW1 — `recommendationResource` unstable source | Task 1 | ✅ |
| NEW2 — `validateSearch` missing | Task 2 | ✅ |
| NEW2 / Part 8 — URL state migration | Task 3 | ✅ |
| FE10 — silent `.slice(0, 6)` | Task 4 | ✅ |
| FE14 — `'blue'` renders black | Task 5 | ✅ |
| FE11 — RadioGroupField/ToggleField label | — | ⚠️ See note below |

**FE11 note:** Audit flagged RadioGroupField/ToggleField but current code at those locations already uses `<fieldset><legend>` and wrapping `<label>` respectively — both correct HTML patterns. Verify by reading `Primitives.tsx:351–398` before starting a separate fix; if the audit was correct, add a Task 6. Not included here because the current code appears compliant.

**Placeholder scan:** No TBD, TODO, or "similar to" references found.

**Type consistency:** `queuePage()`, `activeQueueView()`, `activeQueueSort()`, `operatorLens()`, `appliedQueueSearch()` are all used consistently as accessor functions in Tasks 2 and 3. `setActiveQueueView` etc. are `(v: T) => void` in both the task description and `board-context.tsx`.

**Task ordering:** Task 2 must precede Task 3 (schema required for typed `useSearch`). Tasks 1, 4, 5 are independent.
