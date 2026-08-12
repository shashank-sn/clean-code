package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestRunReleaseGateRequiresSeparateTrustedArtifacts(t *testing.T) {
	path:=filepath.Join(t.TempDir(),"release.json");if err:=os.WriteFile(path,[]byte("{}"),0o600);err!=nil{t.Fatal(err)}
	var stdout,stderr bytes.Buffer
	if code:=run([]string{"release-gate","--input",path},&stdout,&stderr);code!=2||!bytes.Contains(stderr.Bytes(),[]byte("--policy-gates")){t.Fatalf("expected trusted artifact requirement, got %d: %s",code,stderr.String())}
}

type releaseFixture struct{ input,gates,requirements,change,tests,reviews,decisions,approval,signature,key,root string }

func writeReleaseFixture(t *testing.T) releaseFixture {
	t.Helper();root:=t.TempDir();gates:=filepath.Join(root,"gates.json");requirements:=filepath.Join(root,"requirements.md");change:=filepath.Join(root,"change.json");tests:=filepath.Join(root,"tests.json");reviews:=filepath.Join(root,"reviews.json");decisions:=filepath.Join(root,"decisions.json");input:=filepath.Join(root,"binding.json");approvalPath:=filepath.Join(root,"approval.json");signature:=filepath.Join(root,"approval.sig");key:=filepath.Join(root,"approval.pub")
	write:=func(path,body string){if err:=os.WriteFile(path,[]byte(body),0o600);err!=nil{t.Fatal(err)}};digest:=func(body string)string{sum:=sha256.Sum256([]byte(body));return hex.EncodeToString(sum[:])}
	gatesBody:=`{"policy_revision":"policy","required_tests":["acceptance"],"required_reviews":["independent"],"required_decisions":["RELEASE_RISK"]}`;requirementsBody:="R1: release safely\n";changeBody:=`{"schema_version":"1.0.0","base_revision":"base","final_revision":"final","changed_paths":["change.go"]}`
	reqD,changeD,policyD:=digest(requirementsBody),digest(changeBody),digest(gatesBody)
	testsBody:=`{"tests":[{"id":"acceptance","repository":"owner/repo","run_id":"test-run","requirement_ids":["R1"],"revision":"final","status":"PASS","artifact_digest":"artifact","actor_run_id":"test-actor","started_at":"2026-08-12T00:00:00Z","finished_at":"2026-08-12T00:00:00Z"}]}`
	reviewsBody:=fmt.Sprintf(`{"reviews":[{"id":"independent","repository":"owner/repo","reviewer_run_id":"review-run","reviewer_context_id":"review-context","change_author_run_id":"build-run","change_author_context_id":"build-context","base_revision":"base","final_revision":"final","requirement_digest":"%s","change_set_digest":"%s","policy_revision":"%s","reviewed_paths":["change.go"],"status":"PASS"}]}`,reqD,changeD,policyD)
	decisionsBody:=fmt.Sprintf(`{"decisions":[{"repository":"owner/repo","final_revision":"final","requirement_digest":"%s","change_set_digest":"%s","policy_revision":"%s","kind":"RELEASE_RISK","authority":"human","subject":"final","status":"APPROVED"}]}`,reqD,changeD,policyD)
	write(gates,gatesBody);write(requirements,requirementsBody);write(change,changeBody);write(tests,testsBody);write(reviews,reviewsBody);write(decisions,decisionsBody)
	binding:=fmt.Sprintf(`{"schema_version":"1.0.0","repository":"owner/repo","base_revision":"base","final_revision":"final","requirement_digest":"%s","change_set_digest":"%s","policy_revision":"%s","test_attestations_digest":"%s","review_attestations_digest":"%s","decision_attestations_digest":"%s","changed_paths":["change.go"]}`,reqD,changeD,policyD,digest(testsBody),digest(reviewsBody),digest(decisionsBody));write(input,binding)
	manifest:=fmt.Sprintf(`{"schema_version":"1.0.0","repository":"owner/repo","final_revision":"final","binding_digest":"%s","policy_digest":"%s","requirements_digest":"%s","change_set_digest":"%s","test_attestations_digest":"%s","review_attestations_digest":"%s","decision_attestations_digest":"%s"}`,digest(binding),policyD,reqD,changeD,digest(testsBody),digest(reviewsBody),digest(decisionsBody));write(approvalPath,manifest)
	public,private,err:=ed25519.GenerateKey(rand.Reader);if err!=nil{t.Fatal(err)};write(signature,base64.StdEncoding.EncodeToString(ed25519.Sign(private,[]byte(manifest))));write(key,base64.StdEncoding.EncodeToString(public))
	return releaseFixture{input,gates,requirements,change,tests,reviews,decisions,approvalPath,signature,key,root}
}

