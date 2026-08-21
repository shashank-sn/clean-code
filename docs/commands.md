# Command reference

- `version`: print the CLI version.
- `hosts`: list maintained host capability records.
- `agent list|validate [ID]|describe ID [--host ID]|emit ID --mode prompt|json [--host ID] [--output FILE]`: inspect, validate, or emit one portable agent contract.
- `provider validate --manifest FILE|result --input FILE`: validate an offline, trusted-policy provider contract or revision-bound provider result.
- `gauntlet plan|run --manifest FILE --output DIR`: create role packets or evaluate the iterative evidence/telemetry loop.
- `setup --host ID [--output DIR]`: report capabilities and optionally create non-overwriting host instructions.
- `discover [REPO]`: inspect metadata and propose commands without execution.
- `verify [--trusted-policy FILE | --allow-repository-policy] [--output DIR] [REPO]`: execute approved checks and normalize evidence.
- `architecture --policy FILE --graph FILE`: enforce declared component directions, public surfaces, exclusions, exceptions, and cycles.
- `architecture view --policy FILE --graph FILE [--previous FILE] [--producer ID --collection-scope TEXT --evidence-file FILE --evidence-sha256 DIGEST]`: emit a local architecture graph view and optional edge diff. Coverage remains incomplete unless a producer proves its collection scope with a graph-bound evidence file.
- `trace --plan FILE`: validate requirement examples and unit, acceptance, integration, and UI/QA tracks.
- `review --input FILE`: validate independent evidence-backed findings; zero findings are valid.
- `audit --input FILE --output NEW_FILE`: create a revision-bound immutable receipt. Use `--check RECEIPT` in place of `--output` to detect receipt or evidence changes.
- `benchmark --manifest FILE`: report detection, false positives, misses, and correct silence.
- `compare-workflows [--manifest FILE]`: score Clean Code vs Compound Engineering workflow coverage on a fixed rubric.
- `benchmark-full-flow [--manifest FILE] [--repo ROOT]`: run full-flow code-quality benchmark on CE vs CC sample outcomes.
- `learn --proposal FILE`: validate that a policy proposal is reversible, independently reviewed when decided, and unable to suppress protected gates.

All report-producing commands write JSON to standard output and diagnostics to standard error. Usage errors return 2; failed checks or invalid input return 1.
