---
name: clean-orchestrate
description: Coordinate specification, implementation, independent testing, deterministic verification, review, human spot checks, and audit without treating agent narration as proof. Use for multi-part coding work, release qualification, agent collaboration, or any change where correlated mistakes and stale evidence need explicit controls.
---

# Clean Orchestrate

Assign clear ownership and make every completion claim traceable to evidence.

## Workflow

1. Give specification ownership to the requirement source and record stable requirement IDs.
2. Give implementation a bounded requirement, architecture policy, and repository context.
3. Give acceptance and UI/QA authors requirements and public contracts while withholding implementation details where feasible.
4. Run deterministic verification through the final integrating session against the final revision.
5. Give review the diff, requirements, and evidence. Keep the change author separate from approval.
6. Request human spot checks for configured requirement, acceptance, UI/QA, and code-sample boundaries.
7. Reconcile contradictions from source evidence, rerun stale checks, and hand the complete evidence set to audit.

For end-to-end delivery to an open PR, prefer `clean-lfg`, which sequences brainstorm, plan, build, test, verify, review, simplify, ship, watch, audit, and compound. See `docs/shipping-pipeline.md`.

## Host differences

- With subagents, assign explicit file and responsibility ownership and keep independent contexts narrow.
- Without subagents, use separate sessions or invocations for implementation, acceptance, and review, then record that independence was procedural.
- Across IDEs, coding platforms, terminals, and CI, route deterministic work through the same CLI contracts.

## Stop conditions

- Stop when mandatory evidence is missing, stale, or belongs to another revision.
- Stop when an independent role received implementation context that could bias its oracle; record the correlation and replace or supplement that evidence.
- Stop when required human checks remain unperformed.
- Stop when reviewers disagree on a blocking issue until the evidence resolves it.
