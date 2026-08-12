package contracts

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"
)

type Status string

const (
	StatusPass          Status = "PASS"
	StatusFail          Status = "FAIL"
	StatusNotAvailable  Status = "NOT_AVAILABLE"
	StatusNotConfigured Status = "NOT_CONFIGURED"
	StatusNotRun        Status = "NOT_RUN"
	StatusError         Status = "ERROR"
)

var validStatuses = map[Status]struct{}{
	StatusPass: {}, StatusFail: {}, StatusNotAvailable: {},
	StatusNotConfigured: {}, StatusNotRun: {}, StatusError: {},
}

var validArtifactFormats = map[string]bool{
	"": true, "file": true, "json": true, "xml": true,
	"sarif": true, "lcov": true, "text": true,
}

type Evidence struct {
	Summary   string               `json:"summary,omitempty"`
	Artifacts []string             `json:"artifacts,omitempty"`
	Baselines []BaselineComparison `json:"baselines,omitempty"`
	Warnings  []string             `json:"warnings,omitempty"`
}

type BaselineComparison struct {
	Artifact  string  `json:"artifact"`
	Metric    string  `json:"metric"`
	Scope     string  `json:"scope"`
	Current   float64 `json:"current"`
	Baseline  float64 `json:"baseline"`
	Direction string  `json:"direction"`
	Tolerance float64 `json:"tolerance,omitempty"`
	Regressed bool    `json:"regressed"`
}

type CheckResult struct {
	SchemaVersion string    `json:"schema_version"`
	CheckID       string    `json:"check_id"`
	Category      string    `json:"category"`
	Provider      string    `json:"provider"`
	Status        Status    `json:"status"`
	Required      bool      `json:"required"`
	Revision      string    `json:"revision,omitempty"`
	StartedAt     time.Time `json:"started_at"`
	DurationMS    int64     `json:"duration_ms"`
	ExitCode      *int      `json:"exit_code,omitempty"`
	Evidence      Evidence  `json:"evidence,omitempty"`
}

func (result CheckResult) Validate() error {
	if result.SchemaVersion == "" {
		return errors.New("schema_version is required")
	}
	if result.CheckID == "" {
		return errors.New("check_id is required")
	}
	if result.Category == "" {
		return errors.New("category is required")
	}
	if result.Provider == "" {
		return errors.New("provider is required")
	}
	if _, ok := validStatuses[result.Status]; !ok {
		return fmt.Errorf("unknown status %q", result.Status)
	}
	if result.StartedAt.IsZero() {
		return errors.New("started_at is required")
	}
	if result.DurationMS < 0 {
		return errors.New("duration_ms cannot be negative")
	}
	for index, baseline := range result.Evidence.Baselines {
		if strings.TrimSpace(baseline.Artifact) == "" {
			return fmt.Errorf("evidence baseline %d artifact is required", index)
		}
		if strings.TrimSpace(baseline.Metric) == "" {
			return fmt.Errorf("evidence baseline %d metric is required", index)
		}
		if baseline.Direction != "higher" && baseline.Direction != "lower" {
			return fmt.Errorf("evidence baseline %d direction must be higher or lower", index)
		}
		if baseline.Scope != "repository" && baseline.Scope != "changed" {
			return fmt.Errorf("evidence baseline %d scope must be repository or changed", index)
		}
		if baseline.Tolerance < 0 {
			return fmt.Errorf("evidence baseline %d tolerance cannot be negative", index)
		}
		if !finite(baseline.Current) || !finite(baseline.Baseline) || !finite(baseline.Tolerance) {
			return fmt.Errorf("evidence baseline %d values must be finite", index)
		}
	}
	return nil
}

type CommandSpec struct {
	ID               string            `json:"id"`
	Category         string            `json:"category,omitempty"`
	Executable       string            `json:"executable"`
	Args             []string          `json:"args,omitempty"`
	WorkingDir       string            `json:"working_directory,omitempty"`
	TimeoutSec       int               `json:"timeout_seconds,omitempty"`
	MaxOutputBytes   int               `json:"max_output_bytes,omitempty"`
	AllowedExitCodes []int             `json:"allowed_exit_codes,omitempty"`
	Required         bool              `json:"required,omitempty"`
	Env              map[string]string `json:"environment,omitempty"`
	Shell            bool              `json:"shell,omitempty"`
	Artifacts        []ArtifactSpec    `json:"artifacts,omitempty"`
	Baselines        []BaselineSpec    `json:"baselines,omitempty"`
}

type ArtifactSpec struct {
	Path     string `json:"path"`
	Format   string `json:"format,omitempty"`
	Required bool   `json:"required,omitempty"`
	Fresh    bool   `json:"fresh,omitempty"`
	MaxBytes int64  `json:"max_bytes,omitempty"`
}

