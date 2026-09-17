# local reviewer study: 2026-09-17

the candidate met the preregistered local acceptance bar in both trials: 8/8 primary defects found, all four high-impact defects found, 4/4 clean controls silent, zero unsupported findings, and approval withheld on the missing-context probe. **the baseline achieved the same result. this study does not demonstrate a recall improvement or establish that the agents are exceptional reviewers.**

| condition | primary defects | silent controls | unsupported findings | missing-context probe | local bar |
| --- | --- | --- | --- | --- | --- |
| baseline | 8/8 | 4/4 | 0 | incomplete; approval withheld | pass |
| candidate trial 1 | 8/8 | 4/4 | 0 | incomplete; approval withheld | pass |
| candidate trial 2, reversed order | 8/8 | 4/4 | 0 | incomplete; approval withheld | pass |

the demonstrated improvements are in the workflow and its safeguards: a shared causal investigation protocol, read-only product permissions, truthful separation of review passes, explicit coverage and evidence requirements, and a v2 validator that fails closed on incomplete or contradictory records. 35 independent CLI/schema/audit probes pass. the validator checks the record; it does not certify the source code or authenticate declared command execution.

the candidate observations include call-path and assessment summaries and distinguish passing builds from missing behavioral tests. this is a qualitative observation, not a separately preregistered quality score. all primary packets were reviewed; a defective change can have a complete review. production readiness and external CI approval were outside the source-packet task.

## method and limits

the study used `gpt-5.6-luna` with high reasoning in six fresh reviewer contexts: two for the baseline and two for each candidate trial. each context received six assigned source packets; one context per condition also received the missing-context probe. the second candidate trial reversed case order. the same frozen reviewer instructions were used in both candidate trials.

the suite contains eight synthetic go defects and four clean controls. the primary defects cover tenant authorization, concurrent idempotency, partial-write error handling, argument forwarding, weakened readiness checks, escaped separators, expiry boundaries, and shutdown completion. hidden tests reproduce each defect against before/after code and check the controls. reviewers saw requirements, diffs, caller context, and runnable candidate trees, with safe local checks available.

packets, prompts, model settings, and acceptance thresholds were frozen before the valid runs. raw answers were frozen before causal grading. a fresh luna grader received condition aliases and no reviewer prompt files; it matched concrete findings to fixed oracles and assessed the missing-context probe separately. the deterministic scorer validates and counts that adjudication; it cannot prove the grader's semantic claims.

the original wave was discarded before grading because reference context contained duplicate go declarations and lacked the intended module boundaries. original raw outputs remain ungraded in the local study workspace. the correction created runnable candidate trees and module boundaries. every causal oracle test and oracle-manifest byte remained unchanged; the added oracle `go.mod` is recorded separately. the corrected study restarted all conditions in fresh contexts. further pre-score harness repairs tightened hash binding, test-failure classification, and adjudication validation without changing the frozen cases, prompts, or thresholds.

separation was procedural on a shared filesystem, not a sealed security boundary. the final candidate reviewer received an operational reminder to time out hanging reproductions and report genuine gaps; all operational steering is recorded in `dispatch-notes.json`. this small synthetic suite does not establish production-wide accuracy, model independence, native host reload, or cost and latency improvements. a saturated baseline cannot establish a recall improvement.

all conditions used a common observation format so their findings could be compared. this trial therefore does not measure whether a production host autonomously emits a valid v2 artifact. v2 contract validation and local emitted/packed instruction parity were checked separately.

## inspect and reproduce

- [preregistered design](preregistered.json), [raw hashes](raw-evidence-inventory.json), and [packet identities](visible-packet-identities.json)
- [independent adjudication](grading/adjudication.md), [condition mapping](grading/condition-mapping.json), and [exact raw-transcription check](grading/raw-transcription-check.json)
- [baseline score](grading/baseline-score.json), [candidate trial 1](grading/candidate-trial-1-score.json), and [candidate trial 2](grading/candidate-trial-2-score.json)
- [missing-context observations](grading/protocol-probes.json), [discarded-wave record](invalid-fixture-wave.json), and [oracle preservation](oracle-preservation.json)
- [independent core review](core-review/final-core-review.md), [independent harness review](harness-review/final-review.md), and [35 contract probes](verification/contract-probes-final.json)

from the repository root:

```sh
node harness/review-evals/runner.js
node harness/review-evals/runner-tests.js
node harness/review-evals/scorer-tests.js
node harness/review-evals/scorer.js docs/review-quality-study/2026-09-17/grading/baseline.json
node harness/review-evals/scorer.js docs/review-quality-study/2026-09-17/grading/candidate-trial-1.json
node harness/review-evals/scorer.js docs/review-quality-study/2026-09-17/grading/candidate-trial-2.json
```

these commands rerun fixture validation and score the preserved adjudication. they do not launch fresh reviewers. `runner.js --live` explicitly returns `NOT_RUN`. the separate Python contract probe script requires `jsonschema` and a locally built CLI; it is independent acceptance evidence, not a new runtime dependency.
