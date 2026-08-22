---
name: clean-test-writer
description: Author independent unit, acceptance, integration, and UI/QA test tracks from requirements and public contracts, withholding implementation details where feasible so tests cannot mirror implementation mistakes. Use when tests must be decoupled from the code that implements the behavior.
---

# Test Writer

Build independent behavior test tracks that cannot mirror the implementation.

## Workflow

1. Take requirements, public contracts, and acceptance examples. Treat implementation internals as a bias to withhold, not a resource to copy.
2. Write tests that would fail against an implementation that violates the contract, not tests that encode how the implementation happens to work.
3. Cover independent tracks where relevant: unit, acceptance, integration, and UI/QA. Keep each track's oracle separate from the implementation.
4. If a separate context or session is not available, record that independence was procedural, not structural, so reviewers can weigh the correlation.
5. Record evidence locations so verification can trace each track to its command and result.
6. Hand the test plan and tracks to `clean-verify` and `clean-review`.

## Guardrails

- Do not mirror implementation details into tests; a test that asserts current behavior verbatim adds no oracle value.
- Do not weaken an existing failing test to make it pass; that hides the very defect the track exists to catch.
- If given implementation internals that would bias the oracle, record the correlation and mark the evidence accordingly instead of silently proceeding.
- Preserve the exact status vocabulary: PASS, FAIL, NOT_AVAILABLE, NOT_CONFIGURED, NOT_RUN, ERROR.

## Tool-free mode

- If test execution is unavailable, produce the test plan and the exact status for the unavailable capability. Never claim a test ran when it did not.
