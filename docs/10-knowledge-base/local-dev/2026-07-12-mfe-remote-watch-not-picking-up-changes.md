# MFE remote stops picking up file changes, "Module X could not be loaded"

Date: 2026-07-12

## Symptom

After editing a file under `frontend/apps/onboarding` (or a shared file
under `frontend/packages/shared` consumed by a remote), the browser shows
the shell's `RemotePage` error boundary instead of the updated UI:

```text
Module onboarding could not be loaded. The service may be unavailable.
```

`remoteEntry.js` is still reachable (`curl` through APISIX and directly on
the container's own port both return `200`) — this is not the port-forward
issue from `2026-07-11-mfe-remoteentry-empty-response.md`. The remote
container is up and its `vite build --watch` / `vite preview` processes are
both alive (`docker exec <container> ps aux` shows both PIDs).

## Investigation

1. `docker logs <remote-container> --since 90m -t | grep -E "built in|transformed"`
   — showed the container's last rebuild of any kind was from over an hour
   earlier, well before the file edits in question. (A later pass, after
   the real fix below was in place, confirmed `✓ 6 modules transformed` is
   actually a normal, correct incremental rebuild for a small single-file
   edit — Vite/Rollup only reprocesses the affected module subgraph. Don't
   use a small module count alone as a staleness signal; the real tell is
   the total *absence* of any `built in ...` line after an edit, for a
   long time.)
2. Checked all three MFE remotes the same way — same pattern in each:
   `vite build --watch` had gone quiet after the container's first ~10
   minutes of uptime, despite the host repeatedly editing files under
   `apps/<remote>` and `packages/shared` in the time since.
3. Processes inside the container were not crashed (`ps aux` showed both
   the watch and preview processes still running) — the watcher process
   itself is alive, it simply isn't receiving filesystem change
   notifications for the bind-mounted source tree.

## Root Cause

`vite build --watch` relies on inotify to detect file changes. Docker
Desktop's bind-mount implementation on WSL2 does not reliably forward
inotify events from the Windows/WSL2 host filesystem into the Linux
container for every write — this is a known class of issue with bind
mounts on WSL2, not something fixable from inside the app's Vite config.
The watcher process stays alive and appears healthy; it just stops
reacting to new writes after some point, silently. There's no error to
grep for — the absence of a `built in ...` log line after an edit is the
only symptom.

Note this is a **different** root cause from the 2026-07-11 incident (that
one was a host↔container TCP port-forward desync on a long-running
container; this one is inotify events not propagating into an otherwise
perfectly reachable container). Same user-facing symptom class (MFE fails
to load), different fix.

## Fix

**Immediate (unblocks a stuck session):** restart the affected remote
container(s) — forces `dev:<remote>` to rerun `build:<remote>` fresh from
current disk contents:

```bash
docker restart podzone-onboarding-remote-dev podzone-iam-remote-dev podzone-backoffice-remote-dev
```

**Permanent:** switch the watcher to polling instead of relying on
inotify, via chokidar's standard environment variables — chokidar (used
under the hood by both Vite's dev-server watcher and Rollup's `--watch`
mode) reads these directly, no vite.config change needed:

```yaml
environment:
  CHOKIDAR_USEPOLLING: 'true'
  CHOKIDAR_INTERVAL: '300'
```

Added to `backoffice-remote`, `iam-remote`, `onboarding-remote`, and the
HOST `frontend` service in `deployments/docker/services.yml` (the HOST's
Vite dev-server HMR watcher is subject to the identical bind-mount gap,
just less visible since a full page navigation re-fetches source
regardless of whether HMR noticed the change). Requires
`docker compose up -d --build <service>` (recreate, not just `restart`) to
pick up new environment variables.

**Verified working:** edited a file, watched
`docker logs podzone-onboarding-remote-dev --since 30s -t` — the container
auto-rebuilt (`6 modules transformed`, `built in 25ms`) within seconds of
the edit landing on disk, with zero manual restart. Confirmed both for an
edit and for reverting it.

## Prevention

- If a code change doesn't show up in the browser after a refresh, and
  `CHOKIDAR_USEPOLLING` somehow isn't set (e.g. a new service added
  without copying this env block), check
  `docker logs <remote-container> --since <edit-time> -t | grep "built in"`
  — total absence of a build line since the edit means the watcher went
  stale, not that the code fix is wrong.
- Any new Vite-based dev container added to `deployments/docker/services.yml`
  (a new MFE remote, a new dev-mode service) should get the same
  `CHOKIDAR_USEPOLLING`/`CHOKIDAR_INTERVAL` env vars from the start —
  bind-mount inotify gaps on WSL2 are a property of the Docker Desktop
  setup, not of any specific service.
