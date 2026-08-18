---
name: clean-watch-pr
description: Watch an open pull request until CI is green or a blocker is identified. Use after clean-ship when automated checks must pass before merge.
---

# Clean Watch PR

Monitor PR CI and react to failures without silent drift from the verified revision.

## Workflow

1. Resolve the open PR for the current branch or the URL provided.
2. Poll or stream CI status until success, failure, or timeout.
3. On failure, summarize failing jobs with log excerpts and invoke `clean-debug` in pipeline mode when appropriate.
4. Re-run `clean-verify` locally when CI failures indicate missing or stale evidence.
5. Return structured status: green, blocked, or failed with recovery path.

## Safety

- Do not merge automatically unless explicitly authorized.
- Treat flaky CI as a blocker until evidence shows a non-code cause.
