# Audit receipt checklist

Use `harness/schemas/audit-input.schema.json` as the input contract.

1. Bind `repository`, `revision`, and `policy_revision` to the final change.
2. List the final verification report, test plan, independent review, and human spot-check record.
3. Add every executed test track artifact to `supporting_evidence`.
4. Record approved exceptions without hiding gaps.
5. Run `clean-code audit --input audit-input.json --output receipt.json` once; receipts cannot be overwritten.

`complete: true` records that every configured link was present and valid. Untested behavior and general code quality remain outside that claim.
