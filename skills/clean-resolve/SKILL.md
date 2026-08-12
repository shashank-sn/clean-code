---
name: clean-resolve
description: Resolve validated review, automation, or stakeholder feedback against one revision without losing scope or reusing stale evidence. Use after review findings, failed checks, requested changes, or contradictory feedback.
---

# Clean Resolve

Turn feedback into dispositions, bounded changes, and fresh evidence.

## Workflow

1. Bind every feedback item to its source, revision, location, severity, and claimed consequence.
2. Validate the consequence before editing; mark duplicates and non-reproducible items with evidence.
3. Order accepted blockers before improvements and keep unrelated cleanup out.
4. Apply one bounded repair batch and record each item as fixed, rejected with reason, accepted risk, or needs authority.
5. Rerun focused tests and every gate invalidated by the new revision.
6. Return the new revision, dispositions, remaining blockers, and fresh evidence to an independent reviewer.

## Stop conditions

Stop when feedback changes requirements, asks to weaken correctness or safety, conflicts with approved policy, or needs an exception outside the resolving actor's authority.
