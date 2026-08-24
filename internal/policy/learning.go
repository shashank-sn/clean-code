package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

const maxProposalBytes int64 = 5 << 20

type ChangeProposal struct {
	SchemaVersion      string                 `json:"schema_version"`
	ID                 string                 `json:"id"`
	SourceReceipt      string                 `json:"source_receipt"`
	EvidenceHashes     []string               `json:"evidence_hashes"`
	Kind               string                 `json:"kind"`
	TargetClass        string                 `json:"target_class"`
	Effect             string                 `json:"effect"`
	Scope              string                 `json:"scope"`
	Change             string                 `json:"change"`
	Benefit            string                 `json:"benefit"`
	Risks              []string               `json:"risks"`
	Rollback           string                 `json:"rollback"`
	CalibrationFixture string                 `json:"calibration_fixture"`
	Proposer           string                 `json:"proposer"`
	Reviewer           string                 `json:"reviewer,omitempty"`
	Status             string                 `json:"status"`
	EvidenceOrigin     string                 `json:"evidence_origin,omitempty"`
	EvalPromotion      *EvalPromotionEvidence `json:"eval_promotion,omitempty"`
	HumanApprover      string                 `json:"human_approver,omitempty"`
}

// EvalPromotionEvidence makes bottom-up policy learning inspectable. It records
// the minimum evidence needed to propose a reusable rule; it does not grant
// approval or make the rule active.
type EvalPromotionEvidence struct {
	SupportingEvidenceHashes  []string `json:"supporting_evidence_hashes"`
	CleanControlHash          string   `json:"clean_control_hash"`
	HeldOutFixture            string   `json:"held_out_fixture"`
	HeldOutEvidenceHash       string   `json:"held_out_evidence_hash"`
	HeldOutStatus             string   `json:"held_out_status"`
	ExpectedFalsePositiveCost string   `json:"expected_false_positive_cost"`
}

type ProposalReport struct {
	SchemaVersion string   `json:"schema_version"`
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	Issues        []string `json:"issues"`
}

