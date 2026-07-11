# Definition of Ready

## Feature Ready

A feature is ready for implementation only when it has:

- clear business goal;
- defined scope and out of scope;
- actors or personas;
- user flow or use cases;
- functional requirements with stable IDs;
- non-functional requirements when relevant;
- business rules;
- acceptance criteria;
- UI/mockup spec when frontend is involved;
- API/DB/event/permission contract when integration is involved;
- PZEP when the change crosses components;
- component/data ownership identified;
- traceability updated.

## Agent Task Ready

An agent task is ready only when it has:

- requirement or recovery source;
- use case/workflow if behavior is user-facing;
- acceptance criteria;
- PZEP or ADR when applicable;
- target component/module;
- allowed files or folders;
- forbidden changes;
- validation commands;
- expected handoff format.

If any of these are missing, the agent must stop and report the missing docs
instead of coding.
