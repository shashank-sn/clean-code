---
name: clean-review
description: Review code, requirements, architecture, tests, and deterministic evidence with concrete findings and permission to return zero findings. Use after implementation and final verification, during pull-request review, when interpreting metric or tool output, or when checking correctness, dependency direction, maintainability, and operational risk.
---

# Clean Review

Return evidence-backed findings or correct silence.

## Workflow

1. Confirm the revision, changed scope, requirements, verification report, architecture report, and test trace all refer to the final change.
2. Review correctness and requirement conformance before naming or style.
3. Inspect dependency direction, responsibility placement, public boundaries, test strength, failure behavior, and operational risk.
4. Turn tool output into a finding only after establishing its concrete consequence in this change.
5. For every finding, record severity, location or behavior, evidence, consequence, confidence, bounded fix, and disposition.
6. Merge duplicates and resolve conflicts between reviewers using the underlying evidence.
7. Run `clean-code review --input <review.json>`. Preserve an empty findings array when the evidence supports approval.

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
