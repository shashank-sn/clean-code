package review

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateAllowsIndependentCorrectSilence(t *testing.T) {
	report := Evaluate(Input{SchemaVersion: "1.0.0", Revision: "abc", ChangeAuthor: "author", Reviewer: "reviewer", Scope: []string{"change.go"}, Findings: []Finding{}})
	if report.Status != "PASS" || len(report.Findings) != 0 {
		t.Fatalf("expected correct silence, got %+v", report)
	}
}

func TestEvaluateRejectsSelfReview(t *testing.T) {
	report := Evaluate(Input{SchemaVersion: "1.0.0", Revision: "abc", ChangeAuthor: "same", Reviewer: "same", Scope: []string{"change.go"}})
	assertReviewIssue(t, report, "self-review")
}

func TestEvaluateRejectsUnsupportedDuplicateFinding(t *testing.T) {
	finding := validFinding()
	duplicate := finding
	duplicate.ID = "F2"
	duplicate.Evidence = ""
	report := Evaluate(Input{SchemaVersion: "1.0.0", Revision: "abc", ChangeAuthor: "author", Reviewer: "reviewer", Scope: []string{"change.go"}, Findings: []Finding{finding, duplicate}})
	assertReviewIssue(t, report, "unsupported-finding")
	assertReviewIssue(t, report, "duplicate-finding")
}

func TestEvaluateBlocksUnresolvedBlockingFinding(t *testing.T) {
	report := Evaluate(Input{SchemaVersion: "1.0.0", Revision: "abc", ChangeAuthor: "author", Reviewer: "reviewer", Scope: []string{"change.go"}, Findings: []Finding{validFinding()}})
	assertReviewIssue(t, report, "blocking-finding")
}

func TestEvaluateRequiresAcceptedRiskReason(t *testing.T) {
	finding := validFinding()
	finding.Severity = "IMPROVEMENT"
	finding.Disposition = "ACCEPTED_RISK"
	report := Evaluate(Input{SchemaVersion: "1.0.0", Revision: "abc", ChangeAuthor: "author", Reviewer: "reviewer", Scope: []string{"change.go"}, Findings: []Finding{finding}})
	assertReviewIssue(t, report, "missing-resolution")
}

func TestEvaluateV2CompleteArtifactOnly(t *testing.T) {
	report := Evaluate(validV2())
	if report.Status != "PASS" || report.Completion != "COMPLETE" || report.ValidationScope != ArtifactOnly {
		t.Fatalf("expected complete artifact-only report, got %+v", report)
	}
}

func TestEvaluateV2RequiresCoverageAndDimensions(t *testing.T) {
	input := validV2()
	input.Dimensions = nil
	input.Coverage = nil
	input.Completion = "COMPLETE"
	report := Evaluate(input)
	if report.Status == "PASS" || !hasIssue(report, "missing-dimension") || !hasIssue(report, "missing-coverage") {
		t.Fatalf("expected incomplete v2 report, got %+v", report)
	}
}

func TestEvaluateV2RejectsStaleCheck(t *testing.T) {
	input := validV2()
	input.Checks[0].Revision = "old"
	report := Evaluate(input)
	assertReviewIssue(t, report, "revision-mismatch")
}

func TestEvaluateV2BlockingDismissalNeedsResolutionCheck(t *testing.T) {
	input := validV2()
	input.Findings = []Finding{{ID: "F1", Severity: "BLOCKING", Behavior: "bad", Evidence: "evidence", Consequence: "impact", Confidence: "HIGH", Fix: "fix", Disposition: "DISMISSED", ResolutionReason: "rechecked"}}
	report := Evaluate(input)
	assertReviewIssue(t, report, "invalid-resolution-check")
	input.Findings[0].ResolutionCheck = "verification"
	if got := Evaluate(input); got.Status != "PASS" {
		t.Fatalf("expected resolved blocking dismissal to pass, got %+v", got)
	}
}

func TestEvaluateV2CompleteCanReportBlockingDefect(t *testing.T) {
	input := validV2()
	input.Findings = []Finding{validFinding()}
	input.Findings[0].Disposition = "OPEN"
	report := Evaluate(input)
	if report.Completion != "COMPLETE" || report.Status != "FAIL" {
		t.Fatalf("expected complete review with failed status, got %+v", report)
	}
	assertReviewIssue(t, report, "blocking-finding")
}

