# review evaluation fixtures

This pack contains twelve small, self-contained review packets. A packet is the
material available to a reviewer: a requirement, caller/context files, and a
unified diff. The executable oracle and expected outcomes live outside packet
directories under `oracle/`.

The fixtures are synthetic and local. They are intended to compare review
quality, not to act as a sealed security benchmark.

## Reviewer-visible packet paths

The exact paths are listed in `manifest.json` under `modes.live_model_evaluation`.
Each packet is a directory containing `requirement.md`, `context/`, and
`change.diff`; the before/after source trees are also available to the fixture
validator under the case directory.

## Response format

See `response-format.md`. A live evaluator supplies observations; this repository
does not prefill findings for a model.

## Validation

Run `node runner.js` for fixture validation. It compiles and tests every before
and after tree with the separate oracle tests. It does not contact a model or
provider. Live model evaluation is a separate mode described in the manifest.

The pack-level build boundary is checked with `go test -race ./...` from this
directory. To validate one candidate packet, run `node runner.js --case=01`
(substitute `01` through `12`); the reviewer-visible candidate source is at
`NN/packet/candidate/` and the diff is `NN/packet/change.diff`. Run
`node runner-tests.js` for bounded corruption and infrastructure self-tests.
Run `node scorer-tests.js` for adjudication-input validation.
