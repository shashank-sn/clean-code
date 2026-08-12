package integration_test

import (
	"path/filepath"
	"testing"

	"clean-code/internal/discover"
)

func TestMaintainedLanguageFixturesProduceReadOnlyProposals(t *testing.T) {
	fixtures := []struct {
		name      string
		adapterID string
	}{
		{"go", "go"},
		{"javascript", "javascript-typescript"},
		{"python", "python"},
		{"java", "java"},
		{"rust", "rust"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			root := filepath.Join("..", "fixtures", "languages", fixture.name)
			result, err := discover.Inspect(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Commands) != 0 {
				t.Fatalf("adapter proposal gained execution authority: %+v", result.Commands)
			}
			if len(result.Adapters) != 1 || result.Adapters[0].ID != fixture.adapterID || len(result.Adapters[0].ProposedCommands) == 0 {
				t.Fatalf("unexpected adapter result: %+v", result.Adapters)
			}
		})
	}
}

func TestGenericRepositoryWorksWithoutMaintainedAdapter(t *testing.T) {
	result, err := discover.Inspect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !result.GenericCommandsSupported || len(result.Adapters) != 0 {
		t.Fatalf("unexpected generic result: %+v", result)
	}
}
