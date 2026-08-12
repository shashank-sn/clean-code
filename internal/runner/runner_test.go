package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"clean-code/internal/contracts"
)

func TestRunPassesAllowedExitCode(t *testing.T) {
	runner := newTestRunner(t)
	result := runner.Run(context.Background(), helperCommand("pass"), "abc123")

	if result.Status != contracts.StatusPass {
		t.Fatalf("expected PASS, got %s: %s", result.Status, result.Evidence.Summary)
	}
	if result.ExitCode == nil || *result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %v", result.ExitCode)
	}
	if !strings.Contains(result.Evidence.Summary, "helper passed") {
		t.Fatalf("expected helper output, got %q", result.Evidence.Summary)
	}
}

func TestRunMapsNonzeroExitToFail(t *testing.T) {
	runner := newTestRunner(t)
	result := runner.Run(context.Background(), helperCommand("fail"), "abc123")

	if result.Status != contracts.StatusFail {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}
	if result.ExitCode == nil || *result.ExitCode != 3 {
		t.Fatalf("expected exit code 3, got %v", result.ExitCode)
	}
}

func TestRunAcceptsConfiguredNonzeroExitCode(t *testing.T) {
	runner := newTestRunner(t)
	spec := helperCommand("fail")
	spec.AllowedExitCodes = []int{0, 3}
	result := runner.Run(context.Background(), spec, "abc123")

	if result.Status != contracts.StatusPass {
		t.Fatalf("expected configured exit code to pass, got %s", result.Status)
	}
}

func TestRunMapsMissingExecutableToNotAvailable(t *testing.T) {
	runner := newTestRunner(t)
	spec := contracts.CommandSpec{ID: "missing", Executable: "clean-code-command-that-does-not-exist"}
	result := runner.Run(context.Background(), spec, "abc123")

	if result.Status != contracts.StatusNotAvailable {
		t.Fatalf("expected NOT_AVAILABLE, got %s", result.Status)
	}
}

func TestRunHonorsCancellation(t *testing.T) {
	runner := newTestRunner(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	result := runner.Run(ctx, helperCommand("sleep"), "abc123")

	if result.Status != contracts.StatusError {
		t.Fatalf("expected ERROR, got %s", result.Status)
	}
	if !contains(result.Evidence.Warnings, "command timed out or was cancelled") {
		t.Fatalf("expected cancellation warning, got %v", result.Evidence.Warnings)
	}
}

func TestRunRedactsSecretsFromOutput(t *testing.T) {
	runner := newTestRunner(t)
	spec := helperCommand("secret")
	spec.Env["SERVICE_API_TOKEN"] = "super-secret-value"
	result := runner.Run(context.Background(), spec, "abc123")

	if strings.Contains(result.Evidence.Summary, "super-secret-value") {
		t.Fatalf("secret leaked in output: %q", result.Evidence.Summary)
	}
	if !strings.Contains(result.Evidence.Summary, "[REDACTED]") {
		t.Fatalf("expected redaction marker, got %q", result.Evidence.Summary)
	}
}

func TestRunBoundsCapturedOutput(t *testing.T) {
	runner := newTestRunner(t)
	runner.MaxOutputBytes = 32
	result := runner.Run(context.Background(), helperCommand("large"), "abc123")

	if len(result.Evidence.Summary) > 32 {
		t.Fatalf("expected bounded output, got %d bytes", len(result.Evidence.Summary))
	}
	if !contains(result.Evidence.Warnings, "output truncated") {
		t.Fatalf("expected truncation warning, got %v", result.Evidence.Warnings)
	}
}

func TestRunRejectsWorkingDirectoryEscape(t *testing.T) {
	runner := newTestRunner(t)
	spec := helperCommand("pass")
	spec.WorkingDir = ".."
	result := runner.Run(context.Background(), spec, "abc123")

	if result.Status != contracts.StatusError {
		t.Fatalf("expected ERROR, got %s", result.Status)
	}
	if !strings.Contains(result.Evidence.Summary, "working directory") {
		t.Fatalf("expected path error, got %q", result.Evidence.Summary)
	}
}

func TestRunRejectsInvalidCommandBeforeExecution(t *testing.T) {
	runner := newTestRunner(t)
	spec := contracts.CommandSpec{ID: "unsafe", Executable: "sh", Args: []string{"-c", "echo unsafe"}}
	result := runner.Run(context.Background(), spec, "abc123")

	if result.Status != contracts.StatusError {
		t.Fatalf("expected ERROR, got %s", result.Status)
	}
	if result.ExitCode != nil {
		t.Fatalf("invalid command should not execute, got exit code %v", result.ExitCode)
	}
}

func TestRunMapsProcessStartFailureToError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable format behavior differs on Windows")
	}
	runner := newTestRunner(t)
	path := filepath.Join(runner.Root, "invalid-executable")
	if err := os.WriteFile(path, []byte("not an executable format"), 0o755); err != nil {
		t.Fatal(err)
	}
	result := runner.Run(context.Background(), contracts.CommandSpec{ID: "invalid", Executable: path}, "abc123")

	if result.Status != contracts.StatusError {
		t.Fatalf("expected ERROR, got %s", result.Status)
	}
	if result.ExitCode != nil {
		t.Fatalf("start failure should not have exit code, got %v", result.ExitCode)
	}
}

