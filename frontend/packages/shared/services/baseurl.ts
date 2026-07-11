// used for admin side (grpc-gateway)
export const GW_API_URL = import.meta.env.VITE_PODZONE_GW_API_URL || 'http://localhost:8080'

// used for graphql side (Backoffice) — routed through APISIX
// (routes/1015 rewrites /backoffice/graphql -> /query on backoffice-service:8000),
// not called directly against the service's own exposed port.
export const TENANT_GQL_URL = import.meta.env.VITE_PODZONE_GRAPHQL_API_URL || 'http://localhost:9080/backoffice/graphql'

export const ONBOARDING_API_URL = import.meta.env.VITE_PODZONE_ONBOARDING_API_URL || 'http://localhost:8800'
