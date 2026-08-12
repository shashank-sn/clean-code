---
name: clean-discover
description: Inspect a repository read-only to identify languages, project metadata, configuration, and available paths for deterministic quality checks.
---

# Clean Discover

Establish what a repository can verify before selecting tools or gates.

## Workflow

1. Run `clean-code discover <repo>`.
2. Treat the returned language list solely as discovery evidence; support remains language neutral.
3. Read `.clean-code.json` when present and validate every declared command.
4. Report generic command support for every repository.
5. Report matching built-in adapters and their commands as proposals only.
6. Mark richer checks `NOT_AVAILABLE` or `NOT_CONFIGURED` until a compatible installed tool and approved policy exist.
7. Propose configuration changes for human review instead of applying them.

## Safety

- Discovery performs read-only inspection and skips project command execution.
- Never install dependencies or access the network during discovery.
- Keep adapter proposals outside executable policy until a separate trust decision approves them.
- Keep executable and arguments separate. Shell mode requires explicit repository policy.
