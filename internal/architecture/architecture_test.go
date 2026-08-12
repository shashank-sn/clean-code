package architecture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluateAllowsDeclaredInwardDependency(t *testing.T) {
	policy := testPolicy()
	report := Evaluate(policy, Graph{SchemaVersion: "1.0.0", Edges: []Edge{{From: "delivery/http.go", To: "core/usecase.go"}}})
	if report.Status != "PASS" || len(report.Violations) != 0 {
		t.Fatalf("expected pass, got %+v", report)
	}
}

func TestEvaluateRejectsForbiddenOutwardDependency(t *testing.T) {
	report := Evaluate(testPolicy(), Graph{SchemaVersion: "1.0.0", Edges: []Edge{{From: "core/usecase.go", To: "delivery/http.go"}}})
	assertViolation(t, report, "forbidden-dependency")
}

func TestEvaluateRejectsPrivateSurfaceAccess(t *testing.T) {
	report := Evaluate(testPolicy(), Graph{SchemaVersion: "1.0.0", Edges: []Edge{{From: "delivery/http.go", To: "core/internal/helper.go"}}})
	assertViolation(t, report, "private-surface")
}

func TestEvaluateReportsUndeclaredAndAmbiguousPaths(t *testing.T) {
	policy := testPolicy()
	policy.Components = append(policy.Components, Component{ID: "overlap", Paths: []string{"core/**"}})
	report := Evaluate(policy, Graph{SchemaVersion: "1.0.0", Edges: []Edge{
		{From: "unknown/file.go", To: "delivery/http.go"},
		{From: "delivery/http.go", To: "core/usecase.go"},
	}})
	assertViolation(t, report, "undeclared-path")
	assertViolation(t, report, "ambiguous-component")
}

func TestEvaluateHonorsExclusionsAndReasonedExceptions(t *testing.T) {
	policy := testPolicy()
	policy.Exclude = []string{"generated/**"}
	policy.Exceptions = []Exception{{From: "core/legacy.go", To: "delivery/legacy.go", Reason: "remove after migration"}}
	report := Evaluate(policy, Graph{SchemaVersion: "1.0.0", Edges: []Edge{
		{From: "generated/client.go", To: "delivery/http.go"},
		{From: "core/legacy.go", To: "delivery/legacy.go"},
	}})
	if report.Status != "PASS" || report.ExcludedEdges != 1 || report.CheckedEdges != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestEvaluateExceptionCannotHideUndeclaredPath(t *testing.T) {
	policy := testPolicy()
	policy.Exceptions = []Exception{{From: "unknown/**", To: "delivery/**", Reason: "invalid scope"}}
	report := Evaluate(policy, Graph{SchemaVersion: "1.0.0", Edges: []Edge{{From: "unknown/file.go", To: "delivery/http.go"}}})
	assertViolation(t, report, "undeclared-path")
}

func TestEvaluateExceptionStillParticipatesInCycleDetection(t *testing.T) {
	policy := testPolicy()
	policy.Components[0].MayDependOn = []string{"delivery"}
	policy.Exceptions = []Exception{{From: "core/**", To: "delivery/**", Reason: "temporary"}}
	report := Evaluate(policy, Graph{SchemaVersion: "1.0.0", Edges: []Edge{
		{From: "delivery/http.go", To: "core/usecase.go"},
		{From: "core/usecase.go", To: "delivery/http.go"},
	}})
	assertViolation(t, report, "component-cycle")
}

func TestEvaluateReportsComponentCycleWithPath(t *testing.T) {
	policy := testPolicy()
	policy.Components[0].MayDependOn = []string{"delivery"}
	report := Evaluate(policy, Graph{SchemaVersion: "1.0.0", Edges: []Edge{
		{From: "delivery/http.go", To: "core/usecase.go"},
		{From: "core/usecase.go", To: "delivery/http.go"},
	}})
	assertViolation(t, report, "component-cycle")
}

func TestLoadersRejectUnknownFieldsAndSymlinks(t *testing.T) {
	root := t.TempDir()
	policyPath := filepath.Join(root, "policy.json")
	if err := os.WriteFile(policyPath, []byte(`{"schema_version":"1.0.0","components":[],"pretend_pass":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPolicy(policyPath); err == nil {
		t.Fatal("expected unknown field rejection")
	}
	graphPath := filepath.Join(root, "graph.json")
	if err := os.WriteFile(graphPath, []byte(`{"schema_version":"1.0.0","edges":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(root, "graph-link.json")
	if err := os.Symlink(graphPath, linkPath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := LoadGraph(linkPath); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func TestPolicyRejectsUnknownComponentAndMissingExceptionReason(t *testing.T) {
	policy := testPolicy()
	policy.Components[0].MayDependOn = []string{"missing"}
	if err := policy.Validate(); err == nil {
		t.Fatal("expected unknown component rejection")
	}
	policy = testPolicy()
	policy.Exceptions = []Exception{{From: "core/**", To: "delivery/**"}}
	if err := policy.Validate(); err == nil {
		t.Fatal("expected missing reason rejection")
	}
}

func TestPolicyAndGraphRejectPortableAbsoluteAndUnsupportedGlob(t *testing.T) {
	policy := testPolicy()
	policy.Components[0].Paths = []string{"core/**/nested"}
	if err := policy.Validate(); err == nil {
		t.Fatal("expected unsupported double-star glob rejection")
	}
	root := t.TempDir()
	graphPath := filepath.Join(root, "graph.json")
	if err := os.WriteFile(graphPath, []byte(`{"schema_version":"1.0.0","edges":[{"from":"C:\\repo\\file.go","to":"core/usecase.go"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGraph(graphPath); err == nil {
		t.Fatal("expected Windows absolute path rejection")
	}
}

func TestGraphRejectsDuplicateEdge(t *testing.T) {
	root := t.TempDir()
	graphPath := filepath.Join(root, "graph.json")
	if err := os.WriteFile(graphPath, []byte(`{
  "schema_version":"1.0.0",
  "edges":[
    {"from":"delivery/http.go","to":"core/usecase.go"},
    {"from":"delivery/http.go","to":"core/usecase.go"}
  ]
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGraph(graphPath); err == nil {
		t.Fatal("expected duplicate edge rejection")
	}
}

func testPolicy() Policy {
	return Policy{
		SchemaVersion: "1.0.0",
		Components: []Component{
			{ID: "core", Paths: []string{"core/**"}, Public: []string{"core/usecase.go"}},
			{ID: "delivery", Paths: []string{"delivery/**"}, MayDependOn: []string{"core"}},
		},
	}
}

func assertViolation(t *testing.T, report Report, kind string) {
	t.Helper()
	for _, violation := range report.Violations {
		if violation.Kind == kind {
			return
		}
	}
	t.Fatalf("expected %s violation, got %+v", kind, report)
}
