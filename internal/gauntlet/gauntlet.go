package gauntlet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/shashank-sn/clean-code/internal/contracts"
	"github.com/shashank-sn/clean-code/internal/telemetry"
)

const maxManifestBytes int64 = 5 << 20

var roleOrder = []string{"specifier", "implementer", "cleaner", "hardener", "qa", "reviewer"}
var roles = map[string]bool{"specifier": true, "implementer": true, "cleaner": true, "hardener": true, "qa": true, "reviewer": true}
var modes = map[string]bool{"mechanical": true, "native-host": true, "procedural": true}
var storyID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

type Manifest struct {
	SchemaVersion string  `json:"schema_version"`
	Revision      string  `json:"revision"`
	Stories       []Story `json:"stories"`
}

type Story struct {
	ID             string            `json:"id"`
	RequirementIDs []string          `json:"requirement_ids"`
	Budget         telemetry.Budget  `json:"budget,omitempty"`
	Stages         []Stage           `json:"stages"`
	Events         []telemetry.Event `json:"events,omitempty"`
}

type Stage struct {
	Role                 string   `json:"role"`
	Owner                string   `json:"owner"`
	Mode                 string   `json:"mode"`
	AllowedFiles         []string `json:"allowed_files"`
	PublicContracts      []string `json:"public_contracts,omitempty"`
	SourceContext        []string `json:"source_context,omitempty"`
	ExpectedArtifacts    []string `json:"expected_artifacts,omitempty"`
	EvidenceDependencies []string `json:"evidence_dependencies,omitempty"`
	StopCondition        string   `json:"stop_condition"`
}

type WorkPacket struct {
	SchemaVersion        string   `json:"schema_version"`
	StoryID              string   `json:"story_id"`
	Revision             string   `json:"revision"`
	Role                 string   `json:"role"`
	Owner                string   `json:"owner"`
	Mode                 string   `json:"mode"`
	RequirementIDs       []string `json:"requirement_ids"`
	AllowedFiles         []string `json:"allowed_files"`
	PublicContracts      []string `json:"public_contracts,omitempty"`
	SourceContext        []string `json:"source_context,omitempty"`
	ExpectedArtifacts    []string `json:"expected_artifacts,omitempty"`
	EvidenceDependencies []string `json:"evidence_dependencies,omitempty"`
	StopCondition        string   `json:"stop_condition"`
}

type StoryReport struct {
	StoryID        string            `json:"story_id"`
	Decision       string            `json:"decision"`
	StopReason     string            `json:"stop_reason,omitempty"`
	Stages         []StageReport     `json:"stages"`
	Telemetry      telemetry.Summary `json:"telemetry"`
	RefactorPacket *WorkPacket       `json:"refactor_packet,omitempty"`
}

type StageReport struct {
	Role              string           `json:"role"`
	Owner             string           `json:"owner"`
	Mode              string           `json:"mode"`
	Enforced          bool             `json:"enforced_independence"`
	Status            contracts.Status `json:"status"`
	Validated         bool             `json:"validated"`
	ValidationMessage string           `json:"validation_message,omitempty"`
	MissingArtifacts  []string         `json:"missing_artifacts,omitempty"`
	StaleArtifacts    []string         `json:"stale_artifacts,omitempty"`
}

type Report struct {
	SchemaVersion string        `json:"schema_version"`
	Revision      string        `json:"revision"`
	Stories       []StoryReport `json:"stories"`
}

