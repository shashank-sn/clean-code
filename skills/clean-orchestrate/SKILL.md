---
name: clean-orchestrate
description: Coordinate specification, implementation, acceptance, verification, review, decision, and delivery through explicit role identities and capability negotiation. Use for multi-part work, agent collaboration, and release qualification where correlated mistakes or stale evidence must be controlled.
---

# Clean Orchestrate

Build a role graph whose claims can be audited independently.

## Role graph

1. Product authority owns intent, policy, risk, and exceptions.
2. Specification owns stable requirements and acceptance examples.
3. Implementation owns the bounded code change.
4. Acceptance owns public-behavior evidence without implementation-shaped context.
5. Verification owns deterministic execution on the final revision.
6. Review owns risk-driven inspection and blocker disposition.
7. Delivery owns integration and destination evidence, never approval.

## Capability negotiation

Ask the host adapter which capabilities exist: isolated agents, separate sessions, deterministic commands, filesystem access, source control, automation watching, and approval capture. Record each as native, procedural, or unavailable. Missing capability changes the procedure, never the meaning of PASS or independence.

## Workflow

Create a handoff for every edge with requirement IDs, allowed scope, revision, evidence, context lineage, and stop conditions. Reconcile contradictions from source evidence. Re-run invalidated checks after any revision change. A same-identity or implementation-biased fallback may contribute evidence but cannot be labeled independent.

## Stop conditions

Stop on missing authority, self-approval, stale revision identity, incomplete required evidence, unapproved policy weakening, or unresolved blockers.
