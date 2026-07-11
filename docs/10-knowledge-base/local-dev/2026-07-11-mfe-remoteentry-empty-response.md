# MFE remoteEntry.js ERR_EMPTY_RESPONSE / 403 through APISIX

Date: 2026-07-11

## Symptom

Browser console, loading any MFE remote page (IAM, Onboarding, Backoffice):

```text
GET http://localhost:9080/mfe/onboarding/assets/remoteEntry.js net::ERR_EMPTY_RESPONSE
GET http://localhost:9080/mfe/iam/assets/remoteEntry.js net::ERR_EMPTY_RESPONSE
```

Stack trace through `__federation_method_ensure` / `lazyRouteComponent` /
`ErrorBoundary` — confirms it's a Module Federation remote-load failure,
not an application bug.

## Investigation

1. `curl http://localhost:3002/assets/remoteEntry.js` (remote's own Docker
   port, bypassing APISIX) → `200`, real content. Remote dev server itself
   healthy.
2. `curl http://localhost:9080/mfe/iam/assets/remoteEntry.js` (through
   APISIX) → `curl: (56) Recv failure: Connection reset by peer`. Isolated
   the failure to the APISIX hop.
3. `curl http://localhost:9080/` (APISIX root, not even a proxied route) →
   also connection reset. Ruled out a route-specific bug — **all** of
   APISIX's exposed ports were affected (9080, 9180 admin API, 9443).
4. `docker ps` showed `podzone-apisix` "Up 4 hours" vs. the MFE remote
   containers "Up 8 minutes" (recently restarted for unrelated work).
   Suspected a stale host↔container port-forward specific to the
   long-running container — a known Docker Desktop on WSL2 behavior.
5. Confirmed via `docker exec podzone-iam-remote-dev wget http://apisix:9080/...`
   (same Docker network, bypasses the host port-forward entirely) → got a
   real HTTP response (`403 Forbidden`), not a reset. This proved APISIX
   itself was healthy; only the **host-to-container port-forward** was
   broken.
6. The `403` from step 5 was a second, independent problem. Reproduced it
   directly: `curl http://localhost:3002/... -H "Host: apisix"` → `403`,
   vs. no `Host` override → `200`. Vite's `preview` server rejects
   requests whose `Host` header isn't on its allowlist; APISIX's
   `proxy-rewrite` forwards its own Host, not `localhost:<port>`.

## Root Cause

Two independent problems layered on top of each other:

1. **Docker Desktop/WSL2 port-forward desync.** The `podzone-apisix`
   container's host-exposed ports (9080/9180/9443) stopped forwarding
   correctly after ~4 hours of uptime while every other container's ports
   kept working. The container itself was healthy (`docker exec` traffic
   on the internal Docker network worked fine) — only the host→container
   TCP forward was stuck, returning a bare connection reset before nginx
   ever saw the request (confirmed: zero matching entries in APISIX's own
   access log for the failed requests).
2. **Vite preview host-header check.** All three MFE remotes'
   `vite.config.ts` `preview` blocks lacked `allowedHosts`. Vite 5.4+/6
   rejects any request whose `Host` header doesn't match an allowlist by
   default. APISIX's `proxy-rewrite` plugin forwards the original/gateway
   Host header (`apisix`), which Vite preview doesn't recognize, so it
   returned `403` instead of proxying through — this would have kept
   breaking MFE loads even after the port-forward issue was fixed.

## Fix

1. Port-forward: `docker restart podzone-apisix` (also restarted the 3
   remote containers in the same command so config changes below took
   effect immediately: `docker restart podzone-apisix
   podzone-iam-remote-dev podzone-onboarding-remote-dev
   podzone-backoffice-remote-dev`).
2. Vite host check: added `allowedHosts: true` to the `preview` block in
   `frontend/apps/{iam,onboarding,backoffice}/vite.config.ts`. Safe here —
   these are internal Docker-only dev services, not internet-exposed.

Verified: `curl http://localhost:9080/mfe/{iam,onboarding,backoffice}/assets/remoteEntry.js`
all return `200` with real content (2998/3642/4926 bytes respectively).

## Prevention

- If APISIX (or any long-running container) starts silently refusing
  connections on its host-mapped ports while `docker exec` traffic on the
  internal network still works, suspect the WSL2/Docker Desktop
  port-forward before debugging application/route config — restart the
  container first, it's a 10-second check.
- Any new MFE remote's `vite.config.ts` must include `preview.allowedHosts:
  true` from the start — it's exposed through APISIX with a rewritten Host
  header by design, so the default Vite host check will always reject it
  otherwise. Should be added to
  `docs/03-architecture-detail-design/14-mfe-federation-contract.md`'s
  "Rule For New Remotes" list.
