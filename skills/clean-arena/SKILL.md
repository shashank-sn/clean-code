---
name: clean-arena
description: Run a bounded competing-design workflow with independently reasoned options, structured comparison, and a revision-bound decision record. Use for consequential design choices; do not invoke for routine fixes.
---

# Clean Arena

Compare two or more independently reasoned options, then synthesize a reviewable decision record. No automatic policy change.

## When it applies

- Consequential design or high-ambiguity feature/refactor choices.
- Explicit trigger from `clean-route` (`arena_warranted: true`) or the design-decision playbook.

## When it does not

- Routine bugfixes, typo edits, or already-settled local changes.

## Workflow

1. State the decision question, constraints, and owner.
2. Produce at least two candidate options in independent reasoning passes when the host allows isolation.
3. Record evidence, risks, and tradeoffs per candidate. If a candidate cannot run, mark `NOT_AVAILABLE` / `NOT_RUN` / `BLOCKED` with a reason and keep the partial record.
4. Synthesize separately: selected option, why rejected options lost, uncertainty, unresolved assumptions.
5. Validate with `clean-code arena validate --input <record.json>`.
6. Set `policy_change` to `none`. Hand policy proposals only to `clean-learn` under existing approval rules.

## Terminal states

`PASS`, `FAIL`, `NOT_RUN`, `NOT_AVAILABLE`, `BLOCKED`
