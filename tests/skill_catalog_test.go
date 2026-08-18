package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillCatalog(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "clean-") {
			continue
		}
		skillPath := filepath.Join("..", "skills", entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err != nil {
			t.Fatalf("missing SKILL.md for %s: %v", entry.Name(), err)
		}
		skills = append(skills, entry.Name())
	}
	if len(skills) < 20 {
		t.Fatalf("expected at least 20 clean-* skills, got %d: %v", len(skills), skills)
	}
	required := []string{
		"clean-brainstorm", "clean-plan", "clean-debug", "clean-ship",
		"clean-simplify", "clean-compound", "clean-worktree", "clean-watch-pr", "clean-lfg",
	}
	for _, name := range required {
		found := false
		for _, skill := range skills {
			if skill == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing required skill %s", name)
		}
	}
}
