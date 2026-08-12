package runner

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"clean-code/internal/contracts"
)

const defaultTimeout = 5 * time.Minute
const defaultMaxOutputBytes = 1 << 20

var redactedAssignment = regexp.MustCompile(`(?i)(token|secret|password|api[_-]?key|credential)(\s*[:=]\s*)[^\s,;]+`)

type Runner struct {
	Root           string
	MaxOutputBytes int
	Now            func() time.Time
}

func (runner Runner) Run(parent context.Context, spec contracts.CommandSpec, revision string) contracts.CheckResult {
	now := runner.Now
	if now == nil {
		now = time.Now
	}
	started := now().UTC()
	category := spec.Category
	if category == "" {
		category = "command"
	}
	result := contracts.CheckResult{
		SchemaVersion: "1.0.0",
		CheckID:       spec.ID,
		Category:      category,
		Provider:      "generic-command",
		Status:        contracts.StatusError,
		Required:      spec.Required,
		Revision:      revision,
		StartedAt:     started,
	}
	finish := func() contracts.CheckResult {
		result.DurationMS = max(0, now().UTC().Sub(started).Milliseconds())
		return result
	}

	if err := parent.Err(); err != nil {
		result.Status = contracts.StatusNotRun
		result.Evidence.Summary = "command was cancelled before execution"
		return finish()
	}
	if err := spec.Validate(); err != nil {
		result.Evidence.Summary = "invalid command: " + err.Error()
		return finish()
	}

	workingDirectory, err := resolveWorkingDirectory(runner.Root, spec.WorkingDir)
	if err != nil {
		result.Evidence.Summary = err.Error()
		return finish()
	}
	artifactBefore := make(map[string]artifactState, len(spec.Artifacts))
	for _, artifact := range spec.Artifacts {
		state, stateErr := inspectArtifact(runner.Root, artifact.Path, false)
		if stateErr != nil {
			result.Evidence.Summary = stateErr.Error()
			return finish()
		}
		artifactBefore[artifact.Path] = state
	}
	executable, err := exec.LookPath(spec.Executable)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			result.Status = contracts.StatusNotAvailable
			result.Evidence.Summary = fmt.Sprintf("executable %q is not available", spec.Executable)
			return finish()
		}
		result.Evidence.Summary = "resolve executable: " + err.Error()
		return finish()
	}

	timeout := defaultTimeout
	if spec.TimeoutSec > 0 {
		timeout = time.Duration(spec.TimeoutSec) * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	outputLimit := runner.MaxOutputBytes
	if spec.MaxOutputBytes > 0 {
		outputLimit = spec.MaxOutputBytes
	}
	if outputLimit <= 0 {
		outputLimit = defaultMaxOutputBytes
	}
	output := newBoundedBuffer(outputLimit)

	command := exec.CommandContext(ctx, executable, spec.Args...)
	command.Dir = workingDirectory
	command.Env = filteredEnvironment(spec.Env)
	command.Stdout = output
	command.Stderr = output
	command.WaitDelay = 2 * time.Second

	err = command.Run()
	summary := redact(output.String(), spec.Env)
	result.Evidence.Summary = summary
	if output.Truncated() {
		result.Evidence.Warnings = append(result.Evidence.Warnings, "output truncated")
	}
	if ctx.Err() != nil {
		result.Status = contracts.StatusError
		result.Evidence.Warnings = append(result.Evidence.Warnings, "command timed out or was cancelled")
		if result.Evidence.Summary == "" {
			result.Evidence.Summary = ctx.Err().Error()
		}
		return finish()
	}

	exitCode := 0
	if command.ProcessState != nil {
		exitCode = command.ProcessState.ExitCode()
		result.ExitCode = &exitCode
	}
	if err == nil || (command.ProcessState != nil && exitAllowed(exitCode, spec.AllowedExitCodes)) {
		result.Status = contracts.StatusPass
		if artifactErr := validateArtifacts(runner.Root, spec.Artifacts, artifactBefore, &result); artifactErr != nil {
			result.Status = contracts.StatusFail
			result.Evidence.Warnings = append(result.Evidence.Warnings, artifactErr.Error())
		}
		return finish()
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		result.Status = contracts.StatusFail
		if result.Evidence.Summary == "" {
			result.Evidence.Summary = exitError.Error()
		}
		return finish()
	}
	result.Status = contracts.StatusError
	if result.Evidence.Summary == "" {
		result.Evidence.Summary = err.Error()
	}
	return finish()
}

type artifactState struct {
	exists bool
	hash   [sha256.Size]byte
}