func TestEvaluateV2RejectsWhitespaceSelfReview(t *testing.T) {
	input := validV2()
	input.ChangeAuthor = "author "
	input.Reviewer = " author"
	report := Evaluate(input)
	assertReviewIssue(t, report, "self-review")
}

func TestEvaluateV2PreservesFindingSafeguards(t *testing.T) {
	input := validV2()
	negative := validFinding()
	negative.Line = -1
	input.Findings = []Finding{negative}
	report := Evaluate(input)
	assertReviewIssue(t, report, "invalid-line")
	input = validV2()
	input.Findings = []Finding{validFinding(), validFinding()}
	input.Findings[1].ID = "F2"
	report = Evaluate(input)
	assertReviewIssue(t, report, "duplicate-finding")
}

func TestEvaluateV2FailEvidenceIsCompleteButFails(t *testing.T) {
	input := validV2()
	input.Dimensions[0].Status = "FAIL"
	report := Evaluate(input)
	if report.Completion != "COMPLETE" || report.Status != "FAIL" {
		t.Fatalf("expected complete failed assessment, got %+v", report)
	}
	input = validV2()
	input.Checks[0].Status = "FAIL"
	report = Evaluate(input)
	if report.Completion != "COMPLETE" || report.Status != "FAIL" {
		t.Fatalf("expected complete failed required check, got %+v", report)
	}
}

func TestLoadV2RejectsMixedLegacyFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.json")
	data := []byte(`{"schema_version":"1.0.0","revision":"r","change_author":"a","reviewer":"b","scope":["x"],"findings":[],"completion":"COMPLETE"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected mixed-version input to be rejected")
	}
}

func TestLoadV2RejectsNullRequiredBoolean(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.json")
	data, err := json.Marshal(validV2())
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), `"required":true`, `"required":null`, 1))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected null required boolean to be rejected")
	}
}

func TestLoadV2RejectsExplicitZeroFindingLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.json")
	input := validV2()
	finding := validFinding()
	finding.Line = 1
	input.Findings = []Finding{finding}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), `"line":1`, `"line":0`, 1))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected explicitly supplied zero line to be rejected")
	}
}

func validV2() Input {
	dimensions := make([]Dimension, 0, len(reviewDimensions))
	for _, id := range reviewDimensions {
		d := Dimension{ID: id, Status: "PASS", Evidence: []string{"review:" + id}}
		if id == "security" {
			d.Status = "NOT_APPLICABLE"
			d.Evidence = nil
			d.Reason = "no boundary change"
		}
		dimensions = append(dimensions, d)
	}
	return Input{
		SchemaVersion: V2SchemaVersion, BaseRevision: "base", Revision: "candidate", ChangeAuthor: "author", Reviewer: "reviewer",
		Scope: []string{"change.go"}, Requirements: []string{"R1"}, Dimensions: dimensions,
		Coverage:    []Coverage{{Path: "change.go", Status: "REVIEWED", Evidence: "diff"}},
		Checks:      []Check{{ID: "verification", Kind: "test", Required: true, Status: "PASS", Revision: "candidate", Source: "go test", Artifact: "report.json", SHA256: "0000000000000000000000000000000000000000000000000000000000000000"}},
		Limitations: []string{}, Completion: "COMPLETE", Findings: []Finding{},
	}
}

func hasIssue(report Report, kind string) bool {
	for _, issue := range report.Issues {
		if issue.Kind == kind {
			return true
		}
	}
	return false
}

func validFinding() Finding {
	return Finding{
		ID: "F1", Severity: "BLOCKING", File: "change.go", Line: 10,
		Behavior: "request can return stale data", Evidence: "test stale-cache fails",
		Consequence: "users receive obsolete state", Confidence: "HIGH",
		Fix: "invalidate the cache after the write", Disposition: "OPEN",
	}
}

func assertReviewIssue(t *testing.T, report Report, kind string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Kind == kind {
			return
		}
	}
	t.Fatalf("expected %s issue, got %+v", kind, report)
}
