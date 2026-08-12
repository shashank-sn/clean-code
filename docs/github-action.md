# GitHub Action

Use the repository as a composite action to run the same trusted verification contract in GitHub Actions.

```yaml
permissions:
  contents: read

steps:
  - uses: actions/checkout@v4
  - uses: shashank-sn/clean-code@v0.1.0
    with:
      repository: .
      policy: .clean-code.json
      evidence: .clean-code/evidence
```

The action builds the CLI from the pinned action revision. It requires an approved policy path and does not enable repository-policy trust implicitly. A policy mismatch, required non-`PASS` result, incomplete report, or artifact error fails the step.
