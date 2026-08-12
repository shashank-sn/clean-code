package tests_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"clean-code/internal/study"
	"clean-code/internal/history"
	"clean-code/internal/incremental"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func compileSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()
	body,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
	var resource any;if err:=json.Unmarshal(body,&resource);err!=nil{t.Fatal(err)}
	compiler:=jsonschema.NewCompiler()
	if err:=compiler.AddResource("schema.json",resource);err!=nil{t.Fatal(err)}
	schema,err:=compiler.Compile("schema.json");if err!=nil{t.Fatal(err)}
	return schema
}
func validateJSON(t *testing.T,schema *jsonschema.Schema,body []byte)error{t.Helper();var value any;if err:=json.Unmarshal(body,&value);err!=nil{return err};return schema.Validate(value)}
func TestStudySchemasRejectAdversarialNestedValues(t *testing.T){
	root:=filepath.Join("..")
	manifest:=compileSchema(t,filepath.Join(root,"harness/schemas/study-manifest.schema.json"))
	if err:=validateJSON(t,manifest,[]byte(`{"schema_version":"1.0.0","study_id":"s","repository":"o/r","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","minimum_pairs":1,"tasks":[{}]}`));err==nil{t.Fatal("tasks:[{}] must fail")}
	if err:=validateJSON(t,manifest,[]byte(`{"schema_version":"1.0.0","study_id":"s","repository":"o/r","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","minimum_pairs":1,"tasks":[{"id":"t","model":"m","tools":[1],"limit":1,"oracle":"o"}]}`));err==nil{t.Fatal("tools:[1] must fail")}
	result:=compileSchema(t,filepath.Join(root,"harness/schemas/study-result.schema.json"))
	if err:=validateJSON(t,result,[]byte(`{"schema_version":"1.0.0","study_id":"s","manifest_digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","outcomes":[{"repository":"o/r","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","run_id":"r","artifact_digest":"d","started_at":"2026-08-12T00:00:00Z","finished_at":"2026-08-12T00:01:00Z","task_id":"t","arm":"control","model":"m","tools":[1],"limit":1,"oracle":"o","status":"PASS","false_positives":0,"correct_silence":false}]}`));err==nil{t.Fatal("nested result tools:[1] must fail")}
}
func TestCheckedStudyFixturesValidateAndParseRuntime(t *testing.T){
	root:=filepath.Join("..");manifestBody,err:=os.ReadFile(filepath.Join(root,"harness/studies/held-out-v1.json"));if err!=nil{t.Fatal(err)};resultBody,err:=os.ReadFile(filepath.Join(root,"harness/studies/valid-study-result.json"));if err!=nil{t.Fatal(err)}
	if err:=validateJSON(t,compileSchema(t,filepath.Join(root,"harness/schemas/study-manifest.schema.json")),manifestBody);err!=nil{t.Fatal(err)}
	if err:=validateJSON(t,compileSchema(t,filepath.Join(root,"harness/schemas/study-result.schema.json")),resultBody);err!=nil{t.Fatal(err)}
	manifest,err:=study.ParseManifest(manifestBody);if err!=nil{t.Fatal(err)};results,err:=study.ParseResults(resultBody);if err!=nil{t.Fatal(err)}
	if results.ManifestDigest!=study.Digest(manifestBody){t.Fatal("fixture manifest digest mismatch")}
	report,err:=study.Score(manifest,results);if err!=nil{t.Fatal(err)};if report.ClaimAllowed{t.Fatal("empty raw results must remain valid but claim-blocked")}
}

func TestHistoryReceiptSchemaRejectsMalformedNestedContent(t *testing.T){
	schema:=compileSchema(t,filepath.Join("..","harness/schemas/history-receipt.schema.json"))
	cases:=[]string{
		`{"schema_version":"1.0.0","digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","content":{}}`,
		`{"schema_version":"1.0.0","digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","content":{"repository":1,"revision":"r","created_at":"t","signals":[]}}`,
		`{"schema_version":"1.0.0","digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","content":{"repository":"o/r","revision":"r","created_at":"t","signals":[1]}}`,
	}
	for _,body:=range cases{if err:=validateJSON(t,schema,[]byte(body));err==nil{t.Fatalf("expected rejection: %s",body)}}
	if _,err:=history.ParseReceipt([]byte(cases[0]));err!=nil{t.Fatalf("runtime parse should remain structural; Build enforces completeness: %v",err)}
}
func TestIncrementalSchemaRejectsMalformedNestedValues(t *testing.T){
	schema:=compileSchema(t,filepath.Join("..","harness/schemas/incremental-input.schema.json"))
	cases:=[]string{
		`{"schema_version":"1.0.0","changed_paths":[1],"trusted_checks":["x"],"rules":[],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":[1],"rules":[],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":[""],"rules":[],"release":false}`,
		`{"schema_version":"1.0.0","changed_paths":[],"trusted_checks":["x"],"rules":[{}],"release":false}`,
	}
	for _,body:=range cases{if err:=validateJSON(t,schema,[]byte(body));err==nil{t.Fatalf("expected rejection: %s",body)}}
	if _,err:=incremental.Select(incremental.Input{SchemaVersion:"1.0.0",TrustedChecks:[]string{""}});err==nil{t.Fatal("runtime must reject blank trusted id")}
}