func TestRunFailsWhenRequiredArtifactIsMissing(t *testing.T) {
	runner := newTestRunner(t)
	spec := helperCommand("pass")
	spec.Artifacts = []contracts.ArtifactSpec{{Path: "result.json", Format: "json", Required: true, Fresh: true}}
	result := runner.Run(context.Background(), spec, "abc123")

	if result.Status != contracts.StatusFail || len(result.Evidence.Warnings) == 0 {
		t.Fatalf("expected missing artifact failure, got %+v", result)
	}
}

func TestRunAcceptsFreshJSONArtifact(t *testing.T) {
	runner := newTestRunner(t)
	path := filepath.Join(runner.Root, "result.json")
	if err := os.WriteFile(path, []byte(`{"old":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := helperCommand("write-json")
	spec.Env["CLEAN_CODE_TEST_ARTIFACT"] = path
	spec.Artifacts = []contracts.ArtifactSpec{{Path: "result.json", Format: "json", Required: true, Fresh: true}}
	result := runner.Run(context.Background(), spec, "abc123")

	if result.Status != contracts.StatusPass || len(result.Evidence.Artifacts) != 1 {
		t.Fatalf("expected artifact pass, got %+v", result)
	}
}

func TestRunRejectsStaleArtifact(t *testing.T) {
	runner := newTestRunner(t)
	if err := os.WriteFile(filepath.Join(runner.Root, "result.json"), []byte(`{"old":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := helperCommand("pass")
	spec.Artifacts = []contracts.ArtifactSpec{{Path: "result.json", Format: "json", Required: true, Fresh: true}}
	result := runner.Run(context.Background(), spec, "abc123")

	if result.Status != contracts.StatusFail {
		t.Fatalf("expected stale artifact failure, got %+v", result)
	}
}

func newTestRunner(t *testing.T) Runner {
	t.Helper()
	return Runner{Root: t.TempDir(), MaxOutputBytes: 1024}
}

func helperCommand(mode string) contracts.CommandSpec {
	return contracts.CommandSpec{
		ID:         "helper-" + mode,
		Category:   "test",
		Executable: os.Args[0],
		Args:       []string{"-test.run=^TestRunnerHelperProcess$"},
		Env: map[string]string{
			"CLEAN_CODE_TEST_HELPER": "1",
			"CLEAN_CODE_TEST_MODE":   mode,
		},
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestRunnerHelperProcess(t *testing.T) {
	if os.Getenv("CLEAN_CODE_TEST_HELPER") != "1" {
		return
	}
	switch os.Getenv("CLEAN_CODE_TEST_MODE") {
	case "pass":
		fmt.Print("helper passed")
	case "fail":
		fmt.Fprint(os.Stderr, "helper failed")
		os.Exit(3)
	case "sleep":
		time.Sleep(5 * time.Second)
	case "secret":
		fmt.Printf("token=%s", os.Getenv("SERVICE_API_TOKEN"))
	case "large":
		fmt.Print(strings.Repeat("x", 4096))
	case "write-json":
		if err := os.WriteFile(os.Getenv("CLEAN_CODE_TEST_ARTIFACT"), []byte(`{"ok":true}`), 0o600); err != nil {
			os.Exit(4)
		}
	default:
		panic(errors.New("unknown helper mode"))
	}
	os.Exit(0)
}
