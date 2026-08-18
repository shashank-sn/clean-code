package benchmarkflow_test

import (
	"testing"

	slugcc "clean-code/examples/benchmark-flow/outcomes/cc/slug"
	slugce "clean-code/examples/benchmark-flow/outcomes/ce/slug"
)

func TestCEOutcomeCoreCases(t *testing.T) {
	if got := slugce.NormalizeSlug("Hello World"); got != "hello-world" {
		t.Fatalf("ce: got %q", got)
	}
}

func TestCCOutcomeCoreCases(t *testing.T) {
	cases := map[string]string{
		"Hello World": "hello-world",
		"  Foo__Bar!!  ": "foo-bar",
		"---":         "",
	}
	for input, want := range cases {
		if got := slugcc.NormalizeSlug(input); got != want {
			t.Fatalf("cc: NormalizeSlug(%q) = %q want %q", input, got, want)
		}
	}
}
