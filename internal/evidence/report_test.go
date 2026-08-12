package evidence

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"clean-code/internal/contracts"
)

func TestSuccessfulRequiresCompleteRequiredPassingResults(t *testing.T) {
	report := Report{Complete: true, Results: []contracts.CheckResult{{Required: true, Status: contracts.StatusPass}}}
	if !report.Successful() {
		t.Fatal("expected successful report")
	}
	report.Results[0].Status = contracts.StatusFail
	if report.Successful() {
		t.Fatal("required failure must block")
	}
	report.Results[0].Required = false
	if !report.Successful() {
		t.Fatal("optional check failure should remain non-blocking")
	}
	report.PolicyDeltas = []string{"changed"}
	if report.Successful() {
		t.Fatal("policy delta must block")
	}
}

func TestWriteBundleUsesProtectedAtomicFile(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "evidence")
	path, err := WriteBundle(directory, Report{SchemaVersion: "1.0.0", Repository: "repo", Results: []contracts.CheckResult{}})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(directory, "report.json") {
		t.Fatalf("unexpected path %q", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600, got %o", info.Mode().Perm())
	}
}

func TestWriteBundleRejectsSymlinkDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	root := t.TempDir()
	realDirectory := filepath.Join(root, "real")
	if err := os.Mkdir(realDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(realDirectory, link); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteBundle(link, Report{}); err == nil {
		t.Fatal("expected symlink rejection")
	}
}
