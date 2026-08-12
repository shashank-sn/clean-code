package trace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluateCompleteIndependentPlanPasses(t *testing.T) {
	report := Evaluate(completePlan())
	if report.Status != "PASS" || len(report.Issues) != 0 || len(report.Warnings) != 0 {
		t.Fatalf("expected complete plan, got %+v", report)
	}
}

func TestEvaluateReportsMissingExamplesTracksOwnersEvidenceAndReasons(t *testing.T) {
	plan := Plan{
		SchemaVersion: "1.0.0",
		Requirements:  []Requirement{{ID: "R1"}},
		Tracks: []Track{
			{Kind: "unit", RequirementIDs: []string{"R1"}, Status: "PLANNED"},
			{Kind: "acceptance", RequirementIDs: []string{"R1"}, Status: "PASS", Owner: "acceptance"},
			{Kind: "integration", RequirementIDs: []string{"R1"}, Status: "NOT_RUN", Owner: "integration"},
		},
	}
	report := Evaluate(plan)
	for _, kind := range []string{"missing-acceptance-example", "missing-owner", "missing-evidence", "missing-reason", "missing-track"} {
		assertIssue(t, report, kind)
	}
}

func TestEvaluateAcceptsReasonedInapplicableTrack(t *testing.T) {
	plan := completePlan()
	plan.Tracks[3] = Track{Kind: "ui-qa", RequirementIDs: []string{"R1"}, Status: "INAPPLICABLE", Reason: "no user interface", ContextSource: "requirements"}
	if report := Evaluate(plan); report.Status != "PASS" {
		t.Fatalf("expected reasoned inapplicable track to pass: %+v", report)
	}
}

func TestEvaluateWarnsAboutImplementationShapedAcceptanceContext(t *testing.T) {
	plan := completePlan()
	plan.Tracks[1].ContextSource = "implementation"
	report := Evaluate(plan)
	if report.Status != "PASS" || len(report.Warnings) != 1 {
		t.Fatalf("expected correlation warning, got %+v", report)
	}
}

func TestEvaluateReportsUnknownRequirement(t *testing.T) {
	plan := completePlan()
	plan.Tracks[0].RequirementIDs = []string{"R2"}
	report := Evaluate(plan)
	assertIssue(t, report, "unknown-requirement")
	assertIssue(t, report, "missing-track")
}

func TestLoadRejectsUnknownFieldsAndSymlinks(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "plan.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"1.0.0","requirements":[],"tracks":[],"pretend_pass":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown field rejection")
	}
	if err := os.WriteFile(path, []byte(`{"schema_version":"1.0.0","requirements":[],"tracks":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.json")
	if err := os.Symlink(path, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := Load(link); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func completePlan() Plan {
	return Plan{
		SchemaVersion: "1.0.0",
		Requirements:  []Requirement{{ID: "R1", AcceptanceExamples: []string{"given valid input, returns the requested result"}}},
		Tracks: []Track{
			{Kind: "unit", RequirementIDs: []string{"R1"}, Status: "PLANNED", Owner: "unit"},
			{Kind: "acceptance", RequirementIDs: []string{"R1"}, Status: "PLANNED", Owner: "acceptance", ContextSource: "requirements"},
			{Kind: "integration", RequirementIDs: []string{"R1"}, Status: "PLANNED", Owner: "integration"},
			{Kind: "ui-qa", RequirementIDs: []string{"R1"}, Status: "PLANNED", Owner: "qa", ContextSource: "requirements"},
		},
	}
}

func assertIssue(t *testing.T, report Report, kind string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Kind == kind {
			return
		}
	}
	t.Fatalf("expected %s issue, got %+v", kind, report)
}
