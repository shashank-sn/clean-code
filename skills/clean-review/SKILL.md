---
name: clean-review
description: Review code, requirements, architecture, tests, and deterministic evidence with concrete findings, structural simplification, and permission to return zero findings. Use after implementation and final verification, during pull-request review, when interpreting metric or tool output, or when checking correctness, dependency direction, maintainability, and operational risk.
---

# Clean Review

Return evidence-backed findings or correct silence.

## Workflow

1. Confirm the revision, changed scope, requirements, verification report, architecture report, and test trace all refer to the final change.
2. Review correctness and requirement conformance before naming or style.
3. Inspect structural simplification, dependency direction, responsibility placement, public boundaries, test strength, failure behavior, and operational risk.
4. Turn tool output into a finding only after establishing its concrete consequence in this change.
5. For every finding, record severity, location or behavior, evidence, consequence, confidence, bounded fix, and disposition.
6. Merge duplicates and resolve conflicts between reviewers using the underlying evidence.
7. Run `clean-code review --input <review.json>`. Preserve an empty findings array when the evidence supports approval.

## Structural-simplification lens

Use this lens when non-trivial changed code adds control flow, state, a module boundary, or an abstraction. Documentation-only changes do not need this lens. Ask whether the changed scope introduced complexity that a direct, behavior-preserving design can remove:

- Can a branch, wrapper, helper, mode, or layer disappear by using an existing model or canonical utility?
- Did special-case logic enter an unrelated shared path instead of the layer that owns the concept?
- Do repeated conditionals, optional values, casts, or ad-hoc shapes hide a clearer invariant or typed boundary?
- Did the diff duplicate a helper, make a cohesive module harder to scan, or create avoidable sequential or partial-update orchestration?
- Did a changed source file cross 1,000 lines? If so, ask whether a focused decomposition is clearer. This is a review prompt, never a universal size limit.

Do not require a redesign merely because another design is possible. A structural finding needs changed-scope evidence, a concrete maintenance or architecture consequence, and a bounded alternative that preserves the stated behavior. Classify it as `IMPROVEMENT` unless it also violates a requirement, safety property, or declared architecture rule. Do not turn an unconventional style, a metric alone, or an unproven preference into a finding.

When a valid finding warrants a code change, hand it to `clean-refactor` with the protected behavior, the smell and change cost, and the smallest reversible move. Re-verify and re-review the resulting revision.

## Severity

- `BLOCKING`: correctness, safety, requirement, required test, or declared architecture failure.
- `IMPROVEMENT`: bounded maintainability cost with a concrete consequence.
- `ADVISORY`: useful observation that requires no change.

## Guardrails

- Keep authorship and independent approval separate.
- Require reasons for dismissed findings and accepted risks.
- Keep unresolved blocking findings blocking; accepted risk cannot override them.
- Avoid findings based solely on taste, generic advice, a metric threshold, or the existence of unconventional code.
- Re-run final verification after fixes change the revision.
