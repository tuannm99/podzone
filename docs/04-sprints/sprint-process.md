# Sprint Process

Podzone sprints should be vertical-slice based and AI-agent ready.

## Sprint Inputs

Each sprint starts from:

- SRS requirements;
- current status;
- architecture/C4 changes needed;
- design/UI specs;
- API and DB contracts;
- known production blockers.

## Sprint Design Steps

1. Choose one product or platform backbone flow.
2. List SRS requirement IDs covered by the sprint.
3. Identify required design states and contracts.
4. Identify architecture decisions or C4 updates.
5. Break the sprint into vertical slices.
6. Convert slices into agent tasks.
7. Define verification gates.

## Sprint Task Shape

Each task must include:

- SRS ID;
- goal;
- allowed files or module boundary;
- out of scope;
- contracts affected;
- implementation notes;
- tests;
- verification;
- owner agent role.

## Sprint Review

Before closing a sprint:

- update `STATUS_CURENT.md`;
- update `docs/01-srs/traceability-matrix.md`;
- update related architecture docs;
- record known gaps;
- commit with themed messages.

## Agent Assignment Rules

- Backend tasks go to Backend Agent.
- Frontend tasks go to Frontend Agent.
- Contract and architecture tasks go to Architect Agent.
- Risky cross-service changes require Reviewer Agent before commit.
