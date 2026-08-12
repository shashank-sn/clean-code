package trace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const maxPlanBytes int64 = 10 << 20

var trackKinds = []string{"unit", "acceptance", "integration", "ui-qa"}
var trackStatuses = []string{"PLANNED", "PASS", "FAIL", "NOT_AVAILABLE", "NOT_RUN", "INAPPLICABLE"}

type Plan struct {
	SchemaVersion string        `json:"schema_version"`
	Requirements  []Requirement `json:"requirements"`
	Tracks        []Track       `json:"tracks"`
}

type Requirement struct {
	ID                 string   `json:"id"`
	AcceptanceExamples []string `json:"acceptance_examples"`
}

type Track struct {
	Kind           string   `json:"kind"`
	RequirementIDs []string `json:"requirement_ids"`
	Status         string   `json:"status"`
	Owner          string   `json:"owner,omitempty"`
	Evidence       string   `json:"evidence,omitempty"`
	Reason         string   `json:"reason,omitempty"`
	ContextSource  string   `json:"context_source,omitempty"`
}

type Issue struct {
	Kind        string `json:"kind"`
	Requirement string `json:"requirement,omitempty"`
	Track       string `json:"track,omitempty"`
	Summary     string `json:"summary"`
}

type Report struct {
	SchemaVersion string   `json:"schema_version"`
	Status        string   `json:"status"`
	Requirements  int      `json:"requirements"`
	Tracks        int      `json:"tracks"`
	Issues        []Issue  `json:"issues"`
	Warnings      []string `json:"warnings"`
}

