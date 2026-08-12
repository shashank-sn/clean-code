package contracts

import (
	"testing"
	"time"
)

func TestCheckResultValidateAcceptsCompleteResult(t *testing.T) {
	result := CheckResult{
		SchemaVersion: "1.0.0",
		CheckID:       "repo.test",
		Category:      "test",
		Provider:      "generic-command",
		Status:        StatusPass,
		StartedAt:     time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC),
		DurationMS:    12,
		Revision:      "abc123",
	}

	if err := result.Validate(); err != nil {
		t.Fatalf("expected valid result, got %v", err)
	}
}

func TestCheckResultValidateRejectsUnknownStatus(t *testing.T) {
	result := CheckResult{
		SchemaVersion: "1.0.0",
		CheckID:       "repo.test",
		Category:      "test",
		Provider:      "generic-command",
		Status:        Status("MAYBE"),
		StartedAt:     time.Now(),
	}

	if err := result.Validate(); err == nil {
		t.Fatal("expected unknown status to fail validation")
	}
}

func TestCommandSpecRejectsImplicitShell(t *testing.T) {
	spec := CommandSpec{ID: "lint", Executable: "npm && echo unsafe"}

	if err := spec.Validate(); err == nil {
		t.Fatal("expected executable containing shell syntax to fail validation")
	}
}

func TestCommandSpecAcceptsExecutableAndArguments(t *testing.T) {
	spec := CommandSpec{ID: "lint", Executable: "npm", Args: []string{"run", "lint"}}

	if err := spec.Validate(); err != nil {
		t.Fatalf("expected command array to validate, got %v", err)
	}
}

func TestCommandSpecRejectsShellFlagUntilPolicyApprovalExists(t *testing.T) {
	spec := CommandSpec{ID: "lint", Executable: "npm", Args: []string{"run", "lint"}, Shell: true}

	if err := spec.Validate(); err == nil {
		t.Fatal("expected shell mode to fail before policy approval support exists")
	}
}

func TestCommandSpecRejectsShellExecutableCommandMode(t *testing.T) {
	cases := []CommandSpec{
		{ID: "unix", Executable: "sh", Args: []string{"-c", "echo unsafe"}},
		{ID: "windows", Executable: "cmd.exe", Args: []string{"/c", "echo unsafe"}},
		{ID: "windows-path", Executable: `C:\Windows\System32\cmd.exe`, Args: []string{"/c", "echo unsafe"}},
		{ID: "powershell", Executable: "powershell", Args: []string{"-Command", "echo unsafe"}},
	}
	for _, spec := range cases {
		if err := spec.Validate(); err == nil {
			t.Errorf("expected %s command mode to fail", spec.ID)
		}
	}
}
