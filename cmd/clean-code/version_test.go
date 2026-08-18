package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveVersionFromPackageJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"version":"9.9.9"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(root, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(binDir, "clean-code")
	if err := os.WriteFile(exe, []byte("stub"), 0o600); err != nil {
		t.Fatal(err)
	}

	original := version
	version = "0.1.0-dev"
	t.Cleanup(func() { version = original })

	got := resolveVersionForExecutable(exe)
	if got != "9.9.9" {
		t.Fatalf("resolveVersionForExecutable() = %q, want 9.9.9", got)
	}
}

func TestResolveVersionUsesLdflags(t *testing.T) {
	original := version
	version = "1.2.3"
	t.Cleanup(func() { version = original })

	if got := resolveVersion(); got != "1.2.3" {
		t.Fatalf("resolveVersion() = %q, want 1.2.3", got)
	}
}
