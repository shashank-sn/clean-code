package slug

import "testing"

func TestNormalizeSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "spaces", input: "Hello World", want: "hello-world"},
		{name: "trim", input: "  Foo__Bar!!  ", want: "foo-bar"},
		{name: "already clean", input: "already-clean", want: "already-clean"},
		{name: "only hyphens", input: "---", want: ""},
		{name: "empty", input: "", want: ""},
		{name: "whitespace only", input: "   ", want: ""},
		{name: "collapse hyphens", input: "a--b", want: "a-b"},
		{name: "leading trailing hyphens", input: "-hello-", want: "hello"},
		{name: "digits", input: "v2 final", want: "v2-final"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeSlug(tc.input); got != tc.want {
				t.Fatalf("NormalizeSlug(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func FuzzNormalizeSlug(f *testing.F) {
	f.Add("Hello World")
	f.Add("")
	f.Add("---")
	f.Fuzz(func(t *testing.T, input string) {
		got := NormalizeSlug(input)
		if got != "" {
			if got[0] == '-' || got[len(got)-1] == '-' {
				t.Fatalf("leading or trailing hyphen in %q", got)
			}
			if stringsContainsDoubleHyphen(got) {
				t.Fatalf("double hyphen in %q", got)
			}
		}
		for _, b := range []byte(got) {
			if b != '-' && (b < '0' || b > '9') && (b < 'a' || b > 'z') {
				t.Fatalf("invalid byte %q in %q", b, got)
			}
		}
	})
}

func stringsContainsDoubleHyphen(s string) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] == '-' && s[i+1] == '-' {
			return true
		}
	}
	return false
}
