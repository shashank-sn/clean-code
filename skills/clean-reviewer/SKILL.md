---
name: clean-reviewer
description: Review a diff against repo standards and the originating spec with evidence-backed findings, in an isolated context where the change author is not the approver. Use after final verification, during pull-request review, or any time an independent reviewer is required.
---

# Reviewer

Return evidence-backed findings or correct silence, decoupled from the change author.

## Workflow

1. Confirm the revision, changed scope, requirements, verification report, architecture report, and test trace all refer to the final change.
2. Run two internal passes, isolated from each other so neither pollutes the other:
   - **Standards**: does the change follow the repository's coding standards plus a Fowler-smell baseline?
   - **Spec**: does the change faithfully implement the originating issue or spec?
3. Review correctness and requirement conformance before naming or style.
4. Inspect dependency direction, responsibility placement, public boundaries, test strength, failure behavior, and operational risk.
5. Turn tool output into a finding only after establishing its concrete consequence in this change.
6. For every finding, record severity, location or behavior, evidence, consequence, confidence, bounded fix, and disposition.
7. Merge duplicates and resolve conflicts between passes using the underlying evidence.
8. Run `clean-code review --input <review.json>`. Preserve an empty findings array when the evidence supports approval.

## Severity

- `BLOCKING`: correctness, safety, requirement, required test, or declared architecture failure.
- `IMPROVEMENT`: bounded maintainability cost with a concrete consequence.
- `ADVISORY`: useful observation that requires no change.

## Guardrails

- Keep authorship and independent approval separate. Never approve a change you authored or in the same context that authored it.
- Require reasons for dismissed findings and accepted risks.
- Keep unresolved blocking findings blocking; accepted risk cannot override them.
- Avoid findings based solely on taste, generic advice, a metric threshold, or the existence of unconventional code.
- Re-run final verification after fixes change the revision.

## Tool-free mode

- If the reviewer cannot execute `clean-code review`, produce the findings artifact with the exact status that reflects the unavailable capability: `NOT_AVAILABLE`, `NOT_CONFIGURED`, `NOT_RUN`, or `ERROR`. Never claim a review action occurred that did not.
