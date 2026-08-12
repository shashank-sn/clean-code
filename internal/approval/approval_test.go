package approval
import("strings";"testing")
const digest="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
func valid()[]byte{return []byte(`{"schema_version":"1.0.0","repository":"owner/repo","final_revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","binding_digest":"`+digest+`","policy_digest":"`+digest+`","requirements_digest":"`+digest+`","change_set_digest":"`+digest+`","test_attestations_digest":"`+digest+`","review_attestations_digest":"`+digest+`","decision_attestations_digest":"`+digest+`"}`)}
func TestParseValidatesManifest(t *testing.T){if _,err:=Parse(valid());err!=nil{t.Fatal(err)}}
func TestParseRejectsVersionAndMalformedIdentity(t *testing.T){body:=strings.Replace(string(valid()),`"1.0.0"`,`"2.0.0"`,1);if _,err:=Parse([]byte(body));err==nil{t.Fatal("expected version rejection")};body=strings.Replace(string(valid()),"owner/repo","bad",1);if _,err:=Parse([]byte(body));err==nil{t.Fatal("expected repo rejection")}}
func TestParseRejectsMalformedSHAAndDigest(t *testing.T){body:=strings.Replace(string(valid()),strings.Repeat("a",40),"ABC",1);if _,err:=Parse([]byte(body));err==nil{t.Fatal("expected sha rejection")};body=strings.Replace(string(valid()),digest,"bad",1);if _,err:=Parse([]byte(body));err==nil{t.Fatal("expected digest rejection")}}
