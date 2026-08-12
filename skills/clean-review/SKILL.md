---
name: clean-review
description: Run revision-bound independent review with risk-driven agent lenses and permission to return correct silence. Use after implementation and final verification, during change review, or when checking correctness, requirements, architecture, tests, security, maintainability, and operational risk.
---

# Clean Review

Return validated findings or correct silence for one complete change set.

## Review request

Require the requirement contract, base and final revisions, changed-path manifest or diff digest, policy revision, verification results, implementation actor identity, reviewer run identity, and context lineage. Reject missing, stale, partial, or self-authored scope.

## Workflow

1. Select only lenses justified by the change: correctness, requirements, dependency direction, test strength, security, data integrity, concurrency, performance, or operations.
2. Give independent lenses requirements and public contracts while withholding implementation rationale where feasible.
3. Inspect all changed paths and trace each requirement to code and evidence.
4. Turn observations into findings only when a concrete consequence is established.
5. Record severity, location or behavior, evidence, consequence, confidence, bounded fix, and verification.
6. Merge duplicates, validate blockers, and resolve conflicts from primary evidence.
7. Return zero findings when complete scope and valid evidence support it.

## Guardrails

Keep authorship and approval separate by run identity and context lineage, not display name. Accepted risk cannot waive correctness, safety, data integrity, explicit requirements, or declared architecture blockers. Any fix creates a new revision and invalidates prior final approval.
