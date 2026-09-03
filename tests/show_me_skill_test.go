package tests_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowMeSkillKeepsVisualAndEvidenceContract(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "skills", "clean-show-me", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{
		"name: clean-show-me",
		"**Observed** or **Proposed**",
		"Pseudocode",
		"Call tree",
		"Component tree",
		"File tree",
		"```mermaid",
		"```diff",
		"focused HTML artifact",
		"A visual explains; it does not verify",
		"github.com/humanlayer/skills",
	} {
		if !strings.Contains(string(body), phrase) {
			t.Errorf("clean-show-me skill missing contract %q", phrase)
		}
	}
}
