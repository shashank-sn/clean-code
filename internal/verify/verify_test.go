package verify

import (
	"context"
	"testing"
	"time"

	"clean-code/internal/contracts"
)

type fakeRunner struct {
	run []string
}

func (runner *fakeRunner) Run(_ context.Context, command contracts.CommandSpec, revision string) contracts.CheckResult {
	runner.run = append(runner.run, command.ID)
	return contracts.CheckResult{
		SchemaVersion: "1.0.0",
		CheckID:       command.ID,
		Category:      "command",
		Provider:      "fake",
		Status:        contracts.StatusPass,
		Required:      command.Required,
		Revision:      revision,
		StartedAt:     time.Unix(1, 0).UTC(),
	}
}

func TestRunReturnsNotConfiguredWithoutCommands(t *testing.T) {
	report := Service{Now: fixedNow}.Run(context.Background(), "/repo", "abc", nil, nil, "")
	if report.Complete || report.Successful() {
		t.Fatal("empty policy must not report success")
	}
	if len(report.Results) != 1 || report.Results[0].Status != contracts.StatusNotConfigured {
		t.Fatalf("unexpected results: %+v", report.Results)
	}
}

func TestRunUsesProposedPolicyWhenNoTrustedPolicyProvided(t *testing.T) {
	runner := &fakeRunner{}
	report := Service{Runner: runner, Now: fixedNow}.Run(
		context.Background(), "/repo", "abc",
		[]contracts.CommandSpec{{ID: "test", Executable: "test", Required: true}},
		nil, "",
	)
	if !report.Successful() || len(runner.run) != 1 || runner.run[0] != "test" {
		t.Fatalf("unexpected report or execution: %+v %#v", report, runner.run)
	}
}

func TestRunDoesNotExecuteUnapprovedRepositoryPolicy(t *testing.T) {
	runner := &fakeRunner{}
	report := Service{Runner: runner, Now: fixedNow}.Run(
		context.Background(), "/repo", "abc",
		[]contracts.CommandSpec{{ID: "test", Executable: "test", Required: true}},
		nil, "unapproved repository policy",
	)
	if len(runner.run) != 0 {
		t.Fatalf("unapproved command executed: %#v", runner.run)
	}
	if report.Successful() || len(report.Results) != 1 || report.Results[0].CheckID != "policy.approval" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRunBlocksDeltaAndExecutesOnlyTrustedPolicy(t *testing.T) {
	runner := &fakeRunner{}
	trusted := []contracts.CommandSpec{{ID: "trusted", Executable: "safe", Required: true}}
	proposed := []contracts.CommandSpec{{ID: "proposed", Executable: "unsafe", Required: true}}
	report := Service{Runner: runner, Now: fixedNow}.Run(
		context.Background(), "/repo", "abc", proposed, trusted, "/policy.json",
	)
	if report.Successful() || report.Complete {
		t.Fatal("policy drift must block completion")
	}
	if len(runner.run) != 1 || runner.run[0] != "trusted" {
		t.Fatalf("only trusted commands may run, got %#v", runner.run)
	}
	if len(report.PolicyDeltas) != 2 {
		t.Fatalf("expected two deltas, got %#v", report.PolicyDeltas)
	}
}

func TestRunBlocksRepositoryRevisionChange(t *testing.T) {
	runner := &fakeRunner{}
	service := Service{
		Runner:          runner,
		Now:             fixedNow,
		CurrentRevision: func(string) string { return "def" },
	}
	report := service.Run(
		context.Background(), "/repo", "abc",
		[]contracts.CommandSpec{{ID: "test", Executable: "test", Required: true}},
		nil, "",
	)
	if report.Successful() || report.Complete {
		t.Fatal("revision change must block completion")
	}
	if len(report.Results) != 2 || report.Results[1].CheckID != "revision.stable" {
		t.Fatalf("expected integrity result, got %+v", report.Results)
	}
}

func fixedNow() time.Time {
	return time.Unix(10, 0).UTC()
}
