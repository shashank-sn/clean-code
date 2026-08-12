package tests_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPluginManifestPointsToAllSkills(t *testing.T) {
	root := filepath.Join("..")
	body, err := os.ReadFile(filepath.Join(root, ".codex-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Name   string `json:"name"`
		Skills string `json:"skills"`
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "clean-code" || manifest.Skills != "./skills/" {
		t.Fatalf("unexpected plugin manifest: %+v", manifest)
	}
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 11 {
		t.Fatalf("expected eleven skills, got %d", len(entries))
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(root, "skills", entry.Name(), "SKILL.md")); err != nil {
				t.Errorf("missing SKILL.md for %s", entry.Name())
			}
		}
	}
}

func TestRelativeMarkdownLinksResolve(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	linkPattern := regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "dist") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(body), -1) {
			target := strings.Split(match[1], "#")[0]
			if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), filepath.FromSlash(target))); err != nil {
				t.Errorf("%s has broken link %q", path, match[1])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckedInJSONExamplesParse(t *testing.T) {
	root := filepath.Join("..")
	for _, directory := range []string{"tests/fixtures", "harness/config"} {
		err := filepath.WalkDir(filepath.Join(root, directory), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				return nil
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var value any
			if err := json.Unmarshal(body, &value); err != nil {
				t.Errorf("parse %s: %v", path, err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
