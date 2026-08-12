package releasecontract

import (
	"strings"
	"testing"
	"time"
)

func validBinding() Binding {
	now := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	return Binding{
		SchemaVersion: "1.0.0", Repository: "owner/repo", BaseRevision: "base", FinalRevision: "final",
		RequirementDigest: "requirements", ChangeSetDigest: "change", PolicyRevision: "policy",
		Tests: []Test{{ID: "acceptance", RequirementIDs: []string{"R1"}, Revision: "final", Status: "PASS", Required: true, ArtifactDigest: "artifact", ActorRunID: "test-run", StartedAt: now, FinishedAt: now}},
		Reviews: []Review{{ReviewerRunID: "review-run", ChangeAuthorID: "build-run", Revision: "final", ChangeSetDigest: "change", Status: "PASS"}},
		Decisions: []Decision{{Kind: "RELEASE_RISK", Authority: "human", Subject: "final", Status: "APPROVED"}},
	}
}

func TestValidateAcceptsRevisionBoundExecutedEvidence(t *testing.T) {
	if err := validBinding().Validate(time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)); err != nil { t.Fatal(err) }
}

func TestValidateDoesNotPromotePlannedToPass(t *testing.T) {
	b := validBinding()
	b.Tests[0].Status = "PLANNED"
	if err := b.Validate(time.Time{}); err == nil || !strings.Contains(err.Error(), "not an executed pass") { t.Fatalf("expected planned test rejection, got %v", err) }
}

func TestValidateRejectsStaleAndSelfReview(t *testing.T) {
	b := validBinding()
	b.Reviews[0].ReviewerRunID = b.Reviews[0].ChangeAuthorID
	if err := b.Validate(time.Time{}); err == nil || !strings.Contains(err.Error(), "not independent") { t.Fatalf("expected self-review rejection, got %v", err) }
	b = validBinding()
	b.Reviews[0].Revision = "old"
	if err := b.Validate(time.Time{}); err == nil || !strings.Contains(err.Error(), "stale") { t.Fatalf("expected stale review rejection, got %v", err) }
}

func TestValidateRejectsUntypedAndNonWaivableExceptions(t *testing.T) {
	b := validBinding()
	b.Exceptions = []Exception{{Kind: "CORRECTNESS", Approver: "human", Subject: "failing test", Rationale: "ship", Scope: []string{"R1"}, ExpiresAt: time.Now().Add(time.Hour)}}
	if err := b.Validate(time.Now()); err == nil || !strings.Contains(err.Error(), "not waivable") { t.Fatalf("expected non-waivable rejection, got %v", err) }
	b = validBinding()
	b.Decisions[0].Kind = "CODE_APPROVAL"
	if err := b.Validate(time.Now()); err == nil || !strings.Contains(err.Error(), "human decision") { t.Fatalf("expected untyped decision rejection, got %v", err) }
}
