# Podzone Documentation Index

This folder is the canonical engineering documentation for Podzone. Do not use
parallel draft folders as source of truth. Agents and humans should start here
and follow the links below.

## Canonical Flow

```text
00 Governance
  -> 00 Project vision
  -> 01 SRS
  -> 02 Architecture overall
  -> 03 Architecture detail design
  -> 04 Sprint slice
  -> 05 Agent task/process
  -> 06 Recovery evidence
  -> 08 ADR (when an architecture boundary decision is made)
  -> 09 PZEP (when a cross-component feature is proposed)
```

Coding starts only after the target work links back to this chain.

Read [STATUS_CURRENT.md](./STATUS_CURRENT.md) first — it is the living
snapshot of recovery phase, backbone flow status, and known doc debt.

## Start Here

0. [Governance](./00-governance/README.md)
   Agent working rules, ready/done gates, traceability, naming, and review.
1. [Project Vision](./00-project-vision/README.md)
   Product vision, actors, business flows, domain map, and BA requirements.
2. [SRS](./01-srs/README.md)
   System requirements, MVP backbone, and traceability matrix.
3. [Architecture Overall](./02-architecture-overall/README.md)
   C4, system context, containers, data ownership, and runtime sequences.
4. [Architecture Detail Design](./03-architecture-detail-design/README.md)
   Module design, transport contracts, IAM, DDD, frontend, deployment, and OpenAPI.
5. [Sprints](./04-sprints/README.md)
   Small delivery slices that agents can implement.
6. [Process](./05-process/README.md)
   SDLC rules, spec-first vertical slices, task templates, and review checklist.
7. [Recovery](./06-recovery/README.md)
   Current stabilization plan for the unstable agent-expanded codebase.
8. [Problems](./07-problems/)
   Dated issue notes and recovery evidence.
9. [ADR](./08-adr/README.md)
   Architecture Decision Records — service boundaries, data ownership,
   cross-service communication, technology/pattern choices.
10. [PZEP](./09-pzep/README.md)
    Podzone Enhancement Proposals — feature-level design for
    cross-component changes.
11. [Knowledge Base](./10-knowledge-base/README.md)
    Concise runtime/infra incident notes from live debugging — search here
    before re-debugging a symptom that looks familiar.

## Current Recovery Target

The first flow to stabilize is:

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call one protected business API
```

Primary docs:

- [SRS baseline](./01-srs/podzone-srs.md)
- [Traceability matrix](./01-srs/traceability-matrix.md)
- [Recovery plan](./06-recovery/recovery-plan.md)
- [Backbone flow refactor](./06-recovery/backbone-flow-refactor.md)
- [C4 architecture](./02-architecture-overall/01-c4.md)
- [Sprint 0](./04-sprints/sprint-00-foundation.md)

## Agent Rule

An agent task is not ready unless it names:

- governance rule: [Agent Working Rule](./00-governance/agent-working-rule.md);
- SRS requirement ID;
- requirement or recovery doc;
- architecture/C4 doc when relevant;
- API/DB contract when relevant;
- sprint slice or explicit task file;
- allowed files or module boundaries;
- acceptance criteria and verification command.
