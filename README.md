# Clean Code

[![npm version](https://img.shields.io/npm/v/@shashanksn/clean-code)](https://www.npmjs.com/package/@shashanksn/clean-code)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Clean Code is an open-source (MIT) plugin for designing, building, testing, verifying, and shipping maintainable software with coding agents. It works across Codex, Cursor, Claude Code, Copilot, terminal agents, CI pipelines, and a standalone CLI.

Agents forget instructions, mirror mistakes in tests, and narrate success without proof. Clean Code pairs doctrine with **deterministic checks**, **independent test tracks**, **architecture constraints**, **evidence-based review**, **human spot checks**, and **immutable audit receipts**, plus a **full planning-to-PR skill pipeline**.

**Twenty skills**, a Go CLI, five language discovery adapters, generated host instructions, and calibration benchmarks ship in this repository.

**npm:** [@shashanksn/clean-code](https://www.npmjs.com/package/@shashanksn/clean-code) · CLI command: `clean-code`

---

## Install

### npm (recommended)

Published on npm as `@shashanksn/clean-code`. **Node 18+** and **Go 1.22+** are required to run the CLI; if they are missing, install bootstraps them automatically into `~/.clean-code-cli/runtime` (no `brew install` needed).

```bash
npm install -g @shashanksn/clean-code
clean-code version
clean-code setup --host cursor --output /path/to/your/repo
```

In a project:

```bash
npm install @shashanksn/clean-code
npx clean-code discover .
```

Skills install under `node_modules/@shashanksn/clean-code/skills/`.

> npm blocks the unscoped name `clean-code` (too similar to package `cleancode`). The scoped package installs the **`clean-code`** command.

### Go build (from source)

```bash
git clone https://github.com/shashank-stitch/clean-code.git
cd clean-code
go build -o clean-code ./cmd/clean-code
./clean-code version
```

### One-line install (no npm)

Bootstraps Node.js and Go when missing, then builds the CLI.

```bash
curl -fsSL https://raw.githubusercontent.com/shashank-stitch/clean-code/codex/initial-release/scripts/install.sh | bash
export PATH="$HOME/.local/bin:$PATH"
clean-code version
```

### Release binary

Checksummed binaries for macOS, Linux, and Windows: [GitHub Releases](https://github.com/shashank-stitch/clean-code/releases).

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
clean-build + clean-test + clean-verify
clean-review → clean-simplify → clean-ship → clean-watch-pr
clean-audit + clean-compound
```

| Phase | Skills | What happens |
| --- | --- | --- |
| **Plan** | `clean-brainstorm` → `clean-plan` | Scope and requirements, then implementation units and verification contract |
| **Build** | `clean-build` + `clean-test` + `clean-verify` | Small verified changes, independent test tracks, deterministic checks |
| **Ship** | `clean-review` → `clean-simplify` → `clean-ship` → `clean-watch-pr` | Evidence-based review, cleanup, PR, CI watch |
| **Record** | `clean-audit` + `clean-compound` | Immutable audit receipt and durable learnings |

Optional: `clean-setup`, `clean-discover`, `clean-design`, `clean-debug`, `clean-refactor`, `clean-worktree`, `clean-learn`, `clean-orchestrate`.

Autonomous end-to-end: invoke **`clean-lfg`** with your feature description.

---

## Skill map (20 skills)

| Skill | Responsibility |
| --- | --- |
| `clean-setup` | Host integration without changing repo policy |
| `clean-brainstorm` | Requirements-only plans |
| `clean-plan` | Implementation-ready units + verification contract |
| `clean-discover` | Read-only capability discovery |
| `clean-design` | Use cases, boundaries, acceptance, architecture policy |
| `clean-build` | Small verified implementation steps |
| `clean-refactor` | Behavior-preserving structure improvements |
| `clean-debug` | Causal-chain debugging |
| `clean-test` | Independent unit, acceptance, integration, UI/QA tracks |
| `clean-verify` | Deterministic checks + normalized evidence |
| `clean-review` | Evidence-based review; zero findings allowed |
| `clean-simplify` | Behavior-preserving simplification |
| `clean-ship` | Commit, push, PR |
| `clean-watch-pr` | CI watch loop |
| `clean-orchestrate` | Multi-role coordination |
| `clean-lfg` | Full autonomous pipeline |
| `clean-audit` | Immutable release receipts |
| `clean-learn` | Proposal-only policy learning |
| `clean-compound` | `docs/solutions/` learnings |
| `clean-worktree` | Isolated worktrees |

---

## CLI commands

```bash
clean-code version
clean-code hosts
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

See [commands](docs/commands.md), [configuration](docs/configuration.md), and [adapter authoring](docs/adapter-authoring.md).

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
- Mutation, complexity, duplication, and coverage stay **separate evidence** — no universal cleanliness score.
- Acceptance and UI/QA checked **independently** from implementation.
- Human spot checks recorded explicitly.
- Review findings require evidence; **zero findings is valid**.

---

## Repository layout

```text
skills/           # 20 agent skills
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
