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

func TestBottomUpProposalRequiresInspectablePromotionEvidence(t *testing.T) {
	proposal := validProposal()
	proposal.EvidenceOrigin = "BOTTOM_UP"
	proposal.EvalPromotion = &EvalPromotionEvidence{
		SupportingEvidenceHashes: []string{
			"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
		CleanControlHash:          "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		HeldOutFixture:            "harness/calibration/held-out/eval-case.json",
		HeldOutEvidenceHash:       "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		HeldOutStatus:             "PASS",
		ExpectedFalsePositiveCost: "An advisory warning adds one reviewer decision.",
	}
	if issues := ValidateChangeProposal(proposal); len(issues) != 0 {
		t.Fatalf("expected valid bottom-up proposal, got %v", issues)
	}
}

func TestBottomUpProposalRejectsInsufficientOrMissingPromotionEvidence(t *testing.T) {
	proposal := validProposal()
	proposal.EvidenceOrigin = "BOTTOM_UP"
	proposal.EvalPromotion = &EvalPromotionEvidence{}
	issues := ValidateChangeProposal(proposal)
	for _, expected := range []string{
		"eval_promotion requires at least two supporting evidence hashes",
		"eval_promotion clean_control_hash is required",
		"eval_promotion held_out_fixture is required",
		"eval_promotion held_out_evidence_hash is required",
		"eval_promotion held_out_status must be PASS",
		"eval_promotion expected_false_positive_cost is required",
	} {
		if !containsIssue(issues, expected) {
			t.Errorf("expected %q in %v", expected, issues)
		}
	}
}

func TestBottomUpProposalRejectsOverlappingPromotionEvidence(t *testing.T) {
	proposal := validProposal()
	proposal.EvidenceOrigin = "BOTTOM_UP"
	proposal.EvalPromotion = &EvalPromotionEvidence{
		SupportingEvidenceHashes: []string{
			"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
		CleanControlHash:          "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		HeldOutFixture:            "harness/calibration/held-out/eval-case.json",
		HeldOutEvidenceHash:       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		HeldOutStatus:             "PASS",
		ExpectedFalsePositiveCost: "A false positive interrupts a reviewer.",
	}
	issues := ValidateChangeProposal(proposal)
	for _, expected := range []string{
		"eval_promotion clean_control_hash must be distinct from supporting evidence",
		"eval_promotion held_out_evidence_hash must be distinct from supporting evidence",
		"eval_promotion clean_control_hash and held_out_evidence_hash must differ",
	} {
		if !containsIssue(issues, expected) {
			t.Errorf("expected %q in %v", expected, issues)
		}
	}
}

func TestEvidenceOriginRejectsPromotionForTopDownProposal(t *testing.T) {
	proposal := validProposal()
	proposal.EvidenceOrigin = "TOP_DOWN"
	proposal.EvalPromotion = &EvalPromotionEvidence{}
	if !containsIssue(ValidateChangeProposal(proposal), "eval_promotion is only valid for BOTTOM_UP proposals") {
		t.Fatal("expected top-down evidence rejection")
	}
}

func TestApprovedProposalRequiresIndependentHumanApprover(t *testing.T) {
	proposal := validProposal()
	proposal.Status = "APPROVED"
	proposal.Reviewer = "reviewer-a"
	if !containsIssue(ValidateChangeProposal(proposal), "human_approver is required for approved proposals") {
		t.Fatal("expected approved proposal to require human approver")
	}
	proposal.HumanApprover = proposal.Reviewer
	if !containsIssue(ValidateChangeProposal(proposal), "human_approver must be independent of proposer and reviewer") {
		t.Fatal("expected independent human approver rejection")
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
