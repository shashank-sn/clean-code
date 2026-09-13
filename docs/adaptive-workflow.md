# Adaptive workflow

Clean Code can select a bounded workflow shape from task signals without weakening authorization, evidence, or human gates.

## Contracts

| Concern | Skill / CLI | Artifact |
| --- | --- | --- |
| Entry routing | `clean-route` / `clean-code route` | adaptive route decision |
| Competing design | `clean-arena` / `clean-code arena validate` | arena decision record |
| Adversarial probes | `clean-probe` / `clean-code probe` | adversarial probe plan |
| Parallel ownership | `clean-orchestrate` / `clean-code parallel validate` | parallel orchestration record |
| Playbooks | `clean-code playbook list\|show` | `harness/playbooks/*.json` |

Schemas live under `harness/schemas/` (`adaptive-route`, `arena-decision`, `adversarial-probe`, `parallel-orchestration`, `playbook`).

## Rules that do not move

- Route decisions are advisory unless an existing lifecycle contract authorizes execution.
- Never select a model by brand. Never write host or global configuration.
- Prefer `NOT_AVAILABLE`, `NOT_CONFIGURED`, and `NOT_RUN` over invented capability.
- Arena records set `policy_change` to `none`.
- Probe failures feed bounded repair/reverify, not automatic learning.
- Parallel children need owned scope, evidence requests, mutation boundaries, and verified artifacts before synthesis treats them as proof.
- Human approval remains required for release, policy, permissions, and external actions.

## Calendar validity

`clean-code probe calendar --date YYYY-MM-DD` rejects impossible dates such as `2026-02-30`. Native `Date` normalization in some languages can accept that string; the Clean Code probe must not.

## Playbooks

Builtin playbooks: bug investigation, feature delivery, design decision, refactor, code review, and release readiness. Each states when it applies, minimum roles and evidence, whether arena or probes are warranted, parallelism guidance, and terminal states `PASS` / `FAIL` / `NOT_RUN` / `NOT_AVAILABLE` / `BLOCKED`.
