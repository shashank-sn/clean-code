---
name: clean-discover
description: Inspect a repository read-only to identify languages, project metadata, configuration, and available paths for deterministic quality checks.
---

# Clean Discover

Establish what a repository can verify before selecting tools or gates.

## Workflow

1. Run `clean-code discover <repo>`.
2. Treat the returned language list as discovery evidence, not a restriction on support.
3. Read `.clean-code.json` when present and validate every declared command.
4. Report generic command support for every repository.
5. Mark richer checks `NOT_AVAILABLE` or `NOT_CONFIGURED` until a compatible installed tool and policy exist.
6. Propose configuration changes for human review instead of applying them.

## Safety

- Discovery is read-only and does not execute project commands.
- Never install dependencies or access the network during discovery.
- Keep executable and arguments separate. Shell mode requires explicit repository policy.
