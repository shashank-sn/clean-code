package adaptive

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRouteDeterministicForIdenticalInput(t *testing.T) {
	signals := TaskSignals{
		SchemaVersion: SchemaVersion,
		ChangeType:    "feature",
		Risk:          "high",
		Ambiguity:     "medium",
		AffectedBoundaries: []string{"auth", "api"},
		EvidenceNeeds:      []string{"acceptance"},
		HostCapabilities: HostCapabilityView{
			Subagents:        true,
			CommandExecution: true,
			FileEdits:        true,
		},
		ConcurrencyBudget: 3,
	}
	first, err := Route(signals)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Route(signals)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(second)
	if string(a) != string(b) {
		t.Fatalf("route decisions diverged\n%s\n%s", a, b)
	}
	if !first.ArenaWarranted || !first.AdversarialProbesWarranted {
		t.Fatalf("expected arena and probes for high-risk multi-boundary feature: %+v", first)
	}
	if first.Authorization != "advisory" {
		t.Fatalf("expected advisory authorization, got %q", first.Authorization)
	}
}

func TestRoutePrefersUnavailableOverInventedCapability(t *testing.T) {
	signals := TaskSignals{
		SchemaVersion: SchemaVersion,
		ChangeType:    "feature",
		Risk:          "medium",
		Ambiguity:     "low",
		HostCapabilities: HostCapabilityView{
			Subagents:        false,
			CommandExecution: false,
		},
	}
	decision, err := Route(signals)
	if err != nil {
		t.Fatal(err)
	}
	if decision.AllowedParallelism.Allowed {
		t.Fatal("expected parallelism disabled without subagents")
	}
	if !strings.Contains(decision.AllowedParallelism.UnavailableNote, "NOT_AVAILABLE") {
		t.Fatalf("expected NOT_AVAILABLE note, got %#v", decision.AllowedParallelism)
	}
	found := false
	for _, need := range decision.RequiredEvidence {
		if strings.Contains(need, "NOT_AVAILABLE:command_execution") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected command_execution NOT_AVAILABLE evidence marker: %#v", decision.RequiredEvidence)
	}
}

func TestRouteNeverMentionsModelBrand(t *testing.T) {
	for _, change := range []string{"bug", "feature", "design", "refactor", "review", "release"} {
		decision, err := Route(TaskSignals{
			SchemaVersion: SchemaVersion,
			ChangeType:    change,
			Risk:          "high",
			Ambiguity:     "high",
			HostCapabilities: HostCapabilityView{Subagents: true, CommandExecution: true},
		})
		if err != nil {
			t.Fatal(err)
		}
		blob, _ := json.Marshal(decision)
		lower := strings.ToLower(string(blob))
		for _, brand := range []string{"openai", "anthropic", "claude", "gpt-", "gemini", "cursor-model"} {
			if strings.Contains(lower, brand) {
				t.Fatalf("route decision mentioned model brand %q: %s", brand, blob)
			}
		}
	}
}

func TestArenaRejectsRoutineSingleCandidate(t *testing.T) {
	err := ValidateArenaDecision(ArenaDecisionRecord{
		SchemaVersion:    SchemaVersion,
		Revision:         "abc",
		Owner:            "design-owner",
		DecisionQuestion: "which approach?",
		Candidates: []ArenaCandidate{{
			ID: "only", Summary: "one", Evidence: []string{"note"}, Status: TerminalPass,
		}},
		PolicyChange: "none",
		Status:       TerminalPass,
	})
	if err == nil || !strings.Contains(err.Error(), "at least two candidates") {
		t.Fatalf("expected two-candidate rule, got %v", err)
	}
}

func TestArenaAcceptsPartialUnavailableCandidate(t *testing.T) {
	record := ArenaDecisionRecord{
		SchemaVersion:    SchemaVersion,
		Revision:         "rev1",
		Owner:            "owner",
		DecisionQuestion: "storage boundary?",
		Constraints:      []string{"no vendor lock-in in policy"},
		Candidates: []ArenaCandidate{
			{ID: "a", Summary: "repository ports", Evidence: []string{"sketch"}, Risks: []string{"more types"}, Tradeoffs: []string{"clarity vs speed"}, Status: TerminalPass},
			{ID: "b", Summary: "direct db in use case", Status: TerminalNotAvailable, Unavailable: "NOT_AVAILABLE: candidate worker lacked isolated context"},
		},
		SelectedOptionID:   "a",
		SelectionRationale: "keeps policy independent of persistence",
		RejectedRationale:  map[string]string{"b": "unavailable and would pull persistence inward"},
		Uncertainty:        []string{"migration cost"},
		UnresolvedAssumptions: []string{"graph producer coverage"},
		PolicyChange:       "none",
		Status:             TerminalPass,
	}
	if err := ValidateArenaDecision(record); err != nil {
		t.Fatal(err)
	}
	if ArenaTriggerMet(TaskSignals{ChangeType: "bug", Risk: "low", Ambiguity: "low", SchemaVersion: SchemaVersion}) {
		t.Fatal("routine bug must not meet arena trigger")
	}
	if !ArenaTriggerMet(TaskSignals{ChangeType: "design", Risk: "medium", Ambiguity: "low", SchemaVersion: SchemaVersion}) {
		t.Fatal("design change must meet arena trigger")
	}
}

