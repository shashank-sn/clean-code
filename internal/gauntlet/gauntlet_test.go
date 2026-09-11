package gauntlet

import (
	"github.com/shashank-sn/clean-code/internal/contracts"
	"github.com/shashank-sn/clean-code/internal/telemetry"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestValidateStageArtifactsMissingFails(t *testing.T) {
	root := t.TempDir()
	stage := Stage{Role: "specifier", Owner: "spec", Mode: "procedural", ExpectedArtifacts: []string{"docs/spec.md"}}
	report := validateStageArtifacts(stage, root, RevisionInfo{})
	if report.Status != contracts.StatusFail {
		t.Fatalf("expected FAIL, got %s (%s)", report.Status, report.ValidationMessage)
	}
	if len(report.MissingArtifacts) != 1 || report.MissingArtifacts[0] != "docs/spec.md" {
		t.Fatalf("expected docs/spec.md missing: %+v", report.MissingArtifacts)
	}
	if !report.Validated {
		t.Fatal("expected stage to be validated")
	}
	if report.Role != "specifier" || report.Mode != "procedural" {
		t.Fatalf("expected stage identity preserved: %+v", report)
	}
}

func TestValidateStageArtifactsPresentFreshPasses(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(root, "docs", "spec.md")
	if err := os.WriteFile(artifact, []byte("spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	stage := Stage{Role: "specifier", Owner: "spec", Mode: "procedural", ExpectedArtifacts: []string{"docs/spec.md"}}
	revision := RevisionInfo{CommitTime: time.Now().Add(-time.Hour), Resolved: true}
	report := validateStageArtifacts(stage, root, revision)
	if report.Status != contracts.StatusPass {
		t.Fatalf("expected PASS, got %s (%s)", report.Status, report.ValidationMessage)
	}
	if len(report.MissingArtifacts) != 0 || len(report.StaleArtifacts) != 0 {
		t.Fatalf("expected no problems: %+v", report)
	}
}

func TestValidateStageArtifactsSymlinkNotRegularFails(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	if err := os.WriteFile(target, []byte("target"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "link.txt")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	stage := Stage{Role: "implementer", Owner: "build", Mode: "native-host", ExpectedArtifacts: []string{"link.txt"}}
	report := validateStageArtifacts(stage, root, RevisionInfo{})
	if report.Status != contracts.StatusFail {
		t.Fatalf("expected symlink to fail, got %s (%s)", report.Status, report.ValidationMessage)
	}
	if len(report.MissingArtifacts) != 1 {
		t.Fatalf("expected symlink treated as missing/non-regular: %+v", report.MissingArtifacts)
	}
}

func TestValidateStageArtifactsSymlinkParentEscapeFails(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "escaped.txt"), []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape-link")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	stage := Stage{Role: "implementer", Owner: "build", Mode: "native-host", ExpectedArtifacts: []string{"escape-link/escaped.txt"}}
	report := validateStageArtifacts(stage, root, RevisionInfo{})
	if report.Status != contracts.StatusFail {
		t.Fatalf("expected FAIL when a symlinked parent escapes the root, got %s (%s)", report.Status, report.ValidationMessage)
	}
	if len(report.MissingArtifacts) != 1 || report.MissingArtifacts[0] != "escape-link/escaped.txt" {
		t.Fatalf("expected escaping artifact rejected: %+v", report.MissingArtifacts)
	}
}

func TestValidateStageArtifactsStaleFails(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(root, "internal", "x.go")
	if err := os.WriteFile(artifact, []byte("package x"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(artifact, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	stage := Stage{Role: "implementer", Owner: "build", Mode: "native-host", ExpectedArtifacts: []string{"internal/x.go"}}
	revision := RevisionInfo{CommitTime: time.Now(), Resolved: true}
	report := validateStageArtifacts(stage, root, revision)
	if report.Status != contracts.StatusFail {
		t.Fatalf("expected FAIL for stale artifact, got %s (%s)", report.Status, report.ValidationMessage)
	}
	if len(report.StaleArtifacts) != 1 || report.StaleArtifacts[0] != "internal/x.go" {
		t.Fatalf("expected internal/x.go stale: %+v", report.StaleArtifacts)
	}
}

func TestValidateStageArtifactsNoExpectedArtifactsNotRun(t *testing.T) {
	root := t.TempDir()
	stage := Stage{Role: "reviewer", Owner: "review", Mode: "procedural"}
	report := validateStageArtifacts(stage, root, RevisionInfo{})
	if report.Status != contracts.StatusNotRun {
		t.Fatalf("expected NOT_RUN, got %s (%s)", report.Status, report.ValidationMessage)
	}
	if report.Validated {
		t.Fatal("expected stage not validated")
	}
	if report.ValidationMessage == "" {
		t.Fatal("expected a clear reason for NOT_RUN")
	}
}

func TestValidateStageArtifactsUnresolvableRevisionSkipsFreshness(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(root, "docs", "review.md")
	if err := os.WriteFile(artifact, []byte("review"), 0o644); err != nil {
		t.Fatal(err)
	}
	stage := Stage{Role: "reviewer", Owner: "review", Mode: "procedural", ExpectedArtifacts: []string{"docs/review.md"}}
	report := validateStageArtifacts(stage, root, RevisionInfo{})
	if report.Status != contracts.StatusPass {
		t.Fatalf("expected PASS when freshness cannot be checked, got %s (%s)", report.Status, report.ValidationMessage)
	}
}

func TestValidateArtifactsStagesMappedInOrder(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "review.md"), []byte("review"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{SchemaVersion: "1.0.0", Revision: "unresolvable-revision", Stories: []Story{{ID: "story", RequirementIDs: []string{"R1"}, Stages: stages()}}}
	report := ValidateArtifacts(manifest, root)
	if len(report.Stories) != 1 || len(report.Stories[0].Stages) != 6 {
		t.Fatalf("expected one story with six stages: %+v", report)
	}
	for index, stage := range report.Stories[0].Stages {
		if stage.Role != roleOrder[index] {
			t.Fatalf("stage %d role %q out of order", index, stage.Role)
		}
	}
}
