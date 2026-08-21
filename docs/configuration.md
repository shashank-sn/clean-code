# Configuration

Place `.clean-code.json` at the repository root. Start from `harness/config/example.clean-code.json` and validate against `harness/config/config.schema.json`.

Each command declares an executable and argument array, working directory, timeout, output limit, accepted exit codes, requirement level, artifacts, and optional baselines. Shell strings are disabled by default. Discovery proposes adapter commands but never runs or installs them.

Use `clean-code verify --trusted-policy approved.clean-code.json <repo>` for routine checks. `--allow-repository-policy` is an explicit one-run approval for the repository file. New commands, wider permissions, disabled gates, and weaker thresholds remain visible as policy drift.

Statuses retain their exact meaning: `PASS`, `FAIL`, `NOT_AVAILABLE`, `NOT_CONFIGURED`, `NOT_RUN`, `STALE`, and `ERROR`. Optional missing tools stay visible as unavailable or unconfigured results. `STALE` means the evidence belongs to a different revision and cannot satisfy a required check.

Artifacts may be JSON, XML, SARIF, LCOV, text, or opaque files. Set metric, scope, direction, tolerance, and baseline explicitly. Required missing, stale, malformed, or regressed artifacts block verification.
