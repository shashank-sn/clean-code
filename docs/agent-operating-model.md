# Agent operating model

Clean Code treats software delivery as a graph of claims, evidence, and authority. A model response is never completion evidence by itself.

## Role boundaries

| Role | Owns | Cannot approve |
| --- | --- | --- |
| Product authority | intent, policy, risk, exceptions | code correctness evidence |
| Specification | requirements and acceptance examples | its own implementation |
| Implementation | bounded code changes | final review or policy weakening |
| Acceptance | public-behavior test oracles | implementation-shaped assumptions |
| Verification | deterministic execution | whether a gate is required |
| Review | risk-driven findings or correct silence | changes from the same run identity |
| Delivery | integration and destination evidence | implementation or exception approval |

Independence means a separate execution identity plus controlled context lineage. A different display name or model is insufficient. A host without isolated agents can use separate sessions, but same-context fallback remains correlated and is never labeled independent.

## Portable lifecycle

The canonical suite has sixteen responsibilities:

- setup and discovery;
- planning, design, build, refactor, and debug;
- test planning and execution, deterministic verification, and independent review;
- feedback resolution, orchestration, and audit;
- delivery, handoff, and learning.

Host adapters implement filesystem, source-control, automation, approval, and tool invocation mechanics. They cannot change status meanings, evidence requirements, dependency rules, or approval boundaries.

## Human review

Humans approve four things: product intent, policy, release risk, and typed exceptions. They may read code, but release completeness does not require it. Independent agents inspect the complete revision and return evidence-backed findings or correct silence. Deterministic checks and audit bind every accepted claim to the same requirement contract, change-set digest, final revision, policy revision, and evidence set.

## Completion rule

A claim is complete only when required evidence is executed, current, provenance-bound, and approved by the right authority. PLANNED, NOT_RUN, NOT_AVAILABLE, INAPPLICABLE, FAIL, and PASS remain distinct. Missing capability changes the procedure, not the definition of success.
