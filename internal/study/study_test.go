package study

import (
	"strings"
	"testing"
)

const rev = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const digest64 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func fixture() (Manifest, Results) {
	m := Manifest{SchemaVersion: "1.0.0", StudyID: "s", Repository: "o/r", Revision: rev, CaseCorpusDigest: digest64, OracleCorpusDigest: digest64, ModelConfigDigest: digest64, PreregistrationDigest: digest64, MinimumPairs: 1, Tasks: []Task{{ID: "t", Model: "m", Tools: []string{"x"}, Limit: 1, Oracle: "o"}}}
	out := func(arm, run, artifact string) Outcome {
		return Outcome{Repository: "o/r", Revision: rev, RunID: run, RequestDigest: Digest([]byte("request-" + arm)), ArtifactDigest: Digest([]byte(artifact)), StartedAt: "2026-08-12T00:00:00Z", FinishedAt: "2026-08-12T00:01:00Z", TaskID: "t", Arm: arm, Model: "m", Tools: []string{"x"}, Limit: 1, Oracle: "o", CaseCorpusDigest: digest64, OracleCorpusDigest: digest64, ModelConfigDigest: digest64, Status: "PASS"}
	}
	return m, Results{SchemaVersion: "1.0.0", StudyID: "s", ManifestDigest: "digest", Outcomes: []Outcome{out("control", "r1", "a1"), out("workflow", "r2", "a2")}}
}
func TestDistinctRunsAndArtifacts(t *testing.T) {
	m, r := fixture()
	r.Outcomes[1].RunID = "r1"
	if _, err := Score(m, r); err == nil {
		t.Fatal("expected same-run rejection")
	}
	m, r = fixture()
	r.Outcomes[1].ArtifactDigest = "a1"
	if _, err := Score(m, r); err == nil {
		t.Fatal("expected same-artifact rejection")
	}
}
func TestPerfectTieQualifiesExecutionButBlocksComparativeClaim(t *testing.T) {
	m, r := fixture()
	report, err := Score(m, r)
	if err != nil {
		t.Fatal(err)
	}
	if !report.ExecutionValid || report.ClaimAllowed {
		t.Fatalf("descriptive pilot report = %+v", report)
	}
}
func TestOrderedTimestamps(t *testing.T) {
	m, r := fixture()
	r.Outcomes[0].FinishedAt = "2026-08-11T00:00:00Z"
	if _, err := Score(m, r); err == nil {
		t.Fatal("expected timestamp rejection")
	}
}
func TestStrictResultsEOF(t *testing.T) {
	if _, err := ParseResults([]byte(`{"schema_version":"1.0.0","study_id":"s","manifest_digest":"d","outcomes":[]} {}`)); err == nil {
		t.Fatal("expected trailing rejection")
	}
}
func TestCanonicalManifestIdentity(t *testing.T) {
	body := []byte(`{"schema_version":"1.0.0","study_id":"s","repository":"bad","revision":"x","minimum_pairs":1,"tasks":[]}`)
	if _, err := ParseManifest(body); err == nil {
		t.Fatal("expected identity rejection")
	}
}
func TestManifestDigestChangesOnTaskRemovalOrMinimumChange(t *testing.T) {
	base := []byte(`{"tasks":[1,2],"minimum_pairs":2}`)
	removed := []byte(`{"tasks":[1],"minimum_pairs":2}`)
	lowered := []byte(`{"tasks":[1,2],"minimum_pairs":1}`)
	if Digest(base) == Digest(removed) || Digest(base) == Digest(lowered) {
		t.Fatal("manifest replay digest collision")
	}
}
func TestNegativeFalsePositive(t *testing.T) {
	m, r := fixture()
	r.Outcomes[0].FalsePositives = -1
	if _, err := Score(m, r); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("%v", err)
	}
}
func TestVerifyInputsRejectsCorpusAndConfigReplay(t *testing.T) {
	cases := []byte("cases")
	oracle := []byte("oracle")
	config := []byte("config")
	preregistration := []byte("preregistration")
	m, _ := fixture()
	m.CaseCorpusDigest = Digest(cases)
	m.OracleCorpusDigest = Digest(oracle)
	m.ModelConfigDigest = Digest(config)
	m.PreregistrationDigest = Digest(preregistration)
	if err := VerifyInputs(m, cases, oracle, config, preregistration); err != nil {
		t.Fatal(err)
	}
	if err := VerifyInputs(m, []byte("changed"), oracle, config, preregistration); err == nil {
		t.Fatal("expected changed case corpus rejection")
	}
	if err := VerifyInputs(m, cases, []byte("changed"), config, preregistration); err == nil {
		t.Fatal("expected changed oracle corpus rejection")
	}
	if err := VerifyInputs(m, cases, oracle, []byte("changed"), preregistration); err == nil {
		t.Fatal("expected changed model config rejection")
	}
	if err := VerifyInputs(m, cases, oracle, config, []byte("changed")); err == nil {
		t.Fatal("expected changed preregistration rejection")
	}
}
func TestScoreRejectsOutcomeBindingReplay(t *testing.T) {
	m, r := fixture()
	r.Outcomes[0].ModelConfigDigest = Digest([]byte("changed"))
	if _, err := Score(m, r); err == nil {
		t.Fatal("expected outcome config binding rejection")
	}
	m, r = fixture()
	r.Outcomes[0].CaseCorpusDigest = Digest([]byte("changed"))
	if _, err := Score(m, r); err == nil {
		t.Fatal("expected outcome case binding rejection")
	}
	m, r = fixture()
	r.Outcomes[0].OracleCorpusDigest = Digest([]byte("changed"))
	if _, err := Score(m, r); err == nil {
		t.Fatal("expected outcome oracle binding rejection")
	}
}
