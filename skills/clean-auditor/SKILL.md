---
name: clean-auditor
description: Produce the immutable release receipt mapping requirements to evidence, decoupled from implementers so it cannot self-whitewash. Use for release qualification, final evidence reconciliation, and audit receipt generation.
---

# Auditor

Produce an immutable revision-bound receipt, or explicit gaps, decoupled from the implementer.

## Workflow

1. Consume the complete evidence set: verification report, architecture report, review input, test trace, and human spot checks.
2. Reconcile contradictions from source evidence; rerun stale checks rather than trusting their recorded state.
3. Record every human spot check for configured requirement, acceptance, UI/QA, and code-sample boundaries. Never skip or auto-fill one.
4. Map each requirement to its evidence. Preserve approved exceptions and their reasons.
5. Run `clean-code audit --input <evidence> --output RECEIPT.json`. Preserve the exact status vocabulary and approved exceptions.
6. Confirm the receipt is immutable: the same evidence set reproduces the same receipt bytes.

## Guardrails

- Decouple audit from implementation. The auditor must not be the change author or share the author's context.
- Do not let a missing or stale piece of evidence become a silent pass; emit the explicit gap.
- Do not collapse evidence into a cleanliness score. The receipt lists requirement-to-evidence mappings; it does not grade the codebase.
- Keep unresolved blocking findings blocking; accepted risk cannot override them.

## Tool-free mode

- If `clean-code audit` is unavailable, produce the receipt template with the exact unavailable status. Never claim an audit action occurred that did not.
