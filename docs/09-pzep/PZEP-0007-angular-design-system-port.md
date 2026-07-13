# PZEP-0007: Angular design-system port — full `packages/shared/ui` component library

## Status
Done — implemented and verified 2026-07-13/14 (build/test/live-serve level). Not yet committed.

## Date
2026-07-14

## Related Commit
(uncommitted — working tree only, pending user review)

## Requirement Sources
- Business: user directive 2026-07-13/14 — port the entire SolidJS design
  system (`frontend/packages/shared/ui`) to Angular, following a strict
  per-component file convention (`<name>.component.ts` +
  `<name>.component.html` + `<name>.component.scss` if needed — no inline
  templates, no exceptions).
- Feature: [PZEP-0004](./PZEP-0004-solidjs-to-angular-frontend-migration.md)
  Phase 2, [PZEP-0006](./PZEP-0006-angular-base-mfe-scaffold.md) (this PZEP
  continues that one's scaffold work at much larger scope — a separate
  document because the size warrants its own record, not because the intent
  differs).
- Use Cases: any future onboarding page can be built entirely from ported
  components, without inventing new ones or falling back to inline Tailwind
  strings.
- Functional Requirements: none new — no business behavior, pure UI library
  port.
- Acceptance Criteria: see below.
- UI Specs: `docs/03-architecture-detail-design/15-design-system.md` — every
  class string ported byte-for-byte from the Solid source, not redesigned.

## Summary

