package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChangeProposalRequiresIndependentApproval(t *testing.T) {
	proposal := validProposal()
	proposal.Status = "APPROVED"
	proposal.Reviewer = proposal.Proposer
	if !containsIssue(ValidateChangeProposal(proposal), "proposer cannot review their own policy change") {
		t.Fatal("expected self-approval rejection")
	}
}

func TestChangeProposalCannotSuppressProtectedGate(t *testing.T) {
	proposal := validProposal()
	proposal.Kind = "suppression"
	proposal.TargetClass = "requirement"
	if !containsIssue(ValidateChangeProposal(proposal), "protected gates cannot be weakened or suppressed") {
		t.Fatal("expected protected-gate rejection")
	}
}

func TestReversibleQualityProposalIsValid(t *testing.T) {
	if issues := ValidateChangeProposal(validProposal()); len(issues) != 0 {
		t.Fatalf("expected valid proposal, got %v", issues)
	}
}

func TestLoadChangeProposalRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proposal.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"1.0.0","unknown":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadChangeProposal(path); err == nil {
		t.Fatal("expected strict proposal parsing")
	}
}

func validProposal() ChangeProposal {
	return ChangeProposal{
		SchemaVersion: "1.0.0", ID: "P1", SourceReceipt: "receipt.json",
		EvidenceHashes: []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		Kind:           "threshold", TargetClass: "quality", Effect: "WEAKEN", Scope: "complexity.changed",
		Change: "raise advisory threshold by one", Benefit: "remove confirmed false positive",
		Risks: []string{"may reduce sensitivity"}, Rollback: "restore previous threshold",
		CalibrationFixture: "fixtures/complexity-boundary", Proposer: "agent-a", Status: "PROPOSED",
	}
}

func containsIssue(issues []string, target string) bool {
	for _, issue := range issues {
		if issue == target {
			return true
		}
	}
	return false
}