func LoadChangeProposal(path string) (ChangeProposal, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return ChangeProposal{}, fmt.Errorf("inspect policy proposal: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return ChangeProposal{}, errors.New("policy proposal must be a regular file")
	}
	if info.Size() > maxProposalBytes {
		return ChangeProposal{}, fmt.Errorf("policy proposal exceeds %d bytes", maxProposalBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return ChangeProposal{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxProposalBytes+1))
	decoder.DisallowUnknownFields()
	var proposal ChangeProposal
	if err := decoder.Decode(&proposal); err != nil {
		return ChangeProposal{}, fmt.Errorf("parse policy proposal: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ChangeProposal{}, errors.New("parse policy proposal: unexpected trailing JSON value")
	}
	return proposal, nil
}

func EvaluateChangeProposal(proposal ChangeProposal) ProposalReport {
	issues := ValidateChangeProposal(proposal)
	status := "PASS"
	if len(issues) > 0 {
		status = "FAIL"
	}
	return ProposalReport{SchemaVersion: "1.0.0", ID: proposal.ID, Status: status, Issues: issues}
}

// ValidateChangeProposal keeps learning advisory: a proposal may describe a
// policy change, but it cannot self-approve or weaken a protected gate.
func ValidateChangeProposal(proposal ChangeProposal) []string {
	var issues []string
	required := []struct {
		name  string
		value string
	}{
		{"id", proposal.ID}, {"source_receipt", proposal.SourceReceipt}, {"scope", proposal.Scope},
		{"change", proposal.Change}, {"benefit", proposal.Benefit}, {"rollback", proposal.Rollback},
		{"calibration_fixture", proposal.CalibrationFixture}, {"proposer", proposal.Proposer},
	}
	if proposal.SchemaVersion != "1.0.0" {
		issues = append(issues, "unsupported schema version")
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			issues = append(issues, fmt.Sprintf("%s is required", field.name))
		}
	}
	if !oneOfString(proposal.Kind, "rule", "threshold", "adapter", "convention", "suppression", "calibration") {
		issues = append(issues, "invalid proposal kind")
	}
	if !oneOfString(proposal.TargetClass, "quality", "convention", "correctness", "safety", "security", "privacy", "data-integrity", "requirement") {
		issues = append(issues, "invalid target class")
	}
	if !oneOfString(proposal.Effect, "STRENGTHEN", "NEUTRAL", "WEAKEN") {
		issues = append(issues, "invalid proposal effect")
	}
	if !oneOfString(proposal.Status, "PROPOSED", "APPROVED", "REJECTED", "ROLLED_BACK") {
		issues = append(issues, "invalid proposal status")
	}
	if len(proposal.EvidenceHashes) == 0 {
		issues = append(issues, "at least one evidence hash is required")
	}
	seen := map[string]bool{}
	for _, digest := range proposal.EvidenceHashes {
		if !sha256Pattern.MatchString(digest) {
			issues = append(issues, "evidence hashes must be lowercase SHA-256 values")
		}
		if seen[digest] {
			issues = append(issues, "evidence hashes cannot be duplicated")
		}
		seen[digest] = true
	}
	for _, risk := range proposal.Risks {
		if strings.TrimSpace(risk) == "" {
			issues = append(issues, "risks cannot contain empty entries")
			break
		}
	}
	if proposal.Status != "PROPOSED" && strings.TrimSpace(proposal.Reviewer) == "" {
		issues = append(issues, "reviewer is required after proposal")
	}
	if proposal.Reviewer != "" && proposal.Reviewer == proposal.Proposer {
		issues = append(issues, "proposer cannot review their own policy change")
	}
	if proposal.Status == "APPROVED" && strings.TrimSpace(proposal.HumanApprover) == "" {
		issues = append(issues, "human_approver is required for approved proposals")
	}
	if proposal.HumanApprover != "" && (proposal.HumanApprover == proposal.Proposer || proposal.HumanApprover == proposal.Reviewer) {
		issues = append(issues, "human_approver must be independent of proposer and reviewer")
	}
	issues = append(issues, validateEvalPromotion(proposal)...)
	if isProtectedTarget(proposal.TargetClass) && (proposal.Kind == "suppression" || proposal.Effect == "WEAKEN") {
		issues = append(issues, "protected gates cannot be weakened or suppressed")
	}
	return issues
}

func validateEvalPromotion(proposal ChangeProposal) []string {
	if proposal.EvidenceOrigin == "" {
		return nil
	}
	if !oneOfString(proposal.EvidenceOrigin, "TOP_DOWN", "BOTTOM_UP") {
		return []string{"invalid evidence_origin"}
	}
	if proposal.EvidenceOrigin != "BOTTOM_UP" {
		if proposal.EvalPromotion != nil {
			return []string{"eval_promotion is only valid for BOTTOM_UP proposals"}
		}
		return nil
	}
	if proposal.EvalPromotion == nil {
		return []string{"BOTTOM_UP proposals require eval_promotion"}
	}
	evidence := proposal.EvalPromotion
	issues := []string{}
	if len(evidence.SupportingEvidenceHashes) < 2 {
		issues = append(issues, "eval_promotion requires at least two supporting evidence hashes")
	}
	seen := map[string]bool{}
	for _, digest := range evidence.SupportingEvidenceHashes {
		if !sha256Pattern.MatchString(digest) {
			issues = append(issues, "eval_promotion supporting evidence hashes must be lowercase SHA-256 values")
		}
		if seen[digest] {
			issues = append(issues, "eval_promotion supporting evidence hashes cannot be duplicated")
		}
		seen[digest] = true
	}
	for _, field := range []struct{ name, value string }{
		{"clean_control_hash", evidence.CleanControlHash},
		{"held_out_fixture", evidence.HeldOutFixture},
		{"held_out_evidence_hash", evidence.HeldOutEvidenceHash},
	} {
		if strings.TrimSpace(field.value) == "" {
			issues = append(issues, "eval_promotion "+field.name+" is required")
		}
	}
	for _, digest := range []struct{ name, value string }{
		{"clean_control_hash", evidence.CleanControlHash},
		{"held_out_evidence_hash", evidence.HeldOutEvidenceHash},
	} {
		if digest.value != "" && !sha256Pattern.MatchString(digest.value) {
			issues = append(issues, "eval_promotion "+digest.name+" must be a lowercase SHA-256 value")
		}
		if seen[digest.value] {
			issues = append(issues, "eval_promotion "+digest.name+" must be distinct from supporting evidence")
		}
	}
	if evidence.CleanControlHash != "" && evidence.CleanControlHash == evidence.HeldOutEvidenceHash {
		issues = append(issues, "eval_promotion clean_control_hash and held_out_evidence_hash must differ")
	}
	if evidence.HeldOutStatus != "PASS" {
		issues = append(issues, "eval_promotion held_out_status must be PASS")
	}
	if strings.TrimSpace(evidence.ExpectedFalsePositiveCost) == "" {
		issues = append(issues, "eval_promotion expected_false_positive_cost is required")
	}
	return issues
}

func isProtectedTarget(target string) bool {
	return oneOfString(target, "correctness", "safety", "security", "privacy", "data-integrity", "requirement")
}

func oneOfString(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
