---
name: clean-merge-resolver
description: Resolve an in-progress git merge or rebase conflict hunk by hunk, tracing each resolution to both sides' primary sources, then finish the operation. Never abort. Use when git reports conflict or unmerged paths.
---

# Merge Conflict Resolver

Resolve conflicts by intent traced to each side's primary source, then finish the operation.

## Workflow

1. Confirm the operation in progress (merge or rebase). Never `--abort`; the goal is to finish it correctly.
2. Enumerate the unmerged paths with `git status`. Work one conflict hunk at a time.
3. For each hunk, read both parent revisions to identify the intent behind each side:
   - the incoming change and its originating source, and
   - the current branch's change and its originating source.
4. Resolve by intent: combine both sides' intent when they address different concerns, prefer one side when the other is obsolete, or rewrite the hunk when both are superseded. Record the reasoning per hunk.
5. After all hunks resolve and no unmerged paths remain, verify the resolved tree builds and tests pass.
6. Finish the operation only when resolution is complete and verified. Record resolution notes for review.

## Guardrails

- Never resolve a hunk by guessing. If a side's intent cannot be traced to a primary source, stop and surface the ambiguity rather than invent a resolution.
- Never `--abort`. Aborting discards the conflict context and the traceable resolutions.
- Preserve both sides' behavior unless a side is demonstrably obsolete; losing an intended change is a defect.
- Re-run verification after resolution; a conflict merge is not complete until the tree builds and tests pass.

## Tool-free mode

- If git operations are unavailable, report the conflict hunks and the exact unavailable status. Never claim a resolution occurred that did not.
