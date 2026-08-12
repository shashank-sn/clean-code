---
name: clean-test
description: Derive independent unit, acceptance, integration or contract, and UI/QA evidence from requirements and public behavior. Use when planning tests, implementing a feature or bug fix, reviewing test quality, creating Gherkin examples, writing user-interaction procedures, or checking whether implementation-shaped tests became the only oracle.
---

# Clean Test

Build independent evidence around requirements and public contracts.

## Tracks

1. Unit: local policy, calculations, boundaries, and error conditions.
2. Acceptance: executable examples written from requirement inputs and outcomes.
3. Integration or contract: real component interactions, callbacks, persistence, protocols, and failure recovery.
4. UI/QA: user actions, visible results, accessibility state, and supported device or viewport conditions.

## Workflow

1. Give acceptance and UI/QA authors the requirement and public contract while withholding implementation details when feasible.
2. Assign every requirement at least one acceptance example and every applicable track an owner and evidence location.
3. Include happy paths, empty and boundary inputs, error paths, permissions, retries, and cross-layer behavior where applicable.
4. Mark a track inapplicable only with a requirement-specific reason. Missing tools or time leave the track unrun.
5. Run `clean-code trace --plan <test-plan.json>` before claiming test planning is complete.
6. Use mutation testing to probe assertion sensitivity and coverage to locate unexecuted code. Keep their meanings separate.
7. Report correlated assumptions when implementation and tests were authored from the same internal structure.

## Guardrails

- Assert outcomes through public behavior instead of private implementation shape.
- Keep tests fast enough for their feedback tier and deterministic under repetition.
- Search around every discovered bug for nearby failing cases and uncovered branches.
- Preserve ignored examples when they record a real requirement ambiguity; assign an owner for resolution.
- Treat repeated race-test passes as weak evidence. Reduce shared state and isolate concurrency policy.
