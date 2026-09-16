---
name: clean-simplify
description: Simplify recently changed code for clarity and reuse while preserving behavior and structure pins from the plan. Use after implementation before pruning and review.
---

# Clean Simplify

Reduce complexity in the change set without altering verified behavior.

## Workflow

1. Scope to the branch diff or named files from the active plan unit.
2. Identify duplication and naming drift. Prefer consolidating patterns already used in the repo.
3. Preserve `session-settled` or plan Key Decision structure pins — deliberate duplication stays when the plan requires it.
4. Run affected tests and `clean-verify` when configured before finishing.
5. Hand dead, unreferenced, and leftover code to `clean-prune` instead of deciding it here. Simplification keeps behavior; pruning deletes what nothing uses.
6. Leave changes unstaged for `clean-ship` unless the orchestrator owns commits.

## Safety

- No behavior changes disguised as cleanup. Record intentional behavior changes separately.
- Skip when the change is docs-only or trivial (under ~10 substantive code lines).
