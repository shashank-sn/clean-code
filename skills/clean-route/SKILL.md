---
name: clean-route
description: Select a bounded adaptive workflow shape from task signals without choosing models by brand or writing host config. Use at the start of consequential engineering work when change type, risk, ambiguity, boundaries, evidence needs, or host capabilities should drive playbook selection.
---

# Clean Route

Produce an inspectable, deterministic route decision. The decision is advisory unless an existing lifecycle contract authorizes execution.

## Workflow

1. Collect task signals: change type, risk, ambiguity, affected boundaries, evidence needs, host capabilities, concurrency budget.
2. Run `clean-code route --input <signals.json>` (or equivalent in-process `adaptive.Route`).
3. Read the playbook, selected roles, required evidence, arena/probe warrants, parallelism allowance, reasons, and fallback.
4. Prefer `NOT_AVAILABLE`, `NOT_CONFIGURED`, or `NOT_RUN` when a capability is missing. Do not invent checks or model brands.
5. Hand off to the named roles. Do not treat the route itself as proof that work ran.

## Guardrails

- Never select a model by brand name.
- Never write host or global configuration.
- Keep authorization `advisory` unless the caller already has a lifecycle authorization for the next step.
- Safe fallback is the bug-investigation shape when the preferred playbook cannot run.

## Terminal states

`PASS`, `FAIL`, `NOT_RUN`, `NOT_AVAILABLE`, `BLOCKED`
