package architecture

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const maxInputBytes int64 = 10 << 20

var componentID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type Policy struct {
	SchemaVersion string      `json:"schema_version"`
	Exclude       []string    `json:"exclude,omitempty"`
	AllowCycles   bool        `json:"allow_cycles,omitempty"`
	Components    []Component `json:"components"`
	Exceptions    []Exception `json:"exceptions,omitempty"`
}

type Component struct {
	ID          string   `json:"id"`
	Paths       []string `json:"paths"`
	MayDependOn []string `json:"may_depend_on,omitempty"`
	Public      []string `json:"public,omitempty"`
}

type Exception struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type Graph struct {
	SchemaVersion string `json:"schema_version"`
	Edges         []Edge `json:"edges"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Violation struct {
	Kind          string   `json:"kind"`
	From          string   `json:"from,omitempty"`
	To            string   `json:"to,omitempty"`
	FromComponent string   `json:"from_component,omitempty"`
	ToComponent   string   `json:"to_component,omitempty"`
	Path          []string `json:"path,omitempty"`
	Summary       string   `json:"summary"`
}

type Report struct {
	SchemaVersion string      `json:"schema_version"`
	Status        string      `json:"status"`
	CheckedEdges  int         `json:"checked_edges"`
	ExcludedEdges int         `json:"excluded_edges"`
	Violations    []Violation `json:"violations"`
}

func LoadPolicy(path string) (Policy, error) {
	var policy Policy
	if err := loadStrictJSON(path, &policy); err != nil {
		return Policy{}, err
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, fmt.Errorf("validate architecture policy: %w", err)
	}
	return policy, nil
}

func LoadGraph(path string) (Graph, error) {
	var graph Graph
	if err := loadStrictJSON(path, &graph); err != nil {
		return Graph{}, err
	}
	if graph.SchemaVersion != "1.0.0" {
		return Graph{}, fmt.Errorf("validate dependency graph: unsupported schema_version %q", graph.SchemaVersion)
	}
	seen := map[Edge]bool{}
	for index, edge := range graph.Edges {
		if err := validateRelativePath(edge.From); err != nil {
			return Graph{}, fmt.Errorf("validate dependency graph edge %d from: %w", index, err)
		}
		if err := validateRelativePath(edge.To); err != nil {
			return Graph{}, fmt.Errorf("validate dependency graph edge %d to: %w", index, err)
		}
		if seen[edge] {
			return Graph{}, fmt.Errorf("validate dependency graph edge %d: duplicate edge %q -> %q", index, edge.From, edge.To)
		}
		seen[edge] = true
	}
	return graph, nil
}

func (policy Policy) Validate() error {
	if policy.SchemaVersion != "1.0.0" {
		return fmt.Errorf("unsupported schema_version %q", policy.SchemaVersion)
	}
	if len(policy.Components) == 0 {
		return errors.New("at least one component is required")
	}
	components := map[string]bool{}
	for index, component := range policy.Components {
		if !componentID.MatchString(component.ID) || len(component.Paths) == 0 {
			return fmt.Errorf("component %d requires id and paths", index)
		}
		if components[component.ID] {
			return fmt.Errorf("duplicate component id %q", component.ID)
		}
		components[component.ID] = true
		for _, pattern := range append(append([]string{}, component.Paths...), component.Public...) {
			if err := validatePattern(pattern); err != nil {
				return fmt.Errorf("component %q: %w", component.ID, err)
			}
		}
	}
	for _, component := range policy.Components {
		for _, dependency := range component.MayDependOn {
			if !components[dependency] {
				return fmt.Errorf("component %q references unknown dependency %q", component.ID, dependency)
			}
		}
	}
	for index, pattern := range policy.Exclude {
		if err := validatePattern(pattern); err != nil {
			return fmt.Errorf("exclude %d: %w", index, err)
		}
	}
	for index, exception := range policy.Exceptions {
		if strings.TrimSpace(exception.Reason) == "" {
			return fmt.Errorf("exception %d reason is required", index)
		}
		if err := validatePattern(exception.From); err != nil {
			return fmt.Errorf("exception %d from: %w", index, err)
		}
		if err := validatePattern(exception.To); err != nil {
			return fmt.Errorf("exception %d to: %w", index, err)
		}
	}
	return nil
}

func Evaluate(policy Policy, graph Graph) Report {
	report := Report{SchemaVersion: "1.0.0", Status: "PASS", Violations: []Violation{}}
	componentEdges := map[string]map[string]bool{}
	for _, edge := range graph.Edges {
		if matchesAny(policy.Exclude, edge.From) || matchesAny(policy.Exclude, edge.To) {
			report.ExcludedEdges++
			continue
		}
		report.CheckedEdges++
		fromComponents := matchingComponents(policy.Components, edge.From)
		toComponents := matchingComponents(policy.Components, edge.To)
		if len(fromComponents) != 1 || len(toComponents) != 1 {
			report.Violations = append(report.Violations, membershipViolation(edge, fromComponents, toComponents))
			continue
		}
		from := fromComponents[0]
		to := toComponents[0]
		if from.ID == to.ID {
			continue
		}
		if componentEdges[from.ID] == nil {
			componentEdges[from.ID] = map[string]bool{}
		}
		componentEdges[from.ID][to.ID] = true
		if allowedByException(policy.Exceptions, edge) {
			continue
		}
		if !contains(from.MayDependOn, to.ID) {
			report.Violations = append(report.Violations, Violation{
				Kind: "forbidden-dependency", From: edge.From, To: edge.To,
				FromComponent: from.ID, ToComponent: to.ID,
				Summary: fmt.Sprintf("component %q may not depend on component %q", from.ID, to.ID),
			})
		}
		if len(to.Public) > 0 && !matchesAny(to.Public, edge.To) {
			report.Violations = append(report.Violations, Violation{
				Kind: "private-surface", From: edge.From, To: edge.To,
				FromComponent: from.ID, ToComponent: to.ID,
				Summary: fmt.Sprintf("%q is outside component %q's public surface", edge.To, to.ID),
			})
		}
	}
	if !policy.AllowCycles {
		for _, cycle := range componentCycles(componentEdges) {
			report.Violations = append(report.Violations, Violation{Kind: "component-cycle", Path: cycle, Summary: "component dependency cycle: " + strings.Join(cycle, " -> ")})
		}
	}
	sort.Slice(report.Violations, func(i, j int) bool {
		left := report.Violations[i].Kind + report.Violations[i].From + report.Violations[i].To + strings.Join(report.Violations[i].Path, "/")
		right := report.Violations[j].Kind + report.Violations[j].From + report.Violations[j].To + strings.Join(report.Violations[j].Path, "/")
		return left < right
	})
	if len(report.Violations) > 0 {
		report.Status = "FAIL"
	}
	return report
}

func matchingComponents(components []Component, path string) []Component {
	var result []Component
	for _, component := range components {
		if matchesAny(component.Paths, path) {
			result = append(result, component)
		}
	}
	return result
}

func membershipViolation(edge Edge, from, to []Component) Violation {
	violation := Violation{Kind: "undeclared-path", From: edge.From, To: edge.To, Summary: "dependency endpoint does not belong to exactly one declared component"}
	if len(from) > 1 || len(to) > 1 {
		violation.Kind = "ambiguous-component"
		violation.Summary = "dependency endpoint belongs to more than one declared component"
	}
	return violation
}

func allowedByException(exceptions []Exception, edge Edge) bool {
	for _, exception := range exceptions {
		if matches(exception.From, edge.From) && matches(exception.To, edge.To) {
			return true
		}
	}
	return false
}

func matchesAny(patterns []string, path string) bool {
	for _, pattern := range patterns {
		if matches(pattern, path) {
			return true
		}
	}
	return false
}

func matches(pattern, path string) bool {
	pattern = pathpkg.Clean(strings.ReplaceAll(pattern, `\`, "/"))
	path = pathpkg.Clean(strings.ReplaceAll(path, `\`, "/"))
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	matched, _ := pathpkg.Match(pattern, path)
	return matched
}

func componentCycles(edges map[string]map[string]bool) [][]string {
	var cycles [][]string
	state := map[string]int{}
	var stack []string
	seen := map[string]bool{}
	var visit func(string)
	visit = func(node string) {
		state[node] = 1
		stack = append(stack, node)
		neighbors := make([]string, 0, len(edges[node]))
		for neighbor := range edges[node] {
			neighbors = append(neighbors, neighbor)
		}
		sort.Strings(neighbors)
		for _, neighbor := range neighbors {
			if state[neighbor] == 0 {
				visit(neighbor)
			} else if state[neighbor] == 1 {
				start := 0
				for stack[start] != neighbor {
					start++
				}
				cycle := append(append([]string{}, stack[start:]...), neighbor)
				key := canonicalCycle(cycle)
				if !seen[key] {
					seen[key] = true
					cycles = append(cycles, cycle)
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[node] = 2
	}
	nodes := make([]string, 0, len(edges))
	for node := range edges {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	for _, node := range nodes {
		if state[node] == 0 {
			visit(node)
		}
	}
	return cycles
}

func canonicalCycle(cycle []string) string {
	base := cycle[:len(cycle)-1]
	best := ""
	for index := range base {
		rotated := append(append([]string{}, base[index:]...), base[:index]...)
		candidate := strings.Join(rotated, "->")
		if best == "" || candidate < best {
			best = candidate
		}
	}
	return best
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func validatePattern(pattern string) error {
	if strings.TrimSpace(pattern) == "" || portableAbsolute(pattern) {
		return fmt.Errorf("pattern %q must be repository-relative", pattern)
	}
	clean := pathpkg.Clean(strings.ReplaceAll(pattern, `\`, "/"))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("pattern %q escapes repository root", pattern)
	}
	if strings.Contains(strings.TrimSuffix(clean, "/**"), "**") {
		return fmt.Errorf("pattern %q may use ** only as a trailing path segment", pattern)
	}
	if _, err := pathpkg.Match(strings.TrimSuffix(clean, "/**"), "probe"); err != nil {
		return fmt.Errorf("pattern %q is invalid: %w", pattern, err)
	}
	return nil
}

func validateRelativePath(path string) error {
	if strings.TrimSpace(path) == "" || portableAbsolute(path) {
		return fmt.Errorf("path %q must be repository-relative", path)
	}
	clean := pathpkg.Clean(strings.ReplaceAll(path, `\`, "/"))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("path %q escapes repository root", path)
	}
	return nil
}

func portableAbsolute(value string) bool {
	portable := strings.ReplaceAll(value, `\`, "/")
	return strings.HasPrefix(portable, "/") || filepath.VolumeName(value) != "" || len(portable) >= 3 && portable[1] == ':' && portable[2] == '/'
}

func loadStrictJSON(path string, target any) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("inspect %s: input must be a regular file", path)
	}
	if info.Size() > maxInputBytes {
		return fmt.Errorf("inspect %s: input exceeds %d bytes", path, maxInputBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxInputBytes+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("parse %s: unexpected trailing JSON value", path)
		}
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}
