package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxManifestBytes int64 = 1 << 20

func PackageRoot() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("CLEAN_CODE_PACKAGE_ROOT")); configured != "" {
		return validRoot(configured)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	for current := workingDirectory; ; current = filepath.Dir(current) {
		if root, err := validRoot(current); err == nil {
			return root, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return "", errors.New("locate Clean Code package root: set CLEAN_CODE_PACKAGE_ROOT")
}

func validRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve package root: %w", err)
	}
	info, err := os.Lstat(filepath.Join(abs, "skills"))
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("locate Clean Code package root: skills directory is required")
	}
	return abs, nil
}

func LoadAll() (map[string]Package, error) {
	root, err := PackageRoot()
	if err != nil {
		return nil, err
	}
	return LoadAllFrom(root)
}

func LoadAllFrom(root string) (map[string]Package, error) {
	absRoot, err := validRoot(root)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(absRoot, "skills"))
	if err != nil {
		return nil, fmt.Errorf("read skills: %w", err)
	}
	packages := map[string]Package{}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "clean-") {
			continue
		}
		packageDir := filepath.Join(absRoot, "skills", entry.Name())
		loaded, err := loadPackage(packageDir)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", entry.Name(), err)
		}
		if loaded.Descriptor.ID != entry.Name() {
			return nil, fmt.Errorf("load %s: id must match skill directory", entry.Name())
		}
		if _, exists := packages[loaded.Descriptor.ID]; exists {
			return nil, fmt.Errorf("duplicate agent id %q", loaded.Descriptor.ID)
		}
		packages[loaded.Descriptor.ID] = loaded
	}
	if len(packages) == 0 {
		return nil, errors.New("no portable agent packages found")
	}
	for _, id := range sortedKeys(packages) {
		for _, target := range packages[id].Descriptor.HandoffTo {
			if _, exists := packages[target]; !exists {
				return nil, fmt.Errorf("agent %q hands off to unknown agent %q", id, target)
			}
		}
	}
	return packages, nil
}

func Load(id string) (Package, error) {
	packages, err := LoadAll()
	if err != nil {
		return Package{}, err
	}
	loaded, exists := packages[id]
	if !exists {
		return Package{}, fmt.Errorf("unknown agent %q", id)
	}
	return loaded, nil
}

func List() ([]Descriptor, error) {
	packages, err := LoadAll()
	if err != nil {
		return nil, err
	}
	ids := sortedKeys(packages)
	descriptors := make([]Descriptor, 0, len(ids))
	for _, id := range ids {
		descriptors = append(descriptors, packages[id].Descriptor)
	}
	return descriptors, nil
}

func Validate(id string) error {
	packages, err := LoadAll()
	if err != nil {
		return err
	}
	if id != "" {
		if _, exists := packages[id]; !exists {
			return fmt.Errorf("unknown agent %q", id)
		}
	}
	return nil
}

func loadPackage(directory string) (Package, error) {
	manifestPath := filepath.Join(directory, "agent.json")
	info, err := os.Lstat(manifestPath)
	if err != nil {
		return Package{}, fmt.Errorf("inspect manifest: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Package{}, errors.New("inspect manifest: agent.json must be a regular file")
	}
	if info.Size() > maxManifestBytes {
		return Package{}, fmt.Errorf("inspect manifest: exceeds %d bytes", maxManifestBytes)
	}
	file, err := os.Open(manifestPath)
	if err != nil {
		return Package{}, fmt.Errorf("open manifest: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxManifestBytes+1))
	decoder.DisallowUnknownFields()
	var descriptor Descriptor
	if err := decoder.Decode(&descriptor); err != nil {
		return Package{}, fmt.Errorf("parse manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Package{}, errors.New("parse manifest: unexpected trailing JSON value")
		}
		return Package{}, fmt.Errorf("parse manifest: %w", err)
	}
	if err := validateDescriptor(descriptor); err != nil {
		return Package{}, fmt.Errorf("validate manifest: %w", err)
	}
	instructionPath := filepath.Join(directory, descriptor.InstructionFile)
	instructionInfo, err := os.Lstat(instructionPath)
	if err != nil {
		return Package{}, fmt.Errorf("inspect instructions: %w", err)
	}
	if instructionInfo.Mode()&os.ModeSymlink != 0 || !instructionInfo.Mode().IsRegular() {
		return Package{}, errors.New("inspect instructions: SKILL.md must be a regular file")
	}
	instructions, err := os.ReadFile(instructionPath)
	if err != nil {
		return Package{}, fmt.Errorf("read instructions: %w", err)
	}
	return Package{Descriptor: descriptor, Directory: directory, Instructions: string(instructions)}, nil
}

func SortedIDs(packages map[string]Package) []string {
	ids := sortedKeys(packages)
	sort.Strings(ids)
	return ids
}
