---
name: clean-setup
description: Detect a coding host's capabilities and select a native or generic Clean Code integration without changing repository quality policy.
---

# Clean Setup

Connect the Clean Code workflow to the current coding platform, IDE, terminal agent, or automated pipeline.

## Workflow

1. Identify the host only from available environment evidence or an explicit user value.
2. Run `clean-code setup --host <id>` and read the capability result.
3. Use native skills when the capability result confirms them.
4. Use `hosts/generic/AGENTS.md` and CLI commands for unknown hosts or missing native features.
5. State which features are native, procedural, unavailable, or delegated to the CLI.
6. Validate that the selected instructions and CLI are discoverable.

## Safety

- Get permission before installing tools, enabling hooks, editing repository policy, or writing outside the approved integration path.
- When the host supports one context, use separate sessions and record that independence is procedural.
- Keep doctrine, status meanings, and evidence contracts identical across hosts.
