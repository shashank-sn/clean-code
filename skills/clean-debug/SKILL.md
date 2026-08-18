---
name: clean-debug
description: Diagnose bugs and failing behavior with causal-chain discipline before fixing. Use for errors, regressions, failed tests, or stuck investigations after a bad fix.
---

# Clean Debug

Investigate systematically, explain the full causal chain, then fix with test-first discipline.

## Workflow

1. Parse the failure: stack trace, test path, issue reference, or reproduction steps.
2. Reproduce the bug. Trace the code path from trigger to symptom without gaps.
3. Form hypotheses. For uncertain links, make predictions testable in another scenario.
4. Stop at the causal chain gate: do not fix until the chain is complete or labeled as symptom-only with evidence.
5. If fixing: run or add a failing test first, apply one change, rerun affected checks through `clean-verify` when configured.
6. Hand evidence to `clean-review` when the fix touches behavior-bearing code.

## Modes

- **Interactive (default):** ask whether to fix after diagnosis when appropriate.
- **Pipeline (`mode:pipeline`):** fix convergent root causes, defer divergent issues, return structured status.

## Safety

- One hypothesis change at a time. No shotgun debugging.
- Never treat passing tests alone as proof when mutation or acceptance gaps remain.
