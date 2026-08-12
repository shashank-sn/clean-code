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
	SchemaVersion      string   `json:"schema_version"`
	ID                 string   `json:"id"`
	SourceReceipt      string   `json:"source_receipt"`
	EvidenceHashes     []string `json:"evidence_hashes"`
	Kind               string   `json:"kind"`
	TargetClass        string   `json:"target_class"`
	Effect             string   `json:"effect"`
	Scope              string   `json:"scope"`
	Change             string   `json:"change"`
	Benefit            string   `json:"benefit"`
	Risks              []string `json:"risks"`
	Rollback           string   `json:"rollback"`
	CalibrationFixture string   `json:"calibration_fixture"`
	Proposer           string   `json:"proposer"`
	Reviewer           string   `json:"reviewer,omitempty"`
	Status             string   `json:"status"`
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
	if isProtectedTarget(proposal.TargetClass) && (proposal.Kind == "suppression" || proposal.Effect == "WEAKEN") {
		issues = append(issues, "protected gates cannot be weakened or suppressed")
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
