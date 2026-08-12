package adapters

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"clean-code/internal/contracts"
)

//go:embed *.yaml
var files embed.FS

type Definition struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Kind          string     `json:"kind"`
	Trust         string     `json:"trust"`
	Language      string     `json:"language"`
	Markers       []string   `json:"markers"`
	Commands      []Proposal `json:"commands"`
}

type Proposal struct {
	contracts.CommandSpec
	WhenAny []string `json:"when_any,omitempty"`
}

type Match struct {
	ID               string                  `json:"id"`
	Language         string                  `json:"language"`
	Root             string                  `json:"root"`
	MatchedFiles     []string                `json:"matched_files"`
	ProposedCommands []contracts.CommandSpec `json:"proposed_commands,omitempty"`
}

func Catalog() ([]Definition, error) {
	entries, err := files.ReadDir(".")
	if err != nil {
		return nil, err
	}
	definitions := make([]Definition, 0, len(entries))
	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		body, err := files.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read adapter %s: %w", entry.Name(), err)
		}
		var definition Definition
		decoder := json.NewDecoder(strings.NewReader(string(body)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&definition); err != nil {
			return nil, fmt.Errorf("parse adapter %s: %w", entry.Name(), err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			if err == nil {
				return nil, fmt.Errorf("parse adapter %s: unexpected trailing JSON value", entry.Name())
			}
			return nil, fmt.Errorf("parse adapter %s: %w", entry.Name(), err)
		}
		if err := validate(definition); err != nil {
			return nil, fmt.Errorf("validate adapter %s: %w", entry.Name(), err)
		}
		if seen[definition.ID] {
			return nil, fmt.Errorf("duplicate adapter id %q", definition.ID)
		}
		seen[definition.ID] = true
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].ID < definitions[j].ID })
	return definitions, nil
}

func Detect(projectFiles []string) ([]Match, error) {
	definitions, err := Catalog()
	if err != nil {
		return nil, err
	}
	availableByRoot := map[string]map[string][]string{}
	for _, projectFile := range projectFiles {
		portable := filepath.FromSlash(projectFile)
		root := filepath.ToSlash(filepath.Dir(portable))
		if root == "" {
			root = "."
		}
		if availableByRoot[root] == nil {
			availableByRoot[root] = map[string][]string{}
		}
		name := filepath.Base(portable)
		availableByRoot[root][name] = append(availableByRoot[root][name], projectFile)
	}
	var matches []Match
	for _, definition := range definitions {
		roots := make([]string, 0, len(availableByRoot))
		for root := range availableByRoot {
			roots = append(roots, root)
		}
		sort.Strings(roots)
		for _, root := range roots {
			available := availableByRoot[root]
			matchedFiles := matchingFiles(definition.Markers, available)
			if len(matchedFiles) == 0 {
				continue
			}
			match := Match{ID: definition.ID, Language: definition.Language, Root: root, MatchedFiles: matchedFiles}
			for _, proposal := range definition.Commands {
				if len(proposal.WhenAny) == 0 || len(matchingFiles(proposal.WhenAny, available)) > 0 {
					command := proposal.CommandSpec
					if root != "." && command.WorkingDir == "" {
						command.WorkingDir = root
						command.ID += "-" + projectIdentifier(root)
					}
					match.ProposedCommands = append(match.ProposedCommands, command)
				}
			}
			matches = append(matches, match)
		}
	}
	return matches, nil
}

func identifier(value string) string {
	var builder strings.Builder
	lastWasDash := false
	for _, character := range strings.ToLower(value) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			builder.WriteRune(character)
			lastWasDash = false
		} else if builder.Len() > 0 && !lastWasDash {
			builder.WriteByte('-')
			lastWasDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func projectIdentifier(root string) string {
	digest := sha256.Sum256([]byte(root))
	return fmt.Sprintf("%s-%x", identifier(root), digest[:4])
}

func matchingFiles(markers []string, available map[string][]string) []string {
	var result []string
	for _, marker := range markers {
		result = append(result, available[marker]...)
	}
	sort.Strings(result)
	return result
}

func validate(definition Definition) error {
	if definition.SchemaVersion != "1.0.0" {
		return fmt.Errorf("unsupported schema_version %q", definition.SchemaVersion)
	}
	if definition.ID == "" || definition.Kind != "language" || definition.Trust != "builtin" || definition.Language == "" || len(definition.Markers) == 0 {
		return fmt.Errorf("id, builtin language kind, language, and markers are required")
	}
	markers := map[string]bool{}
	for _, marker := range definition.Markers {
		if filepath.Base(marker) != marker || marker == "." || marker == ".." {
			return fmt.Errorf("marker %q must be a filename", marker)
		}
		markers[marker] = true
	}
	commandIDs := map[string]bool{}
	for index, proposal := range definition.Commands {
		if err := proposal.CommandSpec.Validate(); err != nil {
			return fmt.Errorf("command %d: %w", index, err)
		}
		if commandIDs[proposal.ID] {
			return fmt.Errorf("command %d duplicates id %q", index, proposal.ID)
		}
		commandIDs[proposal.ID] = true
		for _, marker := range proposal.WhenAny {
			if !markers[marker] {
				return fmt.Errorf("command %d references unknown marker %q", index, marker)
			}
		}
	}
	return nil
}
