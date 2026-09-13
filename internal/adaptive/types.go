package adaptive

import "github.com/shashank-sn/clean-code/internal/contracts"

const SchemaVersion = "1.0.0"

// Terminal statuses used by playbooks and orchestration records.
const (
	TerminalPass         = "PASS"
	TerminalFail         = "FAIL"
	TerminalNotRun       = "NOT_RUN"
	TerminalNotAvailable = "NOT_AVAILABLE"
	TerminalBlocked      = "BLOCKED"
)

var terminalStatuses = map[string]struct{}{
	TerminalPass: {}, TerminalFail: {}, TerminalNotRun: {},
	TerminalNotAvailable: {}, TerminalBlocked: {},
}

// HostCapabilityView is a read-only projection of host abilities.
// Route decisions must not write host or global configuration.
type HostCapabilityView struct {
	Subagents         bool `json:"subagents"`
	CommandExecution  bool `json:"command_execution"`
	FileEdits         bool `json:"file_edits"`
	BrowserAutomation bool `json:"browser_automation"`
	GitWrite          bool `json:"git_write"`
	PullRequestWrite  bool `json:"pull_request_write"`
}

// TaskSignals are the model-neutral inputs to the adaptive router.
type TaskSignals struct {
	SchemaVersion       string              `json:"schema_version"`
	ChangeType          string              `json:"change_type"`
	Risk                string              `json:"risk"`
	Ambiguity           string              `json:"ambiguity"`
	AffectedBoundaries  []string            `json:"affected_boundaries"`
	EvidenceNeeds       []string            `json:"evidence_needs"`
	HostCapabilities    HostCapabilityView  `json:"host_capabilities"`
	ConcurrencyBudget   int                 `json:"concurrency_budget"`
	AuthorizationMode   string              `json:"authorization_mode,omitempty"`
}

// ParallelAllowance describes what may run concurrently under a route.
type ParallelAllowance struct {
	Allowed         bool     `json:"allowed"`
	MaxConcurrency  int      `json:"max_concurrency"`
	EligibleRoles   []string `json:"eligible_roles"`
	MutationPolicy  string   `json:"mutation_policy"`
	UnavailableNote string   `json:"unavailable_note,omitempty"`
}

// RouteFallback is the safe degraded shape when preferred work cannot run.
type RouteFallback struct {
	PlaybookID string   `json:"playbook_id"`
	Roles      []string `json:"roles"`
	Reason     string   `json:"reason"`
}

// RouteDecision is the typed, inspectable router output. It is advisory
// unless an existing lifecycle contract authorizes execution.
type RouteDecision struct {
	SchemaVersion              string            `json:"schema_version"`
	PlaybookID                 string            `json:"playbook_id"`
	SelectedRoles              []string          `json:"selected_roles"`
	RequiredEvidence           []string          `json:"required_evidence"`
	ArenaWarranted             bool              `json:"arena_warranted"`
	AdversarialProbesWarranted bool              `json:"adversarial_probes_warranted"`
	AllowedParallelism         ParallelAllowance `json:"allowed_parallelism"`
	Reasons                    []string          `json:"reasons"`
	Fallback                   RouteFallback     `json:"fallback"`
	Authorization              string            `json:"authorization"`
	TerminalStates             []string          `json:"terminal_states"`
}

// ArenaCandidate is one independently reasoned option.
type ArenaCandidate struct {
	ID           string   `json:"id"`
	Summary      string   `json:"summary"`
	Evidence     []string `json:"evidence"`
	Risks        []string `json:"risks"`
	Tradeoffs    []string `json:"tradeoffs"`
	Status       string   `json:"status"`
	Unavailable  string   `json:"unavailable_reason,omitempty"`
}

