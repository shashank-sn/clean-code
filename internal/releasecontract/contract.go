package releasecontract

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var executedStatuses = map[string]bool{"PASS": true, "FAIL": true}

type Binding struct {
	SchemaVersion      string     `json:"schema_version"`
	Repository         string     `json:"repository"`
	BaseRevision       string     `json:"base_revision"`
	FinalRevision      string     `json:"final_revision"`
	RequirementDigest  string     `json:"requirement_digest"`
	ChangeSetDigest    string     `json:"change_set_digest"`
	PolicyRevision     string     `json:"policy_revision"`
	Tests              []Test     `json:"tests"`
	Reviews            []Review   `json:"reviews"`
	Decisions          []Decision `json:"decisions"`
	Exceptions         []Exception `json:"exceptions,omitempty"`
}

type Test struct {
	ID             string    `json:"id"`
	RequirementIDs []string  `json:"requirement_ids"`
	Revision       string    `json:"revision"`
	Status         string    `json:"status"`
	ArtifactDigest string    `json:"artifact_digest,omitempty"`
	ActorRunID     string    `json:"actor_run_id,omitempty"`
	StartedAt      time.Time `json:"started_at,omitempty"`
	FinishedAt     time.Time `json:"finished_at,omitempty"`
	Required       bool      `json:"required"`
}

type Review struct {
	ReviewerRunID  string `json:"reviewer_run_id"`
	ChangeAuthorID string `json:"change_author_run_id"`
	Revision       string `json:"revision"`
	ChangeSetDigest string `json:"change_set_digest"`
	Status         string `json:"status"`
}

type Decision struct {
	Kind       string `json:"kind"`
	Authority  string `json:"authority"`
	Subject    string `json:"subject"`
	Status     string `json:"status"`
}

type Exception struct {
	Kind       string    `json:"kind"`
	Approver   string    `json:"approver"`
	Subject    string    `json:"subject"`
	Rationale  string    `json:"rationale"`
	Scope      []string  `json:"scope"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (b Binding) Validate(now time.Time) error {
	required := []string{b.SchemaVersion, b.Repository, b.BaseRevision, b.FinalRevision, b.RequirementDigest, b.ChangeSetDigest, b.PolicyRevision}
	for _, value := range required {
		if strings.TrimSpace(value) == "" {
			return errors.New("release binding metadata is incomplete")
		}
	}
	if b.SchemaVersion != "1.0.0" {
		return errors.New("unsupported release binding schema")
	}
	for _, test := range b.Tests {
		if test.Revision != b.FinalRevision {
			return fmt.Errorf("test %q belongs to another revision", test.ID)
		}
		if test.Required && test.Status != "PASS" {
			return fmt.Errorf("required test %q is not an executed pass", test.ID)
		}
		if executedStatuses[test.Status] && (test.ArtifactDigest == "" || test.ActorRunID == "" || test.StartedAt.IsZero() || test.FinishedAt.Before(test.StartedAt)) {
			return fmt.Errorf("test %q lacks executed evidence", test.ID)
		}
	}
	for _, review := range b.Reviews {
		if review.Revision != b.FinalRevision || review.ChangeSetDigest != b.ChangeSetDigest {
			return errors.New("review is stale or covers another change set")
		}
		if review.ReviewerRunID == "" || review.ReviewerRunID == review.ChangeAuthorID {
			return errors.New("review is not independent")
		}
		if review.Status != "PASS" {
			return errors.New("review is incomplete")
		}
	}
	for _, decision := range b.Decisions {
		if !oneOf(decision.Kind, "INTENT", "POLICY", "RELEASE_RISK") || decision.Authority == "" || decision.Subject == "" || decision.Status != "APPROVED" {
			return errors.New("human decision is invalid or incomplete")
		}
	}
	for _, exception := range b.Exceptions {
		if !oneOf(exception.Kind, "PROCESS", "AVAILABILITY", "TEMPORARY_POLICY") {
			return errors.New("exception kind is not waivable")
		}
		if exception.Approver == "" || exception.Subject == "" || exception.Rationale == "" || len(exception.Scope) == 0 || !exception.ExpiresAt.After(now) {
			return errors.New("exception is invalid or expired")
		}
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate { return true }
	}
	return false
}
