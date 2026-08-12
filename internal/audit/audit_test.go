package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clean-code/internal/contracts"
	"clean-code/internal/evidence"
)

func TestBuildCompleteReceiptHashesEveryEvidenceInput(t *testing.T) {
	manifest := writeAuditFixture(t, true, "abc")
	receipt, err := Build(manifest, fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Complete || len(receipt.Gaps) != 0 || len(receipt.Artifacts) != 7 {
		t.Fatalf("expected complete receipt, got %+v", receipt)
	}
	for _, artifact := range receipt.Artifacts {
		if len(artifact.SHA256) != 64 {
			t.Fatalf("expected SHA-256 digest, got %+v", artifact)
		}
	}
}

func TestBuildReportsStaleVerificationAndMissingHumanCheck(t *testing.T) {
	manifest := writeAuditFixture(t, false, "stale")
	receipt, err := Build(manifest, fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Complete || !hasGap(receipt.Gaps, "verification revision does not match audit revision") || !hasGap(receipt.Gaps, "required ui-qa spot check was not performed") {
		t.Fatalf("expected stale and human gaps, got %+v", receipt)
	}
}

func TestBuildReportsExpiredVerification(t *testing.T) {
	manifest := writeAuditFixture(t, true, "abc")
	var input map[string]any
	body, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	input["max_evidence_age_seconds"] = 1
	writeJSON(t, manifest, input)
	receipt, err := Build(manifest, func() time.Time { return fixedNow().Add(2 * time.Second) })
	if err != nil {
		t.Fatal(err)
	}
	if !hasGap(receipt.Gaps, "verification evidence is outside the permitted age") {
		t.Fatalf("expected expiry gap, got %+v", receipt)
	}
}

func TestWriteCreatesImmutableOwnerOnlyReceipt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt.json")
	receipt := Receipt{SchemaVersion: "1.0.0", Repository: "/repo", Revision: "abc", PolicyRevision: "policy", CreatedAt: fixedNow(), Artifacts: []Artifact{}, Gaps: []string{}, Exceptions: []string{}}
	if err := Write(path, receipt); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, receipt); err == nil {
		t.Fatal("expected immutable output rejection")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected owner-only receipt, got %o", info.Mode().Perm())
	}
}

func TestCheckDetectsChangedEvidence(t *testing.T) {
	manifest := writeAuditFixture(t, true, "abc")
	receipt, err := Build(manifest, fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	receiptPath := filepath.Join(t.TempDir(), "receipt.json")
	if err := Write(receiptPath, receipt); err != nil {
		t.Fatal(err)
	}
	if _, err := Check(manifest, receiptPath); err != nil {
		t.Fatalf("expected matching receipt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(manifest), "unit.json"), []byte(`{"status":"changed"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Check(manifest, receiptPath); err == nil {
		t.Fatal("expected changed evidence detection")
	}
}

func writeAuditFixture(t *testing.T, checked bool, verificationRevision string) string {
	t.Helper()
	root := t.TempDir()
	verification := evidence.Report{
		SchemaVersion: "1.0.0", Repository: "/repo", Revision: verificationRevision,
		PolicySource: "approved.json", StartedAt: fixedNow(), FinishedAt: fixedNow(), Complete: true,
		Results: []contracts.CheckResult{{
			SchemaVersion: "1.0.0", CheckID: "test", Category: "unit-test", Provider: "fixture",
			Status: contracts.StatusPass, Required: true, Revision: verificationRevision, StartedAt: fixedNow(),
		}},
	}
	writeJSON(t, filepath.Join(root, "verification.json"), verification)
	writeJSON(t, filepath.Join(root, "test-plan.json"), map[string]any{
		"schema_version": "1.0.0",
		"requirements":   []any{map[string]any{"id": "R1", "acceptance_examples": []string{"returns the result"}}},
		"tracks": []any{
			map[string]any{"kind": "unit", "requirement_ids": []string{"R1"}, "status": "PASS", "owner": "unit", "evidence": "unit.json"},
			map[string]any{"kind": "acceptance", "requirement_ids": []string{"R1"}, "status": "PASS", "owner": "acceptance", "evidence": "acceptance.json", "context_source": "requirements"},
			map[string]any{"kind": "integration", "requirement_ids": []string{"R1"}, "status": "PASS", "owner": "integration", "evidence": "integration.json"},
			map[string]any{"kind": "ui-qa", "requirement_ids": []string{"R1"}, "status": "INAPPLICABLE", "reason": "no user interface", "context_source": "requirements"},
		},
	})
	writeJSON(t, filepath.Join(root, "unit.json"), map[string]any{"status": "PASS"})
	writeJSON(t, filepath.Join(root, "acceptance.json"), map[string]any{"status": "PASS"})
	writeJSON(t, filepath.Join(root, "integration.json"), map[string]any{"status": "PASS"})
	writeJSON(t, filepath.Join(root, "review.json"), map[string]any{
		"schema_version": "1.0.0", "revision": "abc", "change_author": "author",
		"reviewer": "reviewer", "scope": []string{"change.go"}, "findings": []any{},
	})
	status := "CHECKED"
	reason := ""
	if !checked {
		status = "NOT_CHECKED"
		reason = "manual UI session pending"
	}
	uiCheck := map[string]any{"kind": "ui-qa", "required": true, "status": status}
	if checked {
		uiCheck["scope"] = "primary user flow"
		uiCheck["outcome"] = "matches"
	} else {
		uiCheck["reason"] = reason
	}
	checks := []any{
		map[string]any{"kind": "requirements", "required": true, "status": "CHECKED", "scope": "R1", "outcome": "matches"},
		map[string]any{"kind": "acceptance", "required": true, "status": "CHECKED", "scope": "R1 example", "outcome": "matches"},
		uiCheck,
		map[string]any{"kind": "code-sample", "required": true, "status": "CHECKED", "scope": "change.go", "outcome": "clear"},
	}
	writeJSON(t, filepath.Join(root, "spot-check.json"), map[string]any{
		"schema_version": "1.0.0", "revision": "abc", "reviewer": "human", "checks": checks,
	})
	writeJSON(t, filepath.Join(root, "audit-input.json"), map[string]any{
		"schema_version": "1.0.0", "repository": "/repo", "revision": "abc", "policy_revision": "policy-1",
		"verification": "verification.json", "test_plan": "test-plan.json", "review": "review.json", "spot_check": "spot-check.json",
		"supporting_evidence": []string{"unit.json", "acceptance.json", "integration.json"},
	})
	return filepath.Join(root, "audit-input.json")
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
}

func fixedNow() time.Time {
	return time.Unix(10, 0).UTC()
}

func hasGap(gaps []string, target string) bool {
	for _, gap := range gaps {
		if gap == target {
			return true
		}
	}
	return false
}
