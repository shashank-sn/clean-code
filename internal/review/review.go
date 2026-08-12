package review

import (
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
	SchemaVersion string    `json:"schema_version"`
	Revision      string    `json:"revision"`
	ChangeAuthor  string    `json:"change_author"`
	Reviewer      string    `json:"reviewer"`
	Scope         []string  `json:"scope"`
	Findings      []Finding `json:"findings"`
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
}

type Issue struct {
	Kind    string `json:"kind"`
	Finding string `json:"finding,omitempty"`
	Summary string `json:"summary"`
}

type Report struct {
	SchemaVersion string    `json:"schema_version"`
	Revision      string    `json:"revision"`
	Status        string    `json:"status"`
	Reviewer      string    `json:"reviewer"`
	Scope         []string  `json:"scope"`
	Findings      []Finding `json:"findings"`
	Issues        []Issue   `json:"issues"`
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
	decoder := json.NewDecoder(io.LimitReader(file, maxReviewBytes+1))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("parse review input: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Input{}, errors.New("parse review input: unexpected trailing JSON value")
		}
		return Input{}, fmt.Errorf("parse review input: %w", err)
	}
	return input, nil
}

func Evaluate(input Input) Report {
	report := Report{
		SchemaVersion: "1.0.0", Revision: input.Revision, Status: "PASS",
		Reviewer: input.Reviewer, Scope: input.Scope, Findings: input.Findings, Issues: []Issue{},
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
	if input.Reviewer == input.ChangeAuthor && input.Reviewer != "" {
		report.Issues = append(report.Issues, Issue{Kind: "self-review", Summary: "change author cannot provide independent review approval"})
	}
	ids := map[string]bool{}
	locations := map[string]string{}
	for index, finding := range input.Findings {
		label := finding.ID
		if label == "" {
			label = fmt.Sprintf("index-%d", index)
		}
		if strings.TrimSpace(finding.ID) == "" || strings.TrimSpace(finding.Behavior) == "" || strings.TrimSpace(finding.Evidence) == "" || strings.TrimSpace(finding.Consequence) == "" || strings.TrimSpace(finding.Fix) == "" {
			report.Issues = append(report.Issues, Issue{Kind: "unsupported-finding", Finding: label, Summary: "finding requires id, behavior, evidence, consequence, and bounded fix"})
		}
		if ids[finding.ID] && finding.ID != "" {
			report.Issues = append(report.Issues, Issue{Kind: "duplicate-id", Finding: finding.ID, Summary: "finding id appears more than once"})
		}
		ids[finding.ID] = true
		if finding.Line < 0 {
			report.Issues = append(report.Issues, Issue{Kind: "invalid-line", Finding: label, Summary: "finding line cannot be negative"})
		}
		if !oneOf(finding.Severity, "BLOCKING", "IMPROVEMENT", "ADVISORY") || !oneOf(finding.Confidence, "HIGH", "MEDIUM", "LOW") || !oneOf(finding.Disposition, "OPEN", "APPLIED", "DISMISSED", "ACCEPTED_RISK") {
			report.Issues = append(report.Issues, Issue{Kind: "invalid-enum", Finding: label, Summary: "finding severity, confidence, or disposition is invalid"})
		}
		if (finding.Disposition == "DISMISSED" || finding.Disposition == "ACCEPTED_RISK") && strings.TrimSpace(finding.ResolutionReason) == "" {
			report.Issues = append(report.Issues, Issue{Kind: "missing-resolution", Finding: label, Summary: "dismissed or accepted-risk finding requires a reason"})
		}
		key := strings.Join([]string{finding.File, fmt.Sprint(finding.Line), finding.RuleID, finding.Behavior}, "|")
		if previous, ok := locations[key]; ok {
			report.Issues = append(report.Issues, Issue{Kind: "duplicate-finding", Finding: label, Summary: fmt.Sprintf("finding duplicates %q", previous)})
		} else {
			locations[key] = label
		}
		if finding.Severity == "BLOCKING" && finding.Disposition != "APPLIED" && finding.Disposition != "DISMISSED" {
			report.Issues = append(report.Issues, Issue{Kind: "blocking-finding", Finding: label, Summary: "blocking finding remains unresolved"})
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

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
