# Podzone Agent Rule

Before modifying code, every agent must read:

1. `docs/00-governance/agent-working-rule.md`
2. `agent/SKILL.md`
3. The assigned task file or sprint slice
4. Referenced requirement, architecture, contract, and component docs

Podzone follows **Spec-first Vertical Slice Delivery**:

```text
Business need
  -> Requirement
  -> Use case
  -> Acceptance criteria
  -> PZEP when cross-component
  -> Architecture/component spec
  -> Contract
  -> Agent task
  -> Code
  -> Test
  -> Review
```

## Hard Rules

Do not code if the task does not include:

- requirement or recovery source;
- acceptance criteria;
- PZEP or architecture decision for cross-component work;
- contract link when FE, BE, DB, event, or permission behavior changes;
- allowed files or clear module boundary;
- forbidden changes;
- validation commands.

Do not:

- modify files outside scope;
- invent API DTOs;
- change API contracts without approval;
- change DB schema without DB spec;
- add new package/module without approval;
- add dependency without approval;
- refactor unrelated code;
- put business logic in handlers;
- put business logic in frontend components;
- bypass tenant isolation;
- bypass permission checks.

## Backend Rule

```text
domain -> usecase/interactor -> port -> adapter -> handler
```

- Domain must not import infrastructure.
- Usecase/interactor depends on ports.
- Handler maps transport only.
- Repository handles persistence only.
- Tenant-scoped queries must filter tenant/workspace/store scope.

## Frontend Rule

```text
Page -> ViewModel -> API Client -> Contract Types
```

- Page composes layout.
- Component renders UI.
- ViewModel owns state, actions, loading, and errors.
- API client follows contract.
- No direct fetch in UI components.

## Required Handoff

After implementation, report:

- summary;
- changed files;
- tests/validation run;
- skipped items with reason;
- risks or follow-up.

If required documentation is missing, stop and report what is missing instead of
coding.
