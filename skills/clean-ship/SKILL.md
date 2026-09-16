---
name: clean-ship
description: Commit, push, and open a pull request with a plain-English, human-readable description tied to requirements and evidence. Use when implementation is verified and ready to ship, or when asked to open or update a PR.
---

# Clean Ship

Ship verified work: conventional commits, push, and a PR body a non-author can read and understand.

## Workflow

1. Gather git state: branch, diff, recent commits, default branch, open PR for current branch.
2. If on default branch with changes, create a meaningful feature branch automatically.
3. Confirm the change set was pruned (`clean-prune`) and verified (`clean-verify`) for code changes. Do not ship code changes with unresolved dead code.
4. Stage related files per logical commit (avoid `git add -A`). Run project tests before each commit.
5. Push to origin (or configured remote) with upstream set.
6. Write the PR body using the PR description rules below, then run the self-check before opening or updating the PR.
7. Open or update the PR, then pass the PR URL to `clean-watch-pr` when CI must reach green.

## PR description rules (required)

Write the PR body in **simple human-understandable English**. The test is whether a reader who has never seen this repository understands what changed and how to check it after one read. If a sentence needs a second read, rewrite it.

Open the body with one plain sentence: what this change does, in everyday words, with no acronym and no file path.

Required sections, in this order:

- **What changed** — short bullets naming the concrete behavior, files, or contracts touched. State what the software does now that it did not do before.
- **Why** — short bullets for the user or system reason behind the change.
- **How to verify** — bullets with the exact commands, fixtures, or evidence paths. Link revision-bound verify, review, and audit artifacts when they exist.
- **Gaps** — bullets for human spot checks, release gates, and checks not run (`NOT_RUN` / `NOT_AVAILABLE`).

Style:

- Short sentences. One idea per bullet. Active voice. Prefer bullets over paragraphs.
- Name the real thing: the file, the command, the flag, the endpoint. Do not describe it abstractly.
- Explain a technical term the first time it appears, in the same sentence, using everyday words: "Added a rate limiter (a cap on how many requests one client can send per minute)."
- Expand every acronym on first use.
- No marketing words: "seamless", "robust", "comprehensive", "powerful", "leverage", "utilize", "holistic", "best-in-class", "significant", "various", "orchestrate".
- No filler verdicts: "improves maintainability", "cleanup and improvements", "miscellaneous fixes". Name the concrete change instead.
- Describe the change, not the writing process. No "this PR aims to".
- Keep existing evidence-linking rules: name the revision and distinguish local checks, review, human checks, and release proof.
- Do not claim checks passed without `clean-verify` or equivalent evidence when behavior changed.

## Self-check before opening the PR

1. A reader who has never seen this repo can say what now happens that did not before.
2. Every file, command, flag, and path named in the body is real and current.
3. The verify section has a command a reader can run without asking a question.
4. The gaps section names what is still unproven instead of implying completeness.
5. No banned word appears, no acronym is unexplained, and no sentence needs a second reading.

If any check fails, rewrite the body before opening the PR. Never open a PR with jargon the change did not require.

## Example body

```markdown
Adds a 60-second cooldown after three failed logins.

**What changed**
- After three failed logins in a row, the same account waits 60 seconds before the next attempt (`internal/auth/throttle.go`).
- The login page shows how many seconds remain.

**Why**
- Failed guesses were unlimited, so a script could try passwords as fast as the server allowed.

**How to verify**
- `go test ./internal/auth/...`
- Manual: fail three logins, then confirm the fourth attempt is blocked for one minute.

**Gaps**
- No test for the exact countdown text.
- Not checked on mobile.
```

## Modes

- **Full workflow (default):** commit, push, PR.
- **Description-only:** write or rewrite PR body without committing.
- **Pipeline (`mode:pipeline`):** non-interactive; conservative defaults; no blocking asks.

## Safety

- Never commit secrets, build artifacts, or `.env` files.
- Never open a PR whose body only the change author can follow. Rewrite jargon into plain words.
- PR description must not claim checks passed without `clean-verify` or equivalent evidence when behavior changed.
