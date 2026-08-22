# Changelog

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
