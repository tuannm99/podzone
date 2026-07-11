# 05 Process Docs

This folder defines how Podzone work should move from product intent to
verified code.

Parent index: [Podzone Documentation Index](../README.md).

Use this folder after a requirement or recovery slice exists. Process docs
explain how to turn that slice into agent-safe work.

Start here:

1. `sdlc-operating-model.md`
2. `spec-first-vertical-slice.md`
3. `feature-spec-template.md`
4. `ui-state-spec-template.md`
5. `pzep-template.md`
6. `component-spec-template.md`
7. `api-contract-template.md`
8. `db-contract-template.md`
9. `vertical-slice-breakdown-template.md`
10. `agent-task-template.md`
11. `review-checklist.md`

Related handoff docs:

- `ai-agent-sdlc.md`
- `../00-governance/agent-working-rule.md`
- `../03-architecture-detail-design/16-agent-onboarding.md`
- `../../CLAUDE.md`

## Ready-To-Code Chain

```text
Business need
  -> SRS ID
  -> requirement/recovery doc
  -> acceptance criteria
  -> PZEP when cross-component
  -> architecture or contract doc
  -> sprint slice
  -> agent task template
  -> review checklist
```

Review and discovery tasks may run before this chain is complete, but they must
not change implementation files. Their output should identify the missing
documents required to make the next coding task ready.