func validateArtifacts(root string, specs []contracts.ArtifactSpec, before map[string]artifactState, result *contracts.CheckResult) error {
	for _, spec := range specs {
		after, err := inspectArtifact(root, spec.Path, spec.Format == "json")
		if err != nil {
			if spec.Required {
				return err
			}
			result.Evidence.Warnings = append(result.Evidence.Warnings, err.Error())
			continue
		}
		if !after.exists {
			if spec.Required {
				return fmt.Errorf("required artifact %q was not produced", spec.Path)
			}
			continue
		}
		if spec.Fresh {
			previous := before[spec.Path]
			if previous.exists && previous.hash == after.hash {
				return fmt.Errorf("artifact %q was not refreshed", spec.Path)
			}
		}
		result.Evidence.Artifacts = append(result.Evidence.Artifacts, filepath.ToSlash(spec.Path))
	}
	return nil
}

func inspectArtifact(root, relative string, validateJSON bool) (artifactState, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return artifactState{}, fmt.Errorf("artifact %q: resolve repository root: %w", relative, err)
	}
	realRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return artifactState{}, fmt.Errorf("artifact %q: resolve repository root: %w", relative, err)
	}
	path := filepath.Join(realRoot, filepath.Clean(relative))
	rel, err := filepath.Rel(realRoot, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return artifactState{}, fmt.Errorf("artifact %q escapes repository root", relative)
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return artifactState{}, nil
	}
	if err != nil {
		return artifactState{}, fmt.Errorf("artifact %q: inspect: %w", relative, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return artifactState{}, fmt.Errorf("artifact %q must be a regular file", relative)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return artifactState{}, fmt.Errorf("artifact %q: resolve path: %w", relative, err)
	}
	resolvedRelative, err := filepath.Rel(realRoot, resolved)
	if err != nil || resolvedRelative == ".." || strings.HasPrefix(resolvedRelative, ".."+string(filepath.Separator)) {
		return artifactState{}, fmt.Errorf("artifact %q resolves outside repository root", relative)
	}
	content, err := os.ReadFile(resolved)
	if err != nil {
		return artifactState{}, fmt.Errorf("artifact %q: read: %w", relative, err)
	}
	if validateJSON && !json.Valid(content) {
		return artifactState{}, fmt.Errorf("artifact %q is not valid JSON", relative)
	}
	return artifactState{exists: true, hash: sha256.Sum256(content)}, nil
}

func resolveWorkingDirectory(root, configured string) (string, error) {
	if root == "" {
		return "", errors.New("working directory: repository root is required")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("working directory: resolve repository root: %w", err)
	}
	realRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", fmt.Errorf("working directory: resolve repository root: %w", err)
	}
	workingDirectory := realRoot
	if configured != "" {
		if filepath.IsAbs(configured) {
			return "", errors.New("working directory: absolute paths are not allowed")
		}
		workingDirectory, err = filepath.EvalSymlinks(filepath.Join(realRoot, filepath.Clean(configured)))
		if err != nil {
			return "", fmt.Errorf("working directory: resolve configured path: %w", err)
		}
	}
	relative, err := filepath.Rel(realRoot, workingDirectory)
	if err != nil {
		return "", fmt.Errorf("working directory: compare path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("working directory: configured path escapes repository root")
	}
	info, err := os.Stat(workingDirectory)
	if err != nil {
		return "", fmt.Errorf("working directory: inspect configured path: %w", err)
	}
	if !info.IsDir() {
		return "", errors.New("working directory: configured path is not a directory")
	}
	return workingDirectory, nil
}

func exitAllowed(exitCode int, configured []int) bool {
	if len(configured) == 0 {
		return exitCode == 0
	}
	for _, allowed := range configured {
		if exitCode == allowed {
			return true
		}
	}
	return false
}

func filteredEnvironment(configured map[string]string) []string {
	allowed := []string{"CI", "COMSPEC", "HOME", "LANG", "LC_ALL", "LOGNAME", "PATH", "PATHEXT", "SYSTEMROOT", "TEMP", "TERM", "TMP", "TMPDIR", "USER", "WINDIR"}
	values := map[string]string{}
	for _, key := range allowed {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
	}
	for key, value := range configured {
		values[key] = value
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}

func redact(output string, configured map[string]string) string {
	for key, value := range configured {
		if secretKey(key) && len(value) >= 3 {
			output = strings.ReplaceAll(output, value, "[REDACTED]")
		}
	}
	return redactedAssignment.ReplaceAllString(output, "$1$2[REDACTED]")
}

func secretKey(key string) bool {
	normalized := strings.ToUpper(key)
	for _, marker := range []string{"TOKEN", "SECRET", "PASSWORD", "API_KEY", "APIKEY", "PRIVATE_KEY", "CREDENTIAL"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

type boundedBuffer struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func newBoundedBuffer(limit int) *boundedBuffer {
	return &boundedBuffer{limit: limit}
}

func (buffer *boundedBuffer) Write(data []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	remaining := buffer.limit - buffer.buffer.Len()
	if remaining > 0 {
		writeLength := min(remaining, len(data))
		_, _ = buffer.buffer.Write(data[:writeLength])
	}
	if len(data) > remaining {
		buffer.truncated = true
	}
	return len(data), nil
}

func (buffer *boundedBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buffer.String()
}

func (buffer *boundedBuffer) Truncated() bool {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.truncated
}
