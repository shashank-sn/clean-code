# Independent test planning

The test-plan contract maps each requirement to acceptance examples and four evidence tracks: unit, acceptance, integration or contract, and UI/QA.

Run the trace check before calling the plan complete:

```bash
clean-code trace --plan test-plan.json
```

Every requirement needs coverage from all four tracks. A track can be `INAPPLICABLE` when it includes a requirement-specific reason. Missing tools or unfinished work use `NOT_AVAILABLE` or `NOT_RUN`, which also require reasons and remain visible.

Applicable tracks require an owner. `PASS` and `FAIL` require an evidence location. Acceptance and UI/QA tracks should use `context_source: requirements`; implementation or mixed context produces a warning about correlated assumptions.

Use [the acceptance template](../harness/templates/acceptance.feature) for executable examples and [the QA template](../harness/templates/qa-procedure.md) for user-interaction procedures. Tool choice remains repository-owned.
