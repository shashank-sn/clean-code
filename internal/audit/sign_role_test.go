package audit_test

import (
	"strings"
	"testing"
	"time"

	"github.com/shashank-sn/clean-code/internal/audit"
)

func seed(offset byte) []byte {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 1 + int(offset))
	}
	return seed
}

func roleSigner(t *testing.T, role, revision, evidenceSHA string, seed []byte) audit.RoleSigner {
	t.Helper()
	signer, err := audit.SignRoleWithDigest(role, revision, evidenceSHA, seed)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func TestSignRoleWithDigestIsDeterministic(t *testing.T) {
	first := roleSigner(t, "implementer", "abc", "1111", seed(0))
	second := roleSigner(t, "implementer", "abc", "1111", seed(0))
	if first.KeyID != second.KeyID || first.Signature != second.Signature || first.PayloadSHA256 != second.PayloadSHA256 {
		t.Fatal("signing the same role binding twice must be deterministic")
	}
}

func TestVerifyRoleSignersDistinctKeysPass(t *testing.T) {
	signers := map[string]audit.RoleSigner{
		"implementer":  roleSigner(t, "implementer", "abc", "1111", seed(0)),
		"reviewer":     roleSigner(t, "reviewer", "abc", "2222", seed(1)),
		"spot_checker": roleSigner(t, "spot_checker", "abc", "3333", seed(2)),
	}
	evidence := map[string]string{"implementer": "1111", "reviewer": "2222", "spot_checker": "3333"}
	verified, err := audit.VerifyRoleSigners(signers, "abc", evidence)
	if err != nil {
		t.Fatalf("expected distinct keys to pass: %v", err)
	}
	if len(verified) != 3 {
		t.Fatalf("expected 3 verified signers, got %d", len(verified))
	}
	if verified[0].Role != "implementer" || verified[1].Role != "reviewer" || verified[2].Role != "spot_checker" {
		t.Fatalf("signers must be sorted by role: %+v", verified)
	}
}

func TestVerifyRoleSignersSameKeyForTwoRolesFails(t *testing.T) {
	reused := roleSigner(t, "implementer", "abc", "1111", seed(0))
	reviewerSame := roleSigner(t, "reviewer", "abc", "2222", seed(0))
	if reused.KeyID != reviewerSame.KeyID {
		t.Fatal("fixture bug: seeds must produce the same key")
	}
	signers := map[string]audit.RoleSigner{
		"implementer":  reused,
		"reviewer":     reviewerSame,
		"spot_checker": roleSigner(t, "spot_checker", "abc", "3333", seed(2)),
	}
	evidence := map[string]string{"implementer": "1111", "reviewer": "2222", "spot_checker": "3333"}
	if _, err := audit.VerifyRoleSigners(signers, "abc", evidence); err == nil {
		t.Fatal("expected same key across roles to fail")
	} else if !strings.Contains(err.Error(), "share the same signing key") {
		t.Fatalf("expected duplicate-key error, got: %v", err)
	}
}

func TestVerifyRoleSignersMissingRequiredRoleFails(t *testing.T) {
	signers := map[string]audit.RoleSigner{
		"implementer": roleSigner(t, "implementer", "abc", "1111", seed(0)),
	}
	evidence := map[string]string{"implementer": "1111"}
	if _, err := audit.VerifyRoleSigners(signers, "abc", evidence); err == nil {
		t.Fatal("expected missing reviewer to fail")
	} else if !strings.Contains(err.Error(), "reviewer") {
		t.Fatalf("expected missing-role error naming reviewer, got: %v", err)
	}
}

func TestVerifyRoleSignersTamperedSignatureFails(t *testing.T) {
	signers := map[string]audit.RoleSigner{
		"implementer": roleSigner(t, "implementer", "abc", "1111", seed(0)),
		"reviewer":    roleSigner(t, "reviewer", "abc", "2222", seed(1)),
	}
	evidence := map[string]string{"implementer": "1111", "reviewer": "2222"}
	reused := signers["reviewer"]
	reused.Signature = "AAAA" + reused.Signature[4:]
	signers["reviewer"] = reused
	if _, err := audit.VerifyRoleSigners(signers, "abc", evidence); err == nil {
		t.Fatal("expected tampered signature to fail")
	}
}

func TestVerifyRoleSignersWrongEvidenceDigestFails(t *testing.T) {
	signers := map[string]audit.RoleSigner{
		"implementer": roleSigner(t, "implementer", "abc", "1111", seed(0)),
		"reviewer":    roleSigner(t, "reviewer", "abc", "2222", seed(1)),
	}
	evidence := map[string]string{"implementer": "9999", "reviewer": "2222"}
	if _, err := audit.VerifyRoleSigners(signers, "abc", evidence); err == nil {
		t.Fatal("expected evidence digest mismatch to fail")
	}
}

func TestSignReceiptRejectsAuditorKeyEqualToRoleSigner(t *testing.T) {
	roleSeed := seed(0)
	signer := roleSigner(t, "implementer", "abc", "1111", roleSeed)
	receipt := audit.Receipt{
		SchemaVersion: "1.0.0",
		Repository:    "/repo",
		Revision:      "abc",
		CreatedAt:     time.Unix(10, 0).UTC(),
		Artifacts:     []audit.Artifact{},
		Signers:       []audit.Signer{{Role: "implementer", KeyID: signer.KeyID}},
	}
	if _, err := audit.SignReceipt(receipt, roleSeed); err == nil {
		t.Fatal("expected SignReceipt to reject the auditor key matching a role signer key")
	}
}
