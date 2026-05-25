# Gateway Bootstrap

## APISIX Seed Flow

```mermaid
sequenceDiagram
    participant Init as apisix-init
    participant Admin as APISIX Admin API
    participant GW as gRPC Gateway
    participant BO as Backoffice
    participant UI as Seller Portal UI

    Init->>Admin: PUT plugin_config podzone-edge-defaults
    Init->>Admin: PUT service podzone-grpcgateway
    Init->>Admin: PUT service podzone-backoffice-graphql
    Init->>Admin: PUT service podzone-ui
    Init->>Admin: PUT consumer podzone-dev-edge (jwt-auth)
    Init->>Admin: PUT route /api/*
    Init->>Admin: PUT route /query*
    Init->>Admin: PUT route /*
    Init->>Admin: PUT route /edge/protected/*
    Admin-->>GW: upstream /api/*
    Admin-->>BO: upstream /query*
    Admin-->>UI: upstream /*
```

## Ownership

- `internal/gateway/apisix_conf`
  - APISIX node/admin/etcd base config
- `deployments/docker/apisix-init/seed.sh`
  - local docker seed for services, routes, and example plugins
- `deployments/docker/services.yml`
  - one-shot `apisix-init` job

## Notes

- `jwt-auth` in APISIX is seeded as an edge-level example route and consumer.
- Application JWT validation still happens inside services today.
- When APISIX becomes the main production edge, move route/plugin seed into:
  - Kubernetes Job / Helm hook
  - Terraform + APISIX provider or Admin API bootstrap
  - per-environment declarative manifests
