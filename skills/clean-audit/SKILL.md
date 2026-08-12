---
name: clean-audit
description: Produce an immutable revision-bound receipt by deriving required gates from approved policy and validating provenance. Use before merge, release, handoff, or any claim that a change is complete.
---

# Clean Audit

Preserve evidence, decisions, and gaps without letting evidence choose its own requirements.

## Required bindings

One receipt names the requirement contract digest, base revision, final revision, change-set digest, policy revision, evidence set, actor run identities, context lineage, and typed exceptions.

## Workflow

1. Load required gates and non-waivable invariants from approved policy.
2. Validate that requirements and changed paths have complete trace, test, and review coverage.
3. Require executed revision-bound evidence for each applicable gate; reject planned, unrun, unavailable, stale, or foreign-revision results.
4. Validate independent review by execution identity and controlled context, including correct-silence results.
5. Record human decisions only for intent, policy, risk, and typed exceptions.
6. Reject exceptions without approver, subject, rationale, expiry, and allowed scope; never waive correctness, safety, data integrity, explicit requirements, or declared architecture blockers.
7. Create a new immutable receipt and preserve every blocking gap.

## Integrity

Evidence cannot mark itself required or optional. Existing receipts are never overwritten. Any change to the revision, policy, requirement contract, change set, or evidence produces a new receipt.
