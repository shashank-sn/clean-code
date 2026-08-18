package tests_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNpmPackageManifest(t *testing.T) {
	root := filepath.Join("..")
	body, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		License string `json:"license"`
		Bin     map[string]string `json:"bin"`
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "@shashanksn/clean-code" || manifest.License != "MIT" {
		t.Fatalf("unexpected package manifest: %+v", manifest)
	}
	if manifest.Bin["clean-code"] != "bin/clean-code.js" {
		t.Fatalf("unexpected bin map: %+v", manifest.Bin)
	}
	if strings.Contains(string(body), "postinstall") {
		t.Fatal("package.json must not define postinstall (npm allowScripts blocks it on global install)")
	}

	for _, path := range []string{
		"bin/runtime.js",
		"bin/build-cli.js",
		"bin/package-meta.js",
		"scripts/ensure-runtimes.sh",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected package file %s: %v", path, err)
		}
	}
}
