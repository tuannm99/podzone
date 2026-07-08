# Vertical Slice Breakdown Template

Use this after the requirement, UX state spec, contracts, and DB notes are ready.

````md
# <Feature> Vertical Slice Breakdown

## Source Documents

- Requirement:
- Figma/UI state spec:
- Architecture/ADR:
- API/proto/GraphQL contract:
- DB/migration note:

## Backbone Flow

```text
<actor>
  -> <screen/action>
  -> <API/command>
  -> <domain behavior>
  -> <persistence/event>
  -> <UI result>
```

## Slice List

### Slice 1: <Name>

Goal:

- <One behavior>

Scope:

- Allowed:
  - `<file or package>`
- Out of scope:
  - <non-goal>

Contract:

- API:
- DB:
- Events:
- Permission:

Implementation tasks:

- <task>
- <task>

Tests:

- <test>

Verification:

```bash
<command>
```

Done:

- <criterion>

### Slice 2: <Name>

Repeat the same structure.
````

## Slice Sizing Rules

Good slice size:

- one usecase;
- one endpoint or resolver;
- one UI action;
- one migration and repository path;
- one worker step.

Bad slice size:

- entire service refactor;
- all screens for a module;
- all tables first;
- UI without contract;
- DB without owner service;
- generated code without source contract.

## Common Slice Types

### Contract Slice

Use when API/proto/GraphQL must be locked before implementation.

Includes:

- schema/source contract;
- generated code;
- compile fix;
- contract tests if present.

Does not include:

- broad business implementation.

### Domain Slice

Includes:

- command/query types;
- usecase/interactor;
- domain behavior;
- port mocks;
- unit tests.

Does not include:

- HTTP/gRPC/GraphQL handler unless needed for the slice.

### Persistence Slice

Includes:

- migration;
- repository method;
- indexes;
- repository tests.

Does not include:

- UI or unrelated query paths.

### Transport Slice

Includes:

- handler/resolver;
- mapper;
- error mapping;
- permission guard;
- transport tests.

Does not include:

- business rules in handler/resolver.

### Frontend Slice

Includes:

- service adapter;
- feature ViewModel;
- UI component/panel;
- loading/empty/error/success states;
- build/lint verification.

Does not include:

- backend contract changes unless explicitly included.

### Worker Slice

Includes:

- worker step;
- idempotency;
- retry/failure state;
- event/outbox handling if commit-coupled;
- tests.

Does not include:

- unrelated worker supervision refactor.
