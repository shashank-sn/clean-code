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
      policy-sha256: ${{ vars.CLEAN_CODE_POLICY_SHA256 }}
      evidence: .clean-code/evidence
```

Store `CLEAN_CODE_POLICY_SHA256` as a repository, organization, or protected-environment Actions variable controlled by policy owners. Do not keep the approved digest in a pull-request-editable file.

The action builds the CLI from the pinned action revision. Before any repository command runs, it computes the policy SHA-256 digest and rejects drift. Inputs cross the expression boundary through environment variables and are quoted as data. A digest mismatch, invalid input path, required non-`PASS` result, incomplete report, or artifact error fails the step.
