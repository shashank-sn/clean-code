package adaptive

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxArenaBytes int64 = 2 << 20

// LoadArenaDecision reads a competing-design decision record.
func LoadArenaDecision(path string) (ArenaDecisionRecord, error) {
	record, err := loadJSONFile[ArenaDecisionRecord](path, maxArenaBytes, "arena decision")
	if err != nil {
		return ArenaDecisionRecord{}, err
	}
	if err := ValidateArenaDecision(record); err != nil {
		return ArenaDecisionRecord{}, err
	}
	return record, nil
}

// ValidateArenaDecision enforces the bounded arena contract.
func ValidateArenaDecision(record ArenaDecisionRecord) error {
	if record.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", record.SchemaVersion)
	}
	if strings.TrimSpace(record.Revision) == "" {
		return errors.New("revision is required")
	}
	if strings.TrimSpace(record.Owner) == "" {
		return errors.New("owner is required")
	}
	if strings.TrimSpace(record.DecisionQuestion) == "" {
		return errors.New("decision_question is required")
	}
	if len(record.Candidates) < 2 {
		return errors.New("arena requires at least two candidates")
	}
	ids := map[string]struct{}{}
	runnable := 0
	for i, candidate := range record.Candidates {
		if strings.TrimSpace(candidate.ID) == "" {
			return fmt.Errorf("candidate %d id is required", i)
		}
		if _, exists := ids[candidate.ID]; exists {
			return fmt.Errorf("duplicate candidate id %q", candidate.ID)
		}
		ids[candidate.ID] = struct{}{}
		if strings.TrimSpace(candidate.Summary) == "" {
			return fmt.Errorf("candidate %q summary is required", candidate.ID)
		}
		if _, ok := terminalStatuses[candidate.Status]; !ok {
			return fmt.Errorf("candidate %q has unknown status %q", candidate.ID, candidate.Status)
		}
		if candidate.Status == TerminalNotAvailable || candidate.Status == TerminalNotRun || candidate.Status == TerminalBlocked {
			if strings.TrimSpace(candidate.Unavailable) == "" {
				return fmt.Errorf("candidate %q must record unavailable_reason when status is %s", candidate.ID, candidate.Status)
			}
			continue
		}
		runnable++
		if len(candidate.Evidence) == 0 {
			return fmt.Errorf("candidate %q requires evidence when runnable", candidate.ID)
		}
	}
	if runnable == 0 {
		return errors.New("arena requires at least one runnable candidate result or explicit partial record with selected_option_id empty and status NOT_AVAILABLE")
	}
	if record.PolicyChange != "none" {
		return errors.New("policy_change must be \"none\"; arena records must not alter policy automatically")
	}
	if _, ok := terminalStatuses[record.Status]; !ok {
		return fmt.Errorf("unknown status %q", record.Status)
	}
	if record.Status == TerminalPass || record.Status == TerminalFail {
		if strings.TrimSpace(record.SelectedOptionID) == "" {
			return errors.New("selected_option_id is required for PASS/FAIL arena outcomes")
		}
		if _, ok := ids[record.SelectedOptionID]; !ok {
			return fmt.Errorf("selected_option_id %q is not a candidate", record.SelectedOptionID)
		}
		if strings.TrimSpace(record.SelectionRationale) == "" {
			return errors.New("selection_rationale is required")
		}
		for id := range ids {
			if id == record.SelectedOptionID {
				continue
			}
			if strings.TrimSpace(record.RejectedRationale[id]) == "" {
				return fmt.Errorf("rejected_rationale missing for candidate %q", id)
			}
		}
	}
	return nil
}

// ArenaTriggerMet reports whether the explicit arena threshold is met.
// Routine fixes (low risk, low ambiguity, non-design) do not trigger.
func ArenaTriggerMet(signals TaskSignals) bool {
	return arenaWarranted(signals)
}

func loadJSONFile[T any](path string, maxBytes int64, label string) (T, error) {
	var zero T
	info, err := os.Lstat(path)
	if err != nil {
		return zero, fmt.Errorf("inspect %s: %w", label, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return zero, fmt.Errorf("inspect %s: input must be a regular file", label)
	}
	if info.Size() > maxBytes {
		return zero, fmt.Errorf("inspect %s: input exceeds %d bytes", label, maxBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return zero, fmt.Errorf("open %s: %w", label, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxBytes+1))
	decoder.DisallowUnknownFields()
	var value T
	if err := decoder.Decode(&value); err != nil {
		return zero, fmt.Errorf("parse %s: %w", label, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return zero, fmt.Errorf("parse %s: unexpected trailing JSON value", label)
		}
		return zero, fmt.Errorf("parse %s: %w", label, err)
	}
	return value, nil
}
