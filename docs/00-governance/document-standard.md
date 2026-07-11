# Document Standard

## Requirement Documents

Requirements describe **what** the system must do and why it matters. They must
avoid implementation details unless the detail is a real constraint.

Each feature requirement should include:

- summary;
- actors;
- scope and out of scope;
- workflows or use cases;
- business rules;
- functional requirements;
- non-functional requirements when relevant;
- acceptance criteria;
- UI spec links when frontend is involved;
- contract links when integration is involved;
- traceability.

## PZEP Documents

PZEP describes **how** an approved requirement will be delivered at feature
level.

Minimum sections:

- status;
- requirement sources;
- summary;
- problem;
- goals and non-goals;
- proposed solution;
- affected components;
- runtime flow;
- API/DB/event/permission changes;
- data ownership;
- security and observability;
- alternatives considered;
- test plan;
- agent implementation plan;
- acceptance criteria mapping;
- open questions.

## Component Specs

Component specs describe a service/module boundary. They must include:

- purpose;
- responsibilities;
- non-responsibilities;
- owned data;
- inbound APIs;
- outbound calls;
- dependencies;
- runtime flows;
- failure modes;
- security;
- observability;
- config;
- agent rules.

## UI Specs

UI specs translate Figma/mockups into implementation behavior:

- route;
- purpose;
- permissions;
- components;
- fields and validation;
- actions and API mapping;
- loading, empty, submitting, success, validation error, API error, forbidden;
- related requirements and acceptance criteria.

Agents must not infer behavior directly from Figma without a Markdown UI spec.
