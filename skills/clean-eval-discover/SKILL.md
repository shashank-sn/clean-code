---
name: clean-eval-discover
description: Turn real work outcomes and bounded human judgment into representative, reviewable evaluation candidates. Use when escaped defects, rejected outputs, human edits, review outcomes, or correct silence may reveal a repeatable quality signal; discover and calibrate candidates without approving, activating, or promoting a rule.
---

# Clean Eval Discover

Discover narrow evaluation candidates from real outcomes without mistaking a one-off preference for durable policy.

## Role separation

- The data owner supplies source material, provenance, access constraints, and redaction decisions.
- Independent judges label blinded cases against the stated criterion, with `pass`, `fail`, or `abstain` and concise evidence.
- A human approver accepts, rejects, or requests revision of a candidate. The discoverer and judges cannot approve it.
- A separate promoter may propose activation only through `clean-learn` after held-out calibration evidence exists. This skill never promotes or activates an evaluation.

When one host or person must cover multiple roles, record the lost independence and treat the result as weaker evidence.

## Workflow

1. Collect bounded source outcomes: traces, diffs, review decisions, human edits, rejected outputs, escaped defects, and clean controls. Preserve revision or artifact identity, provenance, permissions, and any redaction decision.
2. Classify each item as a confirmed defect, false positive, false negative, correct silence, accepted outcome, ambiguous case, or unavailable evidence. Do not infer a label from an outcome without recorded human rationale.
3. Build a representative, blinded review set where feasible. Include contrasting clean controls, boundary cases, and intentionally ambiguous cases; keep a held-out calibration set untouched.
4. Ask narrow judges to label one criterion at a time using `pass`, `fail`, or `abstain`, evidence, confidence, and disagreement. Keep implementation authors and promotion decision-makers out of the judging role when possible.
5. Cluster repeated, evidence-backed patterns. For each cluster, draft a candidate evaluation with its observable behavior, scope, source cases, counterexamples, expected false-positive cost, and rollback or retirement condition.
6. Report candidates, disagreement, selection bias, missing evidence, and clean controls to a human approver. Handoff only reviewable candidates and the untouched held-out set to `clean-learn` and the relevant test or review owner.

## Candidate quality

- Separate top-down requirement checks, bottom-up learned candidates, and task-local checklists. A task-local preference is not durable policy.
- Require more than one supporting example before proposing a general candidate, unless an explicit safety or requirement contract already requires the check.
- Keep observations, human judgments, inferences, and unknowns distinct.
- Prefer a small criterion with a measurable outcome over a broad quality score.
- Preserve rejected candidates and correct silences as calibration evidence; do not optimize only for detection.
- Treat judge disagreement and abstention as findings about criterion ambiguity, not as labels to discard.

## Hard boundaries

- Never approve, activate, promote, or silently add an evaluation, gate, threshold, or policy change.
- Never use the held-out calibration set to author or tune a candidate.
- Never turn missing, unavailable, stale, redacted, or unrun evidence into a positive label or success claim.
- Never mix private source material into a portable fixture without the data owner's recorded permission and redaction decision.
- Never suppress correctness, safety, security, privacy, data-integrity, or explicit requirement failures.

## Handoff

Provide: the source inventory, blinded review set, label rationale and disagreement record, candidate evaluations with counterexamples, held-out-set identity, provenance, and explicit unresolved gaps. Use `clean-learn` for any later reversible policy proposal; it still requires separate review and explicit approval.
