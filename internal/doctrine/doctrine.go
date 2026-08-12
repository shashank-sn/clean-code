package doctrine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Rule struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Classification  string   `json:"classification"`
	Summary         string   `json:"summary"`
	Evidence        []string `json:"evidence"`
	Applicability   string   `json:"applicability"`
	DefaultSeverity string   `json:"default_severity"`
	Source          string   `json:"source"`
	Exceptions      []string `json:"exceptions"`
	FalsePositives  []string `json:"false_positives"`
}

var classifications = map[string]bool{
	"deterministic": true,
	"semantic":      true,
	"convention":    true,
	"architectural": true,
}

var severities = map[string]bool{"blocking": true, "review": true, "advisory": true}

func LoadDir(path string) ([]Rule, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read doctrine directory: %w", err)
	}
	rules := []Rule{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read doctrine file %s: %w", entry.Name(), err)
		}
		var fileRules []Rule
		if err := json.Unmarshal(body, &fileRules); err != nil {
			return nil, fmt.Errorf("parse doctrine file %s: %w", entry.Name(), err)
		}
		rules = append(rules, fileRules...)
	}
	if len(rules) == 0 {
		return nil, errors.New("no doctrine rules found")
	}
	if err := Validate(rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func Validate(rules []Rule) error {
	seen := map[string]bool{}
	for index, rule := range rules {
		if rule.ID == "" || rule.Title == "" || rule.Summary == "" || rule.Source == "" || rule.Applicability == "" {
			return fmt.Errorf("rule %d is missing required text", index)
		}
		if seen[rule.ID] {
			return fmt.Errorf("duplicate rule id %q", rule.ID)
		}
		seen[rule.ID] = true
		if !classifications[rule.Classification] {
			return fmt.Errorf("rule %s has unknown classification %q", rule.ID, rule.Classification)
		}
		if !severities[rule.DefaultSeverity] {
			return fmt.Errorf("rule %s has unknown severity %q", rule.ID, rule.DefaultSeverity)
		}
		if len(rule.Evidence) == 0 {
			return fmt.Errorf("rule %s has no evidence requirements", rule.ID)
		}
	}
	return nil
}