Ported all ~50 SolidJS component files under
`frontend/packages/shared/ui/` (`components/common/*.tsx`,
`components/common/display/*.tsx`, `components/dashboard/StatCard.tsx`,
`forms/fields.tsx`'s underlying `Primitives.tsx` field components, and
`toaster/*`) into `frontend-v2/src/app/shared/`, producing **86 component
folders, 98 `.component.ts`/`.component.html` file pairs**, plus 2 supporting
services (`ToasterService`, replacing Solid's module-singleton pattern) and
a handful of shared non-component helpers (`utils.ts`, `field-classes.ts`,
`collection-types.ts`, `display-shared.ts`).

Work was split across 4 parallel batches (dispatched as forks) plus one
batch (Primitives.tsx's field components, StatCard, Toaster, and the root
`App`/`Shell` file-convention alignment) done directly, followed by a single
consolidated build-and-fix pass that resolved all cross-batch integration
issues.

## Mandatory convention established (binding for all future `frontend-v2` components)

- One folder per component: `<kebab-case-name>/<kebab-case-name>.component.ts`
  + `.component.html` (+ `.component.scss` only if genuinely needed — almost
  nothing in this Tailwind-utility-first codebase needs one).
- No inline `template:` strings anywhere, including the root `App` component
  (renamed `app.ts` → `app.component.ts`, template extracted).
- `angular.json`'s `schematics` now defaults `ng generate component` to this
  same shape (`type: "component"`, `addTypeToClassName: false` so class
  names stay bare — `Button`, not `ButtonComponent` — `style: "scss"`,
  `skipTests: true` matching this project's existing "frontend tests
  optional" convention from `SOLID_STYLE_GUIDE.md`).
- No barrel/index re-export files — components are imported directly by
  path, matching the pattern already used for `Button`/`Card`/`Badge` in
  PZEP-0006.

## Problem

PZEP-0006 built exactly 5 shared components (`Button`, `Card`, `Badge`,
`Spinner`, `NavLink`) to prove the pattern. Everything else in the real
design system — collections/tables, overlays, forms, navigation, display
primitives, editors — did not exist in `frontend-v2` yet, which would have
forced every future page to either invent ad hoc markup or wait on a
component-by-component port blocking real feature work. Porting the whole
library up front removes that blocker for the onboarding redesign that
follows.

## Goals
- Every component in `frontend/packages/shared/ui` has a faithful Angular
  equivalent, same Tailwind classes, same visual output.
- The full port compiles clean under `ng build` (confirmed: Angular's
  compiler type-checks the entire `src` tree matched by `tsconfig.app.json`,
  not just files reachable from the app's routes — verified by watching
  errors surface in files nothing currently imports, e.g. `qr-code`,
  `heading`, `collection-controls`, before they were fixed).
- Known real Angular API differences from what was assumed going in are
  documented and corrected in `agent/ANGULAR_STYLE_GUIDE.md` so they aren't
  repeated.

## Non-Goals
- Wiring any of these components into real onboarding pages — that's the
  next PZEP, gated on the still-open redesign-scope question.
- Adding `@angular/cdk` for real focus-trap behavior in `Modal`/`Drawer` —
  explicitly deferred, same as PZEP-0006's original Overlay note.
- Adding CodeMirror or `@toast-ui/editor` — `CodeEditor`, `RichTextEditor`,
  and `MarkdownPreview` are structurally ported but render a plain
  `<textarea>`/raw text instead of a real editor instance, pending a
  dependency decision.
- Reconciling every "assumed API shape" judgment call the parallel batches
  made independently beyond what actually broke the build — see Verification
  Results for what was actually wrong vs. what turned out to already match.

## Real Findings From The Consolidated Build Pass

The build started at **17 errors**, all genuine cross-batch or Angular-API
issues, not sloppy porting:

1. **`resource()`'s option is `params`, not `request`.** Verified against
   the actually-installed `@angular/core` type declarations
   (`node_modules/@angular/core/types/_api-chunk.d.ts`,
   `BaseResourceOptions`/`PromiseResourceOptions`/`ResourceLoaderParams`,
   all tagged `@publicApi 22.0`). This was wrong in **`agent/ANGULAR_STYLE_GUIDE.md`
   itself** (written before any real `resource()` code existed) — both
   examples there are now fixed, and `qr-code.component.ts` (the only
   component that used `resource()`) is fixed. This is the most
   consequential finding of this PZEP: a style-guide error that would have
   propagated into every future resource-based component if not caught here.
2. **Required inputs are not readable at field-initializer OR constructor
   time** (`NG8118`) — two components (`CollectionFilters`, `CollectionToolbar`)
   tried to seed a local writable signal from a required input eagerly.
   Fixed with `linkedSignal()` (Angular 22's supported pattern for "writable,
   but re-derived from a reactive source") in both cases. One judgment call:
   `CollectionToolbar`'s original Solid behavior (`createSignal(props.search)`)
   only reads the initial value and never re-syncs; `linkedSignal` re-syncs
   whenever `search()` changes. Accepted as more correct UX (an externally
   reset search should be reflected), documented inline.
3. **`as` is a reserved word in Angular template expression syntax**
   (used for `@if (...; as x)` aliasing) — `Heading`'s `as` input broke
   `@switch (as())` with a parser error. Renamed to `tag`.
4. **Cross-batch path assumptions** — `CollectionControls` assumed
   `ErrorAlert`/`LoadingInline` lived under a `feedback/` folder; they
   actually live at top-level `error-alert/`/`loading-inline/` folders (each
   batch's own naming convention split multi-export source files
   independently, and this one guessed wrong). Fixed by reconciling against
   what actually exists, not by re-guessing.
5. **`qrcode` was genuinely missing** as a `frontend-v2` dependency — added
   it (`^1.5.4`, matching `frontend/`'s version) plus a ported copy of
   `frontend/src/solid/types/qrcode.d.ts` (the package ships no types).
   Explicit, deliberate dependency addition, not silently worked around.
6. **`aria-label` as a bare attribute on a custom component doesn't reach
   the component's real inner interactive element** — Angular puts
   unbound static attributes on the host tag (`<app-button>`), not
   whatever `<button>`/`<a>` the component's own template renders. Real
   accessibility regression risk if left as-is; fixed by adding an explicit
   `ariaLabel` input to `Button`, forwarded via `[attr.aria-label]` on all
   three of its internal branches (external link / internal link / native
   button).
7. **A stale `federation.config.mjs` reference to the pre-rename `app.ts`**
   surfaced as a hard build failure the moment the cache was cleared — fixed
   (`./src/app/app.ts` → `./src/app/app.component.ts`), independent of the
   component port itself but found during this pass.

After fixes: **0 errors**, clean `ng build`, clean `ng test` (1/1, unchanged
from PZEP-0006 — no new tests were part of this PZEP's scope), clean
`prettier --check` after one `--write` pass (195 files reformatted to
`frontend-v2`'s own `.prettierrc`, which — unlike `frontend/`'s — defaults to
semicolons; confirmed intentional, not a mistake, since it's a separate
project with its own established config).

## Deliberate Simplifications (per-batch judgment calls, not build-breaking, kept as documented deferrals)

- `Modal`/`Drawer` overlays: no real focus-trap/focus-restore (needs
  `@angular/cdk/a11y`, not added). `role`/`aria-modal`/`aria-labelledby`/
  Escape-closes/click-outside-closes are all wired.
- `Accordion`/`ListGroup`/`Navigation`'s `NavItem`: Solid `JSX.Element` props
  for arbitrary rich content/icons were dropped or simplified to `string` —
  no per-array-item `TemplateRef` plumbing exists yet in this codebase, and
  no current consumer needs it.
- `CodeEditor`/`RichTextEditor`/`MarkdownPreview`: structurally ported,
  editor instantiation stubbed pending a CodeMirror/`@toast-ui/editor`
  dependency decision.
- Several `Show when={onCloseHandler}`-style conditional-rendering-by-
  "is a callback passed" patterns became unconditional renders — Angular's
  `output()` has no "is anything subscribed" introspection the way a Solid
  prop can be checked for `undefined`.
- `SearchSelectField` (batch 4) inlined its own field-label markup instead
  of importing the real `FieldLabel` component built in this same session by
  a different batch — left as-is rather than force a dependency between two
  parallel batches' work; worth deduplicating in a later small cleanup pass,
  not urgent.

## Affected Components
- Frontend only: `frontend-v2/src/app/shared/**`, `frontend-v2/angular.json`
  (schematics defaults), `frontend-v2/package.json` (new `qrcode` dependency),
  `frontend-v2/federation.config.mjs` (stale path fix), `frontend-v2/src/app.component.*`
  (rename + real template), `agent/ANGULAR_STYLE_GUIDE.md` (`resource()`
  API correction).

## API / DB / Event / Permission Contract Changes
None.

## Security Considerations
None beyond what PZEP-0006 already covered — this PZEP adds no new runtime
surface, only presentational components and one new npm dependency
(`qrcode`, already vetted and in production use in `frontend/`).

## Test Plan
- Build: `ng build` — 0 errors (was 17, iteratively fixed and re-verified
  after each batch of fixes, not just once at the end).
- Serve: `ng serve --port 3004` — live-served, curl-verified against the
  real bootstrap chunk that Shell/AdminHomePage still render correctly
  after the full port (regression check against PZEP-0006's baseline).
- Test: `ng test --watch=false` — 1/1 passing, unchanged.
- Format: `prettier --check`/`--write` across all new files.
- E2E: none — no browser automation tool available this session, same
  limitation as every prior PZEP in this migration.

## Agent Implementation Plan
- TASK-0001 through TASK-0004: 4 parallel batches (display components;
  structural/layout components + overlays; collection/data/nav components;
  editor/content/typography components) — each self-contained, each
  required to read every source file in full before porting, each
  forbidden from running `ng build`/`serve`/`test` (avoids concurrent cache
  races) or creating barrel files.
- TASK-0005: primary agent ports `Primitives.tsx`'s remaining field
  components (`FieldLabel`, `InputField`, `SelectField`, `TextareaField`,
  `CheckboxField`, `RadioGroupField`, `ToggleField`, `FileInputField`,
  `SearchInputField`, `NumberInputField`, `PhoneInputField`,
  `DateInputField`, `TimeInputField`), `StatCard`, and the toaster
  (`ToasterService` + `ToasterProvider`), plus renames `app.ts` →
  `app.component.ts` for full convention compliance.
- TASK-0006: consolidated build pass — fix all real errors found (see Real
  Findings above), re-run build/test/format until clean.

All 6 tasks done.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: every source component has a ported Angular equivalent (or a documented, explicit reason it doesn't) | TASK-0001–0005 | File-count inventory: 86 folders / 98 component files vs. ~50 source files (1:1 or 1:many where a source file had multiple exports) |
| AC-2: 100% of ported files follow the mandatory 3-file convention, zero inline templates | TASK-0001–0006 | `grep -rl "template: \`" src/app/shared/ui` → 0 results |
| AC-3: full `ng build` is clean | TASK-0006 | `ng build`, 0 errors |
| AC-4: live-serve regression check against PZEP-0006's baseline still passes | TASK-0006 | `ng serve` + curl content check |
| AC-5: `agent/ANGULAR_STYLE_GUIDE.md`'s `resource()` examples match the real installed API | TASK-0006 | Diffed against `node_modules/@angular/core/types/_api-chunk.d.ts` directly |

## Open Questions
- Whether to add `@angular/cdk` now (unblocks real `Modal`/`Drawer` focus
  trap) or defer until a real page actually needs a modal — not decided
  here, deliberately deferred to whichever future PZEP first needs it.
- Whether to add CodeMirror/`@toast-ui/editor` — same deferral, same reason.
- The `SearchSelectField` / `FieldLabel` duplication noted above — small,
  not urgent, worth a follow-up cleanup task rather than reopening this PZEP.
