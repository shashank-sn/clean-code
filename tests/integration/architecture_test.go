package integration_test

import (
	"path/filepath"
	"testing"

	"clean-code/internal/architecture"
)

func TestArchitectureFixturesDistinguishCompliantAndViolatingGraphs(t *testing.T) {
	root := filepath.Join("..", "fixtures", "architecture")
	policy, err := architecture.LoadPolicy(filepath.Join(root, "policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	compliant, err := architecture.LoadGraph(filepath.Join(root, "compliant.json"))
	if err != nil {
		t.Fatal(err)
	}
	violating, err := architecture.LoadGraph(filepath.Join(root, "violating.json"))
	if err != nil {
		t.Fatal(err)
	}
	if report := architecture.Evaluate(policy, compliant); report.Status != "PASS" {
		t.Fatalf("expected compliant graph to pass: %+v", report)
	}
	if report := architecture.Evaluate(policy, violating); report.Status != "FAIL" || len(report.Violations) != 1 || report.Violations[0].Kind != "forbidden-dependency" {
		t.Fatalf("expected exact forbidden dependency: %+v", report)
	}
}
