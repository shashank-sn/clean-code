package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shashank-sn/clean-code/internal/hosts"
)

func TestLoadAllFindsEveryPortableSkillAgent(t *testing.T) {
	packages, err := LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(packages) != 27 {
		t.Fatalf("expected 27 portable agents, got %d", len(packages))
	}
	for _, id := range []string{"clean-lfg", "clean-eval-discover", "clean-reviewer", "clean-test-writer", "clean-auditor", "clean-merge-resolver", "clean-dispatcher", "clean-show-me"} {
		if _, exists := packages[id]; !exists {
			t.Fatalf("%s package is missing", id)
		}
	}
	for id, loaded := range packages {
		if loaded.Descriptor.ID != id || strings.TrimSpace(loaded.Instructions) == "" {
			t.Fatalf("invalid loaded package %q: %+v", id, loaded)
		}
	}
}

func TestShowMeAgentIsPortableAndNativeInCodex(t *testing.T) {
	runtime, err := Describe("clean-show-me", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if runtime.ExecutionMode != "native" || runtime.Agent.Role != "Visualizer" || runtime.Agent.WorkflowPhase != "explain" {
		t.Fatalf("unexpected show-me runtime: %+v", runtime)
	}
	prompt, err := EmitPrompt("clean-show-me", "codex")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"# Clean Show Me", "**Observed**", "**Proposed**", "A visual explains; it does not verify"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("show-me prompt missing %q:\n%s", expected, prompt)
		}
	}
}

func TestDiscoverAgentIsModelNeutralAcrossHosts(t *testing.T) {
	for _, host := range hosts.Catalog() {
		prompt, err := EmitPrompt("clean-discover", host.ID)
		if err != nil {
			t.Fatalf("emit clean-discover for %s: %v", host.ID, err)
		}
		for _, forbidden := range []string{"composer", "grok", "model must stay", "stop if the runtime"} {
			if strings.Contains(strings.ToLower(prompt), forbidden) {
				t.Fatalf("clean-discover prompt for %s contains model gate %q:\n%s", host.ID, forbidden, prompt)
			}
		}
		if !strings.Contains(prompt, "Run on the host-selected model") {
			t.Fatalf("clean-discover prompt for %s omits the model-neutral contract:\n%s", host.ID, prompt)
		}
	}
}

func TestValidateRejectsUnknownAgent(t *testing.T) {
	if err := Validate("clean-missing"); err == nil || !strings.Contains(err.Error(), "unknown agent") {
		t.Fatalf("expected unknown agent error, got %v", err)
	}
}

func TestLoadAllRejectsUnknownManifestField(t *testing.T) {
	root := fixtureRoot(t, `,"unknown":true`)
	if _, err := LoadAllFrom(root); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected strict JSON failure, got %v", err)
	}
}

func TestLoadAllRejectsUnknownHandoff(t *testing.T) {
	root := fixtureRoot(t, `,"handoff_to":["clean-missing"]`)
	if _, err := LoadAllFrom(root); err == nil || !strings.Contains(err.Error(), "unknown agent") {
		t.Fatalf("expected handoff failure, got %v", err)
	}
}

func TestEmitPromptReportsUnavailableCapabilities(t *testing.T) {
	prompt, err := EmitPrompt("clean-build", "generic")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Execution mode: prompt-only", "Unavailable: read_repository, write_repository, execute_commands", "NOT_AVAILABLE", "# Clean Build"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt missing %q:\n%s", expected, prompt)
		}
	}
}

func TestDescribeUsesNativeModeOnlyWhenRequirementsAreSupported(t *testing.T) {
	runtime, err := Describe("clean-orchestrate", "codex")
	if err != nil {
		t.Fatal(err)
	}
	if runtime.ExecutionMode != "native" || len(runtime.UnavailableCapabilities) != 0 {
		t.Fatalf("unexpected Codex runtime: %+v", runtime)
	}
}

func fixtureRoot(t *testing.T, suffix string) string {
	t.Helper()
	root := t.TempDir()
	directory := filepath.Join(root, "skills", "clean-fixture")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "schema_version":"1.0.0",
  "id":"clean-fixture",
  "title":"Fixture",
  "description":"Fixture agent.",
  "instruction_file":"SKILL.md",
  "role":"Fixture",
  "workflow_phase":"test",
  "input":{"required":["input"],"optional":[]},
  "output":{"required":["output"],"optional":[]},
  "evidence_requirements":["evidence"],
  "permissions":["read_repository"],
  "stop_conditions":["stop"],
  "tool_free_mode":{"available":true,"behavior":"report unavailable work","unavailable_statuses":["NOT_AVAILABLE"]},
  "handoff_to":[]` + suffix + `
}`
	if err := os.WriteFile(filepath.Join(directory, "agent.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "SKILL.md"), []byte("# Fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}
