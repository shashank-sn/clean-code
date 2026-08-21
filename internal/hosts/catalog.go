package hosts

import "sort"

type Capabilities struct {
	ID                string `json:"id"`
	DisplayName       string `json:"display_name"`
	NativeSkills      bool   `json:"native_skills"`
	Subagents         bool   `json:"subagents"`
	Hooks             bool   `json:"hooks"`
	BlockingApprovals bool   `json:"blocking_approvals"`
	FileEdits         bool   `json:"file_edits"`
	CommandExecution  bool   `json:"command_execution"`
	BrowserAutomation bool   `json:"browser_automation"`
	BackgroundTasks   bool   `json:"background_tasks"`
	CLI               bool   `json:"cli"`
	RepositoryRead    bool   `json:"repository_read"`
	ContextCapacity   string `json:"context_capacity"`
	FilesystemMode    string `json:"filesystem_mode"`
	NetworkPolicy     string `json:"network_policy"`
	SubagentIsolation bool   `json:"subagent_isolation"`
	SessionReset      bool   `json:"session_reset"`
	StructuredOutput  bool   `json:"structured_output"`
	Integration       string `json:"integration"`
}

var catalog = map[string]Capabilities{
	"generic": {
		ID: "generic", DisplayName: "Generic coding environment", CLI: true,
		ContextCapacity: "host-defined", FilesystemMode: "host-defined", NetworkPolicy: "host-defined",
		Integration: "portable Markdown instructions and standalone CLI",
	},
	"codex": {
		ID: "codex", DisplayName: "Codex", NativeSkills: true, Subagents: true,
		BlockingApprovals: true, FileEdits: true, CommandExecution: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "sandboxed", NetworkPolicy: "approval-gated",
		SubagentIsolation: true, SessionReset: true, StructuredOutput: true,
		Integration: "native skills with AGENTS.md fallback",
	},
	"claude-code": {
		ID: "claude-code", DisplayName: "Claude Code", NativeSkills: true, Subagents: true,
		Hooks: true, BlockingApprovals: true, FileEdits: true, CommandExecution: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "host-defined", NetworkPolicy: "approval-gated",
		SubagentIsolation: true, StructuredOutput: true,
		Integration: "native skills and agent instructions",
	},
	"cursor": {
		ID: "cursor", DisplayName: "Cursor", FileEdits: true, CommandExecution: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "workspace", NetworkPolicy: "host-defined",
		Integration: "generated rules and portable instructions",
	},
	"copilot": {
		ID: "copilot", DisplayName: "GitHub Copilot", FileEdits: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "workspace", NetworkPolicy: "host-defined",
		Integration: "repository instructions and custom agent definitions",
	},
	"gemini-cli": {
		ID: "gemini-cli", DisplayName: "Gemini CLI", BlockingApprovals: true, FileEdits: true,
		CommandExecution: true, BackgroundTasks: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "workspace", NetworkPolicy: "approval-gated",
		StructuredOutput: true,
		Integration:      "terminal-agent instructions and standalone CLI",
	},
	"ide-agent": {
		ID: "ide-agent", DisplayName: "IDE coding agent", FileEdits: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "workspace", NetworkPolicy: "host-defined",
		Integration: "generated rules or portable instructions",
	},
	"windsurf": {
		ID: "windsurf", DisplayName: "Windsurf", FileEdits: true, CommandExecution: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "workspace", NetworkPolicy: "host-defined",
		Integration: "generated workspace rules and standalone CLI",
	},
	"cline": {
		ID: "cline", DisplayName: "Cline", FileEdits: true, CommandExecution: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "workspace", NetworkPolicy: "host-defined",
		Integration: "generated workspace rules and standalone CLI",
	},
	"roo-code": {
		ID: "roo-code", DisplayName: "Roo Code", FileEdits: true, CommandExecution: true, CLI: true, RepositoryRead: true,
		ContextCapacity: "host-defined", FilesystemMode: "workspace", NetworkPolicy: "host-defined",
		Integration: "generated workspace rules and standalone CLI",
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
