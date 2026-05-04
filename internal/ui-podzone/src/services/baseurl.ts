// used for admin side (grpc-gateway)
export const GW_API_URL =
  import.meta.env.VITE_PODZONE_GW_API_URL || 'http://localhost:8080';

// used for graphql side (Backoffice)
export const TENANT_GQL_URL =
  import.meta.env.VITE_PODZONE_GRAPHQL_API_URL ||
  'http://localhost:8000/graphql';

export const BACKOFFICE_API_URL =
  import.meta.env.VITE_PODZONE_BACKOFFICE_API_URL || 'http://localhost:8000';
