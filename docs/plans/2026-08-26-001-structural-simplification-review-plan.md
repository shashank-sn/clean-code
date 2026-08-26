# Structural-simplification review plan

## Status

`implementation-ready`

## Outcome

Release `0.4.2` with a structural-simplification lens in independent review. It must surface a materially simpler, behavior-preserving design when the changed scope supports one, without turning reviewer taste, a line-count threshold, or an unconventional implementation into a hard failure.

## Requirements

- `SSR-1`: Review non-trivial changed code for removable branches, wrappers, layers, duplicate helpers, unclear type or ownership boundaries, and special-case growth.
- `SSR-2`: A structural finding records changed-scope evidence, a concrete maintenance or architecture consequence, and a bounded behavior-preserving alternative.
- `SSR-3`: Crossing 1,000 lines in a source file is a decomposition prompt, not a universal policy violation.
- `SSR-4`: Simplification occurs before final verification and independent review; structural findings route to `clean-refactor`, whose changes require fresh verification and review.
- `SSR-5`: The shipped skill contract and public documentation preserve correct silence and the prohibition on taste-only findings.
- `SSR-6`: All npm version surfaces identify `0.4.2` and the release notes describe this scope.

## Implementation units

### U1 — Review contract

- **Files:** `skills/clean-review/SKILL.md`, `skills/clean-refactor/SKILL.md`
- **Approach:** Add a risk-driven structural-simplification lens, evidence rules, severity boundary, and refactor handoff.
- **Verification:** Review the skill against `SSR-1` through `SSR-5`; run repository tests.

### U2 — Final-revision workflow

- **Files:** `skills/clean-lfg/SKILL.md`, `harness/workflow/shipping-pipeline.json`, `README.md`, `docs/shipping-pipeline.md`, `docs/benchmark-full-flow.md`, `harness/calibration/workflow-comparison.json`
- **Approach:** Place simplification before final verification and review in every canonical workflow surface.
- **Verification:** A regression test confirms the declared stage order.

### U3 — Release metadata and regression protection

- **Files:** `package.json`, `CHANGELOG.md`, `tests/repository_test.go`
- **Approach:** Bump the npm package to `0.4.2`; assert the review and workflow contracts remain shipped.
- **Verification:** `go test ./...`; `npm pack --dry-run`; packed-artifact smoke test.

## Scope boundaries

- No automatic refactors, new universal line limits, score, schema, CLI command, or policy activation.
- No claim that passing tests proves a design is optimal.
- No publish, tag, merge, or global installation in this PR; those occur only after the release decision.

## Definition of done

The skill, refactor handoff, workflow, public documentation, package metadata, release notes, and regression tests agree on the requirements above. The final branch passes the complete local test and package checks, and the PR records that npm publication requires a post-merge matching `v0.4.2` GitHub Release.
