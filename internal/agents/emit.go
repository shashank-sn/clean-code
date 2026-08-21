package agents

import (
	"encoding/json"
	"fmt"
	"strings"
)

func Describe(id, hostID string) (RuntimeDescriptor, error) {
	loaded, err := Load(id)
	if err != nil {
		return RuntimeDescriptor{}, err
	}
	return Runtime(loaded.Descriptor, hostID), nil
}

func EmitJSON(id, hostID string) ([]byte, error) {
	descriptor, err := Describe(id, hostID)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(descriptor, "", "  ")
}

func EmitPrompt(id, hostID string) (string, error) {
	loaded, err := Load(id)
	if err != nil {
		return "", err
	}
	runtime := Runtime(loaded.Descriptor, hostID)
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "# %s\n\n", loaded.Descriptor.Title)
	fmt.Fprintf(&prompt, "Role: %s\nPhase: %s\nExecution mode: %s\n\n", loaded.Descriptor.Role, loaded.Descriptor.WorkflowPhase, runtime.ExecutionMode)
	fmt.Fprintf(&prompt, "## Contract\n\nRequired input: %s\n\nRequired output: %s\n\nEvidence: %s\n\nStop conditions: %s\n\n", strings.Join(loaded.Descriptor.Input.Required, "; "), strings.Join(loaded.Descriptor.Output.Required, "; "), strings.Join(loaded.Descriptor.EvidenceRequirements, "; "), strings.Join(loaded.Descriptor.StopConditions, "; "))
	fmt.Fprintf(&prompt, "## Runtime descriptor\n\n%s\n\n", runtimeSummary(runtime))
	fmt.Fprintf(&prompt, "## Capability boundary\n\nAvailable: %s\n\nUnavailable: %s\n\nWhen a capability is unavailable, %s Status must be one of: %s.\n\n", printable(runtime.AvailableCapabilities), printable(runtime.UnavailableCapabilities), loaded.Descriptor.ToolFreeMode.Behavior, strings.Join(loaded.Descriptor.ToolFreeMode.UnavailableStatuses, ", "))
	fmt.Fprintf(&prompt, "## Handoff\n\nNext agents: %s\n\n## Instructions\n\n%s", printable(loaded.Descriptor.HandoffTo), strings.TrimSpace(loaded.Instructions))
	return prompt.String() + "\n", nil
}

func runtimeSummary(runtime RuntimeDescriptor) string {
	return fmt.Sprintf("Context capacity: %s; filesystem mode: %s; network policy: %s; browser/UI: %t; subagent isolation: %t; session reset: %t; structured output: %t.", runtime.Host.ContextCapacity, runtime.Host.FilesystemMode, runtime.Host.NetworkPolicy, runtime.Host.BrowserAutomation, runtime.Host.SubagentIsolation, runtime.Host.SessionReset, runtime.Host.StructuredOutput)
}

func printable(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}