func Load(path string) (Plan, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Plan{}, fmt.Errorf("inspect test plan: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Plan{}, errors.New("inspect test plan: input must be a regular file")
	}
	if info.Size() > maxPlanBytes {
		return Plan{}, fmt.Errorf("inspect test plan: input exceeds %d bytes", maxPlanBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return Plan{}, fmt.Errorf("open test plan: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxPlanBytes+1))
	decoder.DisallowUnknownFields()
	var plan Plan
	if err := decoder.Decode(&plan); err != nil {
		return Plan{}, fmt.Errorf("parse test plan: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Plan{}, errors.New("parse test plan: unexpected trailing JSON value")
		}
		return Plan{}, fmt.Errorf("parse test plan: %w", err)
	}
	if plan.SchemaVersion != "1.0.0" {
		return Plan{}, fmt.Errorf("validate test plan: unsupported schema_version %q", plan.SchemaVersion)
	}
	return plan, nil
}

func Evaluate(plan Plan) Report {
	report := Report{
		SchemaVersion: "1.0.0", Status: "PASS", Requirements: len(plan.Requirements),
		Tracks: len(plan.Tracks), Issues: []Issue{}, Warnings: []string{},
	}
	requirements := map[string]bool{}
	coverage := map[string]map[string]bool{}
	if plan.SchemaVersion != "1.0.0" {
		report.Issues = append(report.Issues, Issue{Kind: "schema-version", Summary: "unsupported test plan schema version"})
	}
	if len(plan.Requirements) == 0 {
		report.Issues = append(report.Issues, Issue{Kind: "missing-requirement", Summary: "test plan has no requirements"})
	}
	for _, requirement := range plan.Requirements {
		id := strings.TrimSpace(requirement.ID)
		if id == "" {
			report.Issues = append(report.Issues, Issue{Kind: "invalid-requirement", Summary: "requirement id is required"})
			continue
		}
		if requirements[id] {
			report.Issues = append(report.Issues, Issue{Kind: "duplicate-requirement", Requirement: id, Summary: "requirement id appears more than once"})
			continue
		}
		requirements[id] = true
		coverage[id] = map[string]bool{}
		if len(requirement.AcceptanceExamples) == 0 {
			report.Issues = append(report.Issues, Issue{Kind: "missing-acceptance-example", Requirement: id, Summary: "requirement has no acceptance example"})
		}
		for _, example := range requirement.AcceptanceExamples {
			if strings.TrimSpace(example) == "" {
				report.Issues = append(report.Issues, Issue{Kind: "invalid-acceptance-example", Requirement: id, Summary: "acceptance example cannot be empty"})
			}
		}
	}
	for index, track := range plan.Tracks {
		if !contains(trackKinds, track.Kind) {
			report.Issues = append(report.Issues, Issue{Kind: "invalid-track", Track: track.Kind, Summary: fmt.Sprintf("track %d has an unknown kind", index)})
			continue
		}
		if len(track.RequirementIDs) == 0 {
			report.Issues = append(report.Issues, Issue{Kind: "unlinked-track", Track: track.Kind, Summary: "track has no requirement ids"})
		}
		if !contains(trackStatuses, track.Status) {
			report.Issues = append(report.Issues, Issue{Kind: "invalid-status", Track: track.Kind, Summary: fmt.Sprintf("track %d has an unknown status", index)})
		}
		if track.ContextSource != "" && track.ContextSource != "requirements" && track.ContextSource != "implementation" && track.ContextSource != "mixed" {
			report.Issues = append(report.Issues, Issue{Kind: "invalid-context", Track: track.Kind, Summary: fmt.Sprintf("track %d has an unknown context source", index)})
		}
		if track.Status == "INAPPLICABLE" {
			if strings.TrimSpace(track.Reason) == "" {
				report.Issues = append(report.Issues, Issue{Kind: "missing-reason", Track: track.Kind, Summary: "inapplicable track requires a reason"})
			}
		} else if strings.TrimSpace(track.Owner) == "" {
			report.Issues = append(report.Issues, Issue{Kind: "missing-owner", Track: track.Kind, Summary: "applicable track requires an owner"})
		}
		if track.Status == "PASS" || track.Status == "FAIL" {
			if strings.TrimSpace(track.Evidence) == "" {
				report.Issues = append(report.Issues, Issue{Kind: "missing-evidence", Track: track.Kind, Summary: "executed track requires evidence"})
			}
		}
		if track.Status == "NOT_AVAILABLE" || track.Status == "NOT_RUN" {
			if strings.TrimSpace(track.Reason) == "" {
				report.Issues = append(report.Issues, Issue{Kind: "missing-reason", Track: track.Kind, Summary: "unavailable or unrun track requires a reason"})
			}
		}
		if (track.Kind == "acceptance" || track.Kind == "ui-qa") && track.ContextSource != "requirements" {
			report.Warnings = append(report.Warnings, fmt.Sprintf("%s track was authored from %s context", track.Kind, contextLabel(track.ContextSource)))
		}
		for _, requirementID := range track.RequirementIDs {
			if strings.TrimSpace(requirementID) == "" {
				report.Issues = append(report.Issues, Issue{Kind: "invalid-requirement-reference", Track: track.Kind, Summary: "track requirement id cannot be empty"})
				continue
			}
			if !requirements[requirementID] {
				report.Issues = append(report.Issues, Issue{Kind: "unknown-requirement", Requirement: requirementID, Track: track.Kind, Summary: "track references an unknown requirement"})
				continue
			}
			coverage[requirementID][track.Kind] = true
		}
	}
	for requirementID := range requirements {
		for _, kind := range trackKinds {
			if !coverage[requirementID][kind] {
				report.Issues = append(report.Issues, Issue{
					Kind: "missing-track", Requirement: requirementID, Track: kind,
					Summary: fmt.Sprintf("requirement %q has no %s track", requirementID, kind),
				})
			}
		}
	}
	sort.Slice(report.Issues, func(i, j int) bool {
		left := report.Issues[i].Kind + report.Issues[i].Requirement + report.Issues[i].Track
		right := report.Issues[j].Kind + report.Issues[j].Requirement + report.Issues[j].Track
		return left < right
	})
	sort.Strings(report.Warnings)
	if len(report.Issues) > 0 {
		report.Status = "FAIL"
	}
	return report
}

func contextLabel(value string) string {
	if value == "" {
		return "unspecified"
	}
	return value
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
