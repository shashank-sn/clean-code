# Shipping pipeline

Clean Code now ships a full planning-to-PR pipeline comparable to Compound Engineering, with stronger deterministic evidence gates.

## Skill map (CE → Clean Code)

| Compound Engineering | Clean Code | Notes |
| --- | --- | --- |
| `ce-brainstorm` | `clean-brainstorm` | Requirements-only plans |
| `ce-plan` | `clean-plan` | Implementation-ready units + verification contract |
| `ce-work` | `clean-build` + `clean-orchestrate` | Bounded implementation with evidence handoff |
| `ce-debug` | `clean-debug` | Causal-chain debugging |
| `ce-code-review` | `clean-review` | Evidence-based review + zero-finding allowed |
| `ce-simplify-code` | `clean-simplify` | Behavior-preserving cleanup |
| `ce-commit-push-pr` | `clean-ship` | Commit, push, PR with evidence summary |
| `ce-babysit-pr` | `clean-watch-pr` | CI watch loop |
| `ce-compound` | `clean-compound` | `docs/solutions/` + CONCEPTS |
| `ce-worktree` | `clean-worktree` | Isolated worktrees |
| `lfg` | `clean-lfg` | Full autonomous pipeline |

## Clean Code additions (no CE equivalent)

| Skill / CLI | Purpose |
| --- | --- |
| `clean-verify` / `clean-code verify` | Normalized deterministic checks |
| `clean-audit` / `clean-code audit` | Immutable release receipts |
| `clean-learn` | Proposal-only policy learning |
| `clean-design` | Architecture policy + acceptance |
| `clean-test` | Independent test tracks |
| `clean-discover` | Read-only capability discovery |
| `clean-setup` | Host-neutral integration |

## Compare workflows

```bash
go run ./cmd/clean-code compare-workflows
go run ./cmd/clean-code compare-workflows --manifest harness/calibration/workflow-comparison.json
```

The manifest scores both workflows on eighteen dimensions from product brainstorm through benchmark calibration. Scores are rubric-based for workflow coverage, not agent performance claims.

## Autonomous run

Invoke the `clean-lfg` skill with a feature description when you want planning through PR without step-by-step check-ins.

Canonical stage order lives in `harness/workflow/shipping-pipeline.json`.
