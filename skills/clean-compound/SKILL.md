---
name: clean-compound
description: Document a solved problem as durable repo learning under docs/solutions and update CONCEPTS.md vocabulary when present. Use after shipping or fixing something worth remembering.
---

# Clean Compound

Capture knowledge while context is fresh so the next occurrence is faster.

## Workflow

1. Identify the single solved problem for this run (one learning per invocation).
2. Write a structured solution doc under `docs/solutions/` with YAML frontmatter: problem, root cause, fix, verification, and links.
3. If `CONCEPTS.md` exists, add or refine domain terms introduced or clarified by this work.
4. Cross-link to related solutions. Propose `clean-learn` policy updates only as proposals, never direct gate changes.
5. In pipeline mode, run non-interactively and report filed paths.

## Safety

- Do not weaken hard safety gates through compound docs.
- Solutions cite evidence, not agent narration alone.