func TestCalendarProbeRejectsImpossibleDate(t *testing.T) {
	probe := EvaluateCalendarProbe("2026-02-30")
	if probe.Status != TerminalFail {
		t.Fatalf("expected FAIL for 2026-02-30, got %+v", probe)
	}
	if EvaluateCalendarProbe("2026-02-28").Status != TerminalPass {
		t.Fatal("expected PASS for 2026-02-28")
	}
	if EvaluateCalendarProbe("2024-02-29").Status != TerminalPass {
		t.Fatal("expected PASS for leap day 2024-02-29")
	}
	if EvaluateCalendarProbe("2025-02-29").Status != TerminalFail {
		t.Fatal("expected FAIL for non-leap 2025-02-29")
	}
}

func TestCalendarFixtureBeatsNativeDateNormalization(t *testing.T) {
	// JavaScript Date normalizes impossible calendar strings; our probe must not.
	cmd := exec.Command("node", "-e", `const d=new Date("2026-02-30"); if (Number.isNaN(d.getTime())) process.exit(2); console.log(d.toISOString().slice(0,10))`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("node unavailable for Date normalization contrast: %v", err)
	}
	normalized := strings.TrimSpace(string(out))
	if normalized == "2026-02-30" {
		t.Fatalf("expected native Date to normalize away 2026-02-30, got %q", normalized)
	}
	if err := ValidateCalendarDate("2026-02-30"); err == nil {
		t.Fatal("strict calendar validation must reject 2026-02-30 before any success claim")
	}
	if err := ValidateCalendarDate(normalized); err != nil {
		// Normalized value may be a real date (e.g. 2026-03-02); that is fine.
		t.Logf("normalized date %q validation: %v", normalized, err)
	}
}

func TestDeriveDefaultProbesDistinguishesPrimaryAndPostHoc(t *testing.T) {
	plan, err := DeriveDefaultProbes("rev", "high", []string{"AW-3"}, true)
	if err != nil {
		t.Fatal(err)
	}
	var primary, posthoc int
	calendar := false
	for _, probe := range plan.Probes {
		switch probe.Kind {
		case ProbePrimaryAcceptance:
			primary++
		case ProbePostHocDiagnostic:
			posthoc++
		}
		if probe.ID == "calendar-validity" {
			calendar = true
		}
	}
	if primary == 0 || posthoc == 0 || !calendar {
		t.Fatalf("expected primary, post-hoc, and calendar probes: %+v", plan.Probes)
	}
	if err := ValidateAdversarialProbePlan(plan); err != nil {
		t.Fatal(err)
	}
}

func TestParallelRejectsUnverifiedClaimedArtifacts(t *testing.T) {
	err := ValidateParallelOrchestration(ParallelOrchestrationRecord{
		SchemaVersion:  SchemaVersion,
		Revision:       "rev",
		Coordinator:    "clean-orchestrate",
		WhyParallel:    "independent test tracks",
		MaxConcurrency: 2,
		Children: []ParallelChild{{
			ID: "c1", Role: "clean-test", OwnedScope: []string{"tests/"}, EvidenceRequest: "unit report",
			MutationBoundary: "tests-only", Status: TerminalPass,
			ClaimedArtifacts: []string{"tests/out.json"},
		}},
		HumanApprovalsHeld: []string{"release", "policy", "permissions", "external_actions"},
		Status:             TerminalFail,
	})
	if err == nil || !strings.Contains(err.Error(), "not verified") {
		t.Fatalf("expected unverified artifact failure, got %v", err)
	}
}

func TestParallelAcceptsVerifiedChildrenAndRecordsUnavailable(t *testing.T) {
	record := ParallelOrchestrationRecord{
		SchemaVersion:  SchemaVersion,
		Revision:       "rev",
		Coordinator:    "clean-orchestrate",
		WhyParallel:    "route allowed test and probe support in parallel",
		MaxConcurrency: 2,
		Children: []ParallelChild{
			{
				ID: "tests", Role: "clean-test", OwnedScope: []string{"internal/adaptive"},
				EvidenceRequest: "go test output", MutationBoundary: "no production mutation",
				Status: TerminalPass, ClaimedArtifacts: []string{"evidence/unit.txt"}, VerifiedArtifacts: []string{"evidence/unit.txt"},
			},
			{
				ID: "browser", Role: "clean-test", OwnedScope: []string{"ui"},
				EvidenceRequest: "qa procedure", MutationBoundary: "no production mutation",
				Status: TerminalNotAvailable, FailureNote: "NOT_AVAILABLE: browser automation disabled on host",
			},
		},
		SynthesisNotes:     []string{"ignored unavailable browser child"},
		HumanApprovalsHeld: []string{"release", "policy", "permissions", "external_actions"},
		Status:             TerminalPass,
	}
	if err := ValidateParallelOrchestration(record); err != nil {
		t.Fatal(err)
	}
}

func TestPlaybooksCoverRequiredShapes(t *testing.T) {
	required := []string{"bug-investigation", "feature-delivery", "design-decision", "refactor", "code-review", "release-readiness"}
	for _, id := range required {
		playbook, err := GetPlaybook(id)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidatePlaybook(playbook); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
	if len(ListPlaybooks()) != 6 {
		t.Fatalf("expected 6 playbooks, got %d", len(ListPlaybooks()))
	}
}

func TestFixtureRouterChoices(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "adaptive")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "route-") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		signals, err := LoadTaskSignals(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		if _, err := Route(signals); err != nil {
			t.Fatalf("route %s: %v", entry.Name(), err)
		}
	}
}

func TestLoadArenaAndProbeFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "fixtures", "adaptive")
	if _, err := LoadArenaDecision(filepath.Join(root, "arena-decision.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAdversarialProbePlan(filepath.Join(root, "adversarial-probes.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadParallelOrchestration(filepath.Join(root, "parallel-partial.json")); err != nil {
		t.Fatal(err)
	}
}
