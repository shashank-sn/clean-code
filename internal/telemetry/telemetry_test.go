package telemetry

import "testing"

func TestEvaluateStopsForBudgetAndReorganizesForRepeatedEdits(t *testing.T) {
	budget := Evaluate([]Event{{Kind: "turn"}, {Kind: "turn"}}, Budget{MaxTurns: 1})
	if budget.Decision != DecisionStopEscalate || budget.StopReason == "" {
		t.Fatalf("expected budget stop, got %+v", budget)
	}
	reorganize := Evaluate([]Event{{Files: []string{"a.go", "b.go"}}, {Files: []string{"a.go", "b.go"}}}, Budget{})
	if reorganize.Decision != DecisionReorganizeArchitecture {
		t.Fatalf("expected reorganization, got %+v", reorganize)
	}
}

func TestEvaluateRequestsPlanRevisionForRepeatedFailures(t *testing.T) {
	summary := Evaluate([]Event{{CheckID: "test", Result: "FAIL"}, {CheckID: "test", Result: "FAIL"}}, Budget{})
	if summary.Decision != DecisionRevisePlan || summary.FailingCheckCycles != 1 {
		t.Fatalf("expected revision decision, got %+v", summary)
	}
}
