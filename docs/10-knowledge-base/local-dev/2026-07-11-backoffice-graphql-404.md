# Backoffice GraphQL POST /graphql 404

Date: 2026-07-11

## Symptom

Browser console, backoffice orders page (`CollectionControls` component):

```text
POST http://localhost:8000/graphql 404 (Not Found)
```

## Investigation

1. `curl -X POST http://localhost:8000/graphql` → `404`. `curl -X POST
   http://localhost:8000/query` → `200`, real GraphQL response. Backend
   serves at `/query`, confirmed against `deployments/docker/config/
   backoffice.yml` (`query_path: '/query'`) and the gin route registration
   in `internal/backoffice/graphql_wire.go` (`r.POST(p.Cfg.QueryPath,
   ...)`).
2. Traced the wrong `/graphql` path to `frontend/packages/shared/
   services/baseurl.ts`'s hardcoded fallback for `TENANT_GQL_URL`.
3. Checked whether the `VITE_PODZONE_GRAPHQL_API_URL` env var (which
   should override the wrong fallback) was actually reaching the build —
   `deployments/docker/services.yml`'s `backoffice-remote` service had
   **no `environment:` block at all**. The correct value was only set on
   the HOST `frontend` service. Same gap existed on `onboarding-remote`
   (which also reads `GW_API_URL`/`TENANT_GQL_URL` for its settings/home
   debug panels).
4. Also noticed `backoffice-remote` was still on the `serve:backoffice`
   script (build once, no watch) instead of the `dev:*` watch+preview
   pattern already fixed for iam-remote/onboarding-remote in an earlier
   incident — `dev:backoffice` didn't exist in `package.json`.

## Root Cause

Three compounding gaps, all from `backoffice-remote` not having received
the same fixes already applied to `iam-remote`/`onboarding-remote`:

1. `baseurl.ts`'s hardcoded fallback path was simply wrong (`/graphql`
   instead of `/query`) — nobody had hit this fallback in practice before
   because the HOST shell always had the correct env var.
2. `backoffice-remote` had no `environment:` block, so it always fell
   back to the wrong hardcoded default regardless of what
   `services.yml`'s HOST block said.
3. `backoffice-remote` was missing the watch+preview `dev:backoffice`
   script other remotes already had.

## Fix

Routed the GraphQL call through APISIX instead of hitting the backend's
exposed port directly, matching the existing `/mfe/<name>/*` convention
(see `13a60b9`):

- Added `routes/1015` to `deployments/docker/apisix-init/seed.sh`:
  `/backoffice/graphql*` → rewrite → `/query`, on the existing
  `podzone-backoffice-graphql` service (`service_id: 110`, already
  pointed at `backoffice-service:8000` but only exposed at the
  unnamespaced `/query*`).
- `baseurl.ts` fallback → `http://localhost:9080/backoffice/graphql`.
- Added the missing `environment:` block (with the corrected URL) to
  `backoffice-remote` and `onboarding-remote` in `services.yml`.
- Added `dev:backoffice` to `package.json`, switched `backoffice-remote`'s
  command to it.

Verified: `POST /backoffice/graphql` through APISIX (`:9080`) returns a
real GraphQL response (401 unauthenticated with no JWT — expected) instead
of 404. Confirmed working against the live UI.

## Prevention

When adding environment/config to one MFE remote (iam, onboarding,
backoffice), check the other two got the same treatment —
`iam-remote`/`onboarding-remote`/`backoffice-remote` in `services.yml`
should stay structurally parallel. This is the second incident in two
days where backoffice-remote was the odd one out after a fix landed on
the other two first (see
[2026-07-11-mfe-remoteentry-empty-response.md](./2026-07-11-mfe-remoteentry-empty-response.md)
for the first).
