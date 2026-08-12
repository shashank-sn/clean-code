---
name: clean-audit
description: Produce an immutable, revision-bound release receipt from deterministic verification, independent test planning, review, human spot checks, policy identity, exceptions, and evidence hashes. Use before release or merge, when handing work to another team, or when a durable record of checked and unchecked scope is required.
---

# Clean Audit

Create a receipt that preserves both evidence and gaps.

## Workflow

1. Run final verification against the exact repository revision that will ship.
2. Complete test trace and independent review for that revision.
3. Record requirement, acceptance, UI/QA, and code-sample human checks with reviewer, scope, outcome, and reason for anything skipped.
4. Create an audit input that names repository identity, revision, trusted policy revision, evidence files, and approved exceptions.
5. Run `clean-code audit --input <audit-input.json> --output <new-receipt.json>`.
6. Read `complete`, `gaps`, evidence hashes, and exceptions before making a release decision.
7. Run `clean-code audit --input <audit-input.json> --check <receipt.json>` when receiving or reusing a receipt.

## Integrity

- Create a new receipt for every run. Existing receipts remain immutable.
- Keep evidence files local unless the user authorizes transmission.
- Bind verification, review, and spot checks to the audit revision.
- Require all configured human checks before completion.
- Preserve incomplete receipts and their gaps; they document real state even when release remains blocked.
- Rebuild the receipt after any evidence file changes because its hash will change.
