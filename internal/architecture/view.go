package architecture

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// View is an inspectable JSON projection of an architecture policy and graph.
// It is intentionally separate from Report so the existing deterministic
// checker remains stable while hosts decide how to render this local data.
type View struct {
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	Components    []ViewComponent `json:"components"`
	Edges         []ViewEdge      `json:"edges"`
	Violations    []Violation     `json:"violations"`
	Cycles        [][]string      `json:"cycles"`
	Coverage      GraphCoverage   `json:"coverage"`
	Diff          *GraphDiff      `json:"diff,omitempty"`
}

type ViewComponent struct {
	ID             string   `json:"id"`
	Paths          []string `json:"paths"`
	PublicSurfaces []string `json:"public_surfaces"`
	MayDependOn    []string `json:"may_depend_on"`
}

type ViewEdge struct {
	From          string `json:"from"`
	To            string `json:"to"`
	FromComponent string `json:"from_component,omitempty"`
	ToComponent   string `json:"to_component,omitempty"`
	Status        string `json:"status"`
}

// GraphCoverage is conservative by default: no graph is complete merely
// because it was loaded. Complete is allowed only with producer scope evidence.
type GraphCoverage struct {
	Complete        bool     `json:"complete"`
	ProducerID      string   `json:"producer_id,omitempty"`
	CollectionScope string   `json:"collection_scope,omitempty"`
	EvidenceSHA256  string   `json:"evidence_sha256,omitempty"`
	GraphSHA256     string   `json:"graph_sha256,omitempty"`
	UnknownPaths    []string `json:"unknown_paths"`
	UnknownEdges    int      `json:"unknown_edges"`
	ExcludedEdges   int      `json:"excluded_edges"`
}

type CoverageProof struct {
	ProducerID      string
	CollectionScope string
	EvidenceSHA256  string
	EvidencePath    string
}

type coverageEvidence struct {
	ProducerID      string `json:"producer_id"`
	CollectionScope string `json:"collection_scope"`
	GraphSHA256     string `json:"graph_sha256"`
}

type GraphDiff struct {
	AddedEdges   []Edge `json:"added_edges"`
	RemovedEdges []Edge `json:"removed_edges"`
}

func BuildView(policy Policy, graph Graph, previous *Graph) View {
	view, _ := BuildViewWithCoverage(policy, graph, previous, nil)
	return view
}

func BuildViewWithCoverage(policy Policy, graph Graph, previous *Graph, proof *CoverageProof) (View, error) {
	if err := policy.Validate(); err != nil {
		return View{}, fmt.Errorf("validate architecture policy: %w", err)
	}
	if graph.SchemaVersion != "1.0.0" {
		return View{}, fmt.Errorf("validate dependency graph: unsupported schema_version %q", graph.SchemaVersion)
	}
	if previous != nil && previous.SchemaVersion != "1.0.0" {
		return View{}, fmt.Errorf("validate prior dependency graph: unsupported schema_version %q", previous.SchemaVersion)
	}
	coverage, err := coverageFromProof(proof, graph)
	if err != nil {
		return View{}, err
	}
	report := Evaluate(policy, graph)
	view := View{
		SchemaVersion: "1.0.0", Status: report.Status, Violations: append([]Violation{}, report.Violations...),
		Components: componentViews(policy.Components), Coverage: coverage, Edges: []ViewEdge{}, Cycles: [][]string{},
	}
	componentEdges := map[string]map[string]bool{}
	for _, edge := range graph.Edges {
		viewEdge, unknown := classifyViewEdge(policy, edge)
		view.Edges = append(view.Edges, viewEdge)
		if unknown {
			view.Coverage.UnknownEdges++
			view.Coverage.UnknownPaths = append(view.Coverage.UnknownPaths, unknownEndpoints(policy, edge)...)
		}
		if viewEdge.Status == "excluded" {
			view.Coverage.ExcludedEdges++
			continue
		}
		if viewEdge.FromComponent != "" && viewEdge.ToComponent != "" && viewEdge.FromComponent != viewEdge.ToComponent {
			if componentEdges[viewEdge.FromComponent] == nil {
				componentEdges[viewEdge.FromComponent] = map[string]bool{}
			}
			componentEdges[viewEdge.FromComponent][viewEdge.ToComponent] = true
		}
	}
	view.Coverage.UnknownPaths = uniqueSorted(view.Coverage.UnknownPaths)
	view.Cycles = componentCycles(componentEdges)
	if view.Cycles == nil {
		view.Cycles = [][]string{}
	}
	sort.Slice(view.Edges, func(i, j int) bool {
		return view.Edges[i].From+"\x00"+view.Edges[i].To < view.Edges[j].From+"\x00"+view.Edges[j].To
	})
	if previous != nil {
		view.Diff = graphDiff(*previous, graph)
	}
	return view, nil
}

