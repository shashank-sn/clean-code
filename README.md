# Clean Code

[![npm version](https://img.shields.io/npm/v/@shashanksn/clean-code)](https://www.npmjs.com/package/@shashanksn/clean-code)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Clean Code is an open-source (MIT) plugin for designing, building, testing, verifying, and shipping maintainable software with coding agents. It works across Codex, Cursor, Claude Code, Copilot, terminal agents, CI pipelines, and a standalone CLI.

Agents forget instructions, mirror mistakes in tests, and narrate success without proof. Clean Code pairs doctrine with **deterministic checks**, **independent test tracks**, **architecture constraints**, **evidence-based review**, **human spot checks**, and **tamper-evident, signed audit receipts**, plus a **full planning-to-PR skill pipeline**.

**Thirty-one skills**, a Go CLI, five language discovery adapters, generated host instructions, and calibration benchmarks ship in this repository.

**npm:** [@shashanksn/clean-code](https://www.npmjs.com/package/@shashanksn/clean-code) · CLI command: `clean-code`

---

## Install

### npm (recommended)

Published on npm as `@shashanksn/clean-code`. **Node 18+** is required for npm; **Go 1.22+** is bootstrapped automatically on first `clean-code` run if missing (into `~/.clean-code-cli/runtime`). No install scripts — works with npm's default `allowScripts` policy.

```bash
npm install -g @shashanksn/clean-code@latest
export PATH="$(npm prefix -g)/bin:$PATH"
clean-code version
clean-code setup --host cursor --output /path/to/your/repo
```

On macOS, if `clean-code` is not found after install, your npm global bin is not on `PATH`. Add once to `~/.zshrc`:

```bash
export PATH="$(npm prefix -g)/bin:$PATH"
```

Then run `clean-code doctor` to verify install and PATH.

First `clean-code` run may download Go and compile the CLI once (subsequent runs are instant).

> **Upgrading from 0.2.0:** reinstall with `@latest`. Scoped name is `@shashanksn/clean-code` (not `@shashank/clean-code`).

In a project:

```bash
npm install @shashanksn/clean-code
npx clean-code discover .
```

Skills install under `node_modules/@shashanksn/clean-code/skills/`.

**Troubleshooting:** If `clean-code version` shows `0.1.0-dev`, an old Go binary is first on your `PATH` (often `~/.local/bin/clean-code` from the curl installer). Run `which -a clean-code`, then:

```bash
rm -f ~/.local/bin/clean-code   # remove curl-installed shim
hash -r
npm install -g @shashanksn/clean-code@latest
clean-code version
```

> npm blocks the unscoped name `clean-code` (too similar to package `cleancode`). The scoped package installs the **`clean-code`** command.

### Go build (from source)

```bash
git clone https://github.com/shashank-sn/clean-code.git
cd clean-code
go build -o clean-code ./cmd/clean-code
./clean-code version
```

### One-line install (no npm)

Bootstraps Node.js and Go when missing, then builds the CLI.

```bash
curl -fsSL https://raw.githubusercontent.com/shashank-sn/clean-code/v0.7.0/scripts/install.sh | bash
export PATH="$HOME/.local/bin:$PATH"
clean-code version
```

### Release binary

Checksummed binaries for macOS, Linux, and Windows: [GitHub Releases](https://github.com/shashank-sn/clean-code/releases).

### First-time repo setup

```bash
clean-code setup --host cursor --output /path/to/your/repo
cp node_modules/@shashanksn/clean-code/harness/config/defaults.clean-code.json /path/to/your/repo/.clean-code.json
clean-code discover /path/to/your/repo
clean-code verify --allow-repository-policy /path/to/your/repo
```

Host instructions are never overwritten. Unknown hosts get `AGENTS.md`. See [host compatibility](docs/host-compatibility.md).

---

## Step-by-step

Default manual pipeline:

```
clean-brainstorm → clean-plan
clean-build + clean-test → clean-simplify → clean-prune
clean-verify → clean-review → clean-ship → clean-watch-pr
clean-audit → clean-eval-discover? → clean-learn? + clean-compound
```

| Phase | Skills | What happens |
| --- | --- | --- |
| **Plan** | `clean-brainstorm` → `clean-plan` | Scope and requirements, then implementation units and verification contract |
| **Build** | `clean-build` + `clean-test` → `clean-simplify` → `clean-prune` | Small changes, independent test tracks, behavior-preserving cleanup, then dead-code removal with a reference search behind every deletion |
| **Ship** | `clean-verify` → `clean-review` → `clean-ship` → `clean-watch-pr` | Final-revision evidence, review, plain-English PR, CI watch |
| **Record** | `clean-audit` → `clean-eval-discover`? → `clean-learn`? + `clean-compound` | Signed tamper-evident receipt, role-bound signers, external witness; evaluate only confirmed repeated outcomes |

Optional: `clean-setup`, `clean-discover`, `clean-design`, `clean-route`, `clean-arena`, `clean-probe`, `clean-debug`, `clean-refactor`, `clean-worktree`, `clean-show-me`, `clean-eval-discover`, `clean-learn`, `clean-orchestrate`.

`clean-eval-discover` and `clean-learn` are conditional record-stage roles. They do not run for every feature, and neither can activate a policy change.

Autonomous end-to-end: invoke **`clean-lfg`** with your feature description.

---

## Skill map (31 skills)

| Skill | Responsibility |
| --- | --- |
| `clean-setup` | Host integration without changing repo policy |
| `clean-brainstorm` | Requirements-only plans |
| `clean-plan` | Implementation-ready units + verification contract |
| `clean-discover` | Read-only capability discovery |
| `clean-design` | Use cases, boundaries, acceptance, architecture policy |
| `clean-route` | Adaptive playbook selection from task signals |
| `clean-arena` | Bounded competing-design decision records |
| `clean-probe` | Adversarial acceptance probes for medium/high risk |
| `clean-show-me` | Concise, evidence-bounded visual explanations |
| `clean-build` | Small verified implementation steps |
| `clean-refactor` | Behavior-preserving structure improvements |
| `clean-debug` | Causal-chain debugging |
| `clean-test` | Independent unit, acceptance, integration, UI/QA tracks |
| `clean-verify` | Deterministic checks + normalized evidence |
| `clean-review` | Evidence-based structural review; zero findings allowed |
| `clean-simplify` | Behavior-preserving simplification |
| `clean-prune` | Dead, unreachable, and leftover code removal before verification |
| `clean-ship` | Commit, push, plain-English bullet PR |
| `clean-watch-pr` | CI watch loop |
| `clean-orchestrate` | Multi-role coordination |
| `clean-lfg` | Full autonomous pipeline |
| `clean-audit` | Immutable release receipts |
| `clean-eval-discover` | Blinded evaluation-set discovery and candidate calibration |
| `clean-learn` | Proposal-only policy learning |
| `clean-compound` | `docs/solutions/` learnings |
| `clean-worktree` | Isolated worktrees |

### Specialist sub-agents

| Agent | Responsibility |
| --- | --- |
| `clean-reviewer` | Independent evidence-based review, author not approver |
| `clean-test-writer` | Independent test tracks that cannot mirror the implementation |
| `clean-auditor` | Immutable revision-bound receipt, decoupled from implementers |
| `clean-merge-resolver` | Intent-traced conflict resolution that finishes merges, never aborts |
| `clean-dispatcher` | Dispatch each delivery role to a dedicated isolated sub-agent |

---

## CLI commands

```bash
clean-code version
clean-code hosts
clean-code agent list
clean-code agent describe clean-build --host codex
clean-code agent emit clean-lfg --mode prompt --host generic
clean-code provider validate --manifest harness/providers/mutation/provider.json
clean-code gauntlet plan --manifest gauntlet.json --output .clean-code/packets
clean-code gauntlet run --manifest gauntlet.json --repo . --output .clean-code/gauntlet
clean-code setup --host codex [--output DIR]
clean-code discover [REPO]
clean-code verify [--trusted-policy FILE | --allow-repository-policy] [--output DIR] [REPO]
clean-code architecture --policy FILE --graph FILE
clean-code trace --plan FILE
clean-code review --input FILE
clean-code audit --input FILE --output RECEIPT.json
clean-code benchmark --manifest FILE
clean-code compare-workflows [--manifest FILE]
clean-code benchmark-full-flow [--manifest FILE] [--repo ROOT]
clean-code learn --proposal FILE
```

`gauntlet plan` creates role packets; `gauntlet run` is a portable artifact and freshness validation check (exit `0` = all stages PASS, `1` = any stage FAILS, `2` = any stage NOT_RUN because the portable core cannot execute agent stages). Full multi-role execution requires a host adapter — see [agent-quality gauntlet](docs/gauntlet.md).

See [commands](docs/commands.md), [configuration](docs/configuration.md), [portable agents](docs/portable-agents.md), [agent-quality gauntlet](docs/gauntlet.md), and [adapter authoring](docs/adapter-authoring.md).

---

## Benchmarks

```bash
clean-code compare-workflows
clean-code benchmark-full-flow
```

Details: [benchmark-full-flow](docs/benchmark-full-flow.md), [shipping pipeline](docs/shipping-pipeline.md).

---

## Language and platform support

Any repository can declare commands and artifacts. Maintained discovery adapters: **Go, Java, JavaScript/TypeScript, Python, Rust**.

Host instructions for Codex, Claude Code, Cursor, Copilot, Gemini CLI, Windsurf, Cline, Roo Code, and generic agents.

---

## Enforcement model

- Build, test, requirement, and architecture failures can **block** completion when configured.
- Dead code in the change set can **block** ship: `clean-prune` records the reference search behind every deletion and refuses to ship with an unresolved candidate.
- PR bodies must be readable by someone who has never seen the repository; `clean-ship` runs a plain-English self-check before opening.
- Mutation, complexity, duplication, and coverage stay **separate evidence** — no universal cleanliness score.
- Acceptance and UI/QA checked **independently** from implementation.
- Human spot checks recorded explicitly.
- Review findings require evidence; **zero findings is valid**.
- Structural review seeks removable complexity in changed code, but a line-count threshold, metric, or reviewer preference is never a finding by itself.
- Audit receipts are **tamper-evident, not tamper-proof**: an Ed25519 signature binds the receipt to a key you hold, role-bound signers make independence key-verifiable, and an external witness detects regeneration — but a party holding every key (or controlling the whole pipeline) can still produce a consistent-looking receipt. See `docs/commands.md` → *Audit receipts and the threat model* for the full trust boundary.

---

## Repository layout

```text
skills/           # 31 agent skills
cmd/clean-code/   # CLI
internal/         # Runner, verify, audit, benchmark
harness/          # Schemas, adapters, calibration
examples/         # Adoption and benchmark fixtures
docs/             # Architecture, shipping, benchmarks
package.json      # npm: @shashanksn/clean-code
```

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md). Licensed under [MIT](LICENSE).

---

## Inspiration

Robert C. Martin's *Clean Code* and *Clean Architecture* — names, functions, boundaries, tests, dependency direction, and agent supervision with deterministic evidence. Ideas are summarized in original language; this project does not reproduce book text or treat one author's preferences as universal law.
