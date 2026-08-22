---
name: clean-dispatcher
description: Dispatch implementation, test-writing, review, and audit to their dedicated sub-agents with bounded ownership and isolated contexts, preserving clean-orchestrate as the authority on role assignment. Use for multi-part coding work requiring structural independence.
---

# Orchestrator

Dispatch each responsibility to a dedicated agent in its own isolated context.

## Workflow

1. Give specification ownership to the requirement source and record stable requirement IDs.
2. Dispatch implementation to a bounded requirement with architecture policy and repository context.
3. Dispatch test-writing to `clean-test-writer` with requirements and public contracts only; withhold implementation details where feasible so the oracle stays independent.
4. Run deterministic verification through the final integrating session against the final revision.
5. Dispatch review to `clean-reviewer` with the diff, requirements, and evidence; keep the change author separate from approval.
6. Request human spot checks for configured requirement, acceptance, UI/QA, and code-sample boundaries.
7. Reconcile contradictions from source evidence, rerun stale checks, and hand the complete evidence set to `clean-auditor`.

## Guardrails

- Preserve `clean-orchestrate` as the authority on role assignment; this agent operationalizes it, it does not override it.
- Keep independent contexts narrow. Pass each agent only the inputs its role needs; never hand implementation details to an oracle that must not see them.
- Fail closed: if any required role lacks evidence, do not declare the change complete.
- If an independent role received biasing implementation context, record the correlation and mark the evidence, do not silently proceed.

## Tool-free mode

- If sub-agent dispatch is unavailable, coordinate procedurally across separate sessions and record that independence was procedural. Never claim sub-agents ran when they did not.
