# Benchmark method

The benchmark separates defect detection from unsupported findings.

1. Pin the task, repository revision, tool versions, policy, and agent configuration.
2. Keep seeded defects and clean controls balanced; preserve a held-out set for evaluation.
3. Run the same task with and without the workflow under the same limits.
4. Score exact oracle IDs as true positives, false positives, and false negatives. Count clean controls with no finding as correct silence.
5. Publish raw case results and run metadata. Make no improvement claim from the included demonstration manifest.

`harness/calibration/benchmark-manifest.yaml` is JSON-compatible YAML so the standalone binary can parse it without a YAML runtime. Its prefilled observed values are static scorer fixtures and carry no live agent-performance evidence. Live reviewer observations must retain the case packet hash, emitted prompt or instruction hash, model and reasoning setting, tool/check results, base and candidate revisions, raw output, and completion state. Report static fixture validation separately from live review runs; neither a fixture nor a successful JSON validation proves semantic review quality.
