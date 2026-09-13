package adaptive

import (
	"errors"
	"fmt"
	"strings"
)

const maxParallelBytes int64 = 2 << 20

// LoadParallelOrchestration reads a parallel orchestration record.
func LoadParallelOrchestration(path string) (ParallelOrchestrationRecord, error) {
	record, err := loadJSONFile[ParallelOrchestrationRecord](path, maxParallelBytes, "parallel orchestration")
	if err != nil {
		return ParallelOrchestrationRecord{}, err
	}
	if err := ValidateParallelOrchestration(record); err != nil {
		return ParallelOrchestrationRecord{}, err
	}
	return record, nil
}

// ValidateParallelOrchestration enforces ownership and artifact-check rules.
func ValidateParallelOrchestration(record ParallelOrchestrationRecord) error {
	if record.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", record.SchemaVersion)
	}
	if strings.TrimSpace(record.Revision) == "" {
		return errors.New("revision is required")
	}
	if strings.TrimSpace(record.Coordinator) == "" {
		return errors.New("coordinator is required")
	}
	if strings.TrimSpace(record.WhyParallel) == "" {
		return errors.New("why_parallel is required")
	}
	if record.MaxConcurrency < 1 {
		return errors.New("max_concurrency must be at least 1")
	}
	if len(record.Children) == 0 {
		return errors.New("at least one child is required")
	}
	if _, ok := terminalStatuses[record.Status]; !ok {
		return fmt.Errorf("unknown status %q", record.Status)
	}
	ids := map[string]struct{}{}
	for i, child := range record.Children {
		if strings.TrimSpace(child.ID) == "" {
			return fmt.Errorf("child %d id is required", i)
		}
		if _, exists := ids[child.ID]; exists {
			return fmt.Errorf("duplicate child id %q", child.ID)
		}
		ids[child.ID] = struct{}{}
		if strings.TrimSpace(child.Role) == "" {
			return fmt.Errorf("child %q role is required", child.ID)
		}
		if len(child.OwnedScope) == 0 {
			return fmt.Errorf("child %q owned_scope is required", child.ID)
		}
		if strings.TrimSpace(child.EvidenceRequest) == "" {
			return fmt.Errorf("child %q evidence_request is required", child.ID)
		}
		if strings.TrimSpace(child.MutationBoundary) == "" {
			return fmt.Errorf("child %q mutation_boundary is required", child.ID)
		}
		if _, ok := terminalStatuses[child.Status]; !ok {
			return fmt.Errorf("child %q has unknown status %q", child.ID, child.Status)
		}
		if child.Status == TerminalNotAvailable || child.Status == TerminalFail || child.Status == TerminalBlocked {
			if strings.TrimSpace(child.FailureNote) == "" {
				return fmt.Errorf("child %q must record failure_note when status is %s", child.ID, child.Status)
			}
		}
		if len(child.ClaimedArtifacts) > 0 {
			verified := map[string]struct{}{}
			for _, path := range child.VerifiedArtifacts {
				verified[path] = struct{}{}
			}
			for _, claimed := range child.ClaimedArtifacts {
				if _, ok := verified[claimed]; !ok {
					return fmt.Errorf("child %q claimed artifact %q was not verified; synthesis must not treat reports as proof", child.ID, claimed)
				}
			}
		}
	}
	held := map[string]struct{}{}
	for _, item := range record.HumanApprovalsHeld {
		held[item] = struct{}{}
	}
	for _, required := range []string{"release", "policy", "permissions", "external_actions"} {
		if _, ok := held[required]; !ok {
			return fmt.Errorf("human_approvals_held must retain %q", required)
		}
	}
	return nil
}
