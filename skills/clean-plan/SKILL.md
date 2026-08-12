---
name: clean-plan
description: Turn approved intent into an implementation-ready, verifiable plan. Use before multi-file work, architecture changes, risky fixes, migrations, or any task whose requirements, boundaries, and evidence ownership are not already explicit.
---

# Clean Plan

Produce the smallest plan an implementation agent can execute without inventing product decisions.

## Contract

Inputs are approved requirements, constraints, repository context, and decision authority. Output is a revisionable plan with stable requirement IDs, acceptance examples, architecture boundaries, implementation units, evidence owners, and stop conditions.

## Workflow

1. Separate facts, settled decisions, assumptions, and unresolved authority.
2. Map each requirement to observable acceptance examples and forbidden outcomes.
3. Choose the simplest dependency structure that protects higher-level policy from details.
4. Divide work into bounded units with files or surfaces, dependencies, tests, and verification.
5. Assign implementation, acceptance, review, and approval to distinct run identities.
6. Record what would invalidate the plan and who can decide each exception.

## Evidence

A plan is ready only when every requirement maps to a unit and verification gate, every unit names its allowed scope, and no implementation preference is disguised as product authority.

## Stop conditions

Stop for missing product authority, contradictory requirements, unknown destructive scope, or a plan that requires one actor to author and approve the same claim.
