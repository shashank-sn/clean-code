# Changelog

## Unreleased

## 0.8.0 - 2026-09-17

- Add v2 review records that require requirements, scope coverage, revision-bound checks, and explicit limitations. Teach reviewers to trace affected call paths and verify agent claims; retain legacy inputs with unassessed completion.
- Add a portable 12-case review-evaluation pack with frozen reviewer bindings, independent before/after oracles, runner and scorer self-tests, and explicit separation between fixture validation and live model evaluation.
- Record the matched study result honestly: baseline and candidate each recovered 8/8 primary defects with all 4 clean controls silent; the release makes no recall-gain claim.
- Limit fixture Go execution to one processor so the concurrent-charge case reaches its expected assertion; retain runtime-panic rejection and the frozen study inputs.
- Align the npm and Codex plugin metadata and pin the one-line installer to the `v0.8.0` GitHub release tag.

## 0.7.0 - 2026-09-16

- Add `clean-prune`, a dead-code elimination stage that runs after simplification and before final verification: it lists detector candidates, records the reference search behind every deletion, deletes unused symbols, files, flags, and shims outright, and blocks ship while a candidate stays unresolved or a removal has no search behind it.
- Wire `clean-prune` into the canonical shipping pipeline, `clean-lfg`, the feature-delivery, refactor, and release-readiness playbooks, and the workflow comparison rubric (new `dead_code_elimination` dimension).
- Strengthen the `clean-ship` PR contract: one plain opening sentence, terms explained in everyday words on first use, real file and command names, a banned-word list, and a five-point self-check that must pass before a PR opens.
- Flag unselected registry, route, and dispatch entries when pruning, and check them against the configuration and fixtures that select them. Found by running the agent against a fixture with planted dead code: the first version reported nothing for two handler entries that no code or config selected.
- Point the one-line installer and README at the `v0.7.0` tag; both still referenced `v0.5.0`.

## 0.6.0 - 2026-09-13

- Add adaptive workflow contracts: deterministic entry routing, bounded competing-design arena records, adversarial acceptance probes (including strict calendar validity), parallel orchestration ownership records, and developer playbooks.
- Expose `clean-code route|arena|probe|parallel|playbook` plus portable skills `clean-route`, `clean-arena`, and `clean-probe`.
- Require `clean-ship` PR descriptions to use simple English bullets for what changed, why, how to verify, and gaps.
- Publish signed release evidence under `evidence/releases/0.6.0/` for revision `8bd4041` (complete audit receipt with role signers and witness).

## 0.5.0 - 2026-09-11

- Make audit receipts actually tamper-evident: Ed25519-signed with an embedded public key, so editing any signed field fails `audit --check`. Add `keygen` and `sign-role`.
- Add role-bound signers so independence is machine-verifiable: implementer, reviewer, and spot-checker each attest their own evidence with their own key, and the receipt rejects two roles signed by the same key or an auditor key reused for a role.
- Add an external witness (`audit --witness-file`) that binds the receipt hash to a channel outside the repo and catches even a correctly re-signed receipt whose bytes changed.
- Make `gauntlet run` an honest portable artifact + freshness check with documented exit codes (0 all PASS, 1 any FAIL, 2 any NOT_RUN) and reject symlinked paths that escape the repository root.
- Harden the npm-pack test so `go test ./...` passes regardless of installed npm version (no longer parses `npm pack --json` shape).
- Fail closed on toolchain downloads: SHA-256 verification for Go and Node in both the Node runtime and the bash fallback; pin the one-line installer to a release tag instead of a mutable branch.
- Report the real version on every launch path, drop the phantom `go.sum` from the npm package, and warn prominently before `verify --allow-repository-policy` executes repository-declared commands.
- Canonicalize the Go module path to `github.com/shashank-sn/clean-code`.

## 0.4.2 - 2026-08-26

- Add an evidence-based structural-simplification lens to `clean-review`, including code-judo-style questions about removable complexity, special-case growth, ownership boundaries, and decomposition.
- Make 1,000 changed source lines a contextual decomposition prompt, not a universal policy failure; structural findings require a concrete consequence and bounded behavior-preserving alternative.
- Run simplification before final verification and independent review so the review applies to the revision that is shipped.

