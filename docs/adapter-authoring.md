# Adapter authoring

An adapter identifies repository metadata and proposes direct commands while leaving execution authority unchanged.

## Contract

Store built-in adapters in `harness/adapters/`. The `.yaml` files currently contain strict JSON, which is valid YAML and lets the self-contained CLI decode them without another runtime dependency.

Each adapter declares:

- `schema_version`: currently `1.0.0`;
- `id`: stable lowercase identifier;
- `kind`: `language` for maintained language adapters;
- `trust`: `builtin` for files compiled into this CLI;
- `language`: discovery language identifier;
- `markers`: project filenames that activate the adapter;
- `commands`: direct executable and argument-array proposals.

A command can add `when_any` to limit a proposal to one build system. For example, Java proposes Maven only for `pom.xml` and Gradle only for a Gradle build file. Discovery emits separate matches for nested project roots and assigns each proposal a repository-relative working directory and unique identifier.

## Trust boundary

Discovery returns adapter commands under `proposed_commands`, outside the repository policy. A person must add or approve commands in `.clean-code.json`, or provide a separately trusted policy to `clean-code verify`.

Adapters may contain direct executable arrays and bounded artifact declarations. Shell strings, installers, downloads, credentials, network permissions, and third-party executable adapters sit outside the built-in declarative trust tier.

## Artifact formats

Declared commands may validate bounded artifacts in these formats:

- `file`: existence and byte count;
- `text`: byte and line counts;
- `json`: flattened numeric leaf metrics;
- `xml`: flattened numeric text metrics;
- `sarif`: total results and error, warning, or note counts;
- `lcov`: line, branch, and function totals and percentages.

Set `max_bytes` when the default 10 MiB limit is too broad. Set `fresh: true` when the command must replace or update the artifact.

## Baselines

Baseline entries name an artifact metric, its prior value, whether higher or lower is preferable, a non-negative tolerance, and whether regression blocks the check. The CLI records producer-declared scope verbatim, so `scope: changed` requires a metric calculated from the changed files. Coverage, mutation, duplication, and complexity remain separate metrics.

Example:

```json
{
  "artifact": "coverage.lcov",
  "metric": "lines.percent",
  "scope": "repository",
  "direction": "higher",
  "value": 80,
  "tolerance": 0.5,
  "required": true
}
```

## Validation

Run:

```bash
go test ./harness/adapters ./internal/providers ./internal/discover
```

The catalog rejects unknown fields, unsupported versions, duplicate identifiers, and unsafe command shapes. Add detection tests for every new marker and conditional command.
