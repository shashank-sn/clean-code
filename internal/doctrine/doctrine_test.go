package doctrine

import (
	"path/filepath"
	"testing"
)

func TestRepositoryDoctrineLoadsAndHasUniqueValidRules(t *testing.T) {
	path := filepath.Join("..", "..", "harness", "doctrine")
	rules, err := LoadDir(path)
	if err != nil {
		t.Fatalf("load doctrine: %v", err)
	}
	if len(rules) < 10 {
		t.Fatalf("expected representative foundation, got %d rules", len(rules))
	}
}

func TestValidateRejectsDuplicateRuleID(t *testing.T) {
	rule := Rule{
		ID: "CC-TEST-1", Title: "test", Classification: "semantic",
		Summary: "summary", Evidence: []string{"evidence"}, Applicability: "all",
		DefaultSeverity: "review", Source: "source",
	}
	if err := Validate([]Rule{rule, rule}); err == nil {
		t.Fatal("expected duplicate rule id to fail")
	}
}
