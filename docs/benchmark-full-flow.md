# Full-flow benchmark (CE vs Clean Code)

This benchmark runs the same small task through representative Compound Engineering and Clean Code outcomes, then scores them with automated metrics and an independent blind reviewer.

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
2. **Independent reviewer** — blind scores on naming, simplicity, test quality, maintainability (Gemini reviewer, stored in manifest).

Latest blind review winner: **Clean Code (Outcome B)** — decomposed functions and comprehensive tests vs monolithic CE-style outcome.

## Reproduce a live comparison

1. Run CE pipeline (`ce-brainstorm` → `ce-plan` → `ce-work` → `ce-ship`) on the task in an isolated worktree.
2. Run Clean Code pipeline (`clean-brainstorm` → `clean-plan` → `clean-build` → `clean-test` → `clean-verify` → `clean-review` → `clean-ship`) in another worktree.
3. Point `harness/calibration/full-flow-manifest.json` at both package directories.
4. Invoke a different-model reviewer on both outcomes without revealing workflow labels.
5. Run `benchmark-full-flow` to merge auto + reviewer scores.
