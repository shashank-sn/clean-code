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
	if len(entries) != 16 {
		t.Fatalf("expected sixteen skills, got %d", len(entries))
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

func TestRoadmapSchemasMatchRuntimeRequiredFields(t *testing.T) {
	root:=filepath.Join("..")
	checks:=map[string][]string{
		"harness/schemas/study-manifest.schema.json":{"repository","revision"},
		"harness/schemas/study-result.schema.json":{"manifest_digest"},
		"harness/schemas/policy-exception-approval.schema.json":{"organization_digest","repository_digest","command_id","scope","expires_at","approver"},
	}
	for path,required:=range checks{body,err:=os.ReadFile(filepath.Join(root,path));if err!=nil{t.Fatal(err)};var schema struct{Required []string `json:"required"`};if err:=json.Unmarshal(body,&schema);err!=nil{t.Fatal(err)};seen:=map[string]bool{};for _,v:=range schema.Required{seen[v]=true};for _,v:=range required{if !seen[v]{t.Errorf("%s missing required %s",path,v)}}}
	body,err:=os.ReadFile(filepath.Join(root,"harness/schemas/policy-pack.schema.json"));if err!=nil{t.Fatal(err)};if strings.Contains(string(body),`"exceptions"`){t.Fatal("policy pack schema still permits unsigned exceptions")}
}
func TestRoadmapFixturesParseRuntimeShapes(t *testing.T) {
	root:=filepath.Join("..")
	var result struct{SchemaVersion,StudyID,ManifestDigest string;Outcomes []any};body,err:=os.ReadFile(filepath.Join(root,"harness/studies/valid-study-result.json"));if err!=nil{t.Fatal(err)};if err:=json.Unmarshal(body,&result);err!=nil||result.SchemaVersion!="1.0.0"||len(result.ManifestDigest)!=64{t.Fatalf("invalid study fixture: %v",err)}
	var approval struct{SchemaVersion,OrganizationDigest,RepositoryDigest,CommandID,Scope,ExpiresAt,Approver string};body,err=os.ReadFile(filepath.Join(root,"harness/policies/valid-policy-exception-approval.json"));if err!=nil{t.Fatal(err)};if err:=json.Unmarshal(body,&approval);err!=nil||approval.Scope!="required"||approval.Approver==""{t.Fatalf("invalid approval fixture: %v",err)}
}
