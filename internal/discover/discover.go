package discover

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"clean-code/internal/contracts"
)

type Result struct {
	Root                     string                  `json:"root"`
	Languages                []string                `json:"languages"`
	ProjectFiles             []string                `json:"project_files"`
	ConfigurationFound       bool                    `json:"configuration_found"`
	GenericCommandsSupported bool                    `json:"generic_commands_supported"`
	Commands                 []contracts.CommandSpec `json:"commands,omitempty"`
}

type configuration struct {
	SchemaVersion string                  `json:"schema_version"`
	Commands      []contracts.CommandSpec `json:"commands"`
}

var languageMarkers = map[string][]string{
	"go":                    {"go.mod"},
	"java":                  {"pom.xml", "build.gradle", "build.gradle.kts"},
	"javascript-typescript": {"package.json", "deno.json", "bun.lock", "bun.lockb"},
	"python":                {"pyproject.toml", "requirements.txt", "setup.py", "Pipfile"},
	"ruby":                  {"Gemfile"},
	"rust":                  {"Cargo.toml"},
	"swift":                 {"Package.swift"},
	"dotnet":                {"global.json"},
}

var ignoredDirectories = map[string]bool{
	".git": true, ".hg": true, ".svn": true, ".venv": true,
	"bin": true, "build": true, "dist": true, "node_modules": true,
	"target": true, "vendor": true,
}

const maxDiscoveryDepth = 4
const maxConfigurationSize = 1 << 20

func Inspect(root string) (Result, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return Result{}, fmt.Errorf("resolve root: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return Result{}, fmt.Errorf("inspect root: %w", err)
	}
	if !info.IsDir() {
		return Result{}, fmt.Errorf("inspect root: %s is not a directory", absolute)
	}

	languageSet, projectFiles, err := scanProjectFiles(absolute)
	if err != nil {
		return Result{}, err
	}

	languages := make([]string, 0, len(languageSet))
	for language := range languageSet {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	sort.Strings(projectFiles)

	configPath := filepath.Join(absolute, ".clean-code.json")
	commands, err := LoadCommands(configPath)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Root:                     absolute,
		Languages:                languages,
		ProjectFiles:             projectFiles,
		ConfigurationFound:       regularFile(configPath),
		GenericCommandsSupported: true,
		Commands:                 commands,
	}, nil
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func scanProjectFiles(root string) (map[string]bool, []string, error) {
	markers := map[string]string{}
	for language, names := range languageMarkers {
		for _, name := range names {
			markers[name] = language
		}
	}

	languages := map[string]bool{}
	projectFiles := []string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		depth := strings.Count(filepath.ToSlash(relative), "/") + 1
		if entry.IsDir() {
			if ignoredDirectories[entry.Name()] || depth > maxDiscoveryDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if depth > maxDiscoveryDepth || entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		language, ok := markers[entry.Name()]
		if !ok {
			return nil
		}
		languages[language] = true
		projectFiles = append(projectFiles, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("scan repository metadata: %w", err)
	}
	return languages, projectFiles, nil
}

// LoadCommands reads a strict command policy from path without executing it.
func LoadCommands(path string) ([]contracts.CommandSpec, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect configuration: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, errors.New("inspect configuration: .clean-code.json must be a regular file inside the repository")
	}
	if info.Size() > maxConfigurationSize {
		return nil, fmt.Errorf("inspect configuration: .clean-code.json exceeds %d bytes", maxConfigurationSize)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open configuration: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var config configuration
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("parse configuration: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return nil, err
	}
	if config.SchemaVersion != "1.0.0" {
		return nil, fmt.Errorf("parse configuration: unsupported schema_version %q", config.SchemaVersion)
	}
	seen := map[string]bool{}
	for index, command := range config.Commands {
		if err := command.Validate(); err != nil {
			return nil, fmt.Errorf("validate command %d: %w", index, err)
		}
		if seen[command.ID] {
			return nil, fmt.Errorf("validate command %d: duplicate id %q", index, command.ID)
		}
		seen[command.ID] = true
	}
	return config.Commands, nil
}

func ensureEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("parse configuration: %w", err)
	}
	return fmt.Errorf("parse configuration: unexpected trailing JSON value")
}