func runReleaseFixture(f releaseFixture)(int,string){var out,err bytes.Buffer;code:=run([]string{"release-gate","--input",f.input,"--policy-gates",f.gates,"--requirements",f.requirements,"--change-set",f.change,"--root",f.root,"--repository","owner/repo","--test-attestations",f.tests,"--review-attestations",f.reviews,"--decision-attestations",f.decisions,"--approval-manifest",f.approval,"--approval-signature",f.signature,"--trusted-public-key",f.key},&out,&err);return code,err.String()}
func TestReleaseGateRejectsModifiedPolicyArtifact(t *testing.T){f:=writeReleaseFixture(t);if err:=os.WriteFile(f.gates,[]byte(`{"policy_revision":"changed","required_tests":["acceptance"],"required_reviews":["independent"],"required_decisions":["RELEASE_RISK"]}`),0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"trusted artifact digest"){t.Fatalf("got %d %s",code,msg)}}
func TestReleaseGateRejectsModifiedRequirementsArtifact(t *testing.T){f:=writeReleaseFixture(t);if err:=os.WriteFile(f.requirements,[]byte("R2: changed\n"),0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"trusted artifact digest"){t.Fatalf("got %d %s",code,msg)}}
func TestReleaseGateRejectsModifiedChangeSetArtifact(t *testing.T){f:=writeReleaseFixture(t);if err:=os.WriteFile(f.change,[]byte(`{"schema_version":"1.0.0","base_revision":"base","final_revision":"final","changed_paths":["other.go"]}`),0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"trusted artifact digest"){t.Fatalf("got %d %s",code,msg)}}
func TestReleaseGateRejectsRelabeledStaleBinding(t *testing.T){f:=writeReleaseFixture(t);body,err:=os.ReadFile(f.input);if err!=nil{t.Fatal(err)};changed:=bytes.Replace(body,[]byte(`"base_revision":"base"`),[]byte(`"base_revision":"stale"`),1);if err:=os.WriteFile(f.input,changed,0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"approval manifest does not match"){t.Fatalf("got %d %s",code,msg)}}
func TestReleaseGateRejectsWrongActualRootRevision(t *testing.T){f:=writeReleaseFixture(t);code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"repository is not a Git checkout"){t.Fatalf("got %d %s",code,msg)}}

func TestReleaseGateRejectsAdvancedFinalRevisionWithOldAttestations(t *testing.T){f:=writeReleaseFixture(t);body,_:=os.ReadFile(f.input);body=bytes.Replace(body,[]byte(`"final_revision":"final"`),[]byte(`"final_revision":"advanced"`),1);if err:=os.WriteFile(f.input,body,0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"approval manifest does not match"){t.Fatalf("expected signed stale-final rejection, got %d %s",code,msg)}}
func TestReleaseGateRejectsCrossRepositoryReplay(t *testing.T){f:=writeReleaseFixture(t);body,_:=os.ReadFile(f.input);body=bytes.Replace(body,[]byte(`"repository":"owner/repo"`),[]byte(`"repository":"fork/repo"`),1);if err:=os.WriteFile(f.input,body,0o600);err!=nil{t.Fatal(err)};var out,errOut bytes.Buffer;code:=run([]string{"release-gate","--input",f.input,"--policy-gates",f.gates,"--requirements",f.requirements,"--change-set",f.change,"--root",f.root,"--repository","fork/repo","--test-attestations",f.tests,"--review-attestations",f.reviews,"--decision-attestations",f.decisions,"--approval-manifest",f.approval,"--approval-signature",f.signature,"--trusted-public-key",f.key},&out,&errOut);if code!=1||!strings.Contains(errOut.String(),"approval manifest does not match"){t.Fatalf("expected cross-repo rejection, got %d %s",code,errOut.String())}}

func TestReleaseGateRejectsTamperedApprovalSignature(t *testing.T){f:=writeReleaseFixture(t);if err:=os.WriteFile(f.signature,[]byte(base64.StdEncoding.EncodeToString(make([]byte,ed25519.SignatureSize))),0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"signature verification failed"){t.Fatalf("got %d %s",code,msg)}}
func TestReleaseGateRejectsWrongTrustedKey(t *testing.T){f:=writeReleaseFixture(t);public,_,_:=ed25519.GenerateKey(rand.Reader);if err:=os.WriteFile(f.key,[]byte(base64.StdEncoding.EncodeToString(public)),0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"signature verification failed"){t.Fatalf("got %d %s",code,msg)}}
func TestReleaseGateRejectsConsistentForgeryWithOldSignature(t *testing.T){f:=writeReleaseFixture(t);body,_:=os.ReadFile(f.approval);body=bytes.Replace(body,[]byte(`"final_revision":"final"`),[]byte(`"final_revision":"forged"`),1);if err:=os.WriteFile(f.approval,body,0o600);err!=nil{t.Fatal(err)};code,msg:=runReleaseFixture(f);if code!=1||!strings.Contains(msg,"signature verification failed"){t.Fatalf("got %d %s",code,msg)}}
