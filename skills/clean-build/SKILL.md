---
name: clean-build
description: Implement one bounded behavior through short, verified steps while protecting requirements, architecture policy, repository conventions, and unrelated code. Use when adding a feature, fixing a bug, changing behavior, or continuing an implementation from an approved design and test plan.
---

# Clean Build

Implement the smallest behavior that closes one verified requirement gap.

## Workflow

1. Name the requirement, acceptance example, affected component, and completion check.
2. Establish the current result with the cheapest relevant test or reproduction.
3. Add or adjust a behavior-level check before implementation when practical.
4. Change only the files needed for that behavior and keep dependencies within approved boundaries.
5. Run the focused check, then the affected integration and architecture checks.
6. Stop feature work when each new case forces coordinated edits in several places. Name the missing concept and refactor under green tests.
7. Hand off changed files, requirement IDs, checks run, evidence locations, remaining risks, and intentional exceptions. Leave completion to independent verification.

## Guardrails

- Preserve repository style and user-owned edits.
- Prefer direct names, small argument lists, explicit dependencies, positive conditions, and one abstraction level per function.
- Keep framework, transport, database, vendor, and UI details outside stable policy boundaries.
- Make temporary compatibility bridges explicit and remove the old path after all callers move.
- Record unavailable and unrun checks honestly.
