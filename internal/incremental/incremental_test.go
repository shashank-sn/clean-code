package incremental
import"testing"
func TestEmptyAndReleaseUseFull(t *testing.T){in:=Input{SchemaVersion:"1.0.0",TrustedChecks:[]string{"all"}};got,_:=Select(in);if got.Mode!="full"{t.Fatal(got)};in.ChangedPaths=[]string{"a.go"};in.Release=true;got,_=Select(in);if got.Mode!="full"{t.Fatal(got)}}
func TestKnownPathSelectsOnlyTrustedCheck(t *testing.T){in:=Input{SchemaVersion:"1.0.0",ChangedPaths:[]string{"a.go"},TrustedChecks:[]string{"go"},Rules:[]Rule{{CheckID:"go",Patterns:[]string{"*.go"}}}};got,err:=Select(in);if err!=nil||got.Mode!="incremental"||len(got.Checks)!=1{t.Fatalf("%+v %v",got,err)}}
func TestRejectsDuplicateTrustedAndEmptyPatterns(t *testing.T){in:=Input{SchemaVersion:"1.0.0",TrustedChecks:[]string{"x","x"}};if _,err:=Select(in);err==nil{t.Fatal("expected duplicate rejection")};in=Input{SchemaVersion:"1.0.0",TrustedChecks:[]string{"x"},Rules:[]Rule{{CheckID:"x",Patterns:[]string{""}}}};if _,err:=Select(in);err==nil{t.Fatal("expected empty pattern rejection")}}
