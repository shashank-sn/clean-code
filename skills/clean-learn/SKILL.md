---
name: clean-learn
description: Turn confirmed defects, false positives, accepted changes, review outcomes, and benchmark results into reversible policy, threshold, adapter, convention, or calibration proposals. Use after an audit or repeated outcome reveals a useful local rule change; keep every proposal separate from execution and approval.
---

# Clean Learn

Propose narrow local improvements from confirmed outcomes.

## Workflow

1. Start from a confirmed outcome: escaped defect, caught defect, false positive, correct silence, repeated exception, or accepted local convention.
2. Link the source audit receipt, evidence hashes, affected rule or adapter, and observed consequence.
3. Propose one reversible change with scope, expected benefit, risks, migration, rollback, and a calibration fixture.
4. Compare the proposal with the currently trusted policy and flag every new command, permission, suppression, threshold change, and gate-status change.
5. Send the proposal to a separate reviewer. Apply it only after explicit approval.
6. Run existing and new calibration fixtures, then record measured detection and false-positive outcomes.
7. Run `clean-code learn --proposal <proposal.json>` before sending the proposal for approval.

## Hard boundaries

- Never let the proposing agent approve or activate its proposal.
- Never suppress correctness, safety, security, privacy, data-integrity, or explicit requirement failures.
- Never turn missing, unavailable, stale, or unrun evidence into success.
- Never generalize from one stylistic preference into a universal rule.
- Keep rollback data and provenance for every accepted policy change.
