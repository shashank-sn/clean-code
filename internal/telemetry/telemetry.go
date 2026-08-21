package telemetry

import (
	"sort"
	"strings"
	"time"
)

const (
	DecisionContinue               = "continue"
	DecisionRevisePlan             = "revise_plan"
	DecisionReorganizeArchitecture = "reorganize_architecture"
	DecisionStopEscalate           = "stop_escalate"
)

type Budget struct {
	MaxTurns   int `json:"max_turns,omitempty"`
	MaxRetries int `json:"max_retries,omitempty"`
}

type Event struct {
	At               time.Time `json:"at"`
	Kind             string    `json:"kind"`
	Files            []string  `json:"files,omitempty"`
	CheckID          string    `json:"check_id,omitempty"`
	Result           string    `json:"result,omitempty"`
	Retry            bool      `json:"retry,omitempty"`
	ContextExhausted bool      `json:"context_exhausted,omitempty"`
}

type Summary struct {
	Turns              int      `json:"turns"`
	Retries            int      `json:"retries"`
	FailingCheckCycles int      `json:"failing_check_cycles"`
	RepeatedFiles      []string `json:"repeated_files"`
	Decision           string   `json:"decision"`
	StopReason         string   `json:"stop_reason,omitempty"`
}

func Evaluate(events []Event, budget Budget) Summary {
	summary := Summary{Decision: DecisionContinue, RepeatedFiles: []string{}}
	files := map[string]int{}
	failedChecks := map[string]int{}
	for _, event := range events {
		summary.Turns++
		if event.Retry {
			summary.Retries++
		}
		if event.Result == "FAIL" && event.CheckID != "" {
			failedChecks[event.CheckID]++
		}
		for _, file := range event.Files {
			files[file]++
		}
		if event.ContextExhausted {
			summary.Decision, summary.StopReason = DecisionStopEscalate, "context budget exhausted"
		}
	}
	for _, count := range failedChecks {
		if count > 1 {
			summary.FailingCheckCycles += count - 1
		}
	}
	for file, count := range files {
		if count > 1 {
			summary.RepeatedFiles = append(summary.RepeatedFiles, file)
		}
	}
	sort.Strings(summary.RepeatedFiles)
	if summary.StopReason != "" {
		return summary
	}
	if budget.MaxTurns > 0 && summary.Turns > budget.MaxTurns {
		summary.Decision, summary.StopReason = DecisionStopEscalate, "turn budget exhausted"
		return summary
	}
	if budget.MaxRetries > 0 && summary.Retries > budget.MaxRetries {
		summary.Decision, summary.StopReason = DecisionStopEscalate, "retry budget exhausted"
		return summary
	}
	if len(summary.RepeatedFiles) > 1 {
		summary.Decision, summary.StopReason = DecisionReorganizeArchitecture, "repeated coordinated file edits"
		return summary
	}
	if summary.FailingCheckCycles > 0 {
		summary.Decision, summary.StopReason = DecisionRevisePlan, "repeated failing check cycle"
	}
	return summary
}

func ValidDecision(value string) bool {
	return strings.TrimSpace(value) == DecisionContinue || value == DecisionRevisePlan || value == DecisionReorganizeArchitecture || value == DecisionStopEscalate
}
