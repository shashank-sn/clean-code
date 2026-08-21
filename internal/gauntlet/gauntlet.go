package gauntlet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"clean-code/internal/contracts"
	"clean-code/internal/telemetry"
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
	Role     string           `json:"role"`
	Owner    string           `json:"owner"`
	Mode     string           `json:"mode"`
	Enforced bool             `json:"enforced_independence"`
	Status   contracts.Status `json:"status"`
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
