---
name: clean-verify
description: Run repository-declared deterministic checks, enforce trusted execution policy, validate required output artifacts, and preserve revision-bound evidence. Use after implementation or refactoring, before claiming completion, during code review, and in local IDE terminals, coding-agent platforms, CLI sessions, or CI gates.
---

# Clean Verify

Run the same declared checks in every host and report what actually happened.

## Workflow

1. Run `clean-code discover <repo>` and inspect the declared commands without executing them.
2. Use a separately approved policy when available:

   `clean-code verify --trusted-policy <approved-policy.json> --output <evidence-dir> <repo>`

3. Without an approved policy file, require an explicit trust decision before running `clean-code verify --allow-repository-policy --output <evidence-dir> <repo>`.
4. Read the JSON report and the protected `report.json` bundle. Confirm the repository revision, policy source, completion state, policy deltas, and every check result.
5. Report required failures, unavailable tools, missing configuration, timeouts, truncation, and artifact errors explicitly. Never translate them into a pass.

## Gate semantics

- Treat `PASS` as evidence for one declared check at one revision.
- Block whenever a required result has any status except `PASS`.
- Block on `ERROR`, an incomplete report, or any unapproved policy delta.
- Keep optional failures visible without turning them into global failure.
- Treat `NOT_AVAILABLE`, `NOT_CONFIGURED`, and `NOT_RUN` as distinct states.
- Require declared artifacts to exist, match their format, and change when `fresh: true`.
- Compare each declared metric with its own baseline and tolerance. Keep coverage, mutation, duplication, and complexity separate.

## Safety

- Execute argument arrays directly. Never rewrite them as shell strings.
- Never install tools, fetch dependencies, or enable network access to manufacture a pass.
- Never execute a proposed command that differs from the trusted policy. Run only the trusted command set and request separate approval for the delta.
- Never add `--allow-repository-policy` without an explicit trust decision for that repository and command set.
- Keep output bounded and redact credential-shaped values from evidence.
- Keep artifact parsing bounded. Accept only declared file, text, JSON, XML, SARIF, or LCOV formats.
- Keep evidence tied to the reported commit and dirty-worktree state.
- Present deterministic checks alongside acceptance testing, UI/QA procedures, architecture review, mutation testing, and human spot checks.

## Host portability

Use the host's terminal, task runner, or pipeline step to invoke the same `clean-code verify` CLI. Keep check semantics identical across IDEs, coding platforms, operating systems, and languages. If the CLI or a declared executable is unavailable, report that state and the exact missing capability.
