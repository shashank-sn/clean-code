package providers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clean-code/internal/contracts"
)

func TestLoadProviderFixtures(t *testing.T) {
	for _, category := range []string{"complexity", "mutation", "acceptance", "browser-qa", "dependency-graph", "architecture-render"} {
		spec, err := Load(filepath.Join("..", "..", FixturePath(category)))
		if err != nil {
			t.Errorf("load %s fixture: %v", category, err)
			continue
		}
		if spec.Category != category || spec.Command.Category != category {
			t.Errorf("unexpected fixture contract for %s: %#v", category, spec)
		}
	}
}

func TestBrowserQAContractDeclaresInspectableArtifacts(t *testing.T) {
	spec, err := Load(filepath.Join("..", "..", FixturePath("browser-qa")))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"screenshot": false, "trace": false, "log": false, "accessibility": false, "flake-classification": false}
	for _, artifact := range spec.Output.Artifacts {
		if _, ok := want[artifact.ID]; ok {
			want[artifact.ID] = artifact.Required
		}
	}
	for id, found := range want {
		if !found {
			t.Errorf("browser provider is missing required %s artifact", id)
		}
	}
}

func TestLoadProviderRejectsUnknownAndUnsafeFields(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "provider.json")
	fixture, err := os.ReadFile(filepath.Join("..", "..", FixturePath("mutation")))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(fixture, &value); err != nil {
		t.Fatal(err)
	}
	value["unknown"] = true
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown provider property to fail")
	}
	value["policy"].(map[string]any)["network_access"] = true
	delete(value, "unknown")
	encoded, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected network provider to fail")
	}
}

func TestResultRequiresRevisionBoundArtifactEvidence(t *testing.T) {
	result := Result{
		SchemaVersion: providerSchemaVersion, ProviderID: "fixture-mutation", ProviderVersion: "fixture-1", Category: "mutation",
		Revision: "abc123", Status: contracts.StatusPass, StartedAt: time.Now(), Evidence: contracts.Evidence{ArtifactDetails: []contracts.ArtifactProvenance{{
			Path: "reports/mutation.json", SHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			Schema: "clean-code/mutation-result/v1", Revision: "other", Fresh: true,
		}}},
	}
	if err := result.Validate(); err == nil {
		t.Fatal("expected mismatched artifact revision to fail")
	}
	result.Evidence.ArtifactDetails[0].Revision = result.Revision
	result.Status = contracts.StatusStale
	if err := result.Validate(); err != nil {
		t.Fatalf("expected revision-bound stale result to validate: %v", err)
	}
}

func TestPassResultRequiresArtifactEvidence(t *testing.T) {
	result := Result{SchemaVersion: providerSchemaVersion, ProviderID: "fixture-mutation", ProviderVersion: "fixture-1", Category: "mutation", Revision: "abc123", Status: contracts.StatusPass, StartedAt: time.Now()}
	if err := result.Validate(); err == nil {
		t.Fatal("expected pass without artifact evidence to fail")
	}
}

func TestParseResultRejectsUnboundAndMalformedOutput(t *testing.T) {
	valid := []byte(`{
  "schema_version":"1.0.0", "provider_id":"fixture-mutation", "provider_version":"fixture-1",
  "category":"mutation", "revision":"abc123", "status":"NOT_AVAILABLE",
  "started_at":"2026-08-21T00:00:00Z", "duration_ms":0
}`)
	result, err := ParseResult(valid)
	if err != nil {
		t.Fatalf("parse valid result: %v", err)
	}
	if result.Status != contracts.StatusNotAvailable {
		t.Fatalf("unexpected result status: %s", result.Status)
	}
	for _, invalid := range [][]byte{
		[]byte(`{"schema_version":"1.0.0"}`),
		append(valid, []byte(` {}`)...),
		[]byte(`{"schema_version":"1.0.0","provider_id":"fixture-mutation","provider_version":"1","category":"mutation","status":"PASS","started_at":"2026-08-21T00:00:00Z","duration_ms":0}`),
	} {
		if _, err := ParseResult(invalid); err == nil {
			t.Errorf("expected invalid provider output to fail: %s", invalid)
		}
	}
}

func TestProviderSpecRejectsShellMode(t *testing.T) {
	spec := Spec{
		SchemaVersion: providerSchemaVersion, ID: "unsafe-provider", Category: "mutation", Version: "1", Languages: []string{"go"},
		Command: contracts.CommandSpec{ID: "unsafe-provider", Executable: "sh", Args: []string{"-c", "echo unsafe"}},
		Input:   ArtifactContract{Format: "json", Schema: "input/v1"}, Output: ArtifactContract{Format: "json", Schema: "output/v1"},
		Availability: Availability{Mode: "optional", Cost: "none"}, Policy: PolicyRequirements{RequiresTrustedPolicy: true},
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("expected shell provider to fail")
	}
}
