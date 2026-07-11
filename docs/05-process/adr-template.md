# ADR Template

ADR means **Architecture Decision Record**. Use it for a decision about
service boundaries, data ownership, cross-service communication, dependency
direction, or a technology/pattern choice that is expensive to reverse.

Do not create an ADR for implementation detail that stays inside one
component's boundary — that belongs in a component spec
(`docs/05-process/component-spec-template.md`) or inline code comments.

ADRs live in `docs/08-adr/`, filename `ADR-NNNN-short-slug.md`.

```markdown
# ADR-NNNN: <Decision Title>

## Status
Proposed | Accepted | Superseded by ADR-NNNN | Rejected

## Date
<YYYY-MM-DD>

## Related Commit
<commit hash or PR link that implements this decision, if already implemented>

## Context
<What forced this decision. What breaks or stays fragile without it.>

## Decision
<The decision, stated as one clear sentence, then elaborated.>

## Alternatives Considered

### Option A: <name>
Pros:
- ...

Cons:
- ...

### Option B: <name>
Pros:
- ...

Cons:
- ...

## Consequences
<What becomes easier. What becomes harder. What future work this
constrains or unblocks.>

## Rule Of Thumb
<A one-line rule a future agent can apply without re-reading the full
reasoning, if one exists.>
```

Rules:

- No ADR, no architecture boundary change (see
  `docs/00-governance/traceability-rule.md`).
- Backfilling an ADR for a decision already made in code is acceptable —
  set `Related Commit` to the commit that made the change, and prefer this
  over leaving a real decision undocumented.
- Link the ADR from every architecture doc it affects.
