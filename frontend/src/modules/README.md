# UI Module Boundaries

`src/modules` owns product-facing frontend boundaries. Keep route pages and feature-local panels here.

- `shell`: authentication, bootstrap, and app entry flows.
- `onboarding`: workspace and store selection/provisioning surfaces.
- `iam`: centralized IAM console surfaces.
- `backoffice`: tenant/store-scoped operating workspace.

Shared UI primitives stay in `src/solid/components`. Shared transport clients and browser storage helpers stay in `src/services`.
Modules may depend on shared code, but shared code must not import modules.
