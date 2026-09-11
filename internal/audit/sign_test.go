package audit_test

import (
	"testing"
	"time"

	"github.com/shashank-sn/clean-code/internal/audit"
)

func TestSignReceiptIsDeterministic(t *testing.T) {
	seed := testSeed()
	receipt := signableReceipt()

	first, err := audit.SignReceipt(receipt, seed)
	if err != nil {
		t.Fatal(err)
	}
	second, err := audit.SignReceipt(receipt, seed)
	if err != nil {
		t.Fatal(err)
	}

	if first.Signature != second.Signature {
		t.Fatal("signing the same receipt twice must produce the same signature")
	}
	if first.SignedPayloadSHA256 != second.SignedPayloadSHA256 {
		t.Fatal("signing the same receipt twice must produce the same payload digest")
	}
	if first.SigningKeyFingerprint != second.SigningKeyFingerprint {
		t.Fatal("signing the same receipt twice must produce the same key fingerprint")
	}
	if first.SigningPublicKey != second.SigningPublicKey {
		t.Fatal("signing the same receipt twice must produce the same public key")
	}
	if receipt.Signature != "" || receipt.SignedPayloadSHA256 != "" || receipt.SigningKeyFingerprint != "" || receipt.SigningPublicKey != "" {
		t.Fatal("SignReceipt must not mutate its input")
	}
}

func TestSignReceiptIgnoresExistingSignatureFields(t *testing.T) {
	plain := signableReceipt()
	withExistingFields := plain
	withExistingFields.SigningKeyFingerprint = "ignored"
	withExistingFields.SignedPayloadSHA256 = "ignored"
	withExistingFields.Signature = "ignored"
	withExistingFields.SigningPublicKey = "ignored"

	first := mustSign(t, plain)
	second := mustSign(t, withExistingFields)
	if first.Signature != second.Signature || first.SignedPayloadSHA256 != second.SignedPayloadSHA256 {
		t.Fatal("signature fields must not be included in the signed payload")
	}
}

func TestVerifyReceiptAcceptsFreshlySignedReceipt(t *testing.T) {
	signed, err := audit.SignReceipt(signableReceipt(), testSeed())
	if err != nil {
		t.Fatal(err)
	}

	if err := audit.VerifyReceipt(signed); err != nil {
		t.Fatalf("freshly signed receipt should verify: %v", err)
	}
}

func TestVerifyReceiptRejectsRevisionTampering(t *testing.T) {
	signed := mustSign(t, signableReceipt())
	signed.Revision = "different-revision"

	if err := audit.VerifyReceipt(signed); err == nil {
		t.Fatal("expected revision tampering to fail verification")
	}
}

func TestVerifyReceiptRejectsArtifactTampering(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*audit.Receipt)
	}{
		{name: "path", mutate: func(receipt *audit.Receipt) { receipt.Artifacts[0].Path = "changed.json" }},
		{name: "sha", mutate: func(receipt *audit.Receipt) { receipt.Artifacts[1].SHA256 = "changed-sha256" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signed := mustSign(t, signableReceipt())
			tt.mutate(&signed)

			if err := audit.VerifyReceipt(signed); err == nil {
				t.Fatalf("expected artifact %s tampering to fail verification", tt.name)
			}
		})
	}
}

func TestVerifyReceiptRejectsEmbeddedPublicKeyTampering(t *testing.T) {
	signed := mustSign(t, signableReceipt())
	other := mustSignWithSeed(t, signableReceiptWithRevision("other-revision"), alternateSeed())
	signed.SigningPublicKey = other.SigningPublicKey

	if err := audit.VerifyReceipt(signed); err == nil {
		t.Fatal("expected embedded public-key tampering to fail verification")
	}
}

func TestVerifyReceiptRejectsSignedMetadataTampering(t *testing.T) {
	for _, field := range []string{"payload digest", "signature"} {
		t.Run(field, func(t *testing.T) {
			signed := mustSign(t, signableReceipt())
			if field == "payload digest" {
				signed.SignedPayloadSHA256 = "changed-digest"
			} else {
				signed.Signature = "changed-signature"
			}
			if err := audit.VerifyReceipt(signed); err == nil {
				t.Fatalf("expected tampered %s to fail verification", field)
			}
		})
	}
}

func TestVerifyReceiptRejectsUnsignedReceipt(t *testing.T) {
	if err := audit.VerifyReceipt(signableReceipt()); err == nil {
		t.Fatal("expected an unsigned receipt to fail verification")
	}
}

func TestSignReceiptRejectsInvalidSeedLength(t *testing.T) {
	for _, seed := range [][]byte{nil, {}, make([]byte, 31), make([]byte, 33)} {
		if _, err := audit.SignReceipt(signableReceipt(), seed); err == nil {
			t.Fatalf("expected seed length %d to be rejected", len(seed))
		}
	}
}

func signableReceipt() audit.Receipt {
	return signableReceiptWithRevision("revision-1")
}

func signableReceiptWithRevision(revision string) audit.Receipt {
	return audit.Receipt{
		SchemaVersion:  "1.0.0",
		Repository:     "/repo",
		Revision:       revision,
		PolicyRevision: "policy-1",
		CreatedAt:      time.Unix(10, 0).UTC(),
		Artifacts: []audit.Artifact{
			{Path: "verification.json", SHA256: "1111111111111111111111111111111111111111111111111111111111111111"},
			{Path: "review.json", SHA256: "2222222222222222222222222222222222222222222222222222222222222222"},
		},
		Gaps:       []string{"a gap"},
		Exceptions: []string{"an exception"},
	}
}

func testSeed() []byte {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	return seed
}

func alternateSeed() []byte {
	seed := testSeed()
	seed[0]++
	return seed
}

func mustSign(t *testing.T, receipt audit.Receipt) audit.Receipt {
	return mustSignWithSeed(t, receipt, testSeed())
}

func mustSignWithSeed(t *testing.T, receipt audit.Receipt, seed []byte) audit.Receipt {
	t.Helper()
	signed, err := audit.SignReceipt(receipt, seed)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}
