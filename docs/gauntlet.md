# Agent-quality gauntlet

The gauntlet turns a bounded story into revision-bound packets for the `specifier`, `implementer`, `cleaner`, `hardener`, `qa`, and `reviewer` roles.

```bash
clean-code gauntlet plan --manifest gauntlet.json --output .clean-code/packets
clean-code gauntlet run --manifest gauntlet.json --output .clean-code/gauntlet
```

Each packet has requirement IDs, allowed files, public contracts, source context, expected artifacts, evidence dependencies, ownership, and a stop condition. `plan` creates packets. `run` records every portable-core stage as `NOT_RUN` and exits non-zero until a host executes the work and attaches trusted, revision-bound evidence. Stages are marked `mechanical`, `native-host`, or `procedural`; procedural isolation is never reported as enforced independence.

Telemetry stays append-only in the manifest. It records turns, retries, failing-check cycles, repeated file edits, and budget exhaustion. A story resolves to `continue`, `revise_plan`, `reorganize_architecture`, or `stop_escalate`. Reorganization emits a refactor decision packet; feature expansion remains blocked until a person approves it or records an explicit deferred risk.

Provider contracts live in `harness/providers/`. Validate a provider contract or its revision-bound result before adding it to trusted repository policy:

```bash
clean-code provider validate --manifest harness/providers/mutation/provider.json
clean-code provider result --input reports/mutation-result.json
```

Provider contracts are offline, non-installing, and require trusted policy. Missing tools stay `NOT_AVAILABLE`; they never become a synthetic pass.
