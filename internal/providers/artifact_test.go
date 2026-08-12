package providers

import (
	"math"
	"testing"
)

func TestParseJSONFlattensNumericMetrics(t *testing.T) {
	parsed, err := Parse("json", []byte(`{"coverage":{"lines":82.5},"passed":4,"files":[{"missed":2}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Metrics["coverage.lines"] != 82.5 || parsed.Metrics["passed"] != 4 || parsed.Metrics["files[0].missed"] != 2 {
		t.Fatalf("unexpected metrics: %#v", parsed.Metrics)
	}
}

func TestParseSARIFCountsLevels(t *testing.T) {
	parsed, err := Parse("sarif", []byte(`{"version":"2.1.0","runs":[{"results":[{"level":"error"},{"level":"warning"},{}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Metrics["results.total"] != 3 || parsed.Metrics["results.error"] != 1 || parsed.Metrics["results.warning"] != 2 {
		t.Fatalf("unexpected metrics: %#v", parsed.Metrics)
	}
}

func TestParseXMLFlattensNumericLeaves(t *testing.T) {
	parsed, err := Parse("xml", []byte(`<report tests="8"><coverage line-rate="0.75"><missed>2</missed></coverage></report>`))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Metrics["report.@tests"] != 8 || parsed.Metrics["report.coverage.@line-rate"] != 0.75 || parsed.Metrics["report.coverage.missed"] != 2 {
		t.Fatalf("unexpected metrics: %#v", parsed.Metrics)
	}
}

func TestParseLCOVAggregatesRecords(t *testing.T) {
	parsed, err := Parse("lcov", []byte("LF:10\nLH:8\nLF:10\nLH:9\nBRF:4\nBRH:3\n"))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(parsed.Metrics["lines.percent"]-85) > 0.001 || parsed.Metrics["branches.percent"] != 75 {
		t.Fatalf("unexpected metrics: %#v", parsed.Metrics)
	}
}

func TestParseRejectsMalformedArtifacts(t *testing.T) {
	for _, test := range []struct {
		format string
		body   string
	}{{"json", "{"}, {"json", `{"ok":1}{"extra":2}`}, {"sarif", `{}`}, {"xml", "<report>"}, {"xml", `<report score="NaN"/>`}, {"lcov", "TN:test"}, {"lcov", "LF:NaN"}} {
		if _, err := Parse(test.format, []byte(test.body)); err == nil {
			t.Errorf("expected malformed %s artifact to fail", test.format)
		}
	}
}
