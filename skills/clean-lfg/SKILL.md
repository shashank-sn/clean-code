---
name: clean-lfg
description: Run the full autonomous shipping pipeline from planning through implementation, simplification, final verification, review, ship, and CI watch. Use only when the user wants end-to-end delivery to an open PR without step-by-step check-ins.
---

# Clean LFG

Autonomous pipeline: plan → build → simplify → verify → review → ship → watch CI → audit → conditional evaluation discovery.

## Stages (in order)

1. **Plan** — Invoke `clean-brainstorm` when requirements are vague; otherwise `clean-plan` to produce an implementation-ready plan under `docs/plans/`.
2. **Design** — When architecture boundaries change, invoke `clean-design` and record policy artifacts.
3. **Build** — Invoke `clean-build` per plan units with bounded scope. Use `clean-worktree` when isolation is needed.
4. **Test** — Invoke `clean-test` for independent unit, acceptance, integration, and UI/QA tracks.
5. **Simplify** — Invoke `clean-simplify` on the branch diff unless docs-only or trivial.
6. **Verify** — Run `clean-code verify` against the final simplified revision with approved policy. Mandatory evidence must not be stale.
7. **Review** — Invoke `clean-review` on the final diff plus evidence. Apply fixes for blocking findings, then restart verification and review for the new revision.
8. **Ship** — Invoke `clean-ship` with `mode:pipeline` to commit, push, and open PR.
9. **Watch** — Invoke `clean-watch-pr` until CI green or blocker.
10. **Audit** — Invoke `clean-audit` to produce an immutable receipt for the shipped revision.
11. **Evaluate (conditional)** — When confirmed outcomes, repeated human judgment, escaped defects, rejected outputs, or correct silence reveal a candidate rule, invoke `clean-eval-discover`. Preserve clean controls and an untouched held-out set; do not activate a rule in this stage.
12. **Learn (conditional)** — Invoke `clean-learn` only for a reviewable proposal produced from calibrated evidence. Approval remains separate.
13. **Compound** — Invoke `clean-compound` to capture durable learnings.

## Gates

- Stop if mandatory verification fails, required human spot checks are absent, or review reports blocking defects unresolved.
- Stop if `clean-verify` returns ERROR for a required check.
- Skip evaluation discovery when there is no confirmed, bounded outcome to analyze; record that it was inapplicable rather than inventing a candidate.
- Record procedural independence when the host cannot enforce separate agent contexts.

## Comparison to autonomous shipping pipelines

This skill sequences the same planning-to-ship stages many teams run manually, but adds mandatory `clean-verify`, architecture/trace checks, audit receipts, and proposal-only `clean-learn` gates that generic ship skills typically leave optional.
