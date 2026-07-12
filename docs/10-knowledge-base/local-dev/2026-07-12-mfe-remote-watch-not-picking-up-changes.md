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
   — showed the container's last **real** rebuild (full module count, e.g.
   325 for onboarding) was from over an hour earlier, well before the file
   edits in question. One later entry showed `✓ 6 modules transformed` /
   `built in 226ms` — too small to be a genuine rebuild of a page-level
   change, and it didn't correspond to the actual edit made.
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

Restart the affected remote container(s) — forces `dev:<remote>` to rerun
`build:<remote>` fresh from current disk contents, then resume watching
(which will itself go stale again after some time, but starts fresh and
correct):

```bash
docker restart podzone-onboarding-remote-dev podzone-iam-remote-dev podzone-backoffice-remote-dev
```

Verify with the same log check — the module count in the post-restart
`built in ...` line should match a normal full build (compare against a
recent known-good count, or just confirm it's in the hundreds, not single
digits), and `curl http://localhost:9080/mfe/<remote>/assets/remoteEntry.js`
should return `200`.

## Prevention

- If a code change to a shared package or a remote's own files doesn't
  show up in the browser after a normal refresh, check
  `docker logs <remote-container> --since <edit-time> -t | grep "built in"`
  **before** assuming the code is wrong — a missing or suspiciously small
  (`6 modules`) rebuild line means the watcher silently went stale, not
  that the fix didn't work.
- No permanent fix identified this session (would need a different file
  watching strategy inside the container — e.g. polling — which has its
  own cost). Restarting the specific remote(s) touched is the fastest
  workaround; restarting all three is safe and avoids re-diagnosing which
  one went stale.
