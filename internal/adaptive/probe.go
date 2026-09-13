package adaptive

import (
	"errors"
	"fmt"
	"strings"
)

const maxProbeBytes int64 = 2 << 20

// LoadAdversarialProbePlan reads a revision-bound probe plan.
func LoadAdversarialProbePlan(path string) (AdversarialProbePlan, error) {
	plan, err := loadJSONFile[AdversarialProbePlan](path, maxProbeBytes, "adversarial probe plan")
	if err != nil {
		return AdversarialProbePlan{}, err
	}
	if err := ValidateAdversarialProbePlan(plan); err != nil {
		return AdversarialProbePlan{}, err
	}
	return plan, nil
}

// ValidateAdversarialProbePlan enforces probe contract rules.
func ValidateAdversarialProbePlan(plan AdversarialProbePlan) error {
	if plan.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", plan.SchemaVersion)
	}
	if strings.TrimSpace(plan.Revision) == "" {
		return errors.New("revision is required")
	}
	if _, ok := validRisk[plan.Risk]; !ok {
		return fmt.Errorf("unknown risk %q", plan.Risk)
	}
	if plan.Risk == "low" {
		return errors.New("adversarial probe plans are for medium/high risk work")
	}
	if len(plan.Probes) == 0 {
		return errors.New("at least one probe is required")
	}
	if plan.RepairLoop != "bounded_repair_reverify" {
		return errors.New("repair_loop must be bounded_repair_reverify; probes must not mutate policy automatically")
	}
	if _, ok := terminalStatuses[plan.Status]; !ok {
		return fmt.Errorf("unknown status %q", plan.Status)
	}
	ids := map[string]struct{}{}
	primary := 0
	for i, probe := range plan.Probes {
		if strings.TrimSpace(probe.ID) == "" {
			return fmt.Errorf("probe %d id is required", i)
		}
		if _, exists := ids[probe.ID]; exists {
			return fmt.Errorf("duplicate probe id %q", probe.ID)
		}
		ids[probe.ID] = struct{}{}
		if probe.Kind != ProbePrimaryAcceptance && probe.Kind != ProbePostHocDiagnostic {
			return fmt.Errorf("probe %q has unknown kind %q", probe.ID, probe.Kind)
		}
		if probe.Kind == ProbePrimaryAcceptance {
			primary++
		}
		if strings.TrimSpace(probe.DerivedFrom) == "" || strings.TrimSpace(probe.Description) == "" {
			return fmt.Errorf("probe %q requires derived_from and description", probe.ID)
		}
		if _, ok := terminalStatuses[probe.Status]; !ok {
			return fmt.Errorf("probe %q has unknown status %q", probe.ID, probe.Status)
		}
		if !probe.Executable && !probe.HumanNeeded && probe.Status != TerminalNotAvailable && probe.Status != TerminalNotRun {
			return fmt.Errorf("probe %q must be executable, require human judgment, or be NOT_AVAILABLE/NOT_RUN", probe.ID)
		}
		if probe.Status == TerminalPass || probe.Status == TerminalFail {
			if strings.TrimSpace(probe.Evidence) == "" {
				return fmt.Errorf("probe %q requires evidence for PASS/FAIL", probe.ID)
			}
		}
	}
	if primary == 0 {
		return errors.New("plan must include at least one primary_acceptance probe distinct from post_hoc_diagnostic")
	}
	return nil
}

// DeriveDefaultProbes builds a minimal probe set from requirement-oriented hints.
// Callers must still execute checks; this only plans them.
func DeriveDefaultProbes(revision, risk string, requirementIDs []string, includeCalendar bool) (AdversarialProbePlan, error) {
	if _, ok := validRisk[risk]; !ok {
		return AdversarialProbePlan{}, fmt.Errorf("unknown risk %q", risk)
	}
	if risk == "low" {
		return AdversarialProbePlan{}, errors.New("default probes are for medium/high risk")
	}
	if strings.TrimSpace(revision) == "" {
		return AdversarialProbePlan{}, errors.New("revision is required")
	}
	probes := []AdversarialProbe{
		{
			ID:          "boundary-conditions",
			Kind:        ProbePrimaryAcceptance,
			DerivedFrom: "requirement_contract",
			Description: "Exercise declared boundary conditions from the accepted requirement contract",
			Executable:  true,
			Status:      TerminalNotRun,
		},
		{
			ID:          "malformed-inputs",
			Kind:        ProbePrimaryAcceptance,
			DerivedFrom: "requirement_contract",
			Description: "Reject malformed inputs instead of coercing them into success",
			Executable:  true,
			Status:      TerminalNotRun,
		},
		{
			ID:          "idempotence",
			Kind:        ProbePrimaryAcceptance,
			DerivedFrom: "requirement_contract",
			Description: "Repeat the same accepted operation and confirm stable outcomes",
			Executable:  true,
			Status:      TerminalNotRun,
		},
		{
			ID:          "authorization-boundaries",
			Kind:        ProbePrimaryAcceptance,
			DerivedFrom: "authorization",
			Description: "Confirm unauthorized callers cannot perform protected actions",
			Executable:  true,
			Status:      TerminalNotRun,
		},
	}
	if len(requirementIDs) > 0 {
		probes = append(probes, AdversarialProbe{
			ID:          "requirement-ids",
			Kind:        ProbePrimaryAcceptance,
			DerivedFrom: strings.Join(requirementIDs, ","),
			Description: "Map each requirement ID to an executable acceptance example",
			Executable:  true,
			Status:      TerminalNotRun,
		})
	}
	if includeCalendar {
		probes = append(probes, AdversarialProbe{
			ID:          "calendar-validity",
			Kind:        ProbePrimaryAcceptance,
			DerivedFrom: "time_zones_dates",
			Description: "Reject impossible YYYY-MM-DD values such as 2026-02-30 without Date normalization",
			Executable:  true,
			Status:      TerminalNotRun,
		})
	}
	probes = append(probes, AdversarialProbe{
		ID:          "regression-history",
		Kind:        ProbePostHocDiagnostic,
		DerivedFrom: "regression_history",
		Description: "Optional diagnostic against known regressions; must not rewrite primary scores",
		Executable:  false,
		Status:      TerminalNotRun,
		HumanNeeded: true,
	})
	return AdversarialProbePlan{
		SchemaVersion: SchemaVersion,
		Revision:      revision,
		Risk:          risk,
		Probes:        probes,
		RepairLoop:    "bounded_repair_reverify",
		Status:        TerminalNotRun,
	}, nil
}

// EvaluateCalendarProbe runs the strict calendar fixture used by adversarial probes.
func EvaluateCalendarProbe(date string) AdversarialProbe {
	probe := AdversarialProbe{
		ID:          "calendar-validity",
		Kind:        ProbePrimaryAcceptance,
		DerivedFrom: "time_zones_dates",
		Description: "Reject impossible YYYY-MM-DD values without Date normalization",
		Executable:  true,
		HumanNeeded: false,
	}
	if err := ValidateCalendarDate(date); err != nil {
		probe.Status = TerminalFail
		probe.Evidence = err.Error()
		return probe
	}
	probe.Status = TerminalPass
	probe.Evidence = fmt.Sprintf("%s is a real Gregorian calendar date", date)
	return probe
}
