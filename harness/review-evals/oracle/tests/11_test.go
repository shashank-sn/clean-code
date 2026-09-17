package decode

import "testing"
func TestCollectsAllFailures(t *testing.T) {
	values, err := ParseAll([]string{" 7 ", "bad", "also-bad", "9"})
	if err == nil { t.Fatal("expected aggregate error") }
	if len(values) != 2 || values[0] != 7 || values[1] != 9 { t.Fatalf("values=%v", values) }
	if got := err.Error(); len(got) == 0 || !contains(got, "item 1") || !contains(got, "item 2") { t.Fatalf("missing indexed failures: %q", got) }
}
func contains(s, sub string) bool { for i := 0; i+len(sub) <= len(s); i++ { if s[i:i+len(sub)] == sub { return true } }; return false }