func MarshalView(view View) ([]byte, error) {
	return json.MarshalIndent(view, "", "  ")
}

func coverageFromProof(proof *CoverageProof, graph Graph) (GraphCoverage, error) {
	coverage := GraphCoverage{Complete: false, UnknownPaths: []string{}}
	if proof == nil {
		return coverage, nil
	}
	if strings.TrimSpace(proof.ProducerID) == "" || strings.TrimSpace(proof.CollectionScope) == "" || strings.TrimSpace(proof.EvidencePath) == "" || !isSHA256(proof.EvidenceSHA256) {
		return GraphCoverage{}, errors.New("graph coverage proof requires producer id, collection scope, evidence file, and SHA-256 evidence")
	}
	body, err := os.ReadFile(proof.EvidencePath)
	if err != nil {
		return GraphCoverage{}, fmt.Errorf("read graph coverage evidence: %w", err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(body))
	if !strings.EqualFold(digest, proof.EvidenceSHA256) {
		return GraphCoverage{}, errors.New("graph coverage evidence SHA-256 does not match evidence file")
	}
	var evidence coverageEvidence
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		return GraphCoverage{}, fmt.Errorf("parse graph coverage evidence: %w", err)
	}
	if evidence.ProducerID != proof.ProducerID || evidence.CollectionScope != proof.CollectionScope || evidence.GraphSHA256 != graphSHA256(graph) {
		return GraphCoverage{}, errors.New("graph coverage evidence does not bind producer, scope, and graph")
	}
	coverage.Complete = true
	coverage.ProducerID = proof.ProducerID
	coverage.CollectionScope = proof.CollectionScope
	coverage.EvidenceSHA256 = proof.EvidenceSHA256
	coverage.GraphSHA256 = evidence.GraphSHA256
	return coverage, nil
}

func graphSHA256(graph Graph) string {
	body, _ := json.Marshal(graph)
	return fmt.Sprintf("%x", sha256.Sum256(body))
}

func componentViews(components []Component) []ViewComponent {
	views := make([]ViewComponent, 0, len(components))
	for _, component := range components {
		views = append(views, ViewComponent{
			ID: component.ID, Paths: append([]string{}, component.Paths...), PublicSurfaces: append([]string{}, component.Public...),
			MayDependOn: append([]string{}, component.MayDependOn...),
		})
	}
	sort.Slice(views, func(i, j int) bool { return views[i].ID < views[j].ID })
	return views
}

func classifyViewEdge(policy Policy, edge Edge) (ViewEdge, bool) {
	result := ViewEdge{From: edge.From, To: edge.To, Status: "checked"}
	if matchesAny(policy.Exclude, edge.From) || matchesAny(policy.Exclude, edge.To) {
		result.Status = "excluded"
		return result, false
	}
	from := matchingComponents(policy.Components, edge.From)
	to := matchingComponents(policy.Components, edge.To)
	if len(from) != 1 || len(to) != 1 {
		result.Status = "unknown"
		return result, true
	}
	result.FromComponent = from[0].ID
	result.ToComponent = to[0].ID
	return result, false
}

func unknownEndpoints(policy Policy, edge Edge) []string {
	unknown := []string{}
	if len(matchingComponents(policy.Components, edge.From)) != 1 {
		unknown = append(unknown, edge.From)
	}
	if len(matchingComponents(policy.Components, edge.To)) != 1 {
		unknown = append(unknown, edge.To)
	}
	return unknown
}

func graphDiff(previous, current Graph) *GraphDiff {
	prior := map[Edge]bool{}
	for _, edge := range previous.Edges {
		prior[edge] = true
	}
	now := map[Edge]bool{}
	for _, edge := range current.Edges {
		now[edge] = true
	}
	diff := &GraphDiff{AddedEdges: []Edge{}, RemovedEdges: []Edge{}}
	for edge := range now {
		if !prior[edge] {
			diff.AddedEdges = append(diff.AddedEdges, edge)
		}
	}
	for edge := range prior {
		if !now[edge] {
			diff.RemovedEdges = append(diff.RemovedEdges, edge)
		}
	}
	sort.Slice(diff.AddedEdges, func(i, j int) bool { return edgeKey(diff.AddedEdges[i]) < edgeKey(diff.AddedEdges[j]) })
	sort.Slice(diff.RemovedEdges, func(i, j int) bool { return edgeKey(diff.RemovedEdges[i]) < edgeKey(diff.RemovedEdges[j]) })
	return diff
}

func edgeKey(edge Edge) string { return edge.From + "\x00" + edge.To }

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}
