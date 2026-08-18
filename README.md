# Clean Code

Clean Code is an open-source (MIT) plugin for designing, building, testing, verifying, and shipping maintainable software with coding agents. It works across Codex, Cursor, Claude Code, Copilot, terminal agents, CI pipelines, and a standalone CLI.

Agents forget instructions, mirror mistakes in tests, and narrate success without proof. Clean Code pairs doctrine with **deterministic checks**, **independent test tracks**, **architecture constraints**, **evidence-based review**, **human spot checks**, and **immutable audit receipts**, plus a **full planning-to-PR skill pipeline**.

**Twenty skills**, a Go CLI, five language discovery adapters, generated host instructions, and calibration benchmarks ship in this repository.

---

## Install

> **`clean-code-skills` is not on the public npm registry yet.** The commands below work today. After the first npm release, `npm install -g clean-code-skills` will work as documented.

### One-line install (macOS / Linux)

Requires **Go 1.22+** (`brew install go` on macOS).

```bash
curl -fsSL https://raw.githubusercontent.com/shashank-stitch/clean-code/cursor/ce-parity-shipping-pipeline-70e7/scripts/install.sh | bash
```

Then ensure `~/.local/bin` is on your `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
clean-code version
```

### Install from Git (npm, before registry publish)

Requires **Node 18+** and **Go 1.22+**.

```bash
npm install -g "git+https://github.com/shashank-stitch/clean-code.git#cursor/ce-parity-shipping-pipeline-70e7"
clean-code version
clean-code setup --host cursor --output /path/to/your/repo
```

Project-local:

```bash
npm install "git+https://github.com/shashank-stitch/clean-code.git#cursor/ce-parity-shipping-pipeline-70e7"
npx clean-code discover .
```

### npm registry (after first publish)

```bash
npm install -g clean-code-skills
clean-code version
```

Maintainers: set GitHub secret `NPM_TOKEN` and publish with [GitHub Releases](https://github.com/shashank-stitch/clean-code/releases) (workflow `npm-publish.yml`), or run `npm login` and `npm publish --access public` from a clone.

### Go build

```bash
git clone https://github.com/shashank-stitch/clean-code.git
cd clean-code
go build -o clean-code ./cmd/clean-code
./clean-code version
```

### Release binary

Download a checksummed binary for macOS, Linux, or Windows from [GitHub Releases](https://github.com/shashank-stitch/clean-code/releases). No Go install required after download.

### Host instructions

```bash
clean-code setup --host cursor --output /path/to/repository
```

Existing instruction files are never overwritten. Unknown hosts receive the portable `AGENTS.md` fallback. See the [compatibility matrix](docs/host-compatibility.md).

Copy a starter policy into your repo:

```bash
cp harness/config/defaults.clean-code.json /path/to/your/repo/.clean-code.json
```

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
| **Build** | `clean-build` + `clean-test` + `clean-verify` | Small verified changes, independent test tracks, deterministic checks (`clean-code verify`) |
| **Ship** | `clean-review` → `clean-simplify` → `clean-ship` → `clean-watch-pr` | Evidence-based review, behavior-preserving cleanup, PR, CI watch |
| **Record** | `clean-audit` + `clean-compound` | Immutable audit receipt and durable repo learnings |

Optional helpers: `clean-setup`, `clean-discover`, `clean-design`, `clean-debug`, `clean-refactor`, `clean-worktree`, `clean-learn`, `clean-orchestrate`.

Hands-off version of the same path: invoke **`clean-lfg`** with your feature description.

---

## Quick start

```bash
clean-code discover /path/to/your/repo
clean-code verify --allow-repository-policy /path/to/your/repo
```

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

**Clean Code-only capabilities:** `clean-verify`, `clean-audit`, `clean-learn`, architecture/trace CLI checks, and host-neutral evidence contracts.

---

## Compound Engineering mapping

Several Clean Code skills mirror stages found in **Compound Engineering** (a separate Cursor plugin). That mapping is **documentation and benchmark metadata only** — this project does **not** import, call, or require Compound Engineering at runtime.

| Clean Code skill | Analogous CE skill (if you use both) |
| --- | --- |
| `clean-brainstorm` | `ce-brainstorm` |
| `clean-plan` | `ce-plan` |
| `clean-build` | `ce-work` |
| `clean-debug` | `ce-debug` |
| `clean-review` | `ce-code-review` |
| `clean-simplify` | `ce-simplify-code` |
| `clean-ship` | `ce-commit-push-pr` |
| `clean-watch-pr` | `ce-babysit-pr` |
| `clean-compound` | `ce-compound` |
| `clean-worktree` | `ce-worktree` |
| `clean-lfg` | `lfg` |

**Where CE names appear in this repo (not in the Go CLI logic):**

- `harness/workflow/shipping-pipeline.json` — optional `ce_equivalent` labels on stages
- `harness/calibration/workflow-comparison.json` — rubric benchmark fixture
- `harness/calibration/full-flow-manifest.json` — sample outcome comparison
- `docs/shipping-pipeline.md`, `docs/benchmark-full-flow.md` — contributor docs

The `compare-workflows` and `benchmark-full-flow` commands read those JSON fixtures to score workflow coverage. No CE code is loaded.

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
go run ./cmd/clean-code compare-workflows
go run ./cmd/clean-code benchmark-full-flow
go test -race ./examples/benchmark-flow/...
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
package.json      # npm package clean-code-skills
```

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md). Licensed under [MIT](LICENSE).

---

## Inspiration

Robert C. Martin's *Clean Code* and *Clean Architecture* — names, functions, boundaries, tests, dependency direction, and agent supervision with deterministic evidence. Ideas are summarized in original language; this project does not reproduce book text or treat one author's preferences as universal law.
