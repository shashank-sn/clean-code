# Generic example repository

This fixture shows the minimum files to run Clean Code discovery and verification on any repository.

## Files

- `.clean-code.json` — trusted command policy (copy from `harness/config/defaults.clean-code.json` and adapt).
- Optional `.clean-code-architecture.json` — component dependency policy (see `docs/architecture.md`).
- Optional `docs/plans/` — requirements and implementation plans for `clean-brainstorm` and `clean-plan`.

## Quick start

```bash
clean-code discover .
clean-code verify --allow-repository-policy .
clean-code compare-workflows
```

## Shipping pipeline

Run the full planning-to-PR pipeline with the `clean-lfg` skill or follow stages in `harness/workflow/shipping-pipeline.json`.
