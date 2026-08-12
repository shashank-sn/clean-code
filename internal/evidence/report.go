package evidence

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"clean-code/internal/contracts"
)

type Report struct {
	SchemaVersion string                  `json:"schema_version"`
	Repository    string                  `json:"repository"`
	Revision      string                  `json:"revision"`
	PolicySource  string                  `json:"policy_source"`
	StartedAt     time.Time               `json:"started_at"`
	FinishedAt    time.Time               `json:"finished_at"`
	Complete      bool                    `json:"complete"`
	PolicyDeltas  []string                `json:"policy_deltas,omitempty"`
	Results       []contracts.CheckResult `json:"results"`
}

func (report Report) Successful() bool {
	if !report.Complete || len(report.Results) == 0 || len(report.PolicyDeltas) > 0 {
		return false
	}
	for _, result := range report.Results {
		if result.Status == contracts.StatusError || (result.Required && result.Status != contracts.StatusPass) {
			return false
		}
	}
	return true
}

// WriteBundle atomically writes report.json with owner-only permissions.
func WriteBundle(directory string, report Report) (string, error) {
	if directory == "" {
		return "", errors.New("evidence output directory is required")
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return "", fmt.Errorf("resolve evidence output: %w", err)
	}
	if info, err := os.Lstat(absolute); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return "", errors.New("evidence output must be a real directory")
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(absolute, 0o700); err != nil {
			return "", fmt.Errorf("create evidence output: %w", err)
		}
	} else {
		return "", fmt.Errorf("inspect evidence output: %w", err)
	}

	target := filepath.Join(absolute, "report.json")
	if info, err := os.Lstat(target); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("evidence report cannot replace a symbolic link")
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect evidence report: %w", err)
	}
	temporary, err := os.CreateTemp(absolute, ".report-*.json")
	if err != nil {
		return "", fmt.Errorf("create evidence report: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return "", fmt.Errorf("protect evidence report: %w", err)
	}
	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		temporary.Close()
		return "", fmt.Errorf("encode evidence report: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return "", fmt.Errorf("sync evidence report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close evidence report: %w", err)
	}
	if err := os.Rename(temporaryName, target); err != nil {
		return "", fmt.Errorf("publish evidence report: %w", err)
	}
	return target, nil
}
