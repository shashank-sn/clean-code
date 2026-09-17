package review

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const maxReviewBytes int64 = 10 << 20

type Input struct {
	SchemaVersion string      `json:"schema_version"`
	BaseRevision  string      `json:"base_revision,omitempty"`
	Revision      string      `json:"revision"`
	ChangeAuthor  string      `json:"change_author"`
	Reviewer      string      `json:"reviewer"`
	Scope         []string    `json:"scope"`
	Requirements  []string    `json:"requirements,omitempty"`
	Dimensions    []Dimension `json:"dimensions,omitempty"`
	Coverage      []Coverage  `json:"coverage,omitempty"`
	Checks        []Check     `json:"checks,omitempty"`
	Limitations   []string    `json:"limitations"`
	Completion    string      `json:"completion,omitempty"`
	Findings      []Finding   `json:"findings"`
}

const (
	LegacySchemaVersion = "1.0.0"
	V2SchemaVersion     = "2.0.0"
	ArtifactOnly        = "ARTIFACT_ONLY"
)

var reviewDimensions = []string{"correctness", "integration", "tests", "failure_modes", "security", "maintainability"}

type Dimension struct {
	ID         string   `json:"id"`
	Status     string   `json:"status"`
	Evidence   []string `json:"evidence,omitempty"`
	Limitation string   `json:"limitation,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

type Coverage struct {
	Path     string `json:"path"`
	Status   string `json:"status"`
	Evidence string `json:"evidence,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type Check struct {
	ID                string `json:"id"`
	Kind              string `json:"kind"`
	Required          bool   `json:"required"`
	Status            string `json:"status"`
	Revision          string `json:"revision"`
	Source            string `json:"source"`
	Artifact          string `json:"artifact"`
	SHA256            string `json:"sha256"`
	UnavailableReason string `json:"unavailable_reason"`
}

type Finding struct {
	ID               string `json:"id"`
	Severity         string `json:"severity"`
	File             string `json:"file,omitempty"`
	Line             int    `json:"line,omitempty"`
	Behavior         string `json:"behavior"`
	RuleID           string `json:"rule_id,omitempty"`
	Evidence         string `json:"evidence"`
	Consequence      string `json:"consequence"`
	Confidence       string `json:"confidence"`
	Fix              string `json:"fix"`
	Disposition      string `json:"disposition"`
	ResolutionReason string `json:"resolution_reason,omitempty"`
	ResolutionCheck  string `json:"resolution_check,omitempty"`
}

type Issue struct {
	Kind    string `json:"kind"`
	Finding string `json:"finding,omitempty"`
	Summary string `json:"summary"`
}

type Report struct {
	SchemaVersion   string      `json:"schema_version"`
	Revision        string      `json:"revision"`
	Status          string      `json:"status"`
	ValidationScope string      `json:"validation_scope"`
	Completion      string      `json:"completion"`
	BaseRevision    string      `json:"base_revision,omitempty"`
	ChangeAuthor    string      `json:"change_author,omitempty"`
	Reviewer        string      `json:"reviewer"`
	Scope           []string    `json:"scope"`
	Requirements    []string    `json:"requirements,omitempty"`
	Dimensions      []Dimension `json:"dimensions,omitempty"`
	Coverage        []Coverage  `json:"coverage,omitempty"`
	Checks          []Check     `json:"checks,omitempty"`
	Limitations     []string    `json:"limitations"`
	Findings        []Finding   `json:"findings"`
	Issues          []Issue     `json:"issues"`
}

