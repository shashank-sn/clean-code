package workflow

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"clean-code/internal/agents"
)

type shippingPipeline struct {
	SchemaVersion       string          `json:"schema_version"`
	Stages              []pipelineStage `json:"stages"`
	CleanCodeAdvantages []string        `json:"clean_code_advantages"`
}

type pipelineStage struct {
	ID           string  `json:"id"`
	Skill        string  `json:"skill"`
	CLI          string  `json:"cli"`
	CEEquivalent *string `json:"ce_equivalent"`
	When         string  `json:"when"`
}

func TestShippingPipelineReferencesRegisteredAgents(t *testing.T) {
	file, err := os.Open("shipping-pipeline.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var pipeline shippingPipeline
	if err := decoder.Decode(&pipeline); err != nil {
		t.Fatal(err)
	}
	if pipeline.SchemaVersion != "1.0.0" {
		t.Fatalf("unexpected schema version %q", pipeline.SchemaVersion)
	}
	packages, err := agents.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, stage := range pipeline.Stages {
		if stage.ID == "" || stage.Skill == "" {
			t.Fatalf("pipeline stage must have id and skill: %+v", stage)
		}
		if seen[stage.ID] {
			t.Fatalf("duplicate pipeline stage %q", stage.ID)
		}
		seen[stage.ID] = true
		if _, exists := packages[stage.Skill]; !exists {
			t.Fatalf("pipeline stage %q references unregistered agent %q", stage.ID, stage.Skill)
		}
	}
	evalStageSeen := false
	for _, stage := range pipeline.Stages {
		if stage.ID == "eval-discover" {
			evalStageSeen = true
			if stage.Skill != "clean-eval-discover" || !strings.Contains(stage.When, "confirmed outcomes") {
				t.Fatalf("eval-discover must be a conditional clean-eval-discover stage: %+v", stage)
			}
		}
	}
	if !evalStageSeen {
		t.Fatal("shipping pipeline must include the eval-discover stage")
	}
}
