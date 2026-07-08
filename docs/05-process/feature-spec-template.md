# Feature Spec Template

Use this template before assigning implementation work to an agent.

````md
# <Feature Name>

## Status

Draft | Ready | In Progress | Done | Blocked

## Goal

<What should be true for the user or operator after this feature ships?>

## Actors

- <Actor 1>
- <Actor 2>

## User Flow

1. <Step>
2. <Step>
3. <Step>

## Screens

- <Screen or route>
- <Modal/drawer/detail view>

## UI State Spec

| State | Description | Expected UI |
| --- | --- | --- |
| Loading |  |  |
| Empty |  |  |
| Validation error |  |  |
| Permission denied |  |  |
| API error |  |  |
| Success |  |  |

## Fields

| Field | Type | Required | Validation | Notes |
| --- | --- | --- | --- | --- |
|  |  |  |  |  |

## Actions

| Action | Trigger | API/Command | Success | Failure |
| --- | --- | --- | --- | --- |
|  |  |  |  |  |

## Permissions

- Required action:
- Resource scope:
- Backend guard:
- Read-only behavior:

## Business Rules

- <Rule>
- <Rule>

## API Contract

```text
<Method> <Path or RPC>
Request:
Response:
Errors:
````

## Database Contract

Tables or collections:

- `<table_or_collection>`

Indexes:

- `<index>`

Migration required: yes | no

## Events Or Async Work

- Event:
- Outbox required: yes | no
- Worker/consumer:

## Acceptance Criteria

- <Criterion>
- <Criterion>

## Out Of Scope

- <Explicit non-goal>
- <Explicit non-goal>

## Verification

```bash
<commands>
```
```
