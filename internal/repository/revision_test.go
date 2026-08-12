package repository

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRevisionReportsUnknownOutsideGit(t *testing.T) {
	if got := Revision(t.TempDir()); got != "unknown" {
		t.Fatalf("expected unknown, got %q", got)
	}
}

func TestRevisionMarksDirtyWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test")
	path := filepath.Join(root, "tracked.txt")
	if err := os.WriteFile(path, []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "tracked.txt")
	runGit(t, root, "commit", "-qm", "initial")
	clean := Revision(root)
	if clean == "unknown" || strings.Contains(clean, "+dirty") {
		t.Fatalf("unexpected clean revision %q", clean)
	}
	if err := os.WriteFile(path, []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if dirty := Revision(root); dirty != clean+"+dirty" {
		t.Fatalf("expected %q, got %q", clean+"+dirty", dirty)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