## 0.4.1 - 2026-08-24

- Add `clean-eval-discover`, a conditional post-audit workflow role that turns confirmed outcomes into blinded, evidence-backed evaluation candidates without activating policy.
- Require bottom-up policy proposals to carry distinct supporting evidence, a clean control, held-out validation, false-positive cost, rollback, and independent human approval.
- Add portable eval schemas, workflow wiring, and regression coverage that prevents workflow stages from referencing unregistered agents.

## 0.4.0 - 2026-08-22

- Add five specialist sub-agents (`clean-reviewer`, `clean-test-writer`, `clean-auditor`, `clean-merge-resolver`, `clean-dispatcher`) that enforce role independence, intent-traced conflict resolution, and isolated-context dispatch.

## 0.3.0 - Unreleased

- Ship portable, model-neutral agent manifests for all Clean Code skills with host capability descriptors and CLI emitters.
- Add offline, trusted-policy provider contracts, an evidence-gated multi-role gauntlet, and architecture graph views with explicit coverage proof.
- Pack README-linked documentation and benchmark fixtures, then test the extracted npm artifact in CI.

## 0.2.8 - 2026-08-18

- Fix npm `repository`, `bugs`, and `homepage` URLs to `https://github.com/shashank-sn/clean-code`.

## 0.2.7 - 2026-08-18

- Add `clean-code doctor` to diagnose npm global `PATH` and binary location.
- README documents `export PATH="$(npm prefix -g)/bin:$PATH"` after global install on macOS.

## 0.2.6 - 2026-08-18

- Go CLI resolves version from nearby `package.json` when ldflags are not set.
- `version` warns when a curl-installed Go binary shadows the npm wrapper (`~/.local/bin/clean-code`).
- `install.sh` embeds package version in the Go binary at build time.

## 0.2.5 - 2026-08-18

- `clean-code version` reports the npm package version (e.g. `0.2.5`) instead of `0.1.0-dev` on global install.

## 0.2.4 - 2026-08-18

- GitHub Release workflow publishes checksumed binaries and triggers automated npm publish via `NPM_TOKEN`.
- npm publish skips when the package version is already on the registry.

## 0.2.3 - 2026-08-18

- Remove `postinstall` script so global install works without npm `allowScripts` / `--allow-scripts`.
- Bootstrap Go and build the native CLI lazily on first `clean-code` invocation instead.

## 0.2.2 - 2026-08-18

- Auto-install Node.js 20 and Go 1.22 when missing during npm postinstall and `install.sh`.
- Managed runtimes live under `~/.clean-code-cli/runtime` and are prepended to PATH for the CLI.

## 0.2.1 - 2026-08-18

- Publish on npm as `@shashanksn/clean-code` with polished README and install docs.
- Deprecate `clean-code-skills` on npm (renamed package).

## 0.2.0 - Unreleased

- Add shipping pipeline skills: `clean-brainstorm`, `clean-plan`, `clean-debug`, `clean-ship`, `clean-simplify`, `clean-compound`, `clean-worktree`, `clean-watch-pr`, and `clean-lfg`.
- Add `compare-workflows` CLI command and workflow coverage benchmark manifest vs Compound Engineering.
- Add `benchmark-full-flow` CLI with slug normalizer CE vs CC outcomes, automated rubric, and blind Gemini reviewer scores.
- Add npm package `clean-code` with CLI wrapper and CONTRIBUTING guide.
- Rewrite README: install via npm, step-by-step pipeline, CE mapping notes, remove status checklist.
- Add default policy template, generic example, benchmark-flow fixtures, CI template, and shipping pipeline docs.

## 0.1.0 - Unreleased

- Add eleven language-neutral skills for setup, discovery, design, build, refactor, test, verify, review, orchestration, audit, and learning.
- Add a standalone CLI with protected command execution, artifact baselines, architecture checks, traceability, review, immutable receipts, and benchmark scoring.
- Add maintained discovery adapters for Go, Java, JavaScript/TypeScript, Python, and Rust.
- Add generated instructions for major coding-agent hosts, IDE agents, terminals, automated pipelines, and an unknown-host fallback.
