# Full-flow benchmark (CE vs Clean Code)

This command scores checked-in example outcomes with automated metrics and reviewer scores supplied in a manifest. It does not run either agent workflow or invoke an independent reviewer. The bundled outcomes are demonstration fixtures, not a controlled measurement of current agent performance.

## Task

See [../examples/benchmark-flow/task.md](../examples/benchmark-flow/task.md): implement `NormalizeSlug` with table-driven tests and edge-case coverage.

## Outcomes

| Workflow | Path | Style |
| --- | --- | --- |
| Compound Engineering (simulated) | `examples/benchmark-flow/outcomes/ce/slug` | Monolithic function, minimal tests |
| Clean Code | `examples/benchmark-flow/outcomes/cc/slug` | Decomposed functions, table + fuzz tests |

## Run

```bash
go test -race ./examples/benchmark-flow/...
go run ./cmd/clean-code benchmark-full-flow
```

## Scoring

1. **Automated rubric** — tests pass, function size, decomposition, test breadth, fuzz hardening.
2. **Stored reviewer scores** — naming, simplicity, test quality, and maintainability scores read from the manifest. The command does not verify the declared reviewer identity, independence, or blinding.

The bundled scores favor **Clean Code (Outcome B)**. That result describes the supplied examples and scores; it is not evidence that Clean Code outperforms another workflow on new tasks. For actual code-review observations and separate defect oracles, use the [review evaluation pack](../harness/review-evals/README.md).

## Reproduce a live comparison

1. Run CE pipeline (`ce-brainstorm` → `ce-plan` → `ce-work` → `ce-ship`) on the task in an isolated worktree.
2. Run Clean Code pipeline (`clean-brainstorm` → `clean-plan` → `clean-build` → `clean-test` → `clean-simplify` → `clean-prune` → `clean-verify` → `clean-review` → `clean-ship`) in another worktree.
3. Point `harness/calibration/full-flow-manifest.json` at both package directories.
4. Invoke a different-model reviewer on both outcomes without revealing workflow labels.
5. Run `benchmark-full-flow` to merge auto + reviewer scores.
