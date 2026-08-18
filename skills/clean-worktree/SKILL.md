---
name: clean-worktree
description: Create or attach an isolated git worktree for parallel or focused work. Use before large features or when the default branch must stay clean.
---

# Clean Worktree

Isolate work on a dedicated branch without disturbing the current checkout.

## Workflow

1. Detect existing worktrees and whether the harness provides native isolation.
2. Prefer harness-native worktree tools when available.
3. Otherwise create a worktree from the default branch with a meaningful branch name derived from the plan or task.
4. Record the worktree path and branch for the orchestrator.
5. Never nest worktrees inside another agent worktree without harness support.

## Safety

- Do not commit to the default branch from a worktree intended for feature work.
- Clean up worktrees after merge or explicit abandon.
