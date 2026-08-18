---
name: clean-plan
description: Enrich a requirements-only plan into an implementation-ready plan with units, verification contract, files, dependencies, and definition of done. Use after clean-brainstorm or when a feature needs structured breakdown before build.
---

# Clean Plan

Convert product intent into an executable implementation plan with traceable units and verification.

## Workflow

1. Read the input plan or feature description. If the plan is requirements-only, enrich it; if missing, ask for `clean-brainstorm` first.
2. Add Implementation Units with Goal, Files, Approach, Dependencies, Patterns to follow, Test scenarios, and Verification for each unit.
3. Define Verification Contract entries that map requirements to deterministic checks, tests, architecture rules, and human gates.
4. Record Definition of Done, Scope Boundaries, and Deferred to Implementation questions.
5. Set plan metadata `status: implementation-ready` when code execution can begin.
6. Hand off to `clean-build` or `clean-lfg` for execution.

## Architecture and evidence

- Declare architecture policy changes when units cross component boundaries.
- Specify which verification commands, test tracks, and spot checks gate each unit.
- Separate hard gates from review signals. Never collapse evidence into a cleanliness score.

## Safety

- Do not weaken trusted policy or approve new shell commands in the plan without explicit policy-delta approval.
- Keep plans as decision artifacts; progress lives in commits, not checkbox edits during execution.
