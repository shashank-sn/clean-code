package tests_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"clean-code/internal/study"
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
