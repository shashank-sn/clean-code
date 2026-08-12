# Clean Code

Clean Code is a planned, language-neutral and host-neutral plugin for designing, building, testing, and reviewing maintainable software with coding agents. The same system will work through coding platforms, IDE agents, terminal agents, automated pipelines, and a standalone CLI.

Written rules alone are weak enforcement. An agent can forget instructions, misunderstand a requirement, write tests that repeat the same mistake, or report that its own work is clean. This project pairs guidance with deterministic checks, independent test tracks, architecture constraints, evidence-based review, and recorded human spot checks.

The repository contains a working foundation and the complete implementation plan. Setup, discovery, and verification are implemented with shared contracts, a doctrine catalog, a host capability model, protected command execution, policy comparison, artifact validation, and revision-bound evidence. Architecture enforcement, orchestration, audit, and the remaining skills are still planned.

## Current commands

```bash
go run ./cmd/clean-code version
go run ./cmd/clean-code hosts
go run ./cmd/clean-code setup --host codex
go run ./cmd/clean-code discover /path/to/repository
go run ./cmd/clean-code verify --allow-repository-policy --output /path/to/evidence /path/to/repository
go run ./cmd/clean-code verify --trusted-policy /path/to/approved.clean-code.json /path/to/repository
go run ./cmd/clean-code architecture --policy /path/to/policy.json --graph /path/to/graph.json
go run ./cmd/clean-code trace --plan /path/to/test-plan.json
```

Discovery reads project metadata and an optional `.clean-code.json`. It never runs the repository's commands. Built-in adapters for JavaScript/TypeScript, Python, Java, Go, and Rust return read-only command proposals. Verification requires a trusted policy file or an explicit repository-policy approval, then runs each executable with its argument array, timeout, output limit, working directory, exit-code rules, redaction, and optional artifact checks. JSON, XML, SARIF, LCOV, text, and opaque artifacts can feed separate baseline comparisons. The normalized report stays tied to the current commit and dirty-worktree state. See `harness/config/example.clean-code.json` for the configuration shape and [adapter authoring](docs/adapter-authoring.md) for the extension contract.

## Skill map

| Skill | Responsibility |
| --- | --- |
| `clean-setup` | Detect the current coding host and install or generate the correct integration without changing project policy. |
| `clean-discover` | Detect repository structure, commands, languages, tools, and available quality signals. |
| `clean-design` | Turn requirements into use cases, boundaries, dependency rules, and acceptance examples. |
| `clean-build` | Guide small, verified implementation steps while protecting design boundaries. |
| `clean-refactor` | Improve structure through behavior-preserving changes and explicit safety checks. |
| `clean-test` | Coordinate independent unit, acceptance, integration, contract, and UI/QA test tracks. |
| `clean-verify` | Run configured deterministic checks and normalize their evidence. |
| `clean-review` | Review code and architecture from evidence, including returning zero findings when warranted. |
| `clean-orchestrate` | Separate specification, implementation, testing, verification, and review responsibilities. |
| `clean-audit` | Produce a traceable release receipt covering requirements, tests, checks, exceptions, and spot checks. |
| `clean-learn` | Calibrate rules from confirmed outcomes without weakening hard safety constraints. |

## Language support

The core works with any repository that can declare commands and expected artifacts. The target project may use any programming language. Universal support guarantees orchestration, declared commands, normalized evidence, path-based architecture rules, review, and audit. Parser-level analysis remains limited to installed tools and available adapters.

Maintained adapters will add richer discovery and result parsing for common toolchains. Missing tools remain visible as `NOT_AVAILABLE` or `NOT_CONFIGURED`; they are never reported as passing. Repository owners decide which checks are mandatory, advisory, or inapplicable.

## Platform and IDE support

Skills will have one canonical source and generated host adapters. Maintained integrations will cover major coding-agent hosts and IDEs, while `AGENTS.md`, portable Markdown instructions, and a self-contained `clean-code` CLI provide the fallback for any environment that can read instructions or run a supported executable.

Host adapters may change invocation and packaging, but they cannot change doctrine, evidence schemas, gate semantics, or result status. A platform that cannot support agents can still run discovery, verification, architecture checks, and audit generation through the CLI or an automated pipeline.

## Enforcement model

- Build, requirement, test, and declared architecture failures can block completion.
- Mutation, complexity, duplication, and coverage results remain separate evidence instead of collapsing into a universal cleanliness score.
- Universal design principles remain separate from language conventions and tool-specific rules.
- Acceptance tests and UI/QA procedures are checked independently from implementation.
- Human spot checks are recorded, including what was inspected and what was not.
- Review findings require concrete evidence. A clean review may return no findings.

## Inspiration

This project is informed by:

- Robert C. Martin's *Clean Code*, especially its focus on names, functions, boundaries, tests, successive refinement, and continuous cleanup.
- Robert C. Martin's *Clean Architecture*, especially use-case boundaries, dependency direction, component structure, testable policies, and keeping delivery mechanisms outside the core.
- Martin's public description of supervising coding agents with deterministic analysis, mutation testing, unit tests, executable acceptance tests, UI-level QA procedures, independent agents, and human spot checks.
- Huolter's [clean-code-skill](https://github.com/huolter/clean-code-skill), which demonstrates machine-readable heuristics, deterministic repository checks, calibrated review, honest unavailable states, and an evidence-or-silence reviewer.
- [Cucumber/Gherkin](https://cucumber.io/docs/gherkin/) for executable acceptance examples; [PIT](https://pitest.org/), [StrykerJS](https://stryker-mutator.io/docs/stryker-js/introduction/), and [mutmut](https://mutmut.readthedocs.io/) for mutation testing; [PMD Copy/Paste Detector](https://pmd.github.io/pmd/pmd_userdocs_cpd.html) and [jscpd](https://github.com/kucherenko/jscpd) for duplication analysis; and [Playwright](https://playwright.dev/docs/best-practices) for UI-level QA.

The project will summarize and operationalize ideas in original language. Book reproduction and the elevation of one author's preferences into universal law are outside its scope.

## Status

- Done: local Git repository, plugin manifest, product scope, and implementation plan.
- Done: `clean-setup`, `clean-discover`, generic host fallback, evidence schemas, initial doctrine rules, and tested Go CLI.
- Done: `clean-verify`, deterministic command execution, trusted-policy drift blocking, required artifact checks, and protected verification bundles.
- Done: maintained language discovery adapters, bounded artifact parsers, and metric-specific baseline regression gates.
- Done: `clean-design` and generic architecture enforcement for declared component dependencies, public surfaces, exceptions, exclusions, and cycles.
- Done: `clean-build`, `clean-refactor`, and `clean-test`, plus contracts and trace validation for independent unit, acceptance, integration, and UI/QA tracks.
- Next: add evidence-based review, orchestration, human spot checks, and audit receipts.

See [the implementation plan](docs/plans/2026-08-12-001-feat-clean-code-system-plan.md).
