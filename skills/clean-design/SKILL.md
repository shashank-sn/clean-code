---
name: clean-design
description: Turn requirements into use cases, stable policy boundaries, delivery and data boundaries, dependency direction, acceptance examples, and repository-owned architecture constraints. Use before implementing a feature, when splitting components, when a change crosses framework or vendor boundaries, or when architecture drift needs an explicit policy and graph check.
---

# Clean Design

Define enough design intent to guide implementation and deterministic checks.

## Workflow

1. Write the actor, desired outcome, failure behavior, and boundary cases in implementation-neutral language.
2. Identify the business rules that should remain stable when UI, framework, database, vendor, or transport choices change.
3. Name components from repository concepts and assign each one or more repository-relative paths.
4. Declare permitted dependency directions and public surfaces. Record exclusions and temporary exceptions with concrete reasons.
5. Ask the repository's language or build tool to produce a dependency graph in the shared JSON format.
6. Run `clean-code architecture --policy <policy.json> --graph <graph.json>` and treat every reported path as mechanical evidence.
7. Keep semantic design questions for review: responsibility placement, use-case cohesion, volatility, and whether the declared components represent useful concepts.

## Decisions

- Keep inner policy independent of UI, persistence, frameworks, vendors, and delivery mechanisms.
- Pass dependencies through explicit arguments, interfaces, factories, or returned values.
- Put behavior near the data and policy it interprets, unless that would pull presentation or vendor details inward.
- Split components by change pressure and release responsibility. Avoid folders created solely to resemble an architecture diagram.
- Keep architecture policy repository-owned. Discovery may propose boundaries, while a person approves them.

## Evidence limits

The graph checker proves only the supplied file edges against the supplied policy. It cannot prove that the graph producer found every dependency, that a component owns the right responsibility, or that passing structure makes the product correct. Report those limits with the mechanical result.

## Safety

- Reject absolute, escaping, ambiguous, and undeclared paths.
- Require a reason for every exception.
- Keep generated-code and test exceptions narrow and visible.
- Report an exact cycle path instead of a generic cycle count.
