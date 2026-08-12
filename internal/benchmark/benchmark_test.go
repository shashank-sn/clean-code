package benchmark

import "testing"

func TestScoreSeparatesDetectionFromFalsePositivesAndSilence(t *testing.T) {
	manifest := Manifest{SchemaVersion: "1.0.0", Cases: []Case{
		{ID: "caught", Class: "seeded-defect", Expected: []string{"R1"}, Observed: []string{"R1"}},
		{ID: "missed", Class: "seeded-defect", Expected: []string{"A1"}},
		{ID: "quiet", Class: "clean-control"},
		{ID: "noise", Class: "clean-control", Observed: []string{"style"}},
	}}
	if err := Validate(manifest); err != nil {
		t.Fatal(err)
	}
	report := Score(manifest)
	if report.TruePositives != 1 || report.FalsePositives != 1 || report.FalseNegatives != 1 || report.CorrectSilence != 1 {
		t.Fatalf("unexpected score: %+v", report)
	}
}

func TestCleanControlCannotDeclareExpectedDefect(t *testing.T) {
	manifest := Manifest{SchemaVersion: "1.0.0", Cases: []Case{{ID: "bad", Class: "clean-control", Expected: []string{"x"}}}}
	if err := Validate(manifest); err == nil {
		t.Fatal("expected invalid clean control")
	}
}
