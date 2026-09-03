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
	metadata, err := os.ReadFile(filepath.Join("..", "skills", "clean-show-me", "agents", "openai.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{"display_name: \"Clean Show Me\"", "short_description:", "default_prompt:", "$clean-show-me"} {
		if !strings.Contains(string(metadata), phrase) {
			t.Errorf("clean-show-me metadata missing %q", phrase)
		}
	}
}
