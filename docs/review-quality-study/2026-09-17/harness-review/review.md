# review-evals harness review

Scope: `harness/review-evals/runner.js`, `runner-tests.js`, `scorer.js`,
`scorer-tests.js`, `README.md`, and `response-format.md` in the clean-code
worktree. I did not inspect live model results or oracle answer contents.

## Findings

### P1 — fixture validation trusts a mutable hash inventory and does not bind the frozen snapshot

`runner.js:37-52` reads `reviewer-material.sha256` from the same root it is
validating, then trusts every hash in that file. There is no pinned digest for
the inventory and no check against `work/review-final/eval-final-snapshot.json`.
The hard-coded bindings at `runner.js:54-62` cover before/after trees and the
diff, but not the visible requirement/context packet files (and the hidden
oracle manifest is not covered by `frozenInventory`).

Reproduction in a temporary copy: append one line to
`01/packet/requirement.md`, update only that file's line in
`reviewer-material.sha256`, and run `node runner.js --root=<copy> --case=01`.
It exits 0 and reports the case as expected. In a separate temporary copy,
filter the hidden oracle manifest to the four clean cases; the runner exits 0
with `total: 4, failed: 0`. A validator must compare the complete pack to an
immutable expected manifest/snapshot (or embed a trusted digest) and enforce
the expected case set.

### P1 — oracle assertions without a diagnostic are classified as public-test failures

`runner.js:79-83` requires both a named oracle `fail` event and an
`oracle_test.go:<line>:` output diagnostic before returning
`oracle-assertion-fail`. A temporary oracle test containing only `t.Fail()` or
`t.FailNow()` emits the named oracle failure but no matching diagnostic, so
`runGo()` returns `public-test-fail`. The same test with `t.Errorf()` returns
`oracle-assertion-fail`.

Classify runtime/panic failures first, then any named failure from the oracle
test set as `oracle-assertion-fail`; do not make a source-location diagnostic
necessary for the oracle/public distinction.

### P1 — scorer grants primary credit to malformed non-concrete findings

`scorer.js:27-40` validates only `finding_id`, `supported`, `blocking`, and the
oracle mapping. A complete input whose eight primary findings contain only
those four fields is accepted with `primary_recall.found: 8` and
`acceptance: true`. The required review response fields in
`response-format.md:9-15` (`title`, `severity`, `file`, `line`,
`explanation`, and `evidence`) are never checked, so an adjudication can claim
all eight findings without a concrete packet-grounded finding.

Validate the adjudication finding shape and types/ranges before adding an
oracle to `found`; require the concrete evidence fields (and reject malformed
or empty values) while keeping `supported` as the independent adjudication
input.

## Checks

The supplied snapshot hash was checked for all 154 paths: 154 matched, 0
missing, 0 mismatched.

Commands run from `harness/review-evals`:

```text
GOCACHE=/private/tmp/clean-code-review-go-cache \
GOMODCACHE=/private/tmp/clean-code-review-go-mod-cache node runner-tests.js
  {"status":"PASS","tests":8}

GOCACHE=/private/tmp/clean-code-review-go-cache \
GOMODCACHE=/private/tmp/clean-code-review-go-mod-cache node scorer-tests.js
  {"status":"PASS","tests":10}

GOCACHE=/private/tmp/clean-code-review-go-cache \
GOMODCACHE=/private/tmp/clean-code-review-go-mod-cache node runner.js
  total 12, failed 0

node --check runner.js runner-tests.js scorer.js scorer-tests.js
  pass (invoked separately for each file)
```

## Reviewed file hashes

```text
be4dbe113b7aad5113f8efeb8b2e59c7a308da13009bdf2974eac18d5fcb0b15  harness/review-evals/runner.js
ee86fb9f3a09fd96c3baa65765c3b89f59a1f517f0db7f24dab3d7a556ddd066  harness/review-evals/runner-tests.js
c8cf22468b448793eafb292ec5dbaa89c3a4b76b79711cfd03678f73bcbe4542  harness/review-evals/scorer.js
2eb6b7e2d2c3bd6492467a08ce48f241dccac971bd949a106acd7ffb10114881  harness/review-evals/scorer-tests.js
38870dba8189b6b630ab96a19a61dccafac21420c4d20daea9ba8d17ff329711  harness/review-evals/README.md
d86233008dec64c667ffd8aff204dc70fc3f186c74c5cd2723dacb02d4de6e72  harness/review-evals/response-format.md
```

The review is read-only with respect to product code. Temporary copies used
for adversarial checks were removed. No network/provider calls or subagents
were used.
