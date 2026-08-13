# held-out-v1 corpus

`cases.json` is the canonical public byte sequence for the ten blinded review cases. Case order does not encode case class.

Expected outcomes, accepted source locations, semantic scoring tokens, and the scoring table are held outside the repository. `preregistration.json` commits to both canonical byte sequences before execution.

The H01-H10 dry-run cases are excluded. This corpus uses new scenarios and identifiers `held-out-review-01` through `held-out-review-10`.

## Execution outcome

The single preregistered attempt began on 2026-08-13 and stopped at ordinal 10 after `held-out-review-05/workflow` timed out. Nine runs completed; the ten later reserved slots were not filled. The partial request envelopes, raw responses, terminal journal, all 20 reservation records, and `execution-report.json` are preserved.

This attempt is incomplete. It was not scored, `execution_valid` is false, `claim_allowed` is false, and it does not qualify a release.
