# 10 Knowledge Base

Parent index: [Podzone Documentation Index](../README.md).

Concise incident notes from live debugging — runtime/infra issues found by
running the stack, not by code review (those go in `docs/07-problems/`).
One file per incident. Keep entries short: symptom, what was checked, root
cause, fix. Skip sections that don't apply rather than padding them.

Search this folder before debugging a symptom that looks familiar —
Docker/WSL2 networking quirks and MFE gateway issues tend to repeat.

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

## Index

| Date | Title | Area |
|---|---|---|
| 2026-07-11 | [MFE remoteEntry.js ERR_EMPTY_RESPONSE / 403 through APISIX](./2026-07-11-mfe-remoteentry-empty-response.md) | Docker/WSL2, APISIX, Vite preview |
