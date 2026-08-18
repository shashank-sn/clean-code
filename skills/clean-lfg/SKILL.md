---
name: clean-lfg
description: Run the full autonomous shipping pipeline from planning through verified implementation, review, simplify, ship, and CI watch. Use only when the user wants end-to-end delivery to an open PR without step-by-step check-ins.
---

# Clean LFG

Autonomous pipeline: plan → build → verify → review → simplify → ship → watch CI → audit.

## Stages (in order)

1. **Plan** — Invoke `clean-brainstorm` when requirements are vague; otherwise `clean-plan` to produce an implementation-ready plan under `docs/plans/`.
2. **Design** — When architecture boundaries change, invoke `clean-design` and record policy artifacts.
3. **Build** — Invoke `clean-build` per plan units with bounded scope. Use `clean-worktree` when isolation is needed.
4. **Test** — Invoke `clean-test` for independent unit, acceptance, integration, and UI/QA tracks.
5. **Verify** — Run `clean-code verify` against the final revision with approved policy. Mandatory evidence must not be stale.
6. **Review** — Invoke `clean-review` on diff plus evidence. Apply fixes for blocking findings.
7. **Simplify** — Invoke `clean-simplify` on the branch diff unless docs-only or trivial.
8. **Ship** — Invoke `clean-ship` with `mode:pipeline` to commit, push, and open PR.
9. **Watch** — Invoke `clean-watch-pr` until CI green or blocker.
10. **Audit** — Invoke `clean-audit` to produce an immutable receipt for the shipped revision.
11. **Compound** — Invoke `clean-compound` to capture durable learnings.

## Gates

- Stop if mandatory verification fails, required human spot checks are absent, or review reports blocking defects unresolved.
- Stop if `clean-verify` returns ERROR for a required check.
- Record procedural independence when the host cannot enforce separate agent contexts.

## Comparison to autonomous shipping pipelines

This skill sequences the same planning-to-ship stages many teams run manually, but adds mandatory `clean-verify`, architecture/trace checks, audit receipts, and proposal-only `clean-learn` gates that generic ship skills typically leave optional.
