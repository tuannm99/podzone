# Agent Code Review Checklist

Use this checklist when reviewing AI-generated changes.

## Requirement Fit

- Does the diff implement the requested requirement?
- Does it follow the referenced spec or contract?
- Did the agent change anything outside scope?
- Are docs updated when behavior or architecture changed?

## Backend Architecture

- Are controllers thin inbound adapters?
- Is business logic in domain/interactor/usecase code?
- Does domain avoid importing infrastructure, transport, SQL, Fx, HTTP, gRPC,
  GraphQL, or Kafka runtime packages?
- Do usecases depend on output ports instead of adapters?
- Are cross-service calls through gRPC clients, events, or projections?
- Did the change avoid unnecessary new packages or abstractions?
- Are compile-time interface assertions present for important adapters?

## Backend Correctness

- Are tenant/workspace/store scopes enforced where required?
- Are repository queries tenant-aware where required?
- Are transactions used where multiple writes must commit together?
- Are errors mapped consistently and with enough detail?
- Are permission-denied errors expected and user-actionable?
- Are idempotency and retries considered for workers?
- Is outbox used only for commit-coupled side effects?

## Frontend Architecture

- Is the route page only a composition wrapper?
- Does a feature ViewModel own resources, forms, actions, loading, errors, and
  derived state?
- Do panels/components avoid direct service calls?
- Are reactive props read safely without destructuring?
- Are forms modeled as cohesive stores instead of one signal per field?
- Are expected validation and permission errors kept inside the current screen?

## Frontend UX

- Are loading, empty, error, permission denied, and success states handled?
- Do operational lists use backend pagination, search, filters, and sort?
- Does pagination preserve page state and avoid document scroll resets?
- Are long workflows split into tabs, drawers, bounded scroll regions, or
  list/detail views when appropriate?
- Are buttons, links, labels, focus states, and status text accessible?

## Database

- Is the owner service clear?
- Is a migration included when schema changes?
- Are table and column names consistent?
- Are tenant-scoped unique indexes scoped by tenant where required?
- Does soft delete affect uniqueness or filtering?
- Are indexes present for expected query paths?
- Are secrets stored as references rather than plaintext?

## Contracts

- Are API/proto/GraphQL contracts updated intentionally?
- Is generated code updated only when contract sources changed?
- Are DTOs separated from domain entities?
- Are pagination and error shapes consistent with common conventions?
- Are backward compatibility and migration impact documented?

## Testing And Verification

- Are unit tests present for usecase/domain behavior?
- Are repository/API tests added for important persistence or transport changes?
- Were generated mocks updated when ports changed?
- Were scoped test/lint/build commands run?
- Did `git diff --check` pass?
- If a check was skipped, is the reason explicit?

## Handoff Quality

- Did the agent list changed files?
- Did the agent describe behavior changed?
- Did the agent list commands run and results?
- Did the agent call out known gaps or risks?
- Is the suggested commit message small and themed?