type BaselineSpec struct {
	Artifact  string  `json:"artifact"`
	Metric    string  `json:"metric"`
	Scope     string  `json:"scope,omitempty"`
	Direction string  `json:"direction"`
	Value     float64 `json:"value"`
	Tolerance float64 `json:"tolerance,omitempty"`
	Required  bool    `json:"required,omitempty"`
}

func (spec CommandSpec) Validate() error {
	if strings.TrimSpace(spec.ID) == "" {
		return errors.New("command id is required")
	}
	if strings.TrimSpace(spec.Executable) == "" {
		return errors.New("command executable is required")
	}
	if spec.Shell {
		return errors.New("shell mode requires policy approval support that is not implemented")
	}
	if strings.ContainsAny(spec.Executable, ";|&`$<>\n\r") {
		return errors.New("shell syntax is not allowed in executable")
	}
	if invokesShellCommandMode(spec.Executable, spec.Args) {
		return errors.New("shell command mode requires policy approval support that is not implemented")
	}
	if spec.TimeoutSec < 0 {
		return errors.New("timeout_seconds cannot be negative")
	}
	if spec.MaxOutputBytes < 0 {
		return errors.New("max_output_bytes cannot be negative")
	}
	for key := range spec.Env {
		if strings.TrimSpace(key) == "" || strings.ContainsAny(key, "=\x00") {
			return fmt.Errorf("invalid environment key %q", key)
		}
		if protectedEnvironmentKey(key) {
			return fmt.Errorf("environment key %q requires separate execution policy", key)
		}
	}
	seenArtifacts := map[string]bool{}
	for index, artifact := range spec.Artifacts {
		if strings.TrimSpace(artifact.Path) == "" {
			return fmt.Errorf("artifact %d path is required", index)
		}
		if filepath.IsAbs(artifact.Path) {
			return fmt.Errorf("artifact %d path must be repository-relative", index)
		}
		clean := filepath.Clean(artifact.Path)
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("artifact %d path escapes repository root", index)
		}
		if !validArtifactFormats[artifact.Format] {
			return fmt.Errorf("artifact %d has unsupported format %q", index, artifact.Format)
		}
		if artifact.MaxBytes < 0 {
			return fmt.Errorf("artifact %d max_bytes cannot be negative", index)
		}
		if seenArtifacts[clean] {
			return fmt.Errorf("artifact %d duplicates path %q", index, artifact.Path)
		}
		seenArtifacts[clean] = true
	}
	for index, baseline := range spec.Baselines {
		if !seenArtifacts[filepath.Clean(baseline.Artifact)] {
			return fmt.Errorf("baseline %d references undeclared artifact %q", index, baseline.Artifact)
		}
		if strings.TrimSpace(baseline.Metric) == "" {
			return fmt.Errorf("baseline %d metric is required", index)
		}
		if baseline.Scope != "" && baseline.Scope != "repository" && baseline.Scope != "changed" {
			return fmt.Errorf("baseline %d scope must be repository or changed", index)
		}
		if baseline.Direction != "higher" && baseline.Direction != "lower" {
			return fmt.Errorf("baseline %d direction must be higher or lower", index)
		}
		if baseline.Tolerance < 0 {
			return fmt.Errorf("baseline %d tolerance cannot be negative", index)
		}
		if !finite(baseline.Value) || !finite(baseline.Tolerance) {
			return fmt.Errorf("baseline %d values must be finite", index)
		}
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func protectedEnvironmentKey(key string) bool {
	normalized := strings.ToUpper(key)
	protected := map[string]bool{
		"BASH_ENV": true, "CDPATH": true, "COMSPEC": true, "ENV": true,
		"GODEBUG": true, "LD_LIBRARY_PATH": true, "LD_PRELOAD": true,
		"NODE_OPTIONS": true, "PATH": true, "PATHEXT": true,
		"PYTHONPATH": true, "RUBYOPT": true, "SHELLOPTS": true,
		"SYSTEMROOT": true, "WINDIR": true,
	}
	return protected[normalized] || strings.HasPrefix(normalized, "DYLD_")
}

func invokesShellCommandMode(executable string, args []string) bool {
	portablePath := strings.ReplaceAll(executable, `\`, "/")
	name := strings.ToLower(filepath.Base(portablePath))
	for _, arg := range args {
		switch name {
		case "sh", "bash", "zsh", "dash", "ksh":
			if arg == "-c" {
				return true
			}
		case "cmd", "cmd.exe":
			if strings.EqualFold(arg, "/c") {
				return true
			}
		case "powershell", "powershell.exe", "pwsh", "pwsh.exe":
			if strings.EqualFold(arg, "-command") || strings.EqualFold(arg, "-encodedcommand") {
				return true
			}
		}
	}
	return false
}
