package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func resolveVersion() string {
	if version != "" && version != "0.1.0-dev" {
		return version
	}

	exe, err := os.Executable()
	if err != nil {
		return version
	}
	return resolveVersionForExecutable(exe)
}

func resolveVersionForExecutable(exe string) string {
	dir := filepath.Dir(exe)
	for i := 0; i < 6; i++ {
		pkgPath := filepath.Join(dir, "package.json")
		if data, err := os.ReadFile(pkgPath); err == nil {
			var pkg struct {
				Version string `json:"version"`
			}
			if json.Unmarshal(data, &pkg) == nil && pkg.Version != "" {
				return pkg.Version
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return version
}
