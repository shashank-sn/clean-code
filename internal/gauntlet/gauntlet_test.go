package gauntlet

import (
	"clean-code/internal/telemetry"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestPacketsAndReorganization(t *testing.T) {
	manifest := Manifest{SchemaVersion: "1.0.0", Revision: "abc", Stories: []Story{{ID: "story", RequirementIDs: []string{"R1"}, Stages: stages(), Events: []telemetry.Event{{Files: []string{"a.go", "b.go"}}, {Files: []string{"a.go", "b.go"}}}}}}
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(Packets(manifest)) != 6 {
		t.Fatal("expected packets")
	}
	report := Evaluate(manifest)
	if report.Stories[0].Decision != telemetry.DecisionReorganizeArchitecture || report.Stories[0].RefactorPacket == nil {
		t.Fatalf("expected refactor packet: %+v", report)
	}
}

func TestManifestRejectsDuplicateOwner(t *testing.T) {
	stages := stages()
	stages[1].Owner = stages[0].Owner
	manifest := Manifest{SchemaVersion: "1.0.0", Revision: "abc", Stories: []Story{{ID: "story", RequirementIDs: []string{"R1"}, Stages: stages}}}
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected duplicate owner rejection")
	}
}

func TestManifestRejectsUnsafeStoryID(t *testing.T) {
	manifest := Manifest{SchemaVersion: "1.0.0", Revision: "abc", Stories: []Story{{ID: "../outside", RequirementIDs: []string{"R1"}, Stages: stages()}}}
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected unsafe story ID rejection")
	}
}

func TestManifestRejectsOutOfOrderPipeline(t *testing.T) {
	stages := stages()
	stages[0], stages[1] = stages[1], stages[0]
	manifest := Manifest{SchemaVersion: "1.0.0", Revision: "abc", Stories: []Story{{ID: "story", RequirementIDs: []string{"R1"}, Stages: stages}}}
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected out-of-order pipeline rejection")
	}
}

func stages() []Stage {
	return []Stage{
		{Role: "specifier", Owner: "spec", Mode: "procedural", AllowedFiles: []string{"docs/spec.md"}, StopCondition: "spec complete"},
		{Role: "implementer", Owner: "build", Mode: "native-host", AllowedFiles: []string{"internal/x.go"}, StopCondition: "implementation complete"},
		{Role: "cleaner", Owner: "clean", Mode: "procedural", AllowedFiles: []string{"internal/x.go"}, StopCondition: "cleanup complete"},
		{Role: "hardener", Owner: "harden", Mode: "mechanical", AllowedFiles: []string{"internal/x.go"}, StopCondition: "hardening complete"},
		{Role: "qa", Owner: "qa", Mode: "mechanical", AllowedFiles: []string{"tests/x_test.go"}, StopCondition: "qa complete"},
		{Role: "reviewer", Owner: "review", Mode: "procedural", AllowedFiles: []string{"docs/review.md"}, StopCondition: "review complete"},
	}
}

func TestWritePacketsDoesNotPartiallyWriteWhenOutputExists(t *testing.T) {
	directory := t.TempDir()
	existing := filepath.Join(directory, "01-story-specifier.json")
	if err := os.WriteFile(existing, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := []WorkPacket{{StoryID: "story", Role: "specifier"}, {StoryID: "story", Role: "qa"}}
	if err := WritePackets(directory, packets); err == nil {
		t.Fatal("expected existing output rejection")
	}
	if _, err := os.Stat(filepath.Join(directory, "02-story-qa.json")); !os.IsNotExist(err) {
		t.Fatalf("unexpected partial output: %v", err)
	}
}
