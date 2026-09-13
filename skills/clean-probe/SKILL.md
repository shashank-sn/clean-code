---
name: clean-probe
description: Plan and run adversarial acceptance probes after implementation for medium/high-risk work before claiming success. Distinguishes primary acceptance from post-hoc diagnostics and feeds failures into a bounded repair/reverify loop.
---

# Clean Probe

Adversarial acceptance probes catch requirement-boundary failures that happy-path suites miss.

## When it applies

- Medium or high residual risk after implementation.
- Route decision with `adversarial_probes_warranted: true`.
- Release readiness for medium/high-risk revisions.

## When it does not

- Low-risk typo or docs-only changes.
- Replacing human spot checks, audit receipts, or primary verification.

## Workflow

1. Derive probes from the accepted requirement contract, boundaries, malformed inputs, idempotence, dates/time zones, authorization, and regression history.
2. Mark each probe as `primary_acceptance` or `post_hoc_diagnostic`. Never silently alter primary scores after the fact.
3. Prefer executable local evidence. Label unavailable or human-judgment probes explicitly.
4. Always include calendar validity when dates matter: `clean-code probe calendar --date YYYY-MM-DD` must reject impossible values such as `2026-02-30`.
5. Validate the plan with `clean-code probe validate --input <plan.json>`.
6. On failure, run a bounded repair and re-verify loop. Do not auto-learn or mutate policy.

## Terminal states

`PASS`, `FAIL`, `NOT_RUN`, `NOT_AVAILABLE`, `BLOCKED`
