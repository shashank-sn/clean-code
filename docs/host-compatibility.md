# Host compatibility

`clean-code setup --host ID --output REPO` generates the native instruction filename where a stable repository convention exists. Every package carries the same workflow and CLI fallback.

| Host ID | Generated path | Native skills | Subagents | CLI fallback |
| --- | --- | --- | --- | --- |
| `codex` | `AGENTS.md` | yes | yes | yes |
| `claude-code` | `CLAUDE.md` | yes | yes | yes |
| `cursor` | `.cursor/rules/clean-code.mdc` | no claim | no claim | yes |
| `copilot` | `.github/copilot-instructions.md` | no claim | no claim | yes |
| `gemini-cli` | `GEMINI.md` | no claim | no claim | yes |
| `windsurf` | `.windsurf/rules/clean-code.md` | no claim | no claim | yes |
| `cline` | `.clinerules/clean-code.md` | no claim | no claim | yes |
| `roo-code` | `.roo/rules/clean-code.md` | no claim | no claim | yes |
| `ide-agent` | `CLEAN_CODE.md` | no claim | no claim | yes |
| unknown | `AGENTS.md` | no claim | no claim | yes |

Capability records describe what the workflow may use. Permission still comes from the user or repository policy. Platforms without subagents use separate sessions for implementation, acceptance work, and review and record that separation as procedural.

The generated paths follow current primary documentation for [Cursor rules](https://docs.cursor.com/context/rules), [GitHub Copilot instructions](https://docs.github.com/en/copilot/reference/custom-instructions-support), [Claude Code project instructions](https://code.claude.com/docs/en/memory), [Gemini CLI context files](https://github.com/google-gemini/gemini-cli), [Windsurf rules](https://docs.windsurf.com/windsurf/cascade/memories), [Cline rules](https://docs.cline.bot/customization/cline-rules), and [Roo Code custom instructions](https://roocodeinc.github.io/Roo-Code/features/custom-instructions/). Capability flags stay conservative when a feature depends on product version or runtime configuration.
