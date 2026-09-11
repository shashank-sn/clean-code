# Command reference

- `version`: print the CLI version.
- `hosts`: list maintained host capability records.
- `agent list|validate [ID]|describe ID [--host ID]|emit ID --mode prompt|json [--host ID] [--output FILE]`: inspect, validate, or emit one portable agent contract.
- `provider validate --manifest FILE|result --input FILE`: validate an offline, trusted-policy provider contract or revision-bound provider result.
- `gauntlet plan|run --manifest FILE --output DIR [--repo ROOT]`: create role packets, or validate each stage's expected artifacts and freshness against the repository (portable only; full multi-role execution requires a host adapter). Exit `0` = all PASS, `1` = any FAIL, `2` = any stage NOT_RUN.
- `setup --host ID [--output DIR]`: report capabilities and optionally create non-overwriting host instructions.
- `discover [REPO]`: inspect metadata and propose commands without execution.
- `verify [--trusted-policy FILE | --allow-repository-policy] [--output DIR] [REPO]`: execute approved checks and normalize evidence. When `--allow-repository-policy` is set, a prominent warning names the policy file and the number of UNapproved repository-declared commands that will be executed, since running them can execute arbitrary code; only use it on repositories you trust.
- `architecture --policy FILE --graph FILE`: enforce declared component directions, public surfaces, exclusions, exceptions, and cycles.
- `architecture view --policy FILE --graph FILE [--previous FILE] [--producer ID --collection-scope TEXT --evidence-file FILE --evidence-sha256 DIGEST]`: emit a local architecture graph view and optional edge diff. Coverage remains incomplete unless a producer proves its collection scope with a graph-bound evidence file.
- `trace --plan FILE`: validate requirement examples and unit, acceptance, integration, and UI/QA tracks.
- `review --input FILE`: validate independent evidence-backed findings; zero findings are valid.
- `audit --input FILE --output NEW_FILE [--signing-key HEX | --signing-key-file FILE | $CLEAN_CODE_SIGNING_KEY] [--witness-file FILE]`: create a revision-bound audit receipt. Use `--check RECEIPT [--witness-file FILE]` in place of `--output` to detect receipt or evidence changes. See **Audit receipts and the threat model** below.
- `keygen [--output DIR]`: generate an Ed25519 signing keypair and print the private seed, public key, and key fingerprint. `--output` also writes `signing-key.seed` (private, mode 0600) and `signing-key.pub`.
- `sign-role --role implementer|reviewer|spot_checker --revision REV --evidence FILE --signing-key HEX`: create a per-role Ed25519 attestation over `role + revision + SHA-256(evidence file)` for the audit input's `signers` map.
- `benchmark --manifest FILE`: report detection, false positives, misses, and correct silence.
- `compare-workflows [--manifest FILE]`: score Clean Code vs Compound Engineering workflow coverage on a fixed rubric.
- `benchmark-full-flow [--manifest FILE] [--repo ROOT]`: run full-flow code-quality benchmark on CE vs CC sample outcomes.
- `learn --proposal FILE`: validate that a policy proposal is reversible, independently reviewed when decided, and unable to suppress protected gates. Bottom-up proposals additionally require distinct supporting evidence, a clean control, passing held-out evidence, false-positive cost, rollback, and an independent human approver when approved.

All report-producing commands write JSON to standard output and diagnostics to standard error. Usage errors return 2; failed checks or invalid input return 1.

## Audit receipts and the threat model

An audit receipt is **tamper-evident, not tamper-proof**. Its integrity comes from three separate controls; each closes a different hole, and none is sufficient alone:

1. **Signature (immutability against editing).** `audit --signing-key ...` embeds an Ed25519 signature, the signing public key, its fingerprint, and a payload SHA-256 in the receipt. `--check` verifies them, so *any byte change to a signed field fails verification*. The private seed must be held by the auditor (or a trusted party), **not** by the person whose work is being audited — otherwise the receipt proves nothing. Store the seed outside the repository (CI secret, keychain, hardware key).
2. **Role signers (independence).** The audit input may carry a `signers` map produced by `clean-code sign-role`. Each of `implementer`, `reviewer`, and `spot_checker` attests its own evidence file (the verification report, the review input, the spot check). The receipt records them, requires `implementer` and `reviewer` to be present, and rejects any two roles signed by the same key. The auditor attests by signing the receipt itself, and the auditor key must differ from every role signer key. Independence is therefore **machine-verifiable**: a single key holder cannot certify two roles.
3. **External witness (immutability against regeneration).** `--witness-file` appends `SHA-256  <receipt path>` to an append-only file. `--check --witness-file` recomputes the on-disk hash and compares it. Even a *correctly re-signed* receipt fails the witness check if its bytes differ from what was witnessed. Keep the witness file somewhere the auditee cannot rewrite (a different repo, an immutable CI artifact, an append-only log); the tool enforces the hash binding, you choose the durable channel.

**What this does not guarantee.** The TOFU policy model remains: `verify --allow-repository-policy` executes whatever commands a repository's `.clean-code.json` declares, and the checks run on a machine the auditee controls. A signer who holds every key, or a CI pipeline the same operator controls end to end, can still produce a consistent-looking receipt. The controls above make that *detectable from outside* (via the witness and distinct-key records) but they do not make it impossible. Use audit receipts as a forcing function for process and as tamper-evident evidence, not as a standalone security boundary.
