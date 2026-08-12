# Held-out study method

## Preregistration

Study `held-out-v1` evaluates revision `2d68025fa8decaa46067544711dce184be069425` with ten blinded code-review cases. Each case has one seeded oracle: an actionable defect or correct silence. The task IDs, model family, tool access, response limit, oracle version, and minimum pair count are fixed in `harness/studies/held-out-v1.json`.

The earlier H01-H10 session was a harness dry run. It was not preregistered, is excluded from results, and supports no claim.

## Arms

- Control receives the requirement, snippet, and response format only.
- Workflow receives the same material plus the agent-first Clean Code review contract.
- Both arms use model `gpt-5`, no tools, a 120-word limit per case, and oracle `review-oracle-v1`.
- Runs and result artifacts must have distinct immutable identities.

## Scoring

Preserve every raw response. Record PASS, FAIL, or TIMEOUT, false positives, and correct silence. A pair counts only when both arms are present, independently identified, config-matched, and signed. Any failure, timeout, missing pair, config drift, replay, or invalid signature blocks performance claims.

## Authority

A protected Ed25519 public key must be supplied from outside the change checkout. The matching private key signs the exact result bytes after outcomes are compiled. The implementation agent must not control that private key.

## Publication

Publish the preregistration, raw outcomes, scoring report, limitations, model/tool metadata, and excluded runs. Do not claim improvement unless `claim_allowed` is true. A tie or negative result is published as measured.
