package policy

import (
	"fmt"
	"maps"
	"slices"
	"sort"

	"clean-code/internal/contracts"
)

// Compare reports every executable policy difference without exposing values.
func Compare(trusted, proposed []contracts.CommandSpec) []string {
	trustedByID := index(trusted)
	proposedByID := index(proposed)
	ids := make([]string, 0, len(trustedByID)+len(proposedByID))
	seen := map[string]bool{}
	for id := range trustedByID {
		ids = append(ids, id)
		seen[id] = true
	}
	for id := range proposedByID {
		if !seen[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)

	var deltas []string
	for _, id := range ids {
		base, baseOK := trustedByID[id]
		candidate, candidateOK := proposedByID[id]
		switch {
		case !baseOK:
			deltas = append(deltas, fmt.Sprintf("command %q was added", id))
		case !candidateOK:
			deltas = append(deltas, fmt.Sprintf("command %q was removed", id))
		default:
			for _, field := range changedFields(base, candidate) {
				deltas = append(deltas, fmt.Sprintf("command %q changed %s", id, field))
			}
		}
	}
	return deltas
}

func index(commands []contracts.CommandSpec) map[string]contracts.CommandSpec {
	result := make(map[string]contracts.CommandSpec, len(commands))
	for _, command := range commands {
		result[command.ID] = command
	}
	return result
}

func changedFields(a, b contracts.CommandSpec) []string {
	var fields []string
	checks := []struct {
		name    string
		changed bool
	}{
		{"category", a.Category != b.Category},
		{"executable", a.Executable != b.Executable},
		{"arguments", !slices.Equal(a.Args, b.Args)},
		{"working_directory", a.WorkingDir != b.WorkingDir},
		{"timeout_seconds", a.TimeoutSec != b.TimeoutSec},
		{"max_output_bytes", a.MaxOutputBytes != b.MaxOutputBytes},
		{"allowed_exit_codes", !slices.Equal(a.AllowedExitCodes, b.AllowedExitCodes)},
		{"required", a.Required != b.Required},
		{"environment", !maps.Equal(a.Env, b.Env)},
		{"shell", a.Shell != b.Shell},
		{"artifacts", !slices.Equal(a.Artifacts, b.Artifacts)},
	}
	for _, check := range checks {
		if check.changed {
			fields = append(fields, check.name)
		}
	}
	return fields
}
