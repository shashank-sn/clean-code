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
	if len(packages) != 31 {
		t.Fatalf("expected 31 portable agents, got %d", len(packages))
	}
	for _, id := range []string{"clean-lfg", "clean-eval-discover", "clean-reviewer", "clean-test-writer", "clean-auditor", "clean-merge-resolver", "clean-dispatcher", "clean-show-me", "clean-route", "clean-arena", "clean-probe", "clean-prune"} {
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

func TestReviewerPackagesAreReadOnlyAndCarryCanonicalProtocol(t *testing.T) {
	root := filepath.Join("..", "..")
	canonical := mustRead(t, filepath.Join(root, "harness", "review", "protocol.md"))
	for _, id := range []string{"clean-review", "clean-reviewer"} {
		loaded, err := Load(id)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.Join(loaded.Descriptor.Permissions, ","), "write_repository") {
			t.Fatalf("%s grants product write access: %#v", id, loaded.Descriptor.Permissions)
		}
		if !contains(loaded.Descriptor.Permissions, "read_repository") || !contains(loaded.Descriptor.Permissions, "execute_commands") {
			t.Fatalf("%s must declare read and conditional command capability: %#v", id, loaded.Descriptor.Permissions)
		}
		for _, required := range []string{
			"base and candidate revisions plus complete changed-file inventory and diff",
			"original requirements and affected-call-path notes",
			"actual check commands, results, skips, changed assertions, and known gaps",
			"change-author and reviewer/context identities",
		} {
			if !contains(loaded.Descriptor.Input.Required, required) {
				t.Fatalf("%s input omits required review packet field %q: %#v", id, required, loaded.Descriptor.Input.Required)
			}
		}
		for _, required := range []string{
			"v2 review record with reviewed scope, requirements, six dimension assessments, checks, limitations, and completion",
			"requirement-to-call-path-and-test mapping or an explicit evidence gap",
		} {
			if !contains(loaded.Descriptor.Output.Required, required) {
				t.Fatalf("%s output omits required review evidence %q: %#v", id, required, loaded.Descriptor.Output.Required)
			}
		}
		if got := markedSection(loaded.Instructions); got != markedSection(canonical) {
			t.Fatalf("%s protocol diverged from canonical source", id)
		}
		for _, hostID := range []string{"generic", "codex"} {
			prompt, err := EmitPrompt(id, hostID)
			if err != nil {
				t.Fatalf("emit %s/%s: %v", id, hostID, err)
			}
			if got := markedSection(prompt); got != markedSection(canonical) {
				t.Fatalf("emitted %s/%s protocol diverged from canonical source", id, hostID)
			}
			if !strings.Contains(prompt, "product-code write") || !strings.Contains(prompt, "INCOMPLETE") {
				t.Fatalf("emitted %s/%s prompt omitted review boundary or terminal state", id, hostID)
			}
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

func mustRead(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func markedSection(body string) string {
	const start = "<!-- review-protocol:start -->"
	const end = "<!-- review-protocol:end -->"
	from := strings.Index(body, start)
	to := strings.Index(body, end)
	if from < 0 || to < from {
		return ""
	}
	to += len(end)
	return body[from:to]
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
