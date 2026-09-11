# Agent-quality gauntlet

The gauntlet turns a bounded story into revision-bound packets for the `specifier`, `implementer`, `cleaner`, `hardener`, `qa`, and `reviewer` roles.

```bash
clean-code gauntlet plan --manifest gauntlet.json --output .clean-code/packets
clean-code gauntlet run --manifest gauntlet.json --output .clean-code/gauntlet
```

Each packet has requirement IDs, allowed files, public contracts, source context, expected artifacts, evidence dependencies, ownership, and a stop condition. `plan` creates packets and never touches the repository. `run` is a portable artifact-validation check: for every stage it verifies that each `expected_artifacts` file exists under the repository root (`--repo`, default `.`), is a regular file (not a symlink), and was modified at or after the manifest `revision` commit time (freshness is skipped when the revision cannot be resolved). Stages that declare no expected artifacts are reported `NOT_RUN`. The portable core never executes agent stages — `mechanical`/`native-host` stages are agent work — so full multi-role execution requires a host adapter. Stages are marked `mechanical`, `native-host`, or `procedural`; procedural isolation is never reported as enforced independence.

`gauntlet run` exit codes:

- `0` when every stage PASSES artifact validation;
- `1` when any stage FAILS validation;
- `2` when any stage could not be validated (`NOT_RUN`), with a message pointing at host adapters.

Telemetry stays append-only in the manifest. It records turns, retries, failing-check cycles, repeated file edits, and budget exhaustion. A story resolves to `continue`, `revise_plan`, `reorganize_architecture`, or `stop_escalate`. Reorganization emits a refactor decision packet; feature expansion remains blocked until a person approves it or records an explicit deferred risk.

Provider contracts live in `harness/providers/`. Validate a provider contract or its revision-bound result before adding it to trusted repository policy:

```bash
clean-code provider validate --manifest harness/providers/mutation/provider.json
clean-code provider result --input reports/mutation-result.json
```

Provider contracts are offline, non-installing, and require trusted policy. Missing tools stay `NOT_AVAILABLE`; they never become a synthetic pass.
