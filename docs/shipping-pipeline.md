# Shipping pipeline

Clean Code now ships a full planning-to-PR pipeline comparable to Compound Engineering, with stronger deterministic evidence gates.

The final-change sequence is `clean-build` → `clean-test` → `clean-simplify` → `clean-prune` → `clean-verify` → `clean-review` → `clean-ship`. Pruning runs before verification so review and ship cover the revision that actually ships. Any review fix starts a new revision and repeats verification and review.

## Skill map (CE → Clean Code)

| Compound Engineering | Clean Code | Notes |
| --- | --- | --- |
| `ce-brainstorm` | `clean-brainstorm` | Requirements-only plans |
| `ce-plan` | `clean-plan` | Implementation-ready units + verification contract |
| `ce-work` | `clean-build` + `clean-orchestrate` | Bounded implementation with evidence handoff |
| `ce-debug` | `clean-debug` | Causal-chain debugging |
| `ce-code-review` | `clean-review` | Evidence-based review + zero-finding allowed |
| `ce-simplify-code` | `clean-simplify` | Behavior-preserving cleanup |
| — | `clean-prune` | Dead-code elimination with a reference search behind every deletion |
| `ce-commit-push-pr` | `clean-ship` | Commit, push, PR with evidence summary |
| `ce-babysit-pr` | `clean-watch-pr` | CI watch loop |
| `ce-compound` | `clean-compound` | `docs/solutions/` + CONCEPTS |
| `ce-worktree` | `clean-worktree` | Isolated worktrees |
| `lfg` | `clean-lfg` | Full autonomous pipeline |

## Clean Code additions (no CE equivalent)

| Skill / CLI | Purpose |
| --- | --- |
| `clean-route` / `clean-code route` | Adaptive playbook router from task signals (advisory) |
| `clean-arena` / `clean-code arena` | Bounded competing-design decision records |
| `clean-probe` / `clean-code probe` | Adversarial acceptance probes for medium/high risk |
| `clean-code parallel` | Validate parallel orchestration ownership records |
| `clean-code playbook` | Developer playbooks with terminal states |
| `clean-verify` / `clean-code verify` | Normalized deterministic checks |
| `clean-prune` | Dead, unreachable, and leftover code removal before final verification |
| `clean-audit` / `clean-code audit` | Immutable release receipts |
| `clean-eval-discover` | Blinded, bottom-up evaluation discovery from confirmed outcomes; never activates a rule |
| `clean-learn` | Proposal-only policy learning |
| `clean-design` | Architecture policy + acceptance |
| `clean-test` | Independent test tracks |
| `clean-discover` | Read-only capability discovery |
| `clean-setup` | Host-neutral integration |

`clean-ship` must open PRs with plain-English bullet descriptions covering what changed, why, how to verify, and gaps. The body must be readable by someone who has never seen the repository: explain each term in everyday words, name real files and commands, and run the self-check before opening the PR. Do not claim checks passed without verify evidence.

## Compare workflows

```bash
go run ./cmd/clean-code compare-workflows
go run ./cmd/clean-code compare-workflows --manifest harness/calibration/workflow-comparison.json
```

The manifest scores both workflows on nineteen dimensions from product brainstorm through benchmark calibration. Scores are rubric-based for workflow coverage, not agent performance claims.

## Autonomous run

Invoke the `clean-lfg` skill with a feature description when you want planning through PR without step-by-step check-ins. After audit, it conditionally invokes `clean-eval-discover` only when confirmed outcomes or repeated human judgment support a candidate evaluation; `clean-learn` follows only with calibrated evidence and separate approval.

Canonical stage order lives in `harness/workflow/shipping-pipeline.json`.
