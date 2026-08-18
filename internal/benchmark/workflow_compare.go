package benchmark

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

const maxWorkflowManifestBytes int64 = 2 << 20

type WorkflowManifest struct {
	SchemaVersion string `json:"schema_version"`
	Workflows     []Workflow `json:"workflows"`
	Dimensions    []string   `json:"dimensions"`
}

type Workflow struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Scores  map[string]float64 `json:"scores"`
	Skills  []string           `json:"skills,omitempty"`
	Notes   string             `json:"notes,omitempty"`
}

type WorkflowComparison struct {
	SchemaVersion string             `json:"schema_version"`
	Dimensions    []string           `json:"dimensions"`
	Workflows     []WorkflowScore    `json:"workflows"`
	Advantage     map[string]float64 `json:"advantage"`
	Summary       string             `json:"summary"`
}

type WorkflowScore struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Total       float64            `json:"total"`
	MaxPossible float64            `json:"max_possible"`
	Coverage    float64            `json:"coverage"`
	Scores      map[string]float64 `json:"scores"`
	Skills      []string           `json:"skills,omitempty"`
	Notes       string             `json:"notes,omitempty"`
}

func LoadWorkflowManifest(path string) (WorkflowManifest, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return WorkflowManifest{}, fmt.Errorf("inspect workflow manifest: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return WorkflowManifest{}, errors.New("workflow manifest must be a regular file")
	}
	if info.Size() > maxWorkflowManifestBytes {
		return WorkflowManifest{}, fmt.Errorf("workflow manifest exceeds %d bytes", maxWorkflowManifestBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return WorkflowManifest{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxWorkflowManifestBytes+1))
	decoder.DisallowUnknownFields()
	var manifest WorkflowManifest
	if err := decoder.Decode(&manifest); err != nil {
		return WorkflowManifest{}, fmt.Errorf("parse workflow manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return WorkflowManifest{}, errors.New("parse workflow manifest: unexpected trailing JSON value")
	}
	if err := ValidateWorkflowManifest(manifest); err != nil {
		return WorkflowManifest{}, err
	}
	return manifest, nil
}

func ValidateWorkflowManifest(manifest WorkflowManifest) error {
	if manifest.SchemaVersion != "1.0.0" {
		return errors.New("workflow manifest requires schema version 1.0.0")
	}
	if len(manifest.Dimensions) == 0 || len(manifest.Workflows) < 2 {
		return errors.New("workflow manifest requires dimensions and at least two workflows")
	}
	seenDims := map[string]bool{}
	for _, dimension := range manifest.Dimensions {
		if dimension == "" || seenDims[dimension] {
			return errors.New("workflow dimensions must be present and unique")
		}
		seenDims[dimension] = true
	}
	seenIDs := map[string]bool{}
	for _, workflow := range manifest.Workflows {
		if workflow.ID == "" || seenIDs[workflow.ID] {
			return errors.New("workflow ids must be present and unique")
		}
		seenIDs[workflow.ID] = true
		for _, dimension := range manifest.Dimensions {
			score, ok := workflow.Scores[dimension]
			if !ok {
				return fmt.Errorf("workflow %q missing score for dimension %q", workflow.ID, dimension)
			}
			if score < 0 || score > 1 {
				return fmt.Errorf("workflow %q score for %q must be between 0 and 1", workflow.ID, dimension)
			}
		}
	}
	return nil
}

func CompareWorkflows(manifest WorkflowManifest) WorkflowComparison {
	maxPossible := float64(len(manifest.Dimensions))
	workflows := make([]WorkflowScore, 0, len(manifest.Workflows))
	totals := map[string]float64{}

	for _, workflow := range manifest.Workflows {
		total := 0.0
		for _, dimension := range manifest.Dimensions {
			total += workflow.Scores[dimension]
		}
		totals[workflow.ID] = total
		workflows = append(workflows, WorkflowScore{
			ID:          workflow.ID,
			Name:        workflow.Name,
			Total:       total,
			MaxPossible: maxPossible,
			Coverage:    total / maxPossible,
			Scores:      workflow.Scores,
			Skills:      append([]string{}, workflow.Skills...),
			Notes:       workflow.Notes,
		})
	}

	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].Coverage > workflows[j].Coverage
	})

	advantage := map[string]float64{}
	for _, dimension := range manifest.Dimensions {
		var bestScore float64
		for _, workflow := range manifest.Workflows {
			score := workflow.Scores[dimension]
			if score > bestScore {
				bestScore = score
			}
		}
		advantage[dimension] = bestScore
	}

	summary := buildWorkflowSummary(workflows, totals)

	return WorkflowComparison{
		SchemaVersion: "1.0.0",
		Dimensions:    append([]string{}, manifest.Dimensions...),
		Workflows:     workflows,
		Advantage:     advantage,
		Summary:       summary,
	}
}

func buildWorkflowSummary(workflows []WorkflowScore, totals map[string]float64) string {
	if len(workflows) == 0 {
		return "no workflows to compare"
	}
	leader := workflows[0]
	runnerName := "runner"
	runnerCoverage := 0.0
	if len(workflows) > 1 {
		runnerName = workflows[1].Name
		runnerCoverage = workflows[1].Coverage
	}
	cleanTotal, ceTotal := totals["clean-code"], totals["compound-engineering"]
	delta := cleanTotal - ceTotal
	return fmt.Sprintf(
		"%s leads overall coverage at %.0f%%; %s at %.0f%%. Clean Code total %.1f vs Compound Engineering %.1f (delta %.1f on %d dimensions).",
		leader.Name,
		leader.Coverage*100,
		runnerName,
		runnerCoverage*100,
		cleanTotal,
		ceTotal,
		delta,
		int(leader.MaxPossible),
	)
}
