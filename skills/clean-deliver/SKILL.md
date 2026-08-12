---
name: clean-deliver
description: Integrate, publish, and watch a verified revision through host adapters while preserving approval and evidence boundaries. Use when preparing a merge, release, deployment, package, or other externally visible delivery.
---

# Clean Deliver

Move one approved revision to its destination and prove what happened.

## Workflow

1. Confirm the delivery target, exact revision, approved requirement contract, resolved blockers, and immutable audit receipt.
2. Ask the host adapter to perform the authorized integration operation without changing core semantics.
3. Require all destination checks to finish on the delivered revision; do not substitute earlier local evidence.
4. Verify artifact identity, checksums, provenance, and declared platform set where applicable.
5. Watch automation and feedback until success, actionable failure, or an authority boundary is reached.
6. Record the destination identity, final state, artifacts, failures, and rollback or recovery path.

## Guardrails

The delivery actor cannot approve its own implementation or policy exception. Never publish from a dirty, ambiguous, or unverified revision. A partial artifact set is a failed delivery unless approved policy says otherwise.
