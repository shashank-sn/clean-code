package architecture

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildViewShowsComponentsEdgesViolationsCyclesAndUnknownCoverage(t *testing.T) {
	policy := testPolicy()
	policy.Components[0].MayDependOn = []string{"delivery"}
	graph := Graph{SchemaVersion: "1.0.0", Edges: []Edge{
		{From: "delivery/http.go", To: "core/usecase.go"},
		{From: "core/usecase.go", To: "delivery/http.go"},
		{From: "unknown/file.go", To: "delivery/http.go"},
	}}
	view := BuildView(policy, graph, nil)
	if view.Coverage.Complete {
		t.Fatal("view without producer proof must not claim complete graph coverage")
	}
	if len(view.Components) != 2 || len(view.Edges) != 3 || len(view.Violations) == 0 || len(view.Cycles) != 1 {
		t.Fatalf("incomplete architecture view: %+v", view)
	}
	if view.Coverage.UnknownEdges != 1 || len(view.Coverage.UnknownPaths) != 1 || view.Coverage.UnknownPaths[0] != "unknown/file.go" {
		t.Fatalf("expected unknown graph coverage, got %+v", view.Coverage)
	}
	encoded, err := MarshalView(view)
	if err != nil || !strings.Contains(string(encoded), `"public_surfaces"`) {
		t.Fatalf("view must have inspectable JSON: %v %s", err, encoded)
	}
}

func TestBuildViewIncludesOptionalGraphDiff(t *testing.T) {
	policy := testPolicy()
	previous := Graph{SchemaVersion: "1.0.0", Edges: []Edge{{From: "delivery/http.go", To: "core/usecase.go"}}}
	current := Graph{SchemaVersion: "1.0.0", Edges: []Edge{{From: "core/usecase.go", To: "delivery/http.go"}}}
	view := BuildView(policy, current, &previous)
	if view.Diff == nil || len(view.Diff.AddedEdges) != 1 || len(view.Diff.RemovedEdges) != 1 {
		t.Fatalf("expected edge diff, got %+v", view.Diff)
	}
}

func TestBuildViewRequiresEvidenceBeforeClaimingCoverageCompleteness(t *testing.T) {
	graph := Graph{SchemaVersion: "1.0.0", Edges: []Edge{{From: "delivery/http.go", To: "core/usecase.go"}}}
	if _, err := BuildViewWithCoverage(testPolicy(), graph, nil, &CoverageProof{ProducerID: "graph-provider"}); err == nil {
		t.Fatal("expected incomplete coverage proof to fail")
	}
	evidence := []byte(fmt.Sprintf(`{"producer_id":"graph-provider","collection_scope":"repository source files","graph_sha256":"%s"}`, graphSHA256(graph)))
	path := filepath.Join(t.TempDir(), "coverage.json")
	if err := os.WriteFile(path, evidence, 0o644); err != nil {
		t.Fatal(err)
	}
	view, err := BuildViewWithCoverage(testPolicy(), graph, nil, &CoverageProof{ProducerID: "graph-provider", CollectionScope: "repository source files", EvidencePath: path, EvidenceSHA256: fmt.Sprintf("%x", sha256.Sum256(evidence))})
	if err != nil || !view.Coverage.Complete {
		t.Fatalf("expected proven coverage, got view=%+v err=%v", view, err)
	}
	if _, err := BuildViewWithCoverage(testPolicy(), graph, nil, &CoverageProof{ProducerID: "graph-provider", CollectionScope: "repository source files", EvidencePath: path, EvidenceSHA256: strings.Repeat("a", 64)}); err == nil {
		t.Fatal("expected mismatched evidence digest rejection")
	}
}