func Load(path string) (Input, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Input{}, fmt.Errorf("inspect review input: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Input{}, errors.New("inspect review input: input must be a regular file")
	}
	if info.Size() > maxReviewBytes {
		return Input{}, fmt.Errorf("inspect review input: input exceeds %d bytes", maxReviewBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return Input{}, fmt.Errorf("open review input: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxReviewBytes+1))
	if err != nil {
		return Input{}, fmt.Errorf("read review input: %w", err)
	}
	if int64(len(data)) > maxReviewBytes {
		return Input{}, fmt.Errorf("inspect review input: input exceeds %d bytes", maxReviewBytes)
	}
	return Decode(data)
}

// Decode applies the same strict, version-aware interpretation used by Load to
// an already-read artifact. Callers that hash a snapshot can therefore decode
// those exact bytes without a second read.
func Decode(data []byte) (Input, error) {
	var envelope struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return Input{}, fmt.Errorf("parse review input: %w", err)
	}
	if envelope.SchemaVersion == LegacySchemaVersion {
		var legacy legacyInput
		if err := decodeStrict(data, &legacy); err != nil {
			return Input{}, fmt.Errorf("parse review input: %w", err)
		}
		return Input{SchemaVersion: legacy.SchemaVersion, Revision: legacy.Revision, ChangeAuthor: legacy.ChangeAuthor, Reviewer: legacy.Reviewer, Scope: legacy.Scope, Findings: legacy.Findings}, nil
	}
	if envelope.SchemaVersion == V2SchemaVersion {
		var input Input
		if err := decodeStrict(data, &input); err != nil {
			return Input{}, fmt.Errorf("parse review input: %w", err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return Input{}, fmt.Errorf("parse review input: %w", err)
		}
		var decoded any
		if err := json.Unmarshal(data, &decoded); err != nil {
			return Input{}, fmt.Errorf("parse review input: %w", err)
		}
		if containsNull(decoded) {
			return Input{}, errors.New("parse review input: v2 fields cannot be null")
		}
		findings, ok := raw["findings"]
		if !ok || string(findings) == "null" {
			return Input{}, errors.New("parse review input: v2 findings must be a non-null array")
		}
		var rawFindings []map[string]json.RawMessage
		if err := json.Unmarshal(findings, &rawFindings); err != nil {
			return Input{}, fmt.Errorf("parse review input: %w", err)
		}
		for i, rawFinding := range rawFindings {
			if line, ok := rawFinding["line"]; ok {
				var value int
				if err := json.Unmarshal(line, &value); err != nil || value < 1 {
					return Input{}, fmt.Errorf("parse review input: v2 finding %d line must be positive when supplied", i)
				}
			}
		}
		if _, ok := raw["limitations"]; !ok {
			return Input{}, errors.New("parse review input: v2 limitations must be declared")
		}
		var rawChecks []map[string]json.RawMessage
		if checks, ok := raw["checks"]; ok && string(checks) != "null" {
			if err := json.Unmarshal(checks, &rawChecks); err != nil {
				return Input{}, fmt.Errorf("parse review input: %w", err)
			}
		}
		for i, check := range input.Checks {
			if i >= len(rawChecks) {
				break
			}
			if _, ok := rawChecks[i]["required"]; !ok {
				return Input{}, fmt.Errorf("parse review input: check %q must declare required", check.ID)
			}
		}
		return input, nil
	}
	// Preserve the legacy loader's behavior for unknown versions: loading is
	// structural, while Evaluate reports the unsupported version deterministically.
	var legacy legacyInput
	if err := decodeStrict(data, &legacy); err != nil {
		return Input{}, fmt.Errorf("parse review input: %w", err)
	}
	return Input{SchemaVersion: legacy.SchemaVersion, Revision: legacy.Revision, ChangeAuthor: legacy.ChangeAuthor, Reviewer: legacy.Reviewer, Scope: legacy.Scope, Findings: legacy.Findings}, nil
}

type legacyInput struct {
	SchemaVersion string    `json:"schema_version"`
	Revision      string    `json:"revision"`
	ChangeAuthor  string    `json:"change_author"`
	Reviewer      string    `json:"reviewer"`
	Scope         []string  `json:"scope"`
	Findings      []Finding `json:"findings"`
}

func decodeStrict(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}

func Evaluate(input Input) Report {
	if input.SchemaVersion == V2SchemaVersion {
		return evaluateV2(input)
	}
	report := Report{
		SchemaVersion: LegacySchemaVersion, Revision: input.Revision, Status: "PASS", Completion: "NOT_ASSESSED", ValidationScope: ArtifactOnly,
		Reviewer: input.Reviewer, Scope: input.Scope, Limitations: []string{}, Findings: input.Findings, Issues: []Issue{},
	}
	if input.SchemaVersion != "1.0.0" {
		report.Issues = append(report.Issues, Issue{Kind: "schema-version", Summary: "unsupported review schema version"})
	}
	if strings.TrimSpace(input.Revision) == "" || strings.TrimSpace(input.Reviewer) == "" || strings.TrimSpace(input.ChangeAuthor) == "" || len(input.Scope) == 0 {
		report.Issues = append(report.Issues, Issue{Kind: "incomplete-review", Summary: "revision, change author, reviewer, and scope are required"})
	}
	for _, scope := range input.Scope {
		if strings.TrimSpace(scope) == "" {
			report.Issues = append(report.Issues, Issue{Kind: "incomplete-review", Summary: "review scope entries cannot be empty"})
			break
		}
	}
	if sameIdentity(input.Reviewer, input.ChangeAuthor) {
		report.Issues = append(report.Issues, Issue{Kind: "self-review", Summary: "change author cannot provide independent review approval"})
	}
	ids := map[string]bool{}
	locations := map[string]string{}
	for index, finding := range input.Findings {
		label := finding.ID
		if label == "" {
			label = fmt.Sprintf("index-%d", index)
		}
		report.Issues = append(report.Issues, findingIssues(finding, index, true)...)
		if ids[finding.ID] && finding.ID != "" {
			report.Issues = append(report.Issues, Issue{Kind: "duplicate-id", Finding: finding.ID, Summary: "finding id appears more than once"})
		}
		ids[finding.ID] = true
		key := findingLocationKey(finding)
		if previous, ok := locations[key]; ok {
			report.Issues = append(report.Issues, Issue{Kind: "duplicate-finding", Finding: label, Summary: fmt.Sprintf("finding duplicates %q", previous)})
		} else {
			locations[key] = label
		}
	}
	sort.Slice(report.Issues, func(i, j int) bool {
		return report.Issues[i].Kind+report.Issues[i].Finding < report.Issues[j].Kind+report.Issues[j].Finding
	})
	if len(report.Issues) > 0 {
		report.Status = "FAIL"
	}
	return report
}

func evaluateV2(input Input) Report {
	report := Report{
		SchemaVersion: V2SchemaVersion, BaseRevision: input.BaseRevision, Revision: input.Revision,
		ChangeAuthor: input.ChangeAuthor, Reviewer: input.Reviewer, Scope: input.Scope,
		Requirements: input.Requirements, Dimensions: input.Dimensions, Coverage: input.Coverage,
		Checks: input.Checks, Limitations: input.Limitations, Completion: "INCOMPLETE",
		ValidationScope: ArtifactOnly, Status: "FAIL", Findings: input.Findings, Issues: []Issue{},
	}
	issue := func(kind, finding, summary string) {
		report.Issues = append(report.Issues, Issue{Kind: kind, Finding: finding, Summary: summary})
	}
	validateV2Metadata(input, issue)
	validateV2Dimensions(input, issue)
	validateV2Coverage(input, issue)
	checkByID := validateV2Checks(input, issue)
	validateV2Findings(&report, input, checkByID, issue)
	if input.Completion != "COMPLETE" && input.Completion != "INCOMPLETE" {
		issue("invalid-enum", "", "completion must be COMPLETE or INCOMPLETE")
	}
	if input.Completion == "COMPLETE" {
		for _, dimension := range input.Dimensions {
			if dimension.Status != "PASS" && dimension.Status != "FAIL" && dimension.Status != "NOT_APPLICABLE" {
				issue("incomplete-dimension", dimension.ID, "complete review requires assessed dimensions")
			}
		}
		for _, path := range input.Scope {
			if coverage := coverageFor(input.Coverage, path); coverage.Status != "REVIEWED" {
				issue("incomplete-coverage", path, "complete review cannot contain unreviewed scope")
			}
		}
		for _, check := range input.Checks {
			if check.Required && check.Status != "PASS" && check.Status != "FAIL" {
				issue("required-check", check.ID, "complete review requires an executed required check")
			}
		}
	}
	if input.Completion == "COMPLETE" && completionValid(report.Issues) {
		report.Completion = "COMPLETE"
	}
	if len(report.Issues) == 0 && report.Completion == "COMPLETE" && !reviewHasFailures(input) {
		report.Status = "PASS"
	}
	sort.Slice(report.Issues, func(i, j int) bool {
		a, b := report.Issues[i], report.Issues[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Finding != b.Finding {
			return a.Finding < b.Finding
		}
		return a.Summary < b.Summary
	})
	return report
}

func validateV2Metadata(input Input, issue func(string, string, string)) {
	if strings.TrimSpace(input.BaseRevision) == "" || strings.TrimSpace(input.Revision) == "" || strings.TrimSpace(input.ChangeAuthor) == "" || strings.TrimSpace(input.Reviewer) == "" {
		issue("incomplete-review", "", "base revision, revision, change author, and reviewer are required")
	}
	if sameIdentity(input.Reviewer, input.ChangeAuthor) {
		issue("self-review", "", "change author cannot provide independent review approval")
	}
	if len(input.Scope) == 0 {
		issue("incomplete-review", "", "review scope cannot be empty")
	}
	seen := map[string]bool{}
	for _, path := range input.Scope {
		if strings.TrimSpace(path) == "" {
			issue("incomplete-review", "", "review scope entries cannot be empty")
		}
		if seen[path] {
			issue("duplicate-scope", path, "review scope entry appears more than once")
		}
		seen[path] = true
	}
	if len(input.Requirements) == 0 {
		issue("incomplete-review", "", "requirements cannot be empty")
	}
	if input.Limitations == nil {
		issue("incomplete-review", "", "limitations must be declared")
	}
	seen = map[string]bool{}
	for _, requirement := range input.Requirements {
		if strings.TrimSpace(requirement) == "" {
			issue("incomplete-review", "", "requirements cannot be empty")
		}
		if seen[requirement] {
			issue("duplicate-requirement", requirement, "requirement appears more than once")
		}
		seen[requirement] = true
	}
}

func validateV2Dimensions(input Input, issue func(string, string, string)) {
	seen := map[string]bool{}
	for _, dimension := range input.Dimensions {
		if !oneOf(dimension.ID, reviewDimensions...) {
			issue("invalid-dimension", dimension.ID, "dimension id is not supported")
		}
		if seen[dimension.ID] {
			issue("duplicate-dimension", dimension.ID, "dimension appears more than once")
		}
		seen[dimension.ID] = true
		if !oneOf(dimension.Status, "PASS", "FAIL", "INCOMPLETE", "NOT_RUN", "NOT_AVAILABLE", "NOT_CONFIGURED", "STALE", "ERROR", "NOT_APPLICABLE") {
			issue("invalid-enum", dimension.ID, "dimension status is invalid")
		}
		if dimension.Status == "PASS" || dimension.Status == "FAIL" {
			if len(nonEmpty(dimension.Evidence)) == 0 {
				issue("missing-evidence", dimension.ID, "pass or fail dimension requires evidence")
			}
		} else if strings.TrimSpace(dimension.Limitation) == "" && strings.TrimSpace(dimension.Reason) == "" {
			issue("missing-reason", dimension.ID, "incomplete or not-applicable dimension requires a reason")
		}
	}
	for _, id := range reviewDimensions {
		if !seen[id] {
			issue("missing-dimension", id, "required review dimension is missing")
		}
	}
	if len(input.Dimensions) != len(reviewDimensions) {
		issue("invalid-dimension", "", "exactly six review dimensions are required")
	}
}

func validateV2Coverage(input Input, issue func(string, string, string)) {
	seen := map[string]bool{}
	for _, coverage := range input.Coverage {
		if strings.TrimSpace(coverage.Path) == "" {
			issue("incomplete-coverage", "", "coverage path cannot be empty")
		}
		if seen[coverage.Path] {
			issue("duplicate-coverage", coverage.Path, "coverage path appears more than once")
		}
		seen[coverage.Path] = true
		if !oneOf(coverage.Status, "REVIEWED", "NOT_REVIEWED") {
			issue("invalid-enum", coverage.Path, "coverage status is invalid")
		} else if coverage.Status == "REVIEWED" && strings.TrimSpace(coverage.Evidence) == "" {
			issue("missing-evidence", coverage.Path, "reviewed coverage requires evidence")
		} else if coverage.Status == "NOT_REVIEWED" && strings.TrimSpace(coverage.Reason) == "" {
			issue("missing-reason", coverage.Path, "unreviewed coverage requires a reason")
		}
	}
	for _, path := range input.Scope {
		if !seen[path] {
			issue("missing-coverage", path, "every scope path requires exactly one coverage entry")
		}
	}
	if len(input.Coverage) != len(input.Scope) {
		issue("invalid-coverage", "", "coverage must contain exactly one entry per scope path")
	}
}

func validateV2Checks(input Input, issue func(string, string, string)) map[string]Check {
	checksByID := map[string]Check{}
	limitationsPresent := len(nonEmpty(input.Limitations)) > 0
	for _, check := range input.Checks {
		if strings.TrimSpace(check.ID) == "" || strings.TrimSpace(check.Kind) == "" {
			issue("incomplete-check", check.ID, "check requires id and kind")
		}
		if checksByID[check.ID].ID != "" {
			issue("duplicate-check", check.ID, "check id appears more than once")
		}
		checksByID[check.ID] = check
		if !oneOf(check.Status, "PASS", "FAIL", "INCOMPLETE", "NOT_RUN", "NOT_AVAILABLE", "NOT_CONFIGURED", "STALE", "ERROR") {
			issue("invalid-enum", check.ID, "check status is invalid")
			continue
		}
		if strings.TrimSpace(check.Source) == "" {
			issue("incomplete-check", check.ID, "check source is required")
		}
		if check.Status == "PASS" || check.Status == "FAIL" {
			if check.Revision != input.Revision {
				issue("revision-mismatch", check.ID, "executed check revision must match candidate revision")
			}
			if strings.TrimSpace(check.Artifact) == "" || strings.TrimSpace(check.SHA256) == "" {
				issue("incomplete-check", check.ID, "executed check requires artifact and sha256 declarations")
			}
		} else if strings.TrimSpace(check.UnavailableReason) == "" && !limitationsPresent {
			issue("missing-reason", check.ID, "non-executed check requires an unavailable or incomplete reason")
		}
		if strings.TrimSpace(check.SHA256) != "" && !validSHA256(check.SHA256) {
			issue("invalid-check", check.ID, "sha256 must be 64 hexadecimal characters")
		}
	}
	if len(input.Checks) == 0 {
		issue("incomplete-review", "", "checks cannot be empty")
	}
	return checksByID
}

func validateV2Findings(report *Report, input Input, checks map[string]Check, issue func(string, string, string)) {
	if input.Findings == nil {
		issue("incomplete-findings", "", "v2 findings must be a non-null array")
	}
	seenIDs := map[string]bool{}
	locations := map[string]string{}
	for index, finding := range input.Findings {
		if strings.TrimSpace(finding.ID) != "" && seenIDs[finding.ID] {
			issue("duplicate-id", finding.ID, "finding id appears more than once")
		}
		seenIDs[finding.ID] = true
		validateFinding(report, finding, index, input, checks, locations, issue)
	}
}

func completionValid(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Kind != "blocking-finding" {
			return false
		}
	}
	return true
}

func reviewHasFailures(input Input) bool {
	for _, dimension := range input.Dimensions {
		if dimension.Status == "FAIL" {
			return true
		}
	}
	for _, check := range input.Checks {
		if check.Status == "FAIL" {
			return true
		}
	}
	return false
}

func findingIssues(finding Finding, index int, enforceBlocking bool) []Issue {
	label := findingLabel(finding, index)
	issues := []Issue{}
	if strings.TrimSpace(finding.ID) == "" || strings.TrimSpace(finding.Behavior) == "" || strings.TrimSpace(finding.Evidence) == "" || strings.TrimSpace(finding.Consequence) == "" || strings.TrimSpace(finding.Fix) == "" {
		issues = append(issues, Issue{Kind: "unsupported-finding", Finding: label, Summary: "finding requires id, behavior, evidence, consequence, and bounded fix"})
	}
	if finding.Line < 0 {
		issues = append(issues, Issue{Kind: "invalid-line", Finding: label, Summary: "finding line cannot be negative"})
	}
	if !oneOf(finding.Severity, "BLOCKING", "IMPROVEMENT", "ADVISORY") || !oneOf(finding.Confidence, "HIGH", "MEDIUM", "LOW") || !oneOf(finding.Disposition, "OPEN", "APPLIED", "DISMISSED", "ACCEPTED_RISK") {
		issues = append(issues, Issue{Kind: "invalid-enum", Finding: label, Summary: "finding severity, confidence, or disposition is invalid"})
	}
	if (finding.Disposition == "DISMISSED" || finding.Disposition == "ACCEPTED_RISK") && strings.TrimSpace(finding.ResolutionReason) == "" {
		issues = append(issues, Issue{Kind: "missing-resolution", Finding: label, Summary: "dismissed or accepted-risk finding requires a reason"})
	}
	if enforceBlocking && finding.Severity == "BLOCKING" && finding.Disposition != "APPLIED" && finding.Disposition != "DISMISSED" {
		issues = append(issues, Issue{Kind: "blocking-finding", Finding: label, Summary: "blocking finding remains unresolved"})
	}
	return issues
}

func findingLabel(finding Finding, index int) string {
	if finding.ID != "" {
		return finding.ID
	}
	return fmt.Sprintf("index-%d", index)
}

func findingLocationKey(finding Finding) string {
	return strings.Join([]string{finding.File, fmt.Sprint(finding.Line), finding.RuleID, finding.Behavior}, "|")
}

func validateFinding(report *Report, finding Finding, index int, input Input, checks map[string]Check, locations map[string]string, issue func(string, string, string)) {
	for _, findingIssue := range findingIssues(finding, index, true) {
		issue(findingIssue.Kind, findingIssue.Finding, findingIssue.Summary)
	}
	label := findingLabel(finding, index)
	key := findingLocationKey(finding)
	if previous, ok := locations[key]; ok {
		issue("duplicate-finding", label, fmt.Sprintf("finding duplicates %q", previous))
	} else {
		locations[key] = label
	}
	if finding.Severity == "BLOCKING" && finding.Disposition == "ACCEPTED_RISK" {
		issue("blocking-finding", label, "accepted-risk disposition does not resolve a blocking finding")
	}
	if finding.Severity == "BLOCKING" && (finding.Disposition == "APPLIED" || finding.Disposition == "DISMISSED") {
		if strings.TrimSpace(finding.ResolutionReason) == "" {
			issue("missing-resolution", label, "resolved blocking finding requires a reason")
		}
		check, ok := checks[finding.ResolutionCheck]
		if !ok || check.Status != "PASS" || check.Revision != input.Revision {
			issue("invalid-resolution-check", label, "resolved blocking finding requires a successful check on the final revision")
		}
	}
}

func coverageFor(coverage []Coverage, path string) Coverage {
	for _, item := range coverage {
		if item.Path == path {
			return item
		}
	}
	return Coverage{}
}
func nonEmpty(values []string) []string {
	var out []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func containsNull(value any) bool {
	switch value := value.(type) {
	case nil:
		return true
	case []any:
		for _, item := range value {
			if containsNull(item) {
				return true
			}
		}
	case map[string]any:
		for _, item := range value {
			if containsNull(item) {
				return true
			}
		}
	}
	return false
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func sameIdentity(left, right string) bool {
	left, right = strings.TrimSpace(left), strings.TrimSpace(right)
	return left != "" && left == right
}
