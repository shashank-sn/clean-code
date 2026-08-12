# Held-out study method

## Preregistration

Study `held-out-v1` evaluates revision `80c18421cfd2a8f7389c2974ad799d94e2e07f60` with ten blinded code-review cases. Each case has one seeded oracle: an actionable defect or correct silence. The exact case bytes, hidden oracle/scoring bytes, immutable model snapshot, inference settings, prompts, tool access, response limit, execution order, and minimum pair count are committed by SHA-256 in `harness/studies/held-out-v1.json` before execution.

The earlier H01-H10 session was a harness dry run. It was not preregistered, is excluded from results, and supports no claim.

## Arms

- Control receives the requirement, snippet, and response format only.
- Workflow receives the same material plus the agent-first Clean Code review contract.
- Both arms use model `gpt-5-2025-08-07`, no tools, minimal reasoning effort, a 120-word limit per case, and oracle `review-oracle-v1`.
- Arm order alternates by case. The committed runner performs no automatic retries and preserves partial failure evidence.
- Runs and result artifacts must have distinct immutable identities.

## Scoring

Preserve every raw response and bind its SHA-256 digest to the outcome. Reveal the committed oracle only after model execution. Record PASS, FAIL, or TIMEOUT, false positives, and correct silence. A pair counts only when both arms are present, independently identified, corpus/config-matched, and signed. Any failure, timeout, missing pair, config drift, replay, or invalid signature blocks performance claims.

Run `node harness/studies/run-held-out.mjs --validate-only` before the preregistration commit and again immediately before execution. Final scoring must pass the exact case, revealed oracle, and model configuration bytes to `clean-code study`; labels alone are not accepted.

## Authority

A protected Ed25519 public key must be supplied from outside the change checkout. The matching private key signs the exact result bytes after outcomes are compiled. The implementation agent must not control that private key.

## Publication

Publish the preregistration, raw outcomes, scoring report, limitations, model/tool metadata, and excluded runs. Do not claim improvement unless `claim_allowed` is true. A tie or negative result is published as measured.
