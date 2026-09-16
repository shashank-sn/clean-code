package adaptive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func playbookMinimumRoles(id string) []string {
	if playbook, ok := builtinPlaybooks[id]; ok {
		return append([]string{}, playbook.MinimumRoles...)
	}
	return []string{"clean-plan", "clean-build", "clean-verify"}
}

func playbookMinimumEvidence(id string) []string {
	if playbook, ok := builtinPlaybooks[id]; ok {
		return append([]string{}, playbook.MinimumEvidence...)
	}
	return []string{"verification-report"}
}

var builtinPlaybooks = map[string]Playbook{
	"bug-investigation": {
		SchemaVersion:    SchemaVersion,
		ID:               "bug-investigation",
		Title:            "Bug investigation",
		AppliesWhen:      []string{"reproducing a defect", "narrow causal fix"},
		DoesNotApplyWhen: []string{"greenfield features", "release qualification alone"},
		MinimumRoles:     []string{"clean-debug", "clean-build", "clean-test", "clean-verify"},
		MinimumEvidence:  []string{"repro-notes", "failing-test-or-NOT_AVAILABLE", "verification-report"},
		ArenaWarranted:   "no for routine fixes; yes only if multiple plausible architectural causes remain after diagnosis",
		AdversarialProbesWarranted: "yes when residual risk is medium/high after the fix",
		ParallelGuidance: "discovery and regression test authoring may run in parallel when scopes do not share mutable files",
		TerminalStates:   []string{TerminalPass, TerminalFail, TerminalNotRun, TerminalNotAvailable, TerminalBlocked},
	},
	"feature-delivery": {
		SchemaVersion:    SchemaVersion,
		ID:               "feature-delivery",
		Title:            "Feature delivery",
		AppliesWhen:      []string{"new user-visible behavior", "multi-file capability with acceptance criteria"},
		DoesNotApplyWhen: []string{"typo-only edits", "docs-only changes without behavior"},
		MinimumRoles:     []string{"clean-plan", "clean-build", "clean-test", "clean-simplify", "clean-prune", "clean-verify", "clean-review"},
		MinimumEvidence:  []string{"requirements", "verification-report", "review-input"},
		ArenaWarranted:   "yes when ambiguity is medium/high or two or more boundaries are affected",
		AdversarialProbesWarranted: "yes for medium/high risk before claiming success",
		ParallelGuidance: "test-writing and review prep may run in parallel; shared implementation stays serialized or worktree-isolated",
		TerminalStates:   []string{TerminalPass, TerminalFail, TerminalNotRun, TerminalNotAvailable, TerminalBlocked},
	},
	"design-decision": {
		SchemaVersion:    SchemaVersion,
		ID:               "design-decision",
		Title:            "Design decision",
		AppliesWhen:      []string{"consequential design choice with lasting boundary impact"},
		DoesNotApplyWhen: []string{"routine bugfixes", "settled local refactors"},
		MinimumRoles:     []string{"clean-design", "clean-arena", "clean-review"},
		MinimumEvidence:  []string{"arena-decision-record", "architecture-policy-or-NOT_CONFIGURED"},
		ArenaWarranted:   "always for this playbook",
		AdversarialProbesWarranted: "optional; use when the decision encodes acceptance-sensitive behavior",
		ParallelGuidance: "candidate reasoning may run in parallel; synthesis is single-owner and serialized",
		TerminalStates:   []string{TerminalPass, TerminalFail, TerminalNotRun, TerminalNotAvailable, TerminalBlocked},
	},
	"refactor": {
		SchemaVersion:    SchemaVersion,
		ID:               "refactor",
		Title:            "Refactor",
		AppliesWhen:      []string{"behavior-preserving structural cleanup"},
		DoesNotApplyWhen: []string{"features that change contracts", "unverified rewrites"},
		MinimumRoles:     []string{"clean-refactor", "clean-simplify", "clean-prune", "clean-test", "clean-verify", "clean-review"},
		MinimumEvidence:  []string{"behavior-preservation-notes", "verification-report", "review-input"},
		ArenaWarranted:   "yes when choosing among competing structural approaches with boundary impact",
		AdversarialProbesWarranted: "yes for medium/high risk refactors touching auth, dates, or idempotent APIs",
		ParallelGuidance: "characterization tests may run beside planning; mutating refactors stay single-owner",
		TerminalStates:   []string{TerminalPass, TerminalFail, TerminalNotRun, TerminalNotAvailable, TerminalBlocked},
	},
	"code-review": {
		SchemaVersion:    SchemaVersion,
		ID:               "code-review",
		Title:            "Code review",
		AppliesWhen:      []string{"independent review of a finished revision with evidence"},
		DoesNotApplyWhen: []string{"author self-approval", "review without a revision and evidence bundle"},
		MinimumRoles:     []string{"clean-review", "clean-reviewer"},
		MinimumEvidence:  []string{"diff", "requirements", "verification-report"},
		ArenaWarranted:   "no",
		AdversarialProbesWarranted: "review may request probes; it does not replace them",
		ParallelGuidance: "multiple reviewers may read in parallel; blocking dispositions reconcile before ship",
		TerminalStates:   []string{TerminalPass, TerminalFail, TerminalNotRun, TerminalNotAvailable, TerminalBlocked},
	},
	"release-readiness": {
		SchemaVersion:    SchemaVersion,
		ID:               "release-readiness",
		Title:            "Release readiness",
		AppliesWhen:      []string{"preparing a verified revision for PR or release evidence"},
		DoesNotApplyWhen: []string{"exploratory spikes", "unverified drafts"},
		MinimumRoles:     []string{"clean-prune", "clean-verify", "clean-review", "clean-ship", "clean-audit"},
		MinimumEvidence:  []string{"prune-report-or-NOT_RUN", "verification-report", "review-input", "spot-check-or-gap", "audit-receipt-or-NOT_RUN"},
		ArenaWarranted:   "no",
		AdversarialProbesWarranted: "yes for medium/high risk releases before success claims",
		ParallelGuidance: "not for shared release mutations; keep ship/audit serialized",
		TerminalStates:   []string{TerminalPass, TerminalFail, TerminalNotRun, TerminalNotAvailable, TerminalBlocked},
	},
}

