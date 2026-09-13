---
name: clean-orchestrate
description: Coordinate specification, implementation, independent testing, deterministic verification, review, human spot checks, and audit without treating agent narration as proof. Use for multi-part coding work, release qualification, agent collaboration, or any change where correlated mistakes and stale evidence need explicit controls.
---

# Clean Orchestrate

Assign clear ownership and make every completion claim traceable to evidence.

## Workflow

1. Optionally invoke `clean-route` on task signals to select playbook, roles, arena/probe warrants, and parallelism. Treat the route as advisory unless a lifecycle contract authorizes the next step.
2. Give specification ownership to the requirement source and record stable requirement IDs.
3. Give implementation a bounded requirement, architecture policy, and repository context.
4. Give acceptance and UI/QA authors requirements and public contracts while withholding implementation details where feasible.
5. When the route allows parallelism, spawn children only with owned scope, an explicit evidence request, and a mutation boundary. Serialize shared-state changes or use `clean-worktree`. Record why parallelism, the concurrency cap, and every unavailable/failed child in a parallel orchestration record (`clean-code parallel validate`).
6. For consequential design choices, run `clean-arena` before locking the approach.
7. Run deterministic verification through the final integrating session against the final revision. For medium/high risk, run `clean-probe` before claiming success.
8. Give review the diff, requirements, and evidence. Keep the change author separate from approval.
9. Request human spot checks for configured requirement, acceptance, UI/QA, and code-sample boundaries. Preserve human approval for releases, policy, permissions, and external actions.
10. Reconcile contradictions from source evidence, rerun stale checks, and hand the complete evidence set to audit. Do not treat a child report as proof without checking claimed artifacts.
11. After audit, route confirmed repeated outcomes to `clean-eval-discover` when evaluation discovery is warranted. Keep it conditional, preserve a held-out set, and send only calibrated candidates to `clean-learn` for separate approval.

For end-to-end delivery to an open PR, prefer `clean-lfg`, which sequences brainstorm, plan, build, test, verify, review, simplify, ship, watch, audit, conditional evaluation discovery, learn, and compound. See `docs/shipping-pipeline.md`.

## Host differences

- With subagents, assign explicit file and responsibility ownership and keep independent contexts narrow.
- Without subagents, use separate sessions or invocations for implementation, acceptance, and review, then record that independence was procedural.
- Across IDEs, coding platforms, terminals, and CI, route deterministic work through the same CLI contracts.

## Stop conditions

- Stop when mandatory evidence is missing, stale, or belongs to another revision.
- Stop when an independent role received implementation context that could bias its oracle; record the correlation and replace or supplement that evidence.
- Stop when required human checks remain unperformed.
- Stop when reviewers disagree on a blocking issue until the evidence resolves it.
