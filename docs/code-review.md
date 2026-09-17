# Code review operator guide

Run review against a named base and candidate revision with the requirements, complete diff, affected-call-path notes, and revision-bound verification evidence. The reviewer owns a read-only investigation of product code; `execute_commands` is conditional on host capability and safe authorization.

Use the causal protocol in the reviewer skills and emitted host guidance:

1. derive invariants from intent and original requirements, and record base/head revisions, changed-file inventory, author/context identity, and scope;
2. inspect the entire diff, then affected callers, state/data flows, authorization and tenant boundaries, and side effects;
3. prioritize correctness, security/authorization, and integration by impact, then trace success, failure, timeout, retry, rollback, concurrency, duplicate delivery, partial update, and recovery scenarios;
4. map every requirement to call paths and happy, boundary, negative, failure, and recovery tests;
5. challenge each candidate finding with counterevidence, severity, causality, confidence, and the smallest bounded fix;
6. verify agent completion claims against the actual diff, test commands/results/skips/changed assertions, exact revision, and actual runtime state for runtime claims; report findings, residual risks, limitations, and unreviewed scope.

The v2 review record must identify `base_revision`, candidate `revision`, `scope`, `requirements`, six dimension assessments (`correctness`, `integration`, `tests`, `failure_modes`, `security`, `maintainability`), revision-bound check provenance, `limitations`, and `completion`. Use `NOT_APPLICABLE` only with a reason. “No supported defect was found in examined scope” is bounded; zero findings never proves correctness. Only explicitly designated legacy tooling may use v1, marked `NOT_ASSESSED`. A successful JSON validation validates the artifact contract; it does not semantically approve the code.

When separate reviewer contexts are available, use separate standards and specification pass packets. Otherwise run those passes sequentially and record procedural separation and the limitation. Never claim isolation that the host did not provide. Keep structural findings tied to changed-scope evidence, a concrete consequence, and a bounded behavior-preserving alternative. Do not use finding quotas, generic alarms, or style preference as evidence.