// ListPlaybooks returns builtin playbooks sorted by id.
func ListPlaybooks() []Playbook {
	ids := make([]string, 0, len(builtinPlaybooks))
	for id := range builtinPlaybooks {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Playbook, 0, len(ids))
	for _, id := range ids {
		out = append(out, builtinPlaybooks[id])
	}
	return out
}

// GetPlaybook returns one builtin playbook.
func GetPlaybook(id string) (Playbook, error) {
	playbook, ok := builtinPlaybooks[id]
	if !ok {
		return Playbook{}, fmt.Errorf("unknown playbook %q", id)
	}
	return playbook, nil
}

// ValidatePlaybook checks a playbook document.
func ValidatePlaybook(playbook Playbook) error {
	if playbook.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", playbook.SchemaVersion)
	}
	if strings.TrimSpace(playbook.ID) == "" || strings.TrimSpace(playbook.Title) == "" {
		return fmt.Errorf("id and title are required")
	}
	if len(playbook.AppliesWhen) == 0 || len(playbook.DoesNotApplyWhen) == 0 {
		return fmt.Errorf("applies_when and does_not_apply_when are required")
	}
	if len(playbook.MinimumRoles) == 0 || len(playbook.MinimumEvidence) == 0 {
		return fmt.Errorf("minimum_roles and minimum_evidence are required")
	}
	if strings.TrimSpace(playbook.ArenaWarranted) == "" || strings.TrimSpace(playbook.AdversarialProbesWarranted) == "" {
		return fmt.Errorf("arena_warranted and adversarial_probes_warranted are required")
	}
	if strings.TrimSpace(playbook.ParallelGuidance) == "" {
		return fmt.Errorf("parallel_guidance is required")
	}
	if len(playbook.TerminalStates) != 5 {
		return fmt.Errorf("terminal_states must list PASS, FAIL, NOT_RUN, NOT_AVAILABLE, BLOCKED")
	}
	seen := map[string]struct{}{}
	for _, state := range playbook.TerminalStates {
		if _, ok := terminalStatuses[state]; !ok {
			return fmt.Errorf("unknown terminal state %q", state)
		}
		seen[state] = struct{}{}
	}
	for required := range terminalStatuses {
		if _, ok := seen[required]; !ok {
			return fmt.Errorf("terminal_states missing %q", required)
		}
	}
	return nil
}

// LoadPlaybooksDir loads optional JSON playbooks from a directory.
func LoadPlaybooksDir(dir string) ([]Playbook, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var playbooks []Playbook
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var playbook Playbook
		if err := json.Unmarshal(body, &playbook); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		if err := ValidatePlaybook(playbook); err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		playbooks = append(playbooks, playbook)
	}
	sort.Slice(playbooks, func(i, j int) bool { return playbooks[i].ID < playbooks[j].ID })
	return playbooks, nil
}
