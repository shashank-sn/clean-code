package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"clean-code/internal/audit"
)

func TestAuditFixtureProducesCompleteImmutableReceipt(t *testing.T) {
	input := filepath.Join("..", "fixtures", "audit", "audit-input.json")
	receipt, err := audit.Build(input, func() time.Time { return time.Unix(10, 0).UTC() })
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Complete || len(receipt.Artifacts) != 7 || len(receipt.Gaps) != 0 {
		t.Fatalf("unexpected receipt: %+v", receipt)
	}
	output := filepath.Join(t.TempDir(), "receipt.json")
	if err := audit.Write(output, receipt); err != nil {
		t.Fatal(err)
	}
	if err := audit.Write(output, receipt); err == nil {
		t.Fatal("expected immutable output rejection")
	}
}
