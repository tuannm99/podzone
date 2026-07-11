# Component Spec Template

Use this template for service, frontend module, worker, gateway, or shared
package boundaries. Component specs belong under detail design docs, usually
`docs/03-architecture-detail-design`.

```markdown
# Component: <Component Name>

## Purpose
<Why this component exists.>

## Responsibilities
- ...

## Non-Responsibilities
- ...

## Owned Data

| Data / Table / Resource | Access | Notes |
|---|---|---|
| ... | read/write | ... |

## Interfaces

### Inbound APIs

| API | Contract | Caller | Notes |
|---|---|---|---|
| ... | ... | ... | ... |

### Outbound Calls

| Target | Protocol | Reason | Notes |
|---|---|---|---|
| ... | ... | ... | ... |

## Dependencies

| Dependency | Type | Reason |
|---|---|---|
| Postgres | DB | ... |
| Mongo | DB | ... |
| Redis | Cache/Queue | ... |
| Kafka | Event Bus | ... |

## Runtime Flows
- ...

## Failure Modes

| Failure | Expected Behavior |
|---|---|
| ... | ... |

## Security
- Authentication:
- Authorization:
- Permission:
- Tenant/workspace/store isolation:
- Sensitive data:

## Observability
- Logs:
- Metrics:
- Traces:
- Alerts:

## Config
- ...

## Agent Rules
- Do not put business logic in handlers.
- Do not bypass usecase/interactor layer.
- Do not access another component's owned data directly unless explicitly allowed.
- Do not change public contracts without PZEP update.
```