func Load(path string) (Manifest, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("inspect gauntlet manifest: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Manifest{}, errors.New("gauntlet manifest must be a regular file")
	}
	if info.Size() > maxManifestBytes {
		return Manifest{}, fmt.Errorf("gauntlet manifest exceeds %d bytes", maxManifestBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxManifestBytes+1))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse gauntlet manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Manifest{}, errors.New("parse gauntlet manifest: unexpected trailing JSON value")
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (manifest Manifest) Validate() error {
	if manifest.SchemaVersion != "1.0.0" || strings.TrimSpace(manifest.Revision) == "" || len(manifest.Stories) == 0 {
		return errors.New("gauntlet manifest requires schema_version, revision, and stories")
	}
	seenStories := map[string]bool{}
	for _, story := range manifest.Stories {
		if !storyID.MatchString(story.ID) || seenStories[story.ID] || len(story.RequirementIDs) == 0 || len(story.Stages) == 0 {
			return fmt.Errorf("invalid or duplicate story %q", story.ID)
		}
		seenStories[story.ID] = true
		if len(story.Stages) != len(roleOrder) {
			return fmt.Errorf("story %q requires the full six-role pipeline", story.ID)
		}
		owners := map[string]bool{}
		for index, stage := range story.Stages {
			if !roles[stage.Role] || !modes[stage.Mode] || strings.TrimSpace(stage.Owner) == "" || owners[stage.Owner] || strings.TrimSpace(stage.StopCondition) == "" {
				return fmt.Errorf("invalid stage %q in story %q", stage.Role, story.ID)
			}
			if stage.Role != roleOrder[index] {
				return fmt.Errorf("stage %d in story %q must be %q", index+1, story.ID, roleOrder[index])
			}
			if len(stage.AllowedFiles) == 0 {
				return fmt.Errorf("stage %q in story %q requires allowed_files", stage.Role, story.ID)
			}
			owners[stage.Owner] = true
		}
	}
	return nil
}

func Packets(manifest Manifest) []WorkPacket {
	var packets []WorkPacket
	for _, story := range manifest.Stories {
		for _, stage := range story.Stages {
			packets = append(packets, WorkPacket{SchemaVersion: "1.0.0", StoryID: story.ID, Revision: manifest.Revision, Role: stage.Role, Owner: stage.Owner, Mode: stage.Mode, RequirementIDs: append([]string{}, story.RequirementIDs...), AllowedFiles: append([]string{}, stage.AllowedFiles...), PublicContracts: append([]string{}, stage.PublicContracts...), SourceContext: append([]string{}, stage.SourceContext...), ExpectedArtifacts: append([]string{}, stage.ExpectedArtifacts...), EvidenceDependencies: append([]string{}, stage.EvidenceDependencies...), StopCondition: stage.StopCondition})
		}
	}
	return packets
}

func Evaluate(manifest Manifest) Report {
	report := Report{SchemaVersion: "1.0.0", Revision: manifest.Revision, Stories: []StoryReport{}}
	for _, story := range manifest.Stories {
		summary := telemetry.Evaluate(story.Events, story.Budget)
		storyReport := StoryReport{StoryID: story.ID, Decision: summary.Decision, StopReason: summary.StopReason, Telemetry: summary, Stages: []StageReport{}}
		for _, stage := range story.Stages {
			storyReport.Stages = append(storyReport.Stages, StageReport{Role: stage.Role, Owner: stage.Owner, Mode: stage.Mode, Enforced: stage.Mode != "procedural", Status: contracts.StatusNotRun})
		}
		if summary.Decision == telemetry.DecisionReorganizeArchitecture {
			storyReport.RefactorPacket = &WorkPacket{SchemaVersion: "1.0.0", StoryID: story.ID, Revision: manifest.Revision, Role: "cleaner", Owner: "architecture-review", Mode: "procedural", RequirementIDs: append([]string{}, story.RequirementIDs...), AllowedFiles: append([]string{}, summary.RepeatedFiles...), StopCondition: "human approves refactor or records explicit deferred risk"}
		}
		report.Stories = append(report.Stories, storyReport)
	}
	return report
}

// RevisionInfo carries the resolved commit time for a manifest revision.
type RevisionInfo struct {
	CommitTime time.Time
	Resolved   bool
}

// ResolveRevisionInfo resolves revision's commit time in the repository rooted
// at repoRoot using `git -C <root> show -s --format=%cI <revision>`. When the
// revision cannot be resolved (not a git repo, unknown revision, or no commit
// time), Resolved is false.
func ResolveRevisionInfo(repoRoot, revision string) RevisionInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "git", "-C", repoRoot, "show", "-s", "--format=%cI", revision)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &bytes.Buffer{}
	if err := command.Run(); err != nil {
		return RevisionInfo{}
	}
	commitTime, err := time.Parse(time.RFC3339, strings.TrimSpace(output.String()))
	if err != nil {
		return RevisionInfo{}
	}
	return RevisionInfo{CommitTime: commitTime, Resolved: true}
}

