# Task: Gate Platform Navigation By Root Vs Org Role

**Sprint:** Post-backbone (unblocked 2026-07-12 by explicit product decision — backbone verified at API level, not full browser UI, see `docs/06-recovery/backbone-flow-refactor.md`)
**SRS:** SRS-IAM-003
**Design reference:** interactive mockup built during SRS-IAM-003 brainstorming (Artifact, not committed to repo) — Root sees Home/Provisioning/Settings/IAM("ROOT"); Org with no org-management role sees Home/Settings only; Org managing their own org sees Home/Settings/IAM("ORG").
**Status:** Implemented 2026-07-12 —
`frontend/src/modules/shell/usePlatformAccess.ts` (new),
`frontend/src/modules/shell/PodzoneNavbar.tsx` (gated + tagged). `npm run
build` / `lint` / `format:check` all pass.

**API-level verified live** against the running Docker dev stack: called
`GET /auth/v1/iam/organizations?collection.page=1&collection.pageSize=1`
with the dev-bootstrap JWT (`devowner@podzone.dev`, real `platform_owner`)
— response: `canManagePlatform: true`, `pageInfo.total: "5"` (5 real
organizations exist). Confirms the platform-admin branch renders
Provisioning + IAM("ROOT"). The org-only and no-org branches were verified
by reading `internal/iam/controller/grpchandler/organization_methods.go:180-212`
directly (the server-side branch is unambiguous: non-platform-admins get
`ListOrganizationsForUser`, only their own orgs), not exercised live — no
JWT was available for a plain organization_root or plain tenant-owner
account this session.

**Still not done:** browser-rendered confirmation (no headless-browser tool
available this pass) that the sidebar actually hides/shows the right links
for each account type — this is the same gap noted in
backbone-flow-refactor.md.

**Confirmed backend signal (no new API needed):** `POST /auth/v1/iam/organizations`
(`listOrganizations` in `internal/iam/controller/grpchandler/organization_methods.go:180-212`)
already computes `canManagePlatform` via `CheckPlatformPermission(actorUserID, "platform:manage_roles")`,
and branches server-side: platform admins get every organization
(`ListOrganizations`), everyone else gets only orgs they belong to
(`ListOrganizationsForUser`). So `canManagePlatform` (platform-wide) and
`pageInfo.total > 0` (member of at least one org) are both already present
in one existing response — no backend change required.

---

```text
You are working in the Podzone monorepo.

Task:
Add a shell-level platform-access resource that calls listOrganizations
once, and use it to gate the sidebar's Provisioning and IAM nav links by
role: Provisioning only for canManagePlatform; IAM for canManagePlatform
OR member of at least one organization; tag the IAM link "ROOT" or "ORG"
accordingly. Home and Settings stay unconditional.

Goal:
An org/store-owner account (no platform grant, no organization membership)
sees only Home and Settings in the sidebar — no Provisioning or IAM links
they can't use. A platform admin sees all four, IAM tagged ROOT. An
organization root/admin with no platform grant sees Home/Settings/IAM,
IAM tagged ORG.

References:
- CLAUDE.md
- docs/00-governance/agent-working-rule.md
- agent/SOLID_STYLE_GUIDE.md
- docs/01-srs/iam/SRS-IAM-003-platform-vs-organization-administration-surface.md
- frontend/src/modules/shell/PodzoneNavbar.tsx
- frontend/packages/shared/services/iam/organizations.ts (listOrganizations)
- frontend/apps/iam/src/pages/admin-iam/organizations/createOrganizationsState.ts (existing canManagePlatform consumer, pattern reference)

Scope:
You may modify:
- frontend/src/modules/shell/usePlatformAccess.ts (new file)
- frontend/src/modules/shell/PodzoneNavbar.tsx

Out of scope:
- Do not add a route guard/redirect for direct URL access to /admin/provisioning
  or /admin/iam — backend already 403s; this task is nav visibility only
  (per earlier decision: "check ý 1 trước" — redirect guard is a separate,
  not-yet-decided follow-up).
- Do not split /admin/iam into two separate consoles/routes — same route,
  same console, already adapts its own content via canManagePlatform inside
  apps/iam (principals/directory sections). This task only touches the
  shell nav link visibility and its tag, not the console internals.
- Do not touch apps/onboarding or apps/backoffice.
- Do not add a new backend endpoint — listOrganizations already returns
  everything needed.
- Do not change IAM permission model or backend code.

Architecture rules:
- Frontend permission checks are not a security boundary (SRS-IAM-001) —
  this is UX only; the backend already enforces the real boundary
  (confirmed: /admin/provisioning and /admin/iam's underlying APIs 403 for
  unauthorized callers regardless of what the nav shows).
- Match existing shell conventions: PodzoneNavbar.tsx already imports
  services directly (tokenStorage, logout, useTenantWorkspace) rather than
  going through a ViewModel — the shell is documented as not following the
  MVVM pattern yet (design-system.md "MVVM Compliance by Module"). Follow
  that existing convention, don't introduce a new pattern for this one file.

Acceptance criteria:
- usePlatformAccess() calls listOrganizations({page:1, pageSize:1, ...})
  once per session (createResource, not re-fetched on every render) and
  exposes canManagePlatform() and hasAnyOrganization() accessors.
- Sidebar Platform nav shows Home and Settings unconditionally.
- Sidebar Platform nav shows Provisioning only when canManagePlatform() is true.
- Sidebar Platform nav shows IAM when canManagePlatform() is true OR
  hasAnyOrganization() is true; hidden otherwise.
- IAM nav link shows a small "ROOT" tag when canManagePlatform() is true,
  "ORG" tag otherwise (only rendered when the link itself is shown).
- While the platform-access resource is loading, Provisioning and IAM are
  not shown (fail toward hiding, not toward flashing links that then
  disappear).
- npm run build, npm run lint, npm run format:check all pass.

Validation:
- cd frontend && npm run build
- cd frontend && npm run lint
- cd frontend && npm run format:check
- git diff --check

Handoff:
- Changed files
- Behavior changed
- Verification results
- Known gaps (e.g. not yet verified against a live platform_owner vs
  plain-tenant account in the browser — no browser automation tool
  available this session, see backbone-flow-refactor.md)
- Suggested commit message: feat(shell): gate Provisioning/IAM nav links by platform vs org role
```
