package audit

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// roleBinding is the canonical payload an individual role signs. It binds the
// role, the audit revision, and the SHA-256 of that role's evidence file, so a
// signature cannot be reused for another role, revision, or evidence set.
func roleBinding(role, revision, evidenceSHA256 string) []byte {
	return []byte(fmt.Sprintf("role=%s\nrevision=%s\nevidence_sha256=%s\n", role, revision, evidenceSHA256))
}

// RoleSigner is a per-role attestation as supplied on the audit input.
type RoleSigner struct {
	KeyID         string `json:"key_id"`
	PublicKey     string `json:"public_key"`
	Signature     string `json:"signature"`
	PayloadSHA256 string `json:"payload_sha256"`
}

// Signer is the verified per-role attestation recorded on a receipt.
type Signer struct {
	Role          string `json:"role"`
	KeyID         string `json:"key_id"`
	PublicKey     string `json:"public_key"`
	Signature     string `json:"signature"`
	PayloadSHA256 string `json:"payload_sha256"`
}

// SignRole creates a per-role Ed25519 attestation over the role binding for
// revision and the SHA-256 of the evidence file at evidencePath.
func SignRole(role, revision, evidencePath string, seed []byte) (RoleSigner, error) {
	if len(seed) != ed25519.SeedSize {
		return RoleSigner{}, fmt.Errorf("sign role: private key seed must be %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	content, err := os.ReadFile(evidencePath)
	if err != nil {
		return RoleSigner{}, fmt.Errorf("sign role: read evidence: %w", err)
	}
	evidenceDigest := sha256.Sum256(content)
	return SignRoleWithDigest(role, revision, hex.EncodeToString(evidenceDigest[:]), seed)
}

// SignRoleWithDigest creates a per-role attestation given an already-computed
// evidence SHA-256 (hex). Exposed for tests and callers that already have the
// digest.
func SignRoleWithDigest(role, revision, evidenceSHA256 string, seed []byte) (RoleSigner, error) {
	if len(seed) != ed25519.SeedSize {
		return RoleSigner{}, fmt.Errorf("sign role: private key seed must be %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	payload := roleBinding(role, revision, evidenceSHA256)
	digest := sha256.Sum256(payload)
	signature := ed25519.Sign(privateKey, payload)
	fingerprint := sha256.Sum256(publicKey)
	return RoleSigner{
		KeyID:         hex.EncodeToString(fingerprint[:]),
		PublicKey:     base64.StdEncoding.EncodeToString(publicKey),
		Signature:     base64.StdEncoding.EncodeToString(signature),
		PayloadSHA256: hex.EncodeToString(digest[:]),
	}, nil
}

// verifyRoleSigner verifies one role attestation against the expected role,
// revision, and evidence digest.
func verifyRoleSigner(signer RoleSigner, role, revision, evidenceSHA256 string) error {
	if signer.KeyID == "" || signer.PublicKey == "" || signer.Signature == "" || signer.PayloadSHA256 == "" {
		return fmt.Errorf("signer %q metadata is incomplete", role)
	}
	publicKey, err := base64.StdEncoding.DecodeString(signer.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("signer %q has an invalid public key", role)
	}
	fingerprint := sha256.Sum256(publicKey)
	if hex.EncodeToString(fingerprint[:]) != signer.KeyID {
		return fmt.Errorf("signer %q key fingerprint mismatch", role)
	}
	payload := roleBinding(role, revision, evidenceSHA256)
	digest := sha256.Sum256(payload)
	if hex.EncodeToString(digest[:]) != signer.PayloadSHA256 {
		return fmt.Errorf("signer %q payload digest mismatch", role)
	}
	signature, err := base64.StdEncoding.DecodeString(signer.Signature)
	if err != nil {
		return fmt.Errorf("signer %q has an invalid signature encoding", role)
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), payload, signature) {
		return fmt.Errorf("signer %q signature verification failed", role)
	}
	return nil
}

// verifyRoleSigners verifies all supplied evidence-role signers (implementer,
// reviewer, spot_checker) against the role-evidence mapping, enforces
// pairwise-distinct keys, and requires implementer and reviewer to be present.
// It returns the verified signers sorted by role. The auditor attests separately
// by signing the receipt itself (SignReceipt), whose key must differ from every
// role signer here.
func VerifyRoleSigners(signers map[string]RoleSigner, revision string, evidenceSHA256 map[string]string) ([]Signer, error) {
	required := []string{"implementer", "reviewer"}
	seenKeys := map[string]string{}
	var result []Signer
	var problems []string
	for role, signer := range signers {
		evidence, ok := evidenceSHA256[role]
		if !ok {
			problems = append(problems, fmt.Sprintf("unknown signer role %q", role))
			continue
		}
		if err := verifyRoleSigner(signer, role, revision, evidence); err != nil {
			problems = append(problems, err.Error())
			continue
		}
		if existing, dup := seenKeys[signer.KeyID]; dup && existing != role {
			problems = append(problems, fmt.Sprintf("roles %q and %q share the same signing key", existing, role))
			continue
		}
		seenKeys[signer.KeyID] = role
		result = append(result, Signer{Role: role, KeyID: signer.KeyID, PublicKey: signer.PublicKey, Signature: signer.Signature, PayloadSHA256: signer.PayloadSHA256})
	}
	for _, role := range required {
		if _, ok := signers[role]; !ok {
			problems = append(problems, fmt.Sprintf("role signer %q is missing", role))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Role < result[j].Role })
	if len(problems) > 0 {
		return result, fmt.Errorf("role signer verification failed: %s", strings.Join(problems, "; "))
	}
	return result, nil
}

// receiptPayload is the canonical signed form of a Receipt. It intentionally
// omits the four signature fields so the payload never self-references. Field
// order is fixed so the canonical bytes are deterministic.
type receiptPayload struct {
	SchemaVersion  string     `json:"schema_version"`
	Repository     string     `json:"repository"`
	Revision       string     `json:"revision"`
	PolicyRevision string     `json:"policy_revision"`
	CreatedAt      time.Time  `json:"created_at"`
	Complete       bool       `json:"complete"`
	Artifacts      []Artifact `json:"artifacts"`
	Gaps           []string   `json:"gaps"`
	Exceptions     []string   `json:"exceptions"`
	Signers        []Signer   `json:"signers,omitempty"`
}

func (receipt Receipt) payload() receiptPayload {
	return receiptPayload{
		SchemaVersion:  receipt.SchemaVersion,
		Repository:     receipt.Repository,
		Revision:       receipt.Revision,
		PolicyRevision: receipt.PolicyRevision,
		CreatedAt:      receipt.CreatedAt,
		Complete:       receipt.Complete,
		Artifacts:      append([]Artifact{}, receipt.Artifacts...),
		Gaps:           append([]string{}, receipt.Gaps...),
		Exceptions:     append([]string{}, receipt.Exceptions...),
		Signers:        append([]Signer{}, receipt.Signers...),
	}
}

func canonicalPayload(receipt Receipt) ([]byte, error) {
	return json.Marshal(receipt.payload())
}

// SignReceipt signs a receipt with an Ed25519 private-key seed and returns a new
// receipt with the four signature fields populated. The input receipt is never
// mutated. Ed25519 is deterministic, so signing the same receipt with the same
// seed always yields identical signature fields.
func SignReceipt(receipt Receipt, seed []byte) (Receipt, error) {
	if len(seed) != ed25519.SeedSize {
		return Receipt{}, fmt.Errorf("sign receipt: private key seed must be %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	payload, err := canonicalPayload(receipt)
	if err != nil {
		return Receipt{}, fmt.Errorf("sign receipt: canonicalize payload: %w", err)
	}
	digest := sha256.Sum256(payload)
	signature := ed25519.Sign(privateKey, payload)

	signed := receipt
	fingerprint := sha256.Sum256(publicKey)
	auditorKeyID := hex.EncodeToString(fingerprint[:])
	for _, signer := range receipt.Signers {
		if signer.KeyID == auditorKeyID {
			return Receipt{}, fmt.Errorf("sign receipt: the auditor signing key must differ from the %s role signer key", signer.Role)
		}
	}
	signed.SignedPayloadSHA256 = hex.EncodeToString(digest[:])
	signed.Signature = base64.StdEncoding.EncodeToString(signature)
	signed.SigningPublicKey = base64.StdEncoding.EncodeToString(publicKey)
	signed.SigningKeyFingerprint = auditorKeyID
	return signed, nil
}

// VerifyReceipt verifies a receipt's embedded signature and metadata. It is
// self-contained: it trusts only the embedded public key. An unsigned or
// incomplete receipt fails.
func VerifyReceipt(receipt Receipt) error {
	if receipt.Signature == "" {
		return errors.New("verify receipt: receipt is not signed")
	}
	if receipt.SigningPublicKey == "" || receipt.SigningKeyFingerprint == "" || receipt.SignedPayloadSHA256 == "" {
		return errors.New("verify receipt: signature metadata is incomplete")
	}
	publicKey, err := base64.StdEncoding.DecodeString(receipt.SigningPublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return errors.New("verify receipt: invalid embedded public key")
	}
	fingerprint := sha256.Sum256(publicKey)
	if hex.EncodeToString(fingerprint[:]) != receipt.SigningKeyFingerprint {
		return errors.New("verify receipt: signing key fingerprint mismatch")
	}
	payload, err := canonicalPayload(receipt)
	if err != nil {
		return fmt.Errorf("verify receipt: canonicalize payload: %w", err)
	}
	digest := sha256.Sum256(payload)
	if hex.EncodeToString(digest[:]) != receipt.SignedPayloadSHA256 {
		return errors.New("verify receipt: signed payload digest mismatch")
	}
	signature, err := base64.StdEncoding.DecodeString(receipt.Signature)
	if err != nil {
		return errors.New("verify receipt: invalid signature encoding")
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), payload, signature) {
		return errors.New("verify receipt: signature verification failed")
	}
	return nil
}
