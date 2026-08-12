package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"clean-code/internal/evidence"
	"clean-code/internal/review"
	"clean-code/internal/trace"
)

const maxAuditFileBytes int64 = 20 << 20

type Input struct {
	SchemaVersion      string   `json:"schema_version"`
	Repository         string   `json:"repository"`
	Revision           string   `json:"revision"`
	PolicyRevision     string   `json:"policy_revision"`
	Verification       string   `json:"verification"`
	TestPlan           string   `json:"test_plan"`
	Review             string   `json:"review"`
	SpotCheck          string   `json:"spot_check"`
	SupportingEvidence []string `json:"supporting_evidence"`
	MaxEvidenceAgeSec  int64    `json:"max_evidence_age_seconds,omitempty"`
	Exceptions         []string `json:"exceptions,omitempty"`
}

type SpotCheck struct {
	SchemaVersion string       `json:"schema_version"`
	Revision      string       `json:"revision"`
	Reviewer      string       `json:"reviewer"`
	Checks        []HumanCheck `json:"checks"`
}

type HumanCheck struct {
	Kind     string `json:"kind"`
	Required bool   `json:"required"`
	Status   string `json:"status"`
	Scope    string `json:"scope,omitempty"`
	Outcome  string `json:"outcome,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type Artifact struct {
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Receipt struct {
	SchemaVersion  string     `json:"schema_version"`
	Repository     string     `json:"repository"`
	Revision       string     `json:"revision"`
	PolicyRevision string     `json:"policy_revision"`
	CreatedAt      time.Time  `json:"created_at"`
	Complete       bool       `json:"complete"`
	Artifacts      []Artifact `json:"artifacts"`
	Gaps           []string   `json:"gaps"`
	Exceptions     []string   `json:"exceptions"`
}

func Build(inputPath string, now func() time.Time) (Receipt, error) {
	var input Input
	if _, err := loadStrict(inputPath, &input); err != nil {
		return Receipt{}, fmt.Errorf("load audit input: %w", err)
	}
	if input.SchemaVersion != "1.0.0" || strings.TrimSpace(input.Repository) == "" || strings.TrimSpace(input.Revision) == "" || strings.TrimSpace(input.PolicyRevision) == "" {
		return Receipt{}, errors.New("validate audit input: schema version, repository, revision, and policy revision are required")
	}
	for name, path := range map[string]string{"verification": input.Verification, "test plan": input.TestPlan, "review": input.Review, "spot check": input.SpotCheck} {
		if strings.TrimSpace(path) == "" {
			return Receipt{}, fmt.Errorf("validate audit input: %s path is required", name)
		}
	}
	for _, exception := range input.Exceptions {
		if strings.TrimSpace(exception) == "" {
			return Receipt{}, errors.New("validate audit input: exceptions cannot be empty")
		}
	}
	if input.MaxEvidenceAgeSec < 0 {
		return Receipt{}, errors.New("validate audit input: max evidence age cannot be negative")
	}
	clock := now
	if clock == nil {
		clock = time.Now
	}
	auditTime := clock().UTC()
	receipt := Receipt{
		SchemaVersion: "1.0.0", Repository: input.Repository, Revision: input.Revision,
		PolicyRevision: input.PolicyRevision, CreatedAt: auditTime, Artifacts: []Artifact{},
		Gaps: []string{}, Exceptions: append([]string{}, input.Exceptions...),
	}
	base := filepath.Dir(inputPath)
	supporting, err := inspectSupporting(base, input.SupportingEvidence, &receipt)
	if err != nil {
		return Receipt{}, err
	}
	reviewInput, reviewReport, err := inspectReview(base, input.Review, input.Revision, &receipt)
	if err != nil {
		return Receipt{}, err
	}
	if reviewReport.Status != "PASS" {
		receipt.Gaps = append(receipt.Gaps, "independent review is incomplete")
	}
	if err := inspectVerification(base, input.Verification, input, auditTime, &receipt); err != nil {
		return Receipt{}, err
	}
	if err := inspectTestPlan(base, input.TestPlan, supporting, &receipt); err != nil {
		return Receipt{}, err
	}
	if err := inspectSpotCheck(base, input.SpotCheck, input.Revision, reviewInput.ChangeAuthor, &receipt); err != nil {
		return Receipt{}, err
	}
	sort.Strings(receipt.Gaps)
	receipt.Complete = len(receipt.Gaps) == 0
	return receipt, nil
}

func Write(path string, receipt Receipt) error {
	if path == "" {
		return errors.New("audit receipt output is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve audit output: %w", err)
	}
	if _, err := os.Lstat(absolute); err == nil {
		return errors.New("audit receipt already exists; receipts are immutable")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect audit output: %w", err)
	}
	parent := filepath.Dir(absolute)
	parentInfo, err := os.Lstat(parent)
	if err != nil {
		return fmt.Errorf("inspect audit output directory: %w", err)
	}
	if parentInfo.Mode()&os.ModeSymlink != 0 || !parentInfo.IsDir() {
		return errors.New("audit output directory must be a real directory")
	}
	file, err := os.OpenFile(absolute, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create audit receipt: %w", err)
	}
	completed := false
	defer func() {
		if !completed {
			_ = os.Remove(absolute)
		}
	}()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(receipt); err != nil {
		file.Close()
		return fmt.Errorf("encode audit receipt: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync audit receipt: %w", err)
	}
	if err := file.Close(); err != nil {
		return err
	}
	completed = true
	return nil
}

func Check(inputPath, receiptPath string) (Receipt, error) {
	var recorded Receipt
	if _, err := loadStrict(receiptPath, &recorded); err != nil {
		return Receipt{}, fmt.Errorf("load audit receipt: %w", err)
	}
	if recorded.SchemaVersion != "1.0.0" || recorded.CreatedAt.IsZero() {
		return Receipt{}, errors.New("validate audit receipt: metadata is incomplete")
	}
	rebuilt, err := Build(inputPath, func() time.Time { return recorded.CreatedAt })
	if err != nil {
		return Receipt{}, err
	}
	if !reflect.DeepEqual(recorded, rebuilt) {
		return Receipt{}, errors.New("audit receipt does not match the current evidence bundle")
	}
	return recorded, nil
}

func inspectVerification(base, configured string, input Input, now time.Time, receipt *Receipt) error {
	path := resolve(base, configured)
	var report evidence.Report
	digest, err := loadStrict(path, &report)
	if err != nil {
		return fmt.Errorf("load verification report: %w", err)
	}
	receipt.Artifacts = append(receipt.Artifacts, Artifact{Kind: "verification", Path: configured, SHA256: digest})
	if report.Repository != input.Repository {
		receipt.Gaps = append(receipt.Gaps, "verification repository does not match audit input")
	}
	if report.Revision != input.Revision {
		receipt.Gaps = append(receipt.Gaps, "verification revision does not match audit revision")
	}
	if !report.Successful() {
		receipt.Gaps = append(receipt.Gaps, "verification report is incomplete or failing")
	}
	if report.StartedAt.IsZero() || report.FinishedAt.Before(report.StartedAt) {
		receipt.Gaps = append(receipt.Gaps, "verification timestamps are invalid")
	}
	if input.MaxEvidenceAgeSec > 0 && (report.FinishedAt.After(now) || now.Sub(report.FinishedAt) > time.Duration(input.MaxEvidenceAgeSec)*time.Second) {
		receipt.Gaps = append(receipt.Gaps, "verification evidence is outside the permitted age")
	}
	for _, result := range report.Results {
		if err := result.Validate(); err != nil {
			receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("verification result %q is invalid", result.CheckID))
		}
		if result.Revision != "" && result.Revision != report.Revision {
			receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("verification result %q belongs to another revision", result.CheckID))
		}
	}
	return nil
}

func inspectTestPlan(base, configured string, supporting map[string]bool, receipt *Receipt) error {
	path := resolve(base, configured)
	var plan trace.Plan
	digest, err := loadStrict(path, &plan)
	if err != nil {
		return fmt.Errorf("load test plan: %w", err)
	}
	receipt.Artifacts = append(receipt.Artifacts, Artifact{Kind: "test-plan", Path: configured, SHA256: digest})
	if report := trace.Evaluate(plan); report.Status != "PASS" {
		receipt.Gaps = append(receipt.Gaps, "test trace is incomplete")
	}
	for _, track := range plan.Tracks {
		if track.Status == "PASS" || track.Status == "FAIL" {
			if !supporting[filepath.Clean(track.Evidence)] {
				receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("%s track evidence is not included in the receipt", track.Kind))
			}
		}
	}
	return nil
}

func inspectSupporting(base string, configured []string, receipt *Receipt) (map[string]bool, error) {
	seen := map[string]bool{}
	for _, evidencePath := range configured {
		clean := filepath.Clean(evidencePath)
		if strings.TrimSpace(evidencePath) == "" || seen[clean] {
			return nil, fmt.Errorf("supporting evidence path %q is empty or duplicated", evidencePath)
		}
		path := resolve(base, evidencePath)
		content, err := readRegular(path)
		if err != nil {
			return nil, fmt.Errorf("load supporting evidence %q: %w", evidencePath, err)
		}
		digest := sha256.Sum256(content)
		receipt.Artifacts = append(receipt.Artifacts, Artifact{Kind: "supporting", Path: evidencePath, SHA256: hex.EncodeToString(digest[:])})
		seen[clean] = true
	}
	return seen, nil
}

func inspectReview(base, configured, revision string, receipt *Receipt) (review.Input, review.Report, error) {
	path := resolve(base, configured)
	var input review.Input
	digest, err := loadStrict(path, &input)
	if err != nil {
		return review.Input{}, review.Report{}, fmt.Errorf("load review input: %w", err)
	}
	receipt.Artifacts = append(receipt.Artifacts, Artifact{Kind: "review", Path: configured, SHA256: digest})
	if input.Revision != revision {
		receipt.Gaps = append(receipt.Gaps, "review revision does not match audit revision")
	}
	return input, review.Evaluate(input), nil
}

func inspectSpotCheck(base, configured, revision, changeAuthor string, receipt *Receipt) error {
	path := resolve(base, configured)
	var spot SpotCheck
	digest, err := loadStrict(path, &spot)
	if err != nil {
		return fmt.Errorf("load spot check: %w", err)
	}
	receipt.Artifacts = append(receipt.Artifacts, Artifact{Kind: "spot-check", Path: configured, SHA256: digest})
	if spot.SchemaVersion != "1.0.0" || strings.TrimSpace(spot.Reviewer) == "" {
		receipt.Gaps = append(receipt.Gaps, "spot check metadata is incomplete")
	}
	if spot.Revision != revision {
		receipt.Gaps = append(receipt.Gaps, "spot-check revision does not match audit revision")
	}
	if spot.Reviewer == changeAuthor && spot.Reviewer != "" {
		receipt.Gaps = append(receipt.Gaps, "change author cannot supply independent human spot checks")
	}
	want := []string{"requirements", "acceptance", "ui-qa", "code-sample"}
	seen := map[string]bool{}
	for _, check := range spot.Checks {
		if !contains(want, check.Kind) || seen[check.Kind] {
			receipt.Gaps = append(receipt.Gaps, "spot checks contain an unknown or duplicate kind")
			continue
		}
		seen[check.Kind] = true
		if check.Status == "CHECKED" {
			if strings.TrimSpace(check.Scope) == "" || strings.TrimSpace(check.Outcome) == "" {
				receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("checked %s spot check lacks scope or outcome", check.Kind))
			}
		} else if check.Status == "NOT_CHECKED" || check.Status == "INAPPLICABLE" {
			if strings.TrimSpace(check.Reason) == "" {
				receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("%s spot check lacks a reason", check.Kind))
			}
			if check.Required {
				receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("required %s spot check was not performed", check.Kind))
			}
		} else {
			receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("%s spot check has an invalid status", check.Kind))
		}
	}
	for _, kind := range want {
		if !seen[kind] {
			receipt.Gaps = append(receipt.Gaps, fmt.Sprintf("%s spot check is missing", kind))
		}
	}
	return nil
}

func resolve(base, configured string) string {
	if filepath.IsAbs(configured) {
		return configured
	}
	return filepath.Join(base, filepath.Clean(configured))
}

func loadStrict(path string, target any) (string, error) {
	content, err := readRegular(path)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return "", err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return "", errors.New("unexpected trailing JSON value")
		}
		return "", err
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]), nil
}

func readRegular(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, errors.New("evidence input must be a regular file")
	}
	if info.Size() > maxAuditFileBytes {
		return nil, fmt.Errorf("evidence input exceeds %d bytes", maxAuditFileBytes)
	}
	return os.ReadFile(path)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
