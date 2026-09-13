---
name: clean-ship
description: Commit, push, and open a pull request with evidence-linked description. Use when implementation is verified and ready to ship, or when asked to open or update a PR.
---

# Clean Ship

Ship verified work: conventional commits, push, and a PR body tied to requirements and evidence.

## Workflow

1. Gather git state: branch, diff, recent commits, default branch, open PR for current branch.
2. If on default branch with changes, create a meaningful feature branch automatically.
3. Stage related files per logical commit (avoid `git add -A`). Run project tests before each commit.
4. Push to origin (or configured remote) with upstream set.
5. Open or update PR using the PR description rules below.
6. Pass PR URL to `clean-watch-pr` when CI must reach green.

## PR description rules (required)

Write a **simple human-understandable English** PR body with **bullet points**. Reviewers should understand the change without decoding jargon walls.

Required sections:

- **What changed** — short bullets naming the concrete behavior or files/contracts touched
- **Why** — short bullets for the user/system reason
- **How to verify** — bullets for commands, fixtures, or evidence paths; link revision-bound verify/review/audit artifacts when they exist
- **Gaps** — bullets for human spot checks, release gates, or checks not run (`NOT_RUN` / `NOT_AVAILABLE`)

Style:

- Plain English. Short sentences.
- Prefer bullets over dense paragraphs.
- No marketing fluff.
- Do not claim checks passed without `clean-verify` or equivalent evidence when behavior changed.
- Keep existing evidence-linking rules: name the revision and distinguish local checks, review, human checks, and release proof.

## Modes

- **Full workflow (default):** commit, push, PR.
- **Description-only:** write or rewrite PR body without committing.
- **Pipeline (`mode:pipeline`):** non-interactive; conservative defaults; no blocking asks.

## Safety

- Never commit secrets, build artifacts, or `.env` files.
- PR description must not claim checks passed without `clean-verify` or equivalent evidence when behavior changed.
