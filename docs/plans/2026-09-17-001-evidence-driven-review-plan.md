# Evidence-driven code review

Status: implementation-ready

## Objective and observed gaps

Improve the normal Clean Code reviewer workflow so it investigates behavior, finds consequential defects, validates agent completion claims, and reports incomplete coverage honestly. Deliver a pull request, not a release or a global installation.

The current review skills list broad concerns but provide little investigation procedure. The specialist asks for two internally isolated passes, which a single context cannot actually provide. Its manifest permits repository edits but not command execution. The review validator accepts empty findings without review coverage, and the bundled calibration manifest supplies its own observed answers. These are source-observed gaps, not measured explanations for every poor review.

## Requirements

- R1: Both reviewer entry points teach a concrete, bounded investigation: establish intent and invariants; map the diff and affected callers; prioritize risk; trace failure scenarios; inspect test strength; challenge candidate findings; report supported findings and coverage gaps.
- R2: A versioned review contract records base and candidate identity, requirements, reviewed scope, relevant dimensions, checks and their provenance, and limitations. An incomplete review cannot become approval merely because findings are empty. The validator must explicitly distinguish artifact validation from semantic code verification.
- R3: Reviewer permissions, emitted prompts, dispatcher packets, and playbooks support the same procedure. Reviews remain read-only for product code. Test execution is conditional on host capability and authorization. Sequential passes never claim independent contexts.
- R4: A repeatable review evaluation uses actual code changes, executable defect oracles, clean controls, and externally supplied reviewer observations. Prefilled calibration answers remain labeled as validator fixtures rather than live agent performance.
- R5: Preserve v1 input compatibility and existing audit, policy, release, and authorization boundaries. Add hostile and incomplete-input tests for the new contract; keep runtime and JSON Schema aligned.
- R6: Run live baseline and candidate Luna reviews on the preregistered suite, retain raw observations and scoring evidence, and report misses and false positives. Claims are limited to the tested model, cases, instructions, and run conditions.

## Implementation units and ownership

1. **Review behavior and integration.** Skills, agent manifests, portable emission, review playbook, host guidance and focused package tests. One canonical investigation protocol must reach both direct skill and emitted-agent users. Avoid growing a collection of unrelated checklists.
2. **Review contract.** `internal/review/`, review schemas, CLI tests, and a documented example. Add an explicit version 2 assessment alongside legacy version 1. Check scope coverage, evidence-backed dimension dispositions, matching revisions for executed checks, missing/failed required checks, and consistency of completion and limitations. V1 must remain usable but visibly provides legacy artifact validation only. Neither version executes or independently certifies the review's claims.
3. **Independent evaluation.** `harness/review-evals/`: 12 neutral cases, including 8 defects and 4 clean controls; reviewer packets separate from executable oracles; dependency-free fixture validation and explicit observation scoring. The author works without candidate implementation context.
4. **Integration and evidence.** Normal-path documentation, full checks, live matched reviews, independent final diff review, and a revision-bound handoff receipt. Preserve any unperformed human or release checks as gaps.

## Preregistered evaluation

Use fresh Luna contexts with the same case packets and tool access. Freeze the baseline instructions from the original commit and the candidate instructions before live runs. Reviewers may inspect only assigned visible packets; separation is procedural on a shared filesystem. They must not read oracles, labels, other responses, or study results. Record model, reasoning setting, prompt hash, packet hash, raw output, tools/checks, and completion status.

The local acceptance bar is all high-impact defects found, at least 7/8 primary defects found, no unsupported blocking findings, and correct silence on all four clean controls. Evidence-integrity/incomplete-review scenarios must remain incomplete. Grade against the preregistered causal defect oracle; do not grade keyword matches or rewrite an oracle to reward a candidate. Report baseline and candidate separately. A saturated baseline or a small synthetic suite does not establish a universal performance gain. Repeat a candidate run with fresh contexts to assess consistency; disclose any diagnostic repairs and reruns separately.

## Verification contract

- R1/R3: emitted prompt and installed-package checks show that both entry points receive the complete protocol; no fictional isolation or product-code mutation permission.
- R2/R5: Go unit, CLI, schema-alignment, and audit-compatibility cases cover complete zero findings, missing coverage, stale evidence, unavailable checks, invalid enums, contradictions, and legacy behavior.
- R4: every defect is reproduced by an executable oracle; valid controls pass; fixture validation never manufactures live observations.
- R6: raw matched runs and independently checked scoring establish the stated local bar or record failure honestly.
- Final revision: `go test -race ./...`, `go vet ./...`, existing benchmark/calibration/full-flow commands, new fixture and scoring tests, package checks, independent review, and GitHub PR checks.

## Boundaries and completion

No merge, npm publication, installed global configuration change, profile learning, automatic policy activation, or external model/provider service. User-authorized Luna subagents perform implementation and live reviews. Use one Luna coordinator and explicit file ownership; the root supplies the plan and dispatches when Luna cannot spawn.

Completion requires the coherent reviewer behavior and contract changes, passing appropriate deterministic checks, local live-evaluation evidence meeting the stated bar, a resolved independent review, and a verified open PR. The final report must name limitations and cannot promise universally exceptional reviews.
