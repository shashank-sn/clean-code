package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"clean-code/internal/discover"
	"clean-code/internal/hosts"
)

func TestRunSetupUsesGenericFallback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"setup", "--host", "future-ide"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success, got %d: %s", code, stderr.String())
	}
	var result hosts.Capabilities
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.ID != "generic" || !result.CLI {
		t.Fatalf("unexpected fallback: %+v", result)
	}
}

func TestRunSetupWritesHostInstructionsWithoutOverwrite(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	args := []string{"setup", "--host", "cursor", "--output", root}
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("expected setup success, got %d: %s", code, stderr.String())
	}
	path := filepath.Join(root, ".cursor", "rules", "clean-code.mdc")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected generated host instructions: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(args, &stdout, &stderr); code != 1 {
		t.Fatalf("expected overwrite rejection, got %d", code)
	}
}

func TestRunBenchmarkScoresManifest(t *testing.T) {
	manifest := filepath.Join("..", "..", "harness", "calibration", "benchmark-manifest.yaml")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"benchmark", "--manifest", manifest}, &stdout, &stderr); code != 0 {
		t.Fatalf("expected benchmark success, got %d: %s", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"correct_silence": 2`)) {
		t.Fatalf("unexpected benchmark report: %s", stdout.String())
	}
}

func TestRunLearnAcceptsReversibleProposal(t *testing.T) {
	proposal := filepath.Join("..", "..", "harness", "policies", "example-change-proposal.json")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"learn", "--proposal", proposal}, &stdout, &stderr); code != 0 {
		t.Fatalf("expected proposal success, got %d: %s", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"status": "PASS"`)) {
		t.Fatalf("unexpected proposal report: %s", stdout.String())
	}
}

func TestRunDiscoverWritesJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(root+"/go.mod", []byte("module sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"discover", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success, got %d: %s", code, stderr.String())
	}
	var result discover.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Languages) != 1 || result.Languages[0] != "go" {
		t.Fatalf("unexpected discovery result: %+v", result)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("expected usage error, got %d", code)
	}
}

func TestRunRejectsExtraArguments(t *testing.T) {
	tests := [][]string{
		{"version", "extra"},
		{"hosts", "extra"},
		{"setup", "extra"},
		{"discover", ".", "extra"},
		{"verify", ".", "extra"},
		{"architecture", "extra"},
		{"trace", "extra"},
		{"review", "extra"},
		{"audit", "extra"},
		{"release-gate", "extra"},
		{"slop", ".", "extra"},
		{"benchmark", "extra"},
		{"learn", "extra"},
	}
	for _, args := range tests {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Errorf("expected usage error for %v, got %d", args, code)
		}
	}
}

