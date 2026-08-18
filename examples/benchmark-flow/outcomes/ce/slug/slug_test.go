package slug

import "testing"

func TestNormalizeSlug(t *testing.T) {
	if got := NormalizeSlug("Hello World"); got != "hello-world" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeSlug(""); got != "" {
		t.Fatalf("empty got %q", got)
	}
}
