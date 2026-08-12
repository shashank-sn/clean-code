package integration_test

import (
	"path/filepath"
	"testing"

	"clean-code/internal/trace"
)

func TestTraceFixturesDistinguishCompleteAndIncompletePlans(t *testing.T) {
	root := filepath.Join("..", "fixtures", "trace")
	complete, err := trace.Load(filepath.Join(root, "complete.json"))
	if err != nil {
		t.Fatal(err)
	}
	incomplete, err := trace.Load(filepath.Join(root, "incomplete.json"))
	if err != nil {
		t.Fatal(err)
	}
	if report := trace.Evaluate(complete); report.Status != "PASS" {
		t.Fatalf("expected complete plan to pass: %+v", report)
	}
	if report := trace.Evaluate(incomplete); report.Status != "FAIL" || len(report.Issues) != 4 {
		t.Fatalf("expected missing acceptance example and three tracks: %+v", report)
	}
}
