---
name: clean-refactor
description: Improve names, functions, duplication, responsibility placement, dependency direction, boundaries, and module structure through behavior-preserving changes. Use when code works but is hard to change, a new case exposes a missing abstraction, metrics identify review pressure, or architecture dependencies need correction.
---

# Clean Refactor

Improve one structural problem while preserving observable behavior.

## Workflow

1. Run the current unit, acceptance, integration, and architecture checks that protect the target behavior.
2. State one concrete smell and its change cost: duplication, mixed abstraction, hidden order, misplaced responsibility, boundary math, selector branching, or dependency direction.
3. Make one reversible move: rename, extract, move, introduce a compatibility bridge, replace a selector with polymorphism, or centralize boundary logic.
4. Run the smallest affected check after every move and the full configured verification before handoff.
5. Keep temporary awkwardness only while it enables the next safe step. Delete obsolete maps, branches, adapters, comments, and compatibility paths once migration finishes.
6. Separate any discovered behavior change into its own requirement and acceptance evidence.

## Guardrails

- Treat passing tests as refactoring confidence; design quality still requires review.
- Use complexity, duplication, coverage, and mutation results to select inspection targets. Explain the consequence before changing code.
- Keep reporting logic outside domain objects when moving it inward would couple policy to presentation.
- Preserve justified static utilities, formula literals, and fixed-family construction switches when they communicate the design clearly.
- Stop and report the baseline when the target checks start red.
