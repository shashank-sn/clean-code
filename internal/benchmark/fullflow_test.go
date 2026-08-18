package benchmark

import (
	"path/filepath"
	"testing"
)

func TestRunFullFlowBenchmark(t *testing.T) {
	root := repoRoot(t)
	manifest, err := LoadFullFlowManifest(filepath.Join(root, "harness", "calibration", "full-flow-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := RunFullFlow(root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if report.Winner != "clean-code" {
		t.Fatalf("expected clean-code to win, got %q summary=%s", report.Winner, report.Summary)
	}
	if len(report.Outcomes) != 2 {
		t.Fatalf("expected 2 outcomes, got %d", len(report.Outcomes))
	}
	for _, outcome := range report.Outcomes {
		if outcome.WorkflowID == "clean-code" {
			if !outcome.TestsPassed {
				t.Fatal("clean-code outcome tests should pass")
			}
			if outcome.Metrics.TestFunctions < 8 {
				t.Fatalf("expected broad tests, got %d", outcome.Metrics.TestFunctions)
			}
			if outcome.Combined <= 0.7 {
				t.Fatalf("expected high combined score, got %f", outcome.Combined)
			}
		}
	}
}

func TestValidateFullFlowManifestRejectsEmptyTask(t *testing.T) {
	manifest := FullFlowManifest{SchemaVersion: "1.0.0", Outcomes: []FullFlowEntry{{WorkflowID: "a", PackageDir: "x"}, {WorkflowID: "b", PackageDir: "y"}}, Rubric: []RubricItem{{ID: "t", Weight: 1}}}
	if err := ValidateFullFlowManifest(manifest); err == nil {
		t.Fatal("expected task_id error")
	}
}

func TestMeasureFunctions(t *testing.T) {
	lines := []string{
		"package slug",
		"func Small() {",
		"  return",
		"}",
		"func Big() {",
		"  if true {",
		"    return",
		"  }",
		"}",
	}
	lengths := measureFunctions(lines)
	if len(lengths) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(lengths))
	}
}
