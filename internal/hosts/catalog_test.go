package hosts

import "testing"

func TestResolveKnownHost(t *testing.T) {
	capabilities := Resolve("codex")

	if capabilities.ID != "codex" {
		t.Fatalf("expected codex adapter, got %q", capabilities.ID)
	}
	if !capabilities.NativeSkills {
		t.Fatal("expected codex adapter to support native skills")
	}
}

func TestResolveUnknownHostUsesGenericFallback(t *testing.T) {
	capabilities := Resolve("future-ide")

	if capabilities.ID != "generic" {
		t.Fatalf("expected generic fallback, got %q", capabilities.ID)
	}
	if capabilities.NativeSkills || capabilities.Subagents || capabilities.Hooks {
		t.Fatal("generic fallback must not claim native host features")
	}
	if !capabilities.CLI {
		t.Fatal("generic fallback must expose CLI support")
	}
}

func TestCatalogHasUniqueHostIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, capabilities := range Catalog() {
		if seen[capabilities.ID] {
			t.Fatalf("duplicate host id %q", capabilities.ID)
		}
		seen[capabilities.ID] = true
	}
}
