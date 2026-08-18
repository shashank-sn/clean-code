# Clean Code

Clean Code is a language-neutral and host-neutral plugin for designing, building, testing, verifying, and shipping maintainable software with coding agents. It works across Codex, Cursor, Claude Code, Copilot, terminal agents, CI pipelines, and a standalone CLI.

Agents forget instructions, mirror mistakes in tests, and narrate success without proof. Clean Code pairs doctrine with **deterministic checks**, **independent test tracks**, **architecture constraints**, **evidence-based review**, **human spot checks**, and **immutable audit receipts** — then adds a **full planning-to-PR pipeline** comparable to Compound Engineering, with stronger verification gates.

**Twenty skills**, a Go CLI, five language discovery adapters, generated host instructions, calibration benchmarks, and CE workflow comparison tooling ship in this repository.

---

## Quick start

```bash
go build -o clean-code ./cmd/clean-code
clean-code setup --host cursor --output /path/to/your/repo
cp harness/config/defaults.clean-code.json /path/to/your/repo/.clean-code.json   # edit commands
clean-code discover /path/to/your/repo
clean-code verify --allow-repository-policy /path/to/your/repo
```

For autonomous delivery, invoke the **`clean-lfg`** skill with your feature description (see [Autonomous pipeline](#autonomous-pipeline-clean-lfg)).

---

## Install

Build locally:

```bash
go build -o clean-code ./cmd/clean-code
```

Or download a checksummed release binary for macOS, Linux, or Windows (no runtime after build).

Generate host instructions:

```bash
clean-code setup --host cursor --output /path/to/repository
```

Existing instruction files are never overwritten. Unknown hosts receive the portable `AGENTS.md` fallback. See the [compatibility matrix](docs/host-compatibility.md).

---

## Step-by-step workflow (manual)

Use this when you want control at each gate. Each step maps to a skill and optional CLI command.

### 1. Connect the host

| Step | Skill | Action |
| --- | --- | --- |
| Detect platform | `clean-setup` | Run `clean-code setup --host <id> --output <repo>` |
| Discover capabilities | `clean-discover` | Run `clean-code discover <repo>` (read-only, no command execution) |

### 2. Define what to build

| Step | Skill | Output |
| --- | --- | --- |
| Explore scope (if vague) | `clean-brainstorm` | Requirements-only plan in `docs/plans/` |
| Implementation plan | `clean-plan` | Units, verification contract, definition of done |
| Design boundaries | `clean-design` | Use cases, acceptance examples, architecture policy |

### 3. Implement with evidence

| Step | Skill | Rule |
| --- | --- | --- |
| Isolate work (optional) | `clean-worktree` | Feature branch or worktree off default |
| Build in small steps | `clean-build` | One bounded behavior; cheapest feedback first |
| Independent tests | `clean-test` | Unit, acceptance, integration, UI/QA tracks — not implementation-shaped oracles |
| Refactor safely | `clean-refactor` | Green baseline; behavior changes recorded separately |
| Fix failures | `clean-debug` | Full causal chain before fixing |

### 4. Verify and review

| Step | Skill / CLI | Gate |
| --- | --- | --- |
| Run checks | `clean-verify` → `clean-code verify` | Trusted policy required; honest PASS/FAIL/NOT_RUN states |
| Architecture | `clean-code architecture` | Declared component graph vs policy |
| Traceability | `clean-code trace` | Requirements ↔ tests ↔ evidence |
| Review | `clean-review` → `clean-code review` | Evidence-backed findings; **zero findings allowed** |
| Simplify | `clean-simplify` | Behavior-preserving cleanup before ship |

### 5. Ship and record

| Step | Skill | Action |
| --- | --- | --- |
| Commit, push, PR | `clean-ship` | Conventional commits; PR lists verification summary |
| Watch CI | `clean-watch-pr` | Until green or documented blocker |
| Audit receipt | `clean-audit` → `clean-code audit` | Immutable receipt bound to revision |
| Capture learning | `clean-compound` | `docs/solutions/` + `CONCEPTS.md` |
| Policy proposals | `clean-learn` → `clean-code learn` | Proposal-only; cannot weaken hard gates |

### 6. Orchestrate multi-agent work

| Skill | When |
| --- | --- |
| `clean-orchestrate` | Multiple roles (spec, build, test, review) — procedural independence when host lacks subagents |
| `clean-lfg` | Hands-off version of the entire sequence (below) |

Canonical stage order: `harness/workflow/shipping-pipeline.json`. Details: [shipping pipeline](docs/shipping-pipeline.md).

---

## Autonomous pipeline (`clean-lfg`)

Invoke **`clean-lfg`** with a feature description when you want planning through PR without step-by-step check-ins. It runs:

1. `clean-brainstorm` (if needed) → `clean-plan`
2. `clean-design` (when boundaries change)
3. `clean-worktree` (when isolation helps)
4. `clean-build` per plan unit
5. `clean-test` — independent tracks
6. `clean-verify` — mandatory deterministic evidence
7. `clean-review` — apply blocking fixes
8. `clean-simplify` — unless docs-only/trivial
9. `clean-ship` — commit, push, PR (`mode:pipeline`)
10. `clean-watch-pr` — CI to green
11. `clean-audit` — receipt
12. `clean-compound` — durable learnings

**Stops** when mandatory verification fails, human spot checks are missing, or review finds unresolved blocking defects.

Compound Engineering equivalent: `lfg`. Clean Code adds verify, architecture, trace, audit, and learn gates CE does not enforce by default.

---

## Skill map (20 skills)

| Skill | Responsibility | CE equivalent |
| --- | --- | --- |
| `clean-setup` | Host integration without changing repo policy | `ce-setup` |
| `clean-brainstorm` | Requirements-only plans | `ce-brainstorm` |
| `clean-plan` | Implementation-ready units + verification contract | `ce-plan` |
| `clean-discover` | Read-only capability discovery | — |
| `clean-design` | Use cases, boundaries, acceptance, architecture policy | (in `ce-plan`) |
| `clean-build` | Small verified implementation steps | `ce-work` |
| `clean-refactor` | Behavior-preserving structure improvements | — |
| `clean-debug` | Causal-chain debugging | `ce-debug` |
| `clean-test` | Independent unit/acceptance/integration/UI tracks | `ce-test-browser` (partial) |
| `clean-verify` | Deterministic checks + normalized evidence | — |
| `clean-review` | Evidence-based review; zero findings OK | `ce-code-review` |
| `clean-simplify` | Behavior-preserving simplification | `ce-simplify-code` |
| `clean-ship` | Commit, push, PR | `ce-commit-push-pr` |
| `clean-watch-pr` | CI watch loop | `ce-babysit-pr` |
| `clean-orchestrate` | Multi-role coordination | — |
| `clean-lfg` | Full autonomous pipeline | `lfg` |
| `clean-audit` | Immutable release receipts | — |
| `clean-learn` | Proposal-only policy learning | — |
| `clean-compound` | `docs/solutions/` learnings | `ce-compound` |
| `clean-worktree` | Isolated worktrees | `ce-worktree` |

**Clean Code-only:** deterministic verification, architecture enforcement, trace validation, audit receipts, policy learning, human spot-check gates, host-neutral CLI.

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

Discovery reads `.clean-code.json` and never executes repo commands. Verification requires trusted policy or explicit `--allow-repository-policy`. See [commands](docs/commands.md), [configuration](docs/configuration.md), and [adapter authoring](docs/adapter-authoring.md).

Default policy template: `harness/config/defaults.clean-code.json`.

---

## Benchmarks

### Workflow coverage rubric (CE vs Clean Code)

Scores **workflow capability** on 18 dimensions (not agent performance):

```bash
go run ./cmd/clean-code compare-workflows
```

Latest rubric: **Clean Code ~96%** vs **Compound Engineering ~71%** — CC leads on verification, architecture, testing independence, audit, policy learning, human gates, and host neutrality.

### Full-flow code quality benchmark

Same small task, two outcomes, automated metrics + **blind independent reviewer**:

```bash
go test -race ./examples/benchmark-flow/...
go run ./cmd/clean-code benchmark-full-flow
```

Task: [slug normalizer](examples/benchmark-flow/task.md). CE-style vs CC-style implementations live under `examples/benchmark-flow/outcomes/`.

| Source | Winner | Notes |
| --- | --- | --- |
| Automated rubric | Clean Code | More tests, smaller functions, fuzz hardening |
| Blind reviewer (Gemini) | Clean Code (Outcome B) | Naming, simplicity, test quality, maintainability |

Details: [benchmark-full-flow](docs/benchmark-full-flow.md).

### Defect detection scorer

```bash
go run ./cmd/clean-code benchmark --manifest harness/calibration/benchmark-manifest.yaml
```

---

## Language support

Any repository can declare commands and artifacts. Universal support: orchestration, declared commands, path-based architecture rules, review, audit.

Maintained discovery adapters: **Go, Java, JavaScript/TypeScript, Python, Rust**. Missing tools report `NOT_AVAILABLE` or `NOT_CONFIGURED` — never silent pass.

---

## Platform and IDE support

Canonical skills + generated host instructions for Codex, Claude Code, Cursor, Copilot, Gemini CLI, Windsurf, Cline, Roo Code, and generic agents. Host adapters change invocation only — not doctrine, schemas, or gate semantics.

---

## Enforcement model

- Build, test, requirement, and architecture failures can **block** completion when configured.
- Mutation, complexity, duplication, and coverage stay **separate evidence** — no universal cleanliness score.
- Acceptance and UI/QA checked **independently** from implementation.
- Human spot checks recorded explicitly (including what was **not** inspected).
- Review findings require evidence; **clean review may return zero findings**.

---

## Repository layout

```text
skills/           # 20 agent skills (canonical source)
cmd/clean-code/   # Standalone CLI
internal/         # Runner, verify, audit, benchmark, …
harness/          # Schemas, adapters, doctrine, calibration, workflow
examples/         # generic + benchmark-flow fixtures
docs/             # Architecture, shipping, benchmarks, plans
hosts/generic/    # Portable AGENTS.md fallback
```

---

## Status

- **Done:** full skill suite, CLI, host generation, shipping pipeline, workflow + full-flow benchmarks, calibration fixtures, examples.
- **Open before public release:** controlled held-out **agent** study (rubrics measure workflow coverage and sample outcomes; they do not claim production agent uplift).

Implementation plan: [docs/plans/2026-08-12-001-feat-clean-code-system-plan.md](docs/plans/2026-08-12-001-feat-clean-code-system-plan.md).

---

## Inspiration

Robert C. Martin's *Clean Code* and *Clean Architecture* (names, functions, boundaries, tests, dependency direction); Martin's agent-supervision practices; Huolter's [clean-code-skill](https://github.com/huolter/clean-code-skill); Cucumber/Gherkin; PIT, StrykerJS, mutmut; PMD CPD, jscpd; Playwright.

Ideas are summarized in original language — not book reproduction or universal law.
