---
name: clean-brainstorm
description: Explore vague product ideas into a requirements-only plan with scope boundaries, acceptance signals, and explicit non-goals. Use before clean-plan when behavior, users, or success criteria are still open.
---

# Clean Brainstorm

Turn an ambiguous request into durable product intent before implementation planning.

## Workflow

1. Classify scope as lightweight, standard, or deep. Ask one targeted question at a time when ambiguity blocks progress.
2. Scan the repository for adjacent patterns, existing plans, and constraints. Quote verified facts; label unverified assumptions.
3. Resolve product decisions here: actors, outcomes, boundaries, success criteria, and explicit non-goals.
4. Propose 2-3 approaches when multiple directions remain. State a recommendation with tradeoffs.
5. Write or update a requirements-only plan under `docs/plans/` with frontmatter `status: requirements-only` and date.
6. Include Goal Capsule, Product Contract sections, acceptance examples, and Outstanding Questions.
7. Hand off to `clean-plan` when the artifact is ready for implementation enrichment.

## Safety

- Do not implement code or choose libraries unless the brainstorm is explicitly about an architectural decision.
- Do not reproduce copyrighted book text. Summarize principles in original language.
- Record human spot-check needs when requirements affect user-visible behavior.

## Output

A markdown plan path the orchestrator can pass to `clean-plan`.
