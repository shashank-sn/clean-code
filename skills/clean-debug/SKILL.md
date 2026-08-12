---
name: clean-debug
description: Diagnose a reproducible failure through evidence and bounded hypotheses before changing code. Use for bugs, flaky tests, incidents, regressions, performance failures, or confusing behavior whose cause is not already proven.
---

# Clean Debug

Prove the cause, apply the smallest fix, and preserve the evidence.

## Workflow

1. State the observed behavior, expected behavior, environment, revision, and smallest reproduction.
2. Trace inputs and state across public boundaries before inspecting incidental implementation detail.
3. Rank at most three falsifiable hypotheses and choose one discriminating check.
4. Run the check, record the result, and discard disproven hypotheses.
5. Add a failing regression test at the narrowest stable boundary.
6. Change only the proven cause, then run the regression test and affected verification.

## Guardrails

Do not combine diagnosis with speculative cleanup. After three failed hypotheses, name the wrong shared assumption and request a fresh independent diagnosis. Treat disappearance without an explained cause as unresolved.

## Evidence

The debug record contains the reproduction, cause evidence, rejected hypotheses, regression test, changed revision, and verification result.
