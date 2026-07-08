# API Contract Template

Use this before backend and frontend agents implement against an API.

````md
# <Feature> API Contract

## Owner

- Service:
- Bounded context:
- Transport: REST | gRPC | GraphQL | event

## Operation

- Name:
- Method/RPC/field:
- Path/topic:
- Permission:
- Tenant/workspace/store scope:

## Request

```json
{}
````

## Response

```json
{}
```

## Errors

| Condition | Code/status | Message detail | UI behavior |
| --- | --- | --- | --- |
| Validation |  |  |  |
| Permission denied |  | missing permission/resource | stay on current screen |
| Not found |  |  |  |
| Conflict |  |  |  |
| Internal |  |  |  |

## Collection Contract

If this is a list:

- Page:
- Page size:
- Search:
- Filters:
- Sort fields:
- Sort direction:
- Page info:

## Compatibility

- Backward compatible: yes | no
- Migration required:
- Generated code required:

## Tests

- Handler/resolver:
- Usecase:
- Contract:
```
