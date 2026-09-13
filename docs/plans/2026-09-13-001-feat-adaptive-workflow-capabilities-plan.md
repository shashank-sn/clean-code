# Adaptive workflow capabilities plan

## Status

`implementation-ready`

## Outcome

Ship portable adaptive workflow contracts for consequential engineering work: entry routing, bounded competing-design records, adversarial acceptance probes, parallel orchestration with ownership, and developer playbooks. Preserve model-neutral contracts, authorization boundaries, revision-bound evidence, and human gates. Same PR updates `clean-ship` so PR descriptions use simple English bullets.

## Requirements

- `AW-1`: Deterministic adaptive entry router returns an inspectable route decision from task signals (change type, risk, ambiguity, boundaries, evidence needs, host capabilities).
- `AW-2`: Bounded arena / competing-design decision record with explicit trigger threshold, independent candidates, synthesis, uncertainty, and no automatic policy change.
- `AW-3`: Adversarial acceptance probes after implementation for medium/high risk; distinguish primary acceptance from post-hoc diagnostics; calendar-validity rejects impossible `YYYY-MM-DD` (e.g. `2026-02-30`).
- `AW-4`: Parallel orchestration records owned scopes, mutation boundaries, concurrency caps, and unavailable/failed children; synthesis checks claimed artifacts.
- `AW-5`: Concise playbooks for bug, feature, design, refactor, review, and release readiness with terminal states pass/fail/not run/not available/blocked.
- `AW-6`: `clean-ship` requires simple human-understandable English PR bodies with bullets for what changed, why, and how to verify.
- `AW-7`: Lifecycle/authorization equal or stricter; advisory routing unless an existing contract authorizes execution; no model-brand selection; no host/global config writes.

## Evidence class

| Item | Class |
| --- | --- |
| Need for adversarial calendar probes after impossible-date acceptance via native Date normalization | source-derived (observed shared failure class) |
| Adaptive routing / arena / parallel ownership patterns | hypothesis adapted into Clean Code contracts |
| Ship PR plain-English bullets | source-derived (explicit delivery ask) |

## Implementation units

### U1 — Typed schemas and fixtures

- **Files:** `harness/schemas/adaptive-route.schema.json`, `arena-decision.schema.json`, `adversarial-probe.schema.json`, `parallel-orchestration.schema.json`, `playbook.schema.json`; fixtures under `tests/fixtures/adaptive/`
- **Approach:** JSON Schema contracts matching Go types; fixtures for router choices, arena, probes (incl. calendar), partial children, host-capability fallback.
- **Verification:** schema JSON parses; fixture-driven Go tests.

### U2 — Deterministic adaptive package + CLI

- **Files:** `internal/adaptive/*.go`, `cmd/clean-code/main.go`, `docs/commands.md`
- **Approach:** Pure functions for `Route`, arena/probe/parallel/playbook validate; strict calendar date check; CLI: `route`, `arena`, `probe`, `parallel`, `playbook`.
- **Verification:** `go test ./internal/adaptive/...`; CLI smoke via tests.

### U3 — Skills and playbooks

- **Files:** `skills/clean-route/`, `skills/clean-arena/`, `skills/clean-probe/`; `harness/playbooks/*.json`; updates to `clean-orchestrate`, `clean-dispatcher`, `clean-lfg`, `harness/workflow/shipping-pipeline.json`
- **Approach:** Portable agent packages; playbooks as data routed by `clean-route`; orchestrate/dispatcher cite parallel contract.
- **Verification:** `clean-code agent validate`; shipping pipeline references registered agents.

### U4 — clean-ship PR guidance + docs

- **Files:** `skills/clean-ship/SKILL.md`, `skills/clean-ship/agent.json`, `docs/shipping-pipeline.md`, `README.md` (skill count / optional skills)
- **Approach:** Require plain-English bullet PR descriptions with verify/evidence section; keep evidence-linking rules.
- **Verification:** repository string assertions; manual PR body on this branch.

## Scope boundaries

- No model-brand routing, cloud/Slack assumptions, global config writes, or weakened audit/verify/review/release controls.
- No autonomous commit/push/deploy/policy/provider loop beyond existing authorization.
- No benchmark performance claims; no npm publish; no merge.

## Definition of done

Contracts, CLI, skills, playbooks, fixtures, and docs agree on AW-1..AW-7. Local `go test ./...` passes. Open PR closes #13 with plain-English bullets. Banned product name absent from branch diff.
