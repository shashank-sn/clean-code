package benchmark

import (
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func TestCompareWorkflowsFromManifest(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "harness", "calibration", "workflow-comparison.json")
	manifest, err := LoadWorkflowManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	report := CompareWorkflows(manifest)
	if report.SchemaVersion != "1.0.0" {
		t.Fatalf("unexpected schema version %q", report.SchemaVersion)
	}
	if len(report.Workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(report.Workflows))
	}
	if report.Workflows[0].ID != "clean-code" {
		t.Fatalf("expected clean-code to lead, got %q", report.Workflows[0].ID)
	}
	if report.Workflows[0].Coverage <= report.Workflows[1].Coverage {
		t.Fatalf("expected clean-code coverage to exceed compound-engineering")
	}
}

func TestValidateWorkflowManifestRejectsInvalidScore(t *testing.T) {
	manifest := WorkflowManifest{
		SchemaVersion: "1.0.0",
		Dimensions:    []string{"planning"},
		Workflows: []Workflow{
			{ID: "a", Name: "A", Scores: map[string]float64{"planning": 1.5}},
			{ID: "b", Name: "B", Scores: map[string]float64{"planning": 0.5}},
		},
	}
	if err := ValidateWorkflowManifest(manifest); err == nil {
		t.Fatal("expected invalid score error")
	}
}
