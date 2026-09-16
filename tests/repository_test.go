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
	if len(entries) < 20 {
		t.Fatalf("expected at least twenty skills, got %d", len(entries))
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

func TestShipSkillRequiresPlainEnglishBulletPR(t *testing.T) {
	root := filepath.Join("..")
	body, err := os.ReadFile(filepath.Join(root, "skills", "clean-ship", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{
		"## PR description rules (required)",
		"simple human-understandable English",
		"**What changed**",
		"**Why**",
		"**How to verify**",
		"**Gaps**",
		"Do not claim checks passed without `clean-verify`",
		"## Self-check before opening the PR",
		"no sentence needs a second reading",
		"Explain a technical term the first time it appears",
		"Never open a PR whose body only the change author can follow.",
		"Confirm the change set was pruned (`clean-prune`)",
	} {
		if !strings.Contains(string(body), phrase) {
			t.Errorf("clean-ship skill missing PR description contract %q", phrase)
		}
	}
}

func TestPruneSkillKeepsDeletionEvidenceContract(t *testing.T) {
	root := filepath.Join("..")
	body, err := os.ReadFile(filepath.Join(root, "skills", "clean-prune", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{
		"name: clean-prune",
		"## Candidate categories",
		"## Gates",
		"Entries in string-keyed registries, handler maps, route tables, and dispatch tables",
		"Treat detector output as a candidate list, not proof.",
		"Block ship while an UNRESOLVED candidate remains in the change set.",
		"Block any removal that has no recorded reference search.",
		"Do not invent a dead-code score, threshold, or percentage.",
		"never delete on a name heuristic alone",
		"Keep behavior identical.",
	} {
		if !strings.Contains(string(body), phrase) {
			t.Errorf("clean-prune skill missing deletion-evidence contract %q", phrase)
		}
	}
	metadata, err := os.ReadFile(filepath.Join(root, "skills", "clean-prune", "agents", "openai.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{"display_name: \"Clean Prune\"", "short_description:", "default_prompt:", "$clean-prune"} {
		if !strings.Contains(string(metadata), phrase) {
			t.Errorf("clean-prune metadata missing %q", phrase)
		}
	}
}

func TestAdaptivePlaybooksAreDiscoverable(t *testing.T) {
	root := filepath.Join("..")
	entries, err := os.ReadDir(filepath.Join(root, "harness", "playbooks"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 6 {
		t.Fatalf("expected at least 6 playbooks, got %d", len(entries))
	}
}

func TestStructuralReviewSkillKeepsEvidenceBoundary(t *testing.T) {
	root := filepath.Join("..")
	body, err := os.ReadFile(filepath.Join(root, "skills", "clean-review", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{
		"## Structural-simplification lens",
		"non-trivial changed code",
		"changed-scope evidence",
		"1,000 lines",
		"never a universal size limit",
		"Do not turn an unconventional style, a metric alone, or an unproven preference into a finding.",
		"hand it to `clean-refactor`",
	} {
		if !strings.Contains(string(body), phrase) {
			t.Errorf("clean-review skill missing structural-review contract %q", phrase)
		}
	}
}

func TestPackageAndPluginVersionsMatch(t *testing.T) {
	root := filepath.Join("..")
	readVersion := func(path string) string {
		body, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		var manifest struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(body, &manifest); err != nil {
			t.Fatal(err)
		}
		return manifest.Version
	}
	if packageVersion, pluginVersion := readVersion("package.json"), readVersion(filepath.Join(".codex-plugin", "plugin.json")); packageVersion == "" || packageVersion != pluginVersion {
		t.Fatalf("package and plugin versions must match, got package=%q plugin=%q", packageVersion, pluginVersion)
	}
}

func TestShippingPipelineSimplifiesBeforeFinalVerificationAndReview(t *testing.T) {
	root := filepath.Join("..")
	body, err := os.ReadFile(filepath.Join(root, "harness", "workflow", "shipping-pipeline.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pipeline struct {
		Stages []struct {
			ID string `json:"id"`
		} `json:"stages"`
	}
	if err := json.Unmarshal(body, &pipeline); err != nil {
		t.Fatal(err)
	}
	positions := map[string]int{"simplify": -1, "verify": -1, "review": -1}
	for index, stage := range pipeline.Stages {
		positions[stage.ID] = index
	}
	if positions["simplify"] == -1 || positions["verify"] == -1 || positions["review"] == -1 || positions["simplify"] >= positions["verify"] || positions["verify"] >= positions["review"] {
		t.Fatalf("expected simplify before verify before review, got %#v", positions)
	}
}
