# UI State Spec Template

Use this after a Figma screen exists and before assigning frontend work.

```md
# <Feature> UI State Spec

## Source

- Figma link:
- Requirement:
- Route:

## Screen Purpose

<What this screen lets the user do.>

## Route And Access

- Route:
- Layout/shell:
- Required authentication:
- Required permission:
- Store/workspace context:

## Component Inventory

| Component | Responsibility | Shared or feature-local |
| --- | --- | --- |
|  |  |  |

## Fields

| Field | Control | Required | Validation | Error copy |
| --- | --- | --- | --- | --- |
|  |  |  |  |  |

## Actions

| Action | Trigger | API/command | Loading state | Success | Failure |
| --- | --- | --- | --- | --- | --- |
|  |  |  |  |  |  |

## States

| State | Condition | UI behavior |
| --- | --- | --- |
| Initial |  |  |
| Loading |  |  |
| Empty |  |  |
| Submitting |  |  |
| Validation error |  |  |
| Permission denied |  |  |
| API error |  |  |
| Success |  |  |

## Data Dependencies

- Query:
- Mutation:
- Storage adapter:
- Route search params:

## Collection Behavior

If the screen has a list:

- Pagination: server | cursor | bounded client
- Search fields:
- Filters:
- Sort fields:
- Empty state:
- Refetch behavior:

## Responsive Behavior

- Desktop:
- Tablet:
- Mobile:

## Accessibility

- Labels:
- Focus management:
- Keyboard behavior:
- Status announcements:

## Acceptance Criteria

- <criterion>
```
