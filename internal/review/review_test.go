package review

import "testing"

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
