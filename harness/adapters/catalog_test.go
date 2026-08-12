package adapters

import (
	"strings"
	"testing"
)

func TestCatalogContainsMaintainedLanguageAdapters(t *testing.T) {
	definitions, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dotnet", "go", "java", "javascript-typescript", "python", "ruby", "rust", "swift"}
	if len(definitions) != len(want) {
		t.Fatalf("expected %d adapters, got %d", len(want), len(definitions))
	}
	for index := range want {
		if definitions[index].ID != want[index] {
			t.Fatalf("expected %v, got %+v", want, definitions)
		}
	}
}

func TestDetectReturnsOnlyApplicableCommands(t *testing.T) {
	matches, err := Detect([]string{"apps/web/package.json", "services/api/pom.xml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected two matches, got %+v", matches)
	}
	if matches[0].ID != "java" || matches[0].Root != "services/api" || len(matches[0].ProposedCommands) != 1 || !strings.HasPrefix(matches[0].ProposedCommands[0].ID, "maven-test-services-api-") || matches[0].ProposedCommands[0].WorkingDir != "services/api" {
		t.Fatalf("unexpected Java match: %+v", matches[0])
	}
	if matches[1].ID != "javascript-typescript" || matches[1].Root != "apps/web" || len(matches[1].ProposedCommands) != 1 || !strings.HasPrefix(matches[1].ProposedCommands[0].ID, "npm-test-apps-web-") || matches[1].ProposedCommands[0].WorkingDir != "apps/web" {
		t.Fatalf("unexpected JavaScript match: %+v", matches[1])
	}
}

func TestDetectAvoidsNormalizedRootIDCollisions(t *testing.T) {
	matches, err := Detect([]string{"services/a_b/go.mod", "services/a-b/go.mod"})
	if err != nil {
		t.Fatal(err)
	}
	if matches[0].ProposedCommands[0].ID == matches[1].ProposedCommands[0].ID {
		t.Fatalf("normalized roots collided: %+v", matches)
	}
}

func TestDetectKeepsMonorepoProjectsSeparate(t *testing.T) {
	matches, err := Detect([]string{"services/a/go.mod", "services/b/go.mod"})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 || matches[0].Root != "services/a" || matches[1].Root != "services/b" {
		t.Fatalf("unexpected monorepo matches: %+v", matches)
	}
	if matches[0].ProposedCommands[0].ID == matches[1].ProposedCommands[0].ID {
		t.Fatalf("monorepo command ids must be unique: %+v", matches)
	}
}
