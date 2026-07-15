# ADR-0006: Angular Material replaces the ported Tailwind/Flowbite design system for `frontend-v2`

## Status
Accepted

## Date
2026-07-14

## Related Commit
(none yet — decision only; implementation follows in the same session)

## Context

PZEP-0006/PZEP-0007 ported `packages/shared/ui` (the SolidJS design system —
Tailwind + a vanilla-Flowbite-CSS-plugin approach, since `flowbite-angular`
is version-locked below Angular 22) into `frontend-v2` as ~98 hand-written
Angular components. ADR-0005's Decision section explicitly called for
preserving this design system "incrementally, not thrown away and
redesigned."

Building the onboarding backbone's `/login` and `/register` pages (PZEP-0008
M0) surfaced four real bugs in this ported system within ~150 lines of new
template code, using only 6 shared components:

1. `classes()` (the shared class-merge helper, ported verbatim from Solid's
   naive `values.filter(Boolean).join(' ')`) does not resolve conflicting
   Tailwind utilities — a component's own hardcoded base class (`Card`'s
   `bg-white`) silently won over a caller's override (`bg-gray-950`)
   depending on Tailwind's internal CSS generation order, not the class
   list's DOM order. Confirmed by grepping the built stylesheet.
2. A `[&_p]:text-gray-300` arbitrary-variant override, applied to an entire
   `Card`, bled into unrelated nested `<p>` tags (the "01/02/03" tip boxes)
   via plain CSS descendant-selector specificity — not fixable by the
   `classes()` fix above, a markup-pattern flaw.
