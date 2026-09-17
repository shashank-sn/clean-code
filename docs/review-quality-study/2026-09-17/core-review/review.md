# independent review

status: FAIL  
completion: COMPLETE  
validation_scope: ARTIFACT_ONLY  
base: `ba97cb950b0cdc390f36706f6ed5c81533673faf`  
candidate snapshot: `sha256:1235e839ebc5b13e4de0d54927819e422640c278d1317ec37913de5be40470d7`

The 28 paths in `core-snapshot.json` were reviewed against the six requirements in the plan. The snapshot hashes still match. The harness/review-evals path was excluded as instructed, and no live study result or CI/GitHub state is treated as approval evidence.

## findings

### F1 — BLOCKING — whitespace bypasses independent-review identity check

Location: `internal/review/review.go:335`; the audit gate is `internal/audit/audit.go:119-120`.

Cause: v2 compares `input.Reviewer == input.ChangeAuthor` without canonicalizing or rejecting surrounding whitespace.

Consequence: a record with `change_author: "reviewer"` and `reviewer: "reviewer "` is accepted as `status: PASS`, `completion: COMPLETE`, with no issues. Because audit only checks the resulting report status, this can pass the independent-review gate into a complete audit receipt.

Reproduction: `probes/self-review-whitespace.json` was run with the candidate CLI. The captured output is `probes/self-review-whitespace.out.json` and exits 0. The required adversarial check is recorded as `FAIL` in `review.json`.

Bounded fix: canonicalize or reject surrounding whitespace for both identities before comparing them; add v2 hostile-input and audit regression tests.

### F2 — IMPROVEMENT — runtime/schema disagreement on finding line zero

Location: `internal/review/review.go:510`; schema location `harness/schemas/review-finding.schema.json:11`.

Cause: runtime rejects only negative lines, so `line: 0` passes; the published schema declares `minimum: 1`.

Consequence: the same record can pass `clean-code review` and fail a Draft 2020-12 schema validator, making producers and downstream gates disagree about validity.

Reproduction: `probes/line-zero.json` was run with the candidate CLI and exits 0; `probes/line-zero.out.json` captures `status: PASS`.

Bounded fix: make runtime and schema use the same line rule, then add a schema/runtime parity test.

## checks and limits

- PASS: `checks/full-go-test-race.log` from `go test -race ./...`.
- PASS: `checks/full-go-vet.log` from `go vet ./...`.
- PASS: candidate CLI accepted the checked-in v2 example with `validation_scope: ARTIFACT_ONLY`; see `checks/candidate-cli-example.json`.
- FAIL: the F1 self-review boundary probe, as above.
- The CLI validates the artifact contract only; it does not inspect product code, execute declared check sources, open declared artifacts, or verify their hashes.
- Standards and specification passes were sequential in this reviewer context; no separate reviewer context was available.
- The default Go cache path was sandbox-blocked, so checks were rerun with a workspace-local cache under this review artifact directory.
