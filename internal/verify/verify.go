package verify

import (
	"context"
	"time"

	"clean-code/internal/contracts"
	"clean-code/internal/evidence"
	"clean-code/internal/policy"
)

type CommandRunner interface {
	Run(context.Context, contracts.CommandSpec, string) contracts.CheckResult
}

type Service struct {
	Runner          CommandRunner
	Now             func() time.Time
	CurrentRevision func(string) string
}

func (service Service) Run(
	ctx context.Context,
	repository string,
	revision string,
	proposed []contracts.CommandSpec,
	trusted []contracts.CommandSpec,
	trustedSource string,
) evidence.Report {
	now := service.Now
	if now == nil {
		now = time.Now
	}
	started := now().UTC()
	report := evidence.Report{
		SchemaVersion: "1.0.0",
		Repository:    repository,
		Revision:      revision,
		PolicySource:  "repository:.clean-code.json",
		StartedAt:     started,
	}
	commands := proposed
	if trustedSource != "" {
		report.PolicySource = trustedSource
		report.PolicyDeltas = policy.Compare(trusted, proposed)
		commands = trusted
	}
	if len(commands) == 0 {
		checkID := "configuration.commands"
		category := "configuration"
		status := contracts.StatusNotConfigured
		summary := "no verification commands are configured"
		if len(report.PolicyDeltas) > 0 {
			checkID = "policy.approval"
			category = "policy"
			status = contracts.StatusNotRun
			summary = "repository commands were not executed because the policy is not approved"
		}
		report.Results = []contracts.CheckResult{{
			SchemaVersion: "1.0.0",
			CheckID:       checkID,
			Category:      category,
			Provider:      "clean-code",
			Status:        status,
			Required:      true,
			Revision:      revision,
			StartedAt:     started,
			Evidence: contracts.Evidence{
				Summary: summary,
			},
		}}
		report.FinishedAt = now().UTC()
		return report
	}
	if service.Runner == nil {
		report.Results = []contracts.CheckResult{{
			SchemaVersion: "1.0.0",
			CheckID:       "runner.available",
			Category:      "infrastructure",
			Provider:      "clean-code",
			Status:        contracts.StatusError,
			Required:      true,
			Revision:      revision,
			StartedAt:     started,
			Evidence: contracts.Evidence{
				Summary: "verification runner is not configured",
			},
		}}
		report.FinishedAt = now().UTC()
		return report
	}
	report.Results = make([]contracts.CheckResult, 0, len(commands))
	for _, command := range commands {
		report.Results = append(report.Results, service.Runner.Run(ctx, command, revision))
	}
	report.Complete = len(report.PolicyDeltas) == 0
	if service.CurrentRevision != nil {
		finalRevision := service.CurrentRevision(repository)
		if finalRevision != revision {
			report.Complete = false
			report.Results = append(report.Results, contracts.CheckResult{
				SchemaVersion: "1.0.0",
				CheckID:       "revision.stable",
				Category:      "integrity",
				Provider:      "clean-code",
				Status:        contracts.StatusError,
				Required:      true,
				Revision:      revision,
				StartedAt:     now().UTC(),
				Evidence: contracts.Evidence{
					Summary: "repository revision changed during verification",
				},
			})
		}
	}
	report.FinishedAt = now().UTC()
	return report
}
