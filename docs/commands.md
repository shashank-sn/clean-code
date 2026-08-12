# Command reference

- `version`: print the CLI version.
- `hosts`: list maintained host capability records.
- `setup --host ID [--output DIR]`: report capabilities and optionally create non-overwriting host instructions.
- `discover [REPO]`: inspect metadata and propose commands without execution.
- `verify [--trusted-policy FILE | --allow-repository-policy] [--output DIR] [REPO]`: execute approved checks and normalize evidence.
- `architecture --policy FILE --graph FILE`: enforce declared component directions, public surfaces, exclusions, exceptions, and cycles.
- `trace --plan FILE`: validate requirement examples and unit, acceptance, integration, and UI/QA tracks.
- `review --input FILE`: validate independent evidence-backed findings; zero findings are valid.
- `audit --input FILE --output NEW_FILE`: create a revision-bound immutable receipt. Use `--check RECEIPT` in place of `--output` to detect receipt or evidence changes.
- `benchmark --manifest FILE`: report detection, false positives, misses, and correct silence.
- `learn --proposal FILE`: validate that a policy proposal is reversible, independently reviewed when decided, and unable to suppress protected gates.
- `study --manifest FILE --case-corpus FILE --oracle-corpus FILE --model-config FILE --preregistration FILE --results FILE --results-signature FILE --trusted-public-key FILE`: verify committed study inputs and claim policy, signed outcomes, and paired execution identities.

All report-producing commands write JSON to standard output and diagnostics to standard error. Usage errors return 2; failed checks or invalid input return 1.