// ArenaDecisionRecord captures a bounded competing-design decision.
type ArenaDecisionRecord struct {
	SchemaVersion         string           `json:"schema_version"`
	Revision              string           `json:"revision"`
	Owner                 string           `json:"owner"`
	DecisionQuestion      string           `json:"decision_question"`
	Constraints           []string         `json:"constraints"`
	Candidates            []ArenaCandidate `json:"candidates"`
	SelectedOptionID      string           `json:"selected_option_id"`
	SelectionRationale    string           `json:"selection_rationale"`
	RejectedRationale     map[string]string `json:"rejected_rationale"`
	Uncertainty           []string         `json:"uncertainty"`
	UnresolvedAssumptions []string         `json:"unresolved_assumptions"`
	PolicyChange          string           `json:"policy_change"`
	Status                string           `json:"status"`
}

// ProbeKind distinguishes preregistered acceptance from post-hoc diagnostics.
type ProbeKind string

const (
	ProbePrimaryAcceptance ProbeKind = "primary_acceptance"
	ProbePostHocDiagnostic ProbeKind = "post_hoc_diagnostic"
)

// AdversarialProbe is one acceptance or diagnostic check.
type AdversarialProbe struct {
	ID          string    `json:"id"`
	Kind        ProbeKind `json:"kind"`
	DerivedFrom string    `json:"derived_from"`
	Description string    `json:"description"`
	Executable  bool      `json:"executable"`
	Status      string    `json:"status"`
	Evidence    string    `json:"evidence,omitempty"`
	HumanNeeded bool      `json:"human_judgment_required"`
}

// AdversarialProbePlan is the revision-bound probe set.
type AdversarialProbePlan struct {
	SchemaVersion string             `json:"schema_version"`
	Revision      string             `json:"revision"`
	Risk          string             `json:"risk"`
	Probes        []AdversarialProbe `json:"probes"`
	RepairLoop    string             `json:"repair_loop"`
	Status        string             `json:"status"`
}

// ParallelChild describes one owned parallel work unit.
type ParallelChild struct {
	ID               string   `json:"id"`
	Role             string   `json:"role"`
	OwnedScope       []string `json:"owned_scope"`
	EvidenceRequest  string   `json:"evidence_request"`
	MutationBoundary string   `json:"mutation_boundary"`
	Status           string   `json:"status"`
	ClaimedArtifacts []string `json:"claimed_artifacts,omitempty"`
	VerifiedArtifacts []string `json:"verified_artifacts,omitempty"`
	FailureNote      string   `json:"failure_note,omitempty"`
}

// ParallelOrchestrationRecord tracks bounded parallel work without losing ownership.
type ParallelOrchestrationRecord struct {
	SchemaVersion     string          `json:"schema_version"`
	Revision          string          `json:"revision"`
	Coordinator       string          `json:"coordinator"`
	WhyParallel       string          `json:"why_parallel"`
	MaxConcurrency    int             `json:"max_concurrency"`
	Children          []ParallelChild `json:"children"`
	SynthesisNotes    []string        `json:"synthesis_notes"`
	HumanApprovalsHeld []string       `json:"human_approvals_held"`
	Status            string          `json:"status"`
}

// Playbook is a concise developer-facing workflow shape.
type Playbook struct {
	SchemaVersion              string   `json:"schema_version"`
	ID                         string   `json:"id"`
	Title                      string   `json:"title"`
	AppliesWhen                []string `json:"applies_when"`
	DoesNotApplyWhen           []string `json:"does_not_apply_when"`
	MinimumRoles               []string `json:"minimum_roles"`
	MinimumEvidence            []string `json:"minimum_evidence"`
	ArenaWarranted             string   `json:"arena_warranted"`
	AdversarialProbesWarranted string   `json:"adversarial_probes_warranted"`
	ParallelGuidance           string   `json:"parallel_guidance"`
	TerminalStates             []string `json:"terminal_states"`
}

// CheckResultStatus re-exports contracts statuses used by adaptive reports.
func CheckResultStatus(status contracts.Status) string {
	return string(status)
}