// ValidateArtifacts validates every stage's expected_artifacts against the
// repository filesystem rooted at repoRoot. It reuses the Evaluate report
// (telemetry decision, stop reason, refactor packet) and replaces each stage's
// status with the artifact-validation outcome. The portable core never executes
// agent stages, so a stage that declares no expected artifacts stays NOT_RUN.
func ValidateArtifacts(manifest Manifest, repoRoot string) Report {
	report := Evaluate(manifest)
	revision := ResolveRevisionInfo(repoRoot, manifest.Revision)
	for storyIndex := range report.Stories {
		for stageIndex := range report.Stories[storyIndex].Stages {
			report.Stories[storyIndex].Stages[stageIndex] = validateStageArtifacts(manifest.Stories[storyIndex].Stages[stageIndex], repoRoot, revision)
		}
	}
	return report
}

// validateStageArtifacts validates one stage's expected artifacts. PASS requires
// every expected artifact's parent directory to resolve (via symlinks) inside the
// repository root and the final component to be a regular file (never a symlink);
// when the revision resolved, each file must also be fresh (mtime >= commit time).
// Any missing, non-regular, symlink, symlink-escaped, or stale artifact fails the
// stage. A stage with no expected artifacts is NOT_RUN.
func validateStageArtifacts(stage Stage, repoRoot string, revision RevisionInfo) StageReport {
	report := StageReport{Role: stage.Role, Owner: stage.Owner, Mode: stage.Mode, Enforced: stage.Mode != "procedural"}
	if len(stage.ExpectedArtifacts) == 0 {
		report.Status = contracts.StatusNotRun
		report.ValidationMessage = "stage declares no expected artifacts for portable validation"
		return report
	}
	report.Validated = true
	realRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		realRoot, err = filepath.Abs(repoRoot)
		if err != nil {
			realRoot = filepath.Clean(repoRoot)
		}
	}
	for _, artifact := range stage.ExpectedArtifacts {
		path := filepath.Join(repoRoot, artifact)
		if !withinRoot(repoRoot, path) {
			report.MissingArtifacts = append(report.MissingArtifacts, artifact)
			continue
		}
		parent, err := filepath.EvalSymlinks(filepath.Dir(path))
		if err != nil {
			report.MissingArtifacts = append(report.MissingArtifacts, artifact)
			continue
		}
		relative, err := filepath.Rel(realRoot, parent)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			report.MissingArtifacts = append(report.MissingArtifacts, artifact)
			continue
		}
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			report.MissingArtifacts = append(report.MissingArtifacts, artifact)
			continue
		}
		if revision.Resolved && info.ModTime().Before(revision.CommitTime) {
			report.StaleArtifacts = append(report.StaleArtifacts, artifact)
		}
	}
	if len(report.MissingArtifacts) > 0 || len(report.StaleArtifacts) > 0 {
		report.Status = contracts.StatusFail
		report.ValidationMessage = artifactFailureMessage(report)
		return report
	}
	report.Status = contracts.StatusPass
	if revision.Resolved {
		report.ValidationMessage = "all expected artifacts present and fresh"
	} else {
		report.ValidationMessage = "all expected artifacts present and regular (freshness not checked: revision unresolvable)"
	}
	return report
}

func artifactFailureMessage(report StageReport) string {
	var problems []string
	if len(report.MissingArtifacts) > 0 {
		problems = append(problems, fmt.Sprintf("%d missing or non-regular", len(report.MissingArtifacts)))
	}
	if len(report.StaleArtifacts) > 0 {
		problems = append(problems, fmt.Sprintf("%d stale", len(report.StaleArtifacts)))
	}
	return "artifact validation failed: " + strings.Join(problems, ", ")
}

// withinRoot reports whether path stays inside root after cleaning.
func withinRoot(root, path string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func WritePackets(directory string, packets []WorkPacket) error {
	if strings.TrimSpace(directory) == "" {
		return errors.New("gauntlet output directory is required")
	}
	for index, packet := range packets {
		path := filepath.Join(directory, fmt.Sprintf("%02d-%s-%s.json", index+1, packet.StoryID, packet.Role))
		if filepath.Dir(path) != filepath.Clean(directory) {
			return fmt.Errorf("invalid gauntlet packet path for story %q", packet.StoryID)
		}
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("gauntlet packet already exists: %s", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	for index, packet := range packets {
		path := filepath.Join(directory, fmt.Sprintf("%02d-%s-%s.json", index+1, packet.StoryID, packet.Role))
		body, err := json.MarshalIndent(packet, "", "  ")
		if err != nil {
			return err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(append(body, '\n'))
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func SortPackets(packets []WorkPacket) {
	sort.Slice(packets, func(i, j int) bool { return packets[i].StoryID+packets[i].Role < packets[j].StoryID+packets[j].Role })
}
