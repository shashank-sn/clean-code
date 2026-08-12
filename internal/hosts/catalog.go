package hosts

import "sort"

type Capabilities struct {
	ID                string `json:"id"`
	DisplayName       string `json:"display_name"`
	NativeSkills      bool   `json:"native_skills"`
	Subagents         bool   `json:"subagents"`
	Hooks             bool   `json:"hooks"`
	BlockingApprovals bool   `json:"blocking_approvals"`
	CommandExecution  bool   `json:"command_execution"`
	CLI               bool   `json:"cli"`
	Integration       string `json:"integration"`
}

var catalog = map[string]Capabilities{
	"generic": {
		ID: "generic", DisplayName: "Generic coding environment", CLI: true,
		Integration: "portable Markdown instructions and standalone CLI",
	},
	"codex": {
		ID: "codex", DisplayName: "Codex", NativeSkills: true, Subagents: true,
		Hooks: true, BlockingApprovals: true, CommandExecution: true, CLI: true,
		Integration: "native skills with AGENTS.md fallback",
	},
	"claude-code": {
		ID: "claude-code", DisplayName: "Claude Code", NativeSkills: true, Subagents: true,
		Hooks: true, BlockingApprovals: true, CommandExecution: true, CLI: true,
		Integration: "native skills and agent instructions",
	},
	"cursor": {
		ID: "cursor", DisplayName: "Cursor", CommandExecution: true, CLI: true,
		Integration: "generated rules and portable instructions",
	},
	"copilot": {
		ID: "copilot", DisplayName: "GitHub Copilot", CLI: true,
		Integration: "repository instructions and custom agent definitions",
	},
	"gemini-cli": {
		ID: "gemini-cli", DisplayName: "Gemini CLI", CommandExecution: true, CLI: true,
		Integration: "terminal-agent instructions and standalone CLI",
	},
	"ide-agent": {
		ID: "ide-agent", DisplayName: "IDE coding agent", CLI: true,
		Integration: "generated rules or portable instructions",
	},
}

func Resolve(id string) Capabilities {
	if capabilities, ok := catalog[id]; ok {
		return capabilities
	}
	return catalog["generic"]
}

func Catalog() []Capabilities {
	ids := make([]string, 0, len(catalog))
	for id := range catalog {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]Capabilities, 0, len(ids))
	for _, id := range ids {
		result = append(result, catalog[id])
	}
	return result
}
