package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryJSONArtifactsAreValid(t *testing.T) {
	root := filepath.Join("..", "..")
	paths := []string{
		filepath.Join(root, ".codex-plugin", "plugin.json"),
		filepath.Join(root, "hosts", "host-capabilities.schema.json"),
		filepath.Join(root, "harness", "config", "config.schema.json"),
		filepath.Join(root, "harness", "config", "example.clean-code.json"),
	}

	err := filepath.WalkDir(filepath.Join(root, "harness", "schemas"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", path, err)
			continue
		}
		var value any
		if err := json.Unmarshal(body, &value); err != nil {
			t.Errorf("parse %s: %v", path, err)
		}
	}
}
