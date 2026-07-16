export const environment = {
  production: true,
  apiBaseUrl: 'http://localhost:8080',
  onboardingApiUrl: 'http://localhost:8800',
  // The SolidJS host (`frontend/`) still serves the store-scoped Backoffice
  // at `/t/:tenantId?storeId=...` — frontend-v2 is not mounted into that
  // shell yet (PZEP-0004 Phase 2+), so "open Backoffice" is a full,
  // cross-app navigation to this origin, not an internal route.
  backofficeBaseUrl: 'http://localhost:3000',
};