func TestRunReviewAcceptsIndependentZeroFindings(t *testing.T) {
	root := t.TempDir()
	input := root + "/review.json"
	if err := os.WriteFile(input, []byte(`{
  "schema_version":"1.0.0",
  "revision":"abc",
  "change_author":"author",
  "reviewer":"reviewer",
  "scope":["change.go"],
  "findings":[]
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"review", "--input", input}, &stdout, &stderr)
	if code != 0 || !bytes.Contains(stdout.Bytes(), []byte(`"findings": []`)) {
		t.Fatalf("expected correct silence, got %d: %s\n%s", code, stderr.String(), stdout.String())
	}
}

func TestRunTraceReportsMissingTrack(t *testing.T) {
	root := t.TempDir()
	plan := root + "/test-plan.json"
	if err := os.WriteFile(plan, []byte(`{
  "schema_version":"1.0.0",
  "requirements":[{"id":"R1","acceptance_examples":["returns the result"]}],
  "tracks":[{"kind":"unit","requirement_ids":["R1"],"status":"PLANNED","owner":"unit"}]
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"trace", "--plan", plan}, &stdout, &stderr)
	if code != 1 || !bytes.Contains(stdout.Bytes(), []byte(`"kind": "missing-track"`)) {
		t.Fatalf("expected trace failure, got %d: %s\n%s", code, stderr.String(), stdout.String())
	}
}

func TestRunArchitectureReportsViolation(t *testing.T) {
	root := t.TempDir()
	policy := root + "/policy.json"
	graph := root + "/graph.json"
	if err := os.WriteFile(policy, []byte(`{
  "schema_version":"1.0.0",
  "components":[
    {"id":"core","paths":["core/**"]},
    {"id":"delivery","paths":["delivery/**"],"may_depend_on":["core"]}
  ]
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(graph, []byte(`{
  "schema_version":"1.0.0",
  "edges":[{"from":"core/usecase.go","to":"delivery/http.go"}]
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"architecture", "--policy", policy, "--graph", graph}, &stdout, &stderr)
	if code != 1 || !bytes.Contains(stdout.Bytes(), []byte(`"kind": "forbidden-dependency"`)) {
		t.Fatalf("expected architecture failure, got %d: %s\n%s", code, stderr.String(), stdout.String())
	}
}

func TestRunVerifyReportsNotConfigured(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run([]string{"verify", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected incomplete verification, got %d: %s", code, stderr.String())
	}
	var report struct {
		Complete bool `json:"complete"`
		Results  []struct {
			Status string `json:"status"`
		} `json:"results"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Complete || len(report.Results) != 1 || report.Results[0].Status != "NOT_CONFIGURED" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRunVerifyPassesAndWritesEvidenceBundle(t *testing.T) {
	root := t.TempDir()
	configuration := `{
  "schema_version": "1.0.0",
  "commands": [{"id":"go-version","executable":"go","args":["version"],"required":true}]
}`
	if err := os.WriteFile(root+"/.clean-code.json", []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	evidenceDirectory := root + "/evidence"
	var stdout, stderr bytes.Buffer
	code := run([]string{"verify", "--allow-repository-policy", "--output", evidenceDirectory, root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected passing verification, got %d: %s\n%s", code, stderr.String(), stdout.String())
	}
	var report struct {
		Complete bool `json:"complete"`
		Results  []struct {
			Status string `json:"status"`
		} `json:"results"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.Complete || len(report.Results) != 1 || report.Results[0].Status != "PASS" {
		t.Fatalf("unexpected report: %+v", report)
	}
	if _, err := os.Stat(evidenceDirectory + "/report.json"); err != nil {
		t.Fatalf("expected evidence report: %v", err)
	}
}

func TestRunVerifyDoesNotExecuteUnapprovedRepositoryPolicy(t *testing.T) {
	root := t.TempDir()
	configuration := `{
  "schema_version": "1.0.0",
  "commands": [{"id":"unapproved","executable":"go","args":["version"],"required":true}]
}`
	if err := os.WriteFile(root+"/.clean-code.json", []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"verify", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected policy block, got %d: %s", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"check_id": "policy.approval"`)) ||
		!bytes.Contains(stdout.Bytes(), []byte(`"status": "NOT_RUN"`)) {
		t.Fatalf("expected unexecuted policy result: %s", stdout.String())
	}
}

func TestRunVerifyRejectsMissingTrustedPolicy(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run([]string{"verify", "--trusted-policy", root + "/missing.json", root}, &stdout, &stderr)
	if code != 1 || !bytes.Contains(stderr.Bytes(), []byte("inspect trusted policy")) {
		t.Fatalf("expected missing policy error, got %d: %s", code, stderr.String())
	}
}

func TestRunSlopStopsAfterSecondPass(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.go", "b.go", "c.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("package sample\n// TODO unbounded\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var firstOut, firstErr bytes.Buffer
	if code := run([]string{"slop", root}, &firstOut, &firstErr); code != 1 {
		t.Fatalf("expected a repair result, got %d: %s", code, firstErr.String())
	}
	previous := filepath.Join(t.TempDir(), "first.json")
	if err := os.WriteFile(previous, firstOut.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	var secondOut, secondErr bytes.Buffer
	if code := run([]string{"slop", "--previous", previous, root}, &secondOut, &secondErr); code != 1 {
		t.Fatalf("expected escalation result, got %d: %s", code, secondErr.String())
	}
	if !bytes.Contains(secondOut.Bytes(), []byte(`"status": "ESCALATE"`)) ||
		!bytes.Contains(secondOut.Bytes(), []byte("Stop rewriting")) {
		t.Fatalf("expected bounded stop report: %s", secondOut.String())
	}
}

func TestRunReleaseGateFailsClosedOnEmptyEvidence(t *testing.T) {
	path:=filepath.Join(t.TempDir(),"release.json")
	body:=[]byte(`{"schema_version":"1.0.0","repository":"owner/repo","base_revision":"base","final_revision":"final","requirement_digest":"requirements","change_set_digest":"change","policy_revision":"policy","policy_gates":{"policy_revision":"policy","required_tests":["acceptance"],"required_reviews":["independent"],"required_decisions":["RELEASE_RISK"]},"changed_paths":["change.go"],"tests":[],"reviews":[],"decisions":[]}`)
	if err:=os.WriteFile(path,body,0o600);err!=nil{t.Fatal(err)}
	var stdout,stderr bytes.Buffer
	if code:=run([]string{"release-gate","--input",path},&stdout,&stderr);code!=1||!bytes.Contains(stderr.Bytes(),[]byte("required test")){t.Fatalf("expected fail-closed gate, got %d: %s",code,stderr.String())}
}
