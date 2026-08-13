package tests_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"clean-code/internal/history"
	"clean-code/internal/incremental"
	"clean-code/internal/study"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func compileSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var resource any
	if err := json.Unmarshal(body, &resource); err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", resource); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
func validateJSON(t *testing.T, schema *jsonschema.Schema, body []byte) error {
	t.Helper()
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return err
	}
	return schema.Validate(value)
}
func TestStudySchemasRejectAdversarialNestedValues(t *testing.T) {
	root := filepath.Join("..")
	manifest := compileSchema(t, filepath.Join(root, "harness/schemas/study-manifest.schema.json"))
	if err := validateJSON(t, manifest, []byte(`{"schema_version":"1.0.0","study_id":"s","repository":"o/r","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","case_corpus_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","oracle_corpus_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","model_config_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","preregistration_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","minimum_pairs":1,"tasks":[{}]}`)); err == nil {
		t.Fatal("tasks:[{}] must fail")
	}
	if err := validateJSON(t, manifest, []byte(`{"schema_version":"1.0.0","study_id":"s","repository":"o/r","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","case_corpus_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","oracle_corpus_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","model_config_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","preregistration_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","minimum_pairs":1,"tasks":[{"id":"t","model":"m","tools":[1],"limit":1,"oracle":"o"}]}`)); err == nil {
		t.Fatal("tools:[1] must fail")
	}
	result := compileSchema(t, filepath.Join(root, "harness/schemas/study-result.schema.json"))
	if err := validateJSON(t, result, []byte(`{"schema_version":"1.0.0","study_id":"s","manifest_digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","outcomes":[{"repository":"o/r","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","run_id":"r","artifact_digest":"d","started_at":"2026-08-12T00:00:00Z","finished_at":"2026-08-12T00:01:00Z","task_id":"t","arm":"control","model":"m","tools":[1],"limit":1,"oracle":"o","status":"PASS","false_positives":0,"correct_silence":false}]}`)); err == nil {
		t.Fatal("nested result tools:[1] must fail")
	}
	if err := validateJSON(t, manifest, []byte(`{"schema_version":"1.0.0","study_id":"s","repository":"o/r","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","minimum_pairs":1,"tasks":[{"id":"t","model":"m","tools":["none"],"limit":1,"oracle":"o"}]}`)); err == nil {
		t.Fatal("missing corpus and config digests must fail")
	}
}
func TestCheckedStudyFixturesValidateAndParseRuntime(t *testing.T) {
	root := filepath.Join("..")
	manifestBody, err := os.ReadFile(filepath.Join(root, "harness/studies/held-out-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	resultBody, err := os.ReadFile(filepath.Join(root, "harness/studies/valid-study-result.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateJSON(t, compileSchema(t, filepath.Join(root, "harness/schemas/study-manifest.schema.json")), manifestBody); err != nil {
		t.Fatal(err)
	}
	if err := validateJSON(t, compileSchema(t, filepath.Join(root, "harness/schemas/study-result.schema.json")), resultBody); err != nil {
		t.Fatal(err)
	}
	manifest, err := study.ParseManifest(manifestBody)
	if err != nil {
		t.Fatal(err)
	}
	results, err := study.ParseResults(resultBody)
	if err != nil {
		t.Fatal(err)
	}
	if results.ManifestDigest != study.Digest(manifestBody) {
		t.Fatal("fixture manifest digest mismatch")
	}
	report, err := study.Score(manifest, results)
	if err != nil {
		t.Fatal(err)
	}
	if report.ClaimAllowed {
		t.Fatal("empty raw results must remain valid but claim-blocked")
	}
}
func TestCheckedStudyModelConfigPinsSnapshot(t *testing.T) {
	root := filepath.Join("..")
	body, err := os.ReadFile(filepath.Join(root, "harness/studies/held-out-v1/model-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateJSON(t, compileSchema(t, filepath.Join(root, "harness/schemas/study-model-config.schema.json")), body); err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(body, &config); err != nil {
		t.Fatal(err)
	}
	if config["model"] != "gpt-5-2025-08-07" || config["store"] != false {
		t.Fatalf("model config is not immutable and non-retained: %+v", config)
	}
}

func TestCheckedStudyCorpusAndPreregistrationValidate(t *testing.T) {
	root := filepath.Join("..")
	casesBody, err := os.ReadFile(filepath.Join(root, "harness/studies/held-out-v1/cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	preBody, err := os.ReadFile(filepath.Join(root, "harness/studies/held-out-v1/preregistration.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateJSON(t, compileSchema(t, filepath.Join(root, "harness/schemas/study-case-corpus.schema.json")), casesBody); err != nil {
		t.Fatal(err)
	}
	if err := validateJSON(t, compileSchema(t, filepath.Join(root, "harness/schemas/study-preregistration.schema.json")), preBody); err != nil {
		t.Fatal(err)
	}
	var prereg struct {
		Claims struct {
			StudyType               string `json:"study_type"`
			ComparativeClaimAllowed bool   `json:"comparative_claim_allowed"`
		} `json:"claims"`
		Commitments struct {
			Cases  string `json:"case_corpus_sha256"`
			Oracle string `json:"oracle_scoring_sha256"`
		} `json:"commitments"`
	}
	if err := json.Unmarshal(preBody, &prereg); err != nil {
		t.Fatal(err)
	}
	if prereg.Commitments.Cases != study.Digest(casesBody) {
		t.Fatal("case corpus commitment mismatch")
	}
	if len(prereg.Commitments.Oracle) != 64 {
		t.Fatal("oracle commitment is not SHA-256")
	}
	if prereg.Claims.StudyType != "descriptive_pilot" || prereg.Claims.ComparativeClaimAllowed {
		t.Fatal("descriptive pilot must block comparative claims")
	}
}

func TestStudyPreregistrationSchemaRejectsEmptyContracts(t *testing.T) {
	schema := compileSchema(t, filepath.Join("..", "harness/schemas/study-preregistration.schema.json"))
	body := []byte(`{"schema_version":"1.0.0","study_id":"s","repository":"o/r","target_revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","corpus":{},"scoring":{},"canonicalization":{},"commitments":{"case_corpus_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","oracle_scoring_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},"exclusions":{}}`)
	if err := validateJSON(t, schema, body); err == nil {
		t.Fatal("empty preregistration contracts must fail")
	}
}

func TestHistoryReceiptSchemaRejectsMalformedNestedContent(t *testing.T) {
	schema := compileSchema(t, filepath.Join("..", "harness/schemas/history-receipt.schema.json"))
	cases := []string{
		`{"schema_version":"1.0.0","digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","content":{}}`,
		`{"schema_version":"1.0.0","digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","content":{"repository":1,"revision":"r","created_at":"t","signals":[]}}`,
		`{"schema_version":"1.0.0","digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","content":{"repository":"o/r","revision":"r","created_at":"t","signals":[1]}}`,
	}
	for _, body := range cases {
		if err := validateJSON(t, schema, []byte(body)); err == nil {
			t.Fatalf("expected rejection: %s", body)
		}
	}
	if _, err := history.ParseReceipt([]byte(cases[0])); err == nil {
		t.Fatal("runtime must reject omitted signals")
	}
}
func TestHistoryReceiptSchemaParseAndBuildFixture(t *testing.T) {
	content := history.Content{Repository: "o/r", Revision: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", CreatedAt: "2026-08-12T00:00:00Z", Signals: []history.Signal{{Name: "coverage", Value: 80, Scale: "percent", Provenance: "ci"}}}
	body, err := json.Marshal(history.Receipt{SchemaVersion: "1.0.0", Digest: history.Digest(content), Content: content})
	if err != nil {
		t.Fatal(err)
	}
	if err := validateJSON(t, compileSchema(t, filepath.Join("..", "harness/schemas/history-receipt.schema.json")), body); err != nil {
		t.Fatal(err)
	}
	receipt, err := history.ParseReceipt(body)
	if err != nil {
		t.Fatal(err)
	}
	report, err := history.Build([]history.Receipt{receipt})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Trends) != 1 || len(report.Trends[0].Points) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
}
func TestIncrementalSchemasRejectMalformedNestedValues(t *testing.T) {
	paths := []string{filepath.Join("..", "harness/schemas/incremental-input.schema.json"), filepath.Join("..", "harness/schemas/impact-map.schema.json")}
	cases := []string{
		`{"schema_version":"1.0.0","changed_paths":[1],"trusted_checks":["x"],"rules":[],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":[1],"rules":[],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":["   "],"rules":[],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":["x"],"rules":[{"check_id":"   ","patterns":["*.go"]}],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":["x"],"rules":[{"check_id":"x","patterns":[" \t "]}],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":["x"],"rules":[{}],"release":false}`,
	}
	for _, path := range paths {
		schema := compileSchema(t, path)
		for _, body := range cases {
			if err := validateJSON(t, schema, []byte(body)); err == nil {
				t.Fatalf("%s expected rejection: %s", path, body)
			}
		}
	}
}
func TestIncrementalSchemasLoadAndSelectFixture(t *testing.T) {
	body := []byte(`{"schema_version":"1.0.0","changed_paths":["internal/a.go"],"trusted_checks":["go-test"],"rules":[{"check_id":"go-test","patterns":["internal/*.go"]}],"release":false}`)
	for _, name := range []string{"incremental-input.schema.json", "impact-map.schema.json"} {
		if err := validateJSON(t, compileSchema(t, filepath.Join("..", "harness/schemas", name)), body); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	input, err := incremental.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	selection, err := incremental.Select(input)
	if err != nil {
		t.Fatal(err)
	}
	if selection.Mode != "incremental" || len(selection.Checks) != 1 || selection.Checks[0] != "go-test" {
		t.Fatalf("unexpected selection: %+v", selection)
	}
}