3. `Button`'s `<ng-content />`, duplicated across its three `@if`/`@else
   if`/`@else` branches (routerLink / external anchor / native button) for
   the polymorphic-wrapper pattern, silently failed to project into the
   external-anchor branch specifically — confirmed via live DOM inspection
   (`<a ...><!--container--></a>`, no projected text at all) while the same
   pattern worked in the native-`<button>` branch.
4. A static (non-conditional) `errorText` string on a field component
   rendered its error `<span>` unconditionally, regardless of the paired
   `[error]` boolean — a usage bug in this session's own new code, not the
   ported library, but symptomatic of the same class of easy-to-miss issue.

Items 1-3 were fixed at the root (a `tailwind-merge`-based `classes()`, a
narrower override scope, an `ngTemplateOutlet`-based `Button` template) and
verified via build + live DOM re-inspection. But the other ~92 of 98 ported
components have never been exercised in a real browser — PZEP-0007's own
verification was `ng build`/`ng test`/`prettier` only, never a click-through.
The bug rate found in a small, simple surface is a real signal that similar
undiscovered issues likely exist elsewhere in the port.

## Decision

`frontend-v2` adopts **Angular Material + SCSS** as its complete styling
approach, replacing Tailwind CSS and the vanilla-Flowbite-CSS-plugin
approach outright. This **revises** ADR-0005's "design system preserved,
re-implemented incrementally" clause for `frontend-v2` specifically —
`frontend/` (SolidJS) is unaffected and keeps its own Tailwind/Flowbite-derived
`packages/shared/ui`.

Revised 2026-07-15, per explicit user direction: replace Tailwind/Flowbite
**entirely**, not incrementally alongside it. Scope:

1. `ng add @angular/material` — theme, typography, and animations
   provider setup, SCSS as the project's stylesheet language.
2. Remove Tailwind CSS and the Flowbite CSS plugin from `frontend-v2`
   entirely: `tailwind.config`, PostCSS config, Tailwind directives in
   global styles, the `tailwind-merge` dependency and its `classes()`
   usage. Global styles move from `styles.css` to `styles.scss` with the
   Material theme.
3. Of the 98 ported components under `shared/ui`, only 13 directories are
   actually imported anywhere in the app (traced via
   `grep -rhoE "from '[^']*shared/ui/[^']*'"`): `button`, `card`,
   `error-alert`, `input-field`, `section-lead`, `nav-link`,
   `toaster-provider`, `empty-block`, `list-group`, `loading-block`,
   `spinner` (used by `button`), `field-label` (used by `input-field`),
   `loading-inline` (used by `loading-block`). These are rewritten on
   Material primitives + SCSS: `Card` → `mat-card`, `Button` →
   `mat-button`/`mat-flat-button`/`mat-icon-button` as appropriate,
   `InputField`+`FieldLabel` → `mat-form-field` + `matInput` +
   `mat-label`, `ErrorAlert` → Material-styled inline error, `ListGroup` →
   `mat-list`, `LoadingBlock`/`EmptyBlock`/`LoadingInline`/`Spinner` →
   `mat-progress-spinner` + SCSS empty-state pattern, `ToasterProvider` →
   kept as a custom signal-driven overlay (no direct Material
   equivalent for a stacked-toast list; `MatSnackBar` shows one at a
   time), restyled with SCSS instead of Tailwind utility classes.
   `SectionLead` and `NavLink` have no Material equivalent (marketing
   copy layout / plain routed link) and stay custom, SCSS-styled.
4. The remaining ~85 ported component directories that nothing imports
   are **deleted**, not migrated — confirmed dead code (traced the same
   way), and converting unused components to Material now is wasted
   effort per explicit user direction. Any of these get re-created on
   Material primitives on demand, when a future milestone actually needs
   that component — not ported forward speculatively.

## Alternatives Considered

### Option A: Keep the ported Tailwind/Flowbite system, audit it
Pros:
- No visual divergence from `frontend/`'s design system during the
  migration window.
- No new dependency; the 3 root-cause fixes already made (twMerge,
  narrower override scope, `ngTemplateOutlet`) directly reduce the same bug
  classes across all 98 components, not just the 2 pages tested.

Cons:
- Rejected by user — after the concrete bug count on a small surface, user
  judged the ongoing maintenance/audit burden not worth it.

### Option B: Angular Material (this decision)
Pros:
- Official, Angular-team-maintained, battle-tested — removes an entire
  class of "we wrote this ourselves and it has an edge case" risk.
- Ships accessibility (focus-trap, ARIA) already built in — directly
  resolves the Modal/Drawer focus-trap gap PZEP-0007 flagged as needing
  `@angular/cdk` (Material is built on CDK).
- Matches "built-in for Angular" as closely as anything not literally
  shipped in `@angular/core` itself.

Cons:
- Visual language diverges from `frontend/`'s Tailwind/Flowbite design
  system permanently — accepted explicitly by user (2026-07-14, reaffirmed
  2026-07-15 as a full replacement, not a phased coexistence) with this
  tradeoff stated up front.
- New dependency surface (`@angular/material`, `@angular/cdk`) and a theme
  file to maintain; Tailwind/PostCSS/`tailwind-merge` removed entirely.
- The ~85 unused ported components are deleted rather than kept as a
  fallback — re-creating any of them later means writing fresh Material
  markup, not resurrecting the old file. Accepted per explicit user
  direction (2026-07-15): they were already dead code, and converting
  unused surface has no payoff.

### Option C: Keep custom visual components, add `@angular/cdk` only for missing behavior
Pros:
- Smallest change — fixes the concrete a11y gap (focus-trap) without a
  visual rewrite.

Cons:
- Rejected by user — does not address the review-burden/bug-rate concern
  that drove this decision; only patches one specific symptom (missing
  focus-trap), not the broader "hand-written component correctness" risk.

Chosen: **B**, per explicit user direction (2026-07-14).

## Consequences

- `frontend-v2` and `frontend/` now have permanently different UI languages
  (Tailwind/Flowbite-derived vs. Material), not just for a migration
  window. This is a product-visible tradeoff, not just internal — flag it
  if/when a screen is shown to anyone outside the immediate dev loop.
- `agent/ANGULAR_STYLE_GUIDE.md` needs a follow-up update to reflect: SCSS
  (not Tailwind) as the stylesheet language, and Material-specific
  component guidance replacing the ported Flowbite notes in its "Reject"
  table.
- The `classes()`/`tailwind-merge` helper and the `ngTemplateOutlet`-based
  `Button` content-projection fix are removed along with Tailwind — the
  latter's *pattern* (single shared `<ng-template>` for a polymorphic
  wrapper's projected content) is still correct practice and should be
  reused in the Material `Button` rewrite even though the specific file is
  rewritten from scratch.
- Future PZEPs touching `frontend-v2` UI must use Material components +
  SCSS for any surface being newly built; do not add Tailwind utility
  classes or port another Flowbite component from `frontend/` into
  `frontend-v2` going forward.
- `docs/03-architecture-detail-design/15-design-system.md` (the Tailwind
  design-system doc) stays authoritative for `frontend/` only; it no longer
  describes `frontend-v2`'s target UI at all, not even transitionally.
- Any of the ~85 deleted ported components is recreated from scratch on
  Material + SCSS if a future milestone needs it — there is no fallback
  copy to resurrect.

## Rule Of Thumb

New `frontend-v2` UI work uses Angular Material + SCSS only; Tailwind CSS,
the Flowbite plugin, and `tailwind-merge` are fully removed from
`frontend-v2` and must not be reintroduced. A component that doesn't exist
yet gets written fresh on Material primitives when actually needed, not
ported from `frontend/`'s Tailwind/Flowbite version.
