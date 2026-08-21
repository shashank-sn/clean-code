# Portable agents

Every `skills/clean-*/` directory contains a model-neutral `agent.json` beside its `SKILL.md`. The manifest is the portable contract; `SKILL.md` remains the instruction source. This lets a host generate native artifacts without changing the agent's evidence, permission, or stop rules.

```bash
clean-code agent list
clean-code agent validate
clean-code agent describe clean-build --host codex
clean-code agent emit clean-lfg --mode prompt --host generic
clean-code agent emit clean-build --mode json --host cursor --output clean-build.json
```

The runtime descriptor declares context capacity, filesystem mode, network policy, browser/UI capability, subagent isolation, session reset support, and structured-output support. It reports capabilities rather than assuming a model can edit files, execute commands, create subagents, or open a PR. Prompt-only hosts get the full instruction contract and must return `NOT_AVAILABLE`, `NOT_CONFIGURED`, `NOT_RUN`, `STALE`, or `ERROR` for unavailable operations.

The command never grants permissions. The host and repository policy remain the authority for file edits, commands, network access, Git writes, and PR creation.
