# 10 Knowledge Base

Parent index: [Podzone Documentation Index](../README.md).

Concise incident notes from live debugging — runtime/infra issues found by
running the stack, not by code review (those go in `docs/07-problems/`).
Split by environment, since the same symptom can have a different cause
depending on where it happens (a Docker/WSL2 port-forward quirk only
exists in local dev; a staging/prod incident is never that).

One file per incident, inside the environment folder it happened in. Keep
entries short: symptom, what was checked, root cause, fix. Skip sections
that don't apply rather than padding them.

Search the relevant environment folder before debugging a symptom that
looks familiar — infra quirks tend to repeat within an environment, not
across them.

## Environments

| Environment | Folder | Notes |
|---|---|---|
| Local dev | [`local-dev/`](./local-dev/README.md) | Docker Compose on a developer machine (WSL2, Docker Desktop). Most entries live here today. |
| Staging | [`staging/`](./staging/README.md) | No incidents recorded yet — staging isn't running. |
| Production | [`prod/`](./prod/README.md) | No incidents recorded yet — prod doesn't exist yet. |

## Entry Template

```markdown
# <Short Title>

Date: YYYY-MM-DD

## Symptom

<Exact error message/log line. Copy-paste, don't paraphrase.>

## Investigation

<Ordered list of what was checked and what each check showed. Only the
steps that actually narrowed the cause — not a transcript of every
command run.>

## Root Cause

<One or two sentences. Name the actual mechanism, not just "it was broken".>

## Fix

<Exact commands/config change. If more than one distinct problem was
found, number them.>

## Prevention

<Optional — only if there's a concrete guard to add (a doc rule, a config
default, a check). Skip if the fix is a one-off.>
```
