# Release evidence for 0.6.0

Bound revision: `8bd4041d459662ec2b11c3f3594fb2a3f614e37f` (PR #14 feature tip).

- `audit-receipt.json` — signed, complete receipt (`complete: true`)
- `audit-input.json` — input with role signers
- `witness.jsonl` — external witness binding for `audit --check`
- `public-keys/` — role and auditor public keys (private seeds are not in git)

Verify:

```bash
go run ./cmd/clean-code audit \
  --input evidence/releases/0.6.0/audit-input.json \
  --check evidence/releases/0.6.0/audit-receipt.json \
  --witness-file evidence/releases/0.6.0/witness.jsonl
```
