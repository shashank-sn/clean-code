---
name: clean-test
description: Plan and validate independent unit, acceptance, integration or contract, and UI evidence from requirements and public behavior. Use when planning tests, implementing a feature or bug fix, reviewing test quality, or qualifying a revision for release.
---

# Clean Test

Keep test planning and executed release evidence as separate claims.

## Planning mode

Map every requirement to acceptance examples and applicable tracks. Each track records owner, stable boundary, planned procedure, expected outcome, and reason for any INAPPLICABLE status. A valid plan may contain PLANNED; it never claims execution.

## Execution mode

Bind every result to requirement IDs, final revision, test artifact digest, command or procedure provenance, actor run identity, start and finish time, and outcome. PLANNED, NOT_RUN, NOT_AVAILABLE, INAPPLICABLE, FAIL, and PASS remain distinct. Required release evidence accepts only executed PASS, or policy-approved INAPPLICABLE.

## Tracks

Use unit for local policy, acceptance for requirement outcomes, integration or contract for real boundaries, and UI for user-visible behavior. Assert public outcomes rather than private implementation shape. Record correlated context when implementation and tests share assumptions.

## Guardrails

Coverage locates execution and mutation probes assertion sensitivity; neither proves correctness. Missing tools or time are NOT_AVAILABLE or NOT_RUN, never INAPPLICABLE. Any revision change makes prior final execution stale.
