package adaptive

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const maxSignalBytes int64 = 1 << 20

var validChangeTypes = map[string]struct{}{
	"bug": {}, "feature": {}, "design": {}, "refactor": {}, "review": {}, "release": {},
}

var validRisk = map[string]struct{}{
	"low": {}, "medium": {}, "high": {},
}

var validAmbiguity = map[string]struct{}{
	"low": {}, "medium": {}, "high": {},
}

// LoadTaskSignals reads and validates router input JSON.
func LoadTaskSignals(path string) (TaskSignals, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return TaskSignals{}, fmt.Errorf("inspect task signals: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return TaskSignals{}, errors.New("inspect task signals: input must be a regular file")
	}
	if info.Size() > maxSignalBytes {
		return TaskSignals{}, fmt.Errorf("inspect task signals: input exceeds %d bytes", maxSignalBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return TaskSignals{}, fmt.Errorf("open task signals: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxSignalBytes+1))
	decoder.DisallowUnknownFields()
	var signals TaskSignals
	if err := decoder.Decode(&signals); err != nil {
		return TaskSignals{}, fmt.Errorf("parse task signals: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return TaskSignals{}, errors.New("parse task signals: unexpected trailing JSON value")
		}
		return TaskSignals{}, fmt.Errorf("parse task signals: %w", err)
	}
	if err := ValidateTaskSignals(signals); err != nil {
		return TaskSignals{}, err
	}
	return signals, nil
}

// ValidateTaskSignals enforces model-neutral router input rules.
func ValidateTaskSignals(signals TaskSignals) error {
	if signals.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", signals.SchemaVersion)
	}
	if _, ok := validChangeTypes[signals.ChangeType]; !ok {
		return fmt.Errorf("unknown change_type %q", signals.ChangeType)
	}
	if _, ok := validRisk[signals.Risk]; !ok {
		return fmt.Errorf("unknown risk %q", signals.Risk)
	}
	if _, ok := validAmbiguity[signals.Ambiguity]; !ok {
		return fmt.Errorf("unknown ambiguity %q", signals.Ambiguity)
	}
	if signals.ConcurrencyBudget < 0 {
		return errors.New("concurrency_budget cannot be negative")
	}
	if signals.AuthorizationMode != "" && signals.AuthorizationMode != "advisory" && signals.AuthorizationMode != "lifecycle_authorized" {
		return fmt.Errorf("unknown authorization_mode %q", signals.AuthorizationMode)
	}
	for _, need := range signals.EvidenceNeeds {
		if strings.TrimSpace(need) == "" {
			return errors.New("evidence_needs entries must be non-empty")
		}
	}
	for _, boundary := range signals.AffectedBoundaries {
		if strings.TrimSpace(boundary) == "" {
			return errors.New("affected_boundaries entries must be non-empty")
		}
	}
	return nil
}

// Route selects a bounded workflow shape. It never selects a model by brand
// and never writes host or global configuration.
func Route(signals TaskSignals) (RouteDecision, error) {
	if err := ValidateTaskSignals(signals); err != nil {
		return RouteDecision{}, err
	}

	playbookID := playbookForChange(signals.ChangeType)
	roles := append([]string{}, playbookMinimumRoles(playbookID)...)
	evidence := uniqueStrings(append(playbookMinimumEvidence(playbookID), signals.EvidenceNeeds...))
	reasons := []string{
		fmt.Sprintf("change_type=%s maps to playbook %s", signals.ChangeType, playbookID),
		fmt.Sprintf("risk=%s ambiguity=%s", signals.Risk, signals.Ambiguity),
	}

	arena := arenaWarranted(signals)
	probes := probesWarranted(signals)
	if arena {
		roles = uniqueStrings(append(roles, "clean-arena", "clean-design"))
		reasons = append(reasons, "arena warranted: consequential design choice with medium/high ambiguity or boundary impact")
	}
	if probes {
		roles = uniqueStrings(append(roles, "clean-probe", "clean-verify"))
		reasons = append(reasons, "adversarial probes warranted: medium/high risk after implementation")
	}
	if len(signals.AffectedBoundaries) > 0 {
		roles = uniqueStrings(append(roles, "clean-design"))
		reasons = append(reasons, "affected boundaries require design/policy attention")
	}
	if signals.ChangeType == "release" {
		roles = uniqueStrings(append(roles, "clean-verify", "clean-review", "clean-ship", "clean-audit"))
		evidence = uniqueStrings(append(evidence, "verification-report", "review-input", "audit-receipt"))
	}

	auth := "advisory"
	if signals.AuthorizationMode == "lifecycle_authorized" {
		auth = "lifecycle_authorized"
		reasons = append(reasons, "caller declared lifecycle_authorized; execution still requires matching existing contracts")
	} else {
		reasons = append(reasons, "route remains advisory unless an existing lifecycle contract authorizes execution")
	}

	parallel := parallelAllowance(signals, roles)
	if !signals.HostCapabilities.Subagents && parallel.Allowed {
		parallel.Allowed = false
		parallel.MaxConcurrency = 1
		parallel.UnavailableNote = "NOT_AVAILABLE: host lacks subagents; use procedural separate sessions"
		reasons = append(reasons, "host lacks subagents; parallelism degraded to procedural handoffs")
	}
	if !signals.HostCapabilities.CommandExecution {
		evidence = uniqueStrings(append(evidence, "NOT_AVAILABLE:command_execution"))
		reasons = append(reasons, "host lacks command_execution; prefer NOT_AVAILABLE over invented checks")
	}

	sort.Strings(roles)
	sort.Strings(evidence)

	return RouteDecision{
		SchemaVersion:              SchemaVersion,
		PlaybookID:                 playbookID,
		SelectedRoles:              roles,
		RequiredEvidence:           evidence,
		ArenaWarranted:             arena,
		AdversarialProbesWarranted: probes,
		AllowedParallelism:         parallel,
		Reasons:                    reasons,
		Fallback: RouteFallback{
			PlaybookID: "bug-investigation",
			Roles:      []string{"clean-debug", "clean-verify"},
			Reason:     "safe fallback when preferred playbook cannot run",
		},
		Authorization:  auth,
		TerminalStates: []string{TerminalPass, TerminalFail, TerminalNotRun, TerminalNotAvailable, TerminalBlocked},
	}, nil
}

func playbookForChange(changeType string) string {
	switch changeType {
	case "bug":
		return "bug-investigation"
	case "feature":
		return "feature-delivery"
	case "design":
		return "design-decision"
	case "refactor":
		return "refactor"
	case "review":
		return "code-review"
	case "release":
		return "release-readiness"
	default:
		return "bug-investigation"
	}
}

func arenaWarranted(signals TaskSignals) bool {
	if signals.ChangeType != "design" && signals.ChangeType != "feature" && signals.ChangeType != "refactor" {
		return false
	}
	if signals.ChangeType == "design" {
		return true
	}
	if signals.Ambiguity == "high" {
		return true
	}
	if signals.Risk == "high" && len(signals.AffectedBoundaries) > 0 {
		return true
	}
	if signals.Ambiguity == "medium" && len(signals.AffectedBoundaries) >= 2 {
		return true
	}
	return false
}

func probesWarranted(signals TaskSignals) bool {
	if signals.Risk == "medium" || signals.Risk == "high" {
		return signals.ChangeType == "feature" || signals.ChangeType == "bug" || signals.ChangeType == "refactor" || signals.ChangeType == "release"
	}
	return false
}

func parallelAllowance(signals TaskSignals, roles []string) ParallelAllowance {
	budget := signals.ConcurrencyBudget
	if budget == 0 {
		budget = 2
	}
	if budget > 4 {
		budget = 4
	}
	eligible := intersect(roles, []string{"clean-discover", "clean-test", "clean-test-writer", "clean-review", "clean-probe"})
	allowed := len(eligible) >= 2 && (signals.Risk != "high" || signals.HostCapabilities.Subagents)
	if signals.ChangeType == "release" {
		allowed = false
		return ParallelAllowance{
			Allowed:        false,
			MaxConcurrency: 1,
			EligibleRoles:  nil,
			MutationPolicy: "serialized",
			UnavailableNote: "release readiness keeps shared-state mutations serialized",
		}
	}
	mutation := "serialized_shared_state"
	if signals.HostCapabilities.Subagents {
		mutation = "isolated_worktrees_or_serialized_shared_state"
	}
	if !allowed {
		return ParallelAllowance{
			Allowed:        false,
			MaxConcurrency: 1,
			EligibleRoles:  eligible,
			MutationPolicy: mutation,
		}
	}
	return ParallelAllowance{
		Allowed:        true,
		MaxConcurrency: budget,
		EligibleRoles:  eligible,
		MutationPolicy: mutation,
	}
}

func intersect(left, right []string) []string {
	set := map[string]struct{}{}
	for _, value := range left {
		set[value] = struct{}{}
	}
	out := make([]string, 0)
	for _, value := range right {
		if _, ok := set[value]; ok {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
