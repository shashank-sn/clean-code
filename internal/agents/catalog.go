// Package agents loads and renders model-neutral Clean Code agent packages.
package agents

import (
	"fmt"
	"sort"

	"clean-code/internal/hosts"
)

const SchemaVersion = "1.0.0"

var validPermissions = map[string]bool{
	"read_repository": true, "write_repository": true, "execute_commands": true,
	"network": true, "browser_automation": true, "git_write": true,
	"pull_request_write": true, "subagents": true,
}

var validUnavailableStatuses = map[string]bool{
	"NOT_AVAILABLE": true, "NOT_CONFIGURED": true, "NOT_RUN": true, "STALE": true, "ERROR": true,
}

type IO struct {
	Required []string `json:"required"`
	Optional []string `json:"optional"`
}

type ToolFreeMode struct {
	Available           bool     `json:"available"`
	Behavior            string   `json:"behavior"`
	UnavailableStatuses []string `json:"unavailable_statuses"`
}

type Descriptor struct {
	SchemaVersion        string       `json:"schema_version"`
	ID                   string       `json:"id"`
	Title                string       `json:"title"`
	Description          string       `json:"description"`
	InstructionFile      string       `json:"instruction_file"`
	Role                 string       `json:"role"`
	WorkflowPhase        string       `json:"workflow_phase"`
	Input                IO           `json:"input"`
	Output               IO           `json:"output"`
	EvidenceRequirements []string     `json:"evidence_requirements"`
	Permissions          []string     `json:"permissions"`
	StopConditions       []string     `json:"stop_conditions"`
	ToolFreeMode         ToolFreeMode `json:"tool_free_mode"`
	HandoffTo            []string     `json:"handoff_to"`
}

type Package struct {
	Descriptor   Descriptor `json:"descriptor"`
	Directory    string     `json:"-"`
	Instructions string     `json:"-"`
}

type RuntimeDescriptor struct {
	Agent                   Descriptor         `json:"agent"`
	Host                    hosts.Capabilities `json:"host"`
	AvailableCapabilities   []string           `json:"available_capabilities"`
	UnavailableCapabilities []string           `json:"unavailable_capabilities"`
	ExecutionMode           string             `json:"execution_mode"`
}

func Runtime(agent Descriptor, hostID string) RuntimeDescriptor {
	host := hosts.Resolve(hostID)
	available := make([]string, 0, len(agent.Permissions))
	unavailable := make([]string, 0, len(agent.Permissions))
	for _, permission := range agent.Permissions {
		if hostSupports(host, permission) {
			available = append(available, permission)
		} else {
			unavailable = append(unavailable, permission)
		}
	}
	mode := "procedural"
	if len(unavailable) > 0 {
		mode = "prompt-only"
	}
	if host.NativeSkills && len(unavailable) == 0 {
		mode = "native"
	}
	return RuntimeDescriptor{Agent: agent, Host: host, AvailableCapabilities: available, UnavailableCapabilities: unavailable, ExecutionMode: mode}
}

func hostSupports(host hosts.Capabilities, permission string) bool {
	switch permission {
	case "read_repository":
		return host.RepositoryRead
	case "write_repository":
		return host.FileEdits
	case "execute_commands":
		return host.CommandExecution
	case "network":
		return false
	case "browser_automation":
		return host.BrowserAutomation
	case "git_write", "pull_request_write":
		return host.CommandExecution
	case "subagents":
		return host.Subagents
	default:
		return false
	}
}

func sortedKeys(values map[string]Package) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func validateDescriptor(descriptor Descriptor) error {
	if descriptor.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", descriptor.SchemaVersion)
	}
	if descriptor.ID == "" || descriptor.Title == "" || descriptor.Description == "" || descriptor.Role == "" || descriptor.WorkflowPhase == "" {
		return fmt.Errorf("id, title, description, role, and workflow_phase are required")
	}
	if descriptor.InstructionFile != "SKILL.md" {
		return fmt.Errorf("instruction_file must be SKILL.md")
	}
	if len(descriptor.Input.Required) == 0 || len(descriptor.Output.Required) == 0 || len(descriptor.EvidenceRequirements) == 0 || len(descriptor.StopConditions) == 0 {
		return fmt.Errorf("input, output, evidence_requirements, and stop_conditions require at least one entry")
	}
	for _, permission := range descriptor.Permissions {
		if !validPermissions[permission] {
			return fmt.Errorf("unsupported permission %q", permission)
		}
	}
	if !descriptor.ToolFreeMode.Available || descriptor.ToolFreeMode.Behavior == "" || len(descriptor.ToolFreeMode.UnavailableStatuses) == 0 {
		return fmt.Errorf("tool_free_mode requires available, behavior, and unavailable_statuses")
	}
	for _, status := range descriptor.ToolFreeMode.UnavailableStatuses {
		if !validUnavailableStatuses[status] {
			return fmt.Errorf("unsupported tool-free status %q", status)
		}
	}
	return nil
}
