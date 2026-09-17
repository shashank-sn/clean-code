package export

import "testing"

type failingDestination struct{ calls int }
func (d *failingDestination) Put(v string) error { d.calls++; if d.calls == 2 { return errWrite }; return nil }
type sentinel string
func (e sentinel) Error() string { return string(e) }
var errWrite error = sentinel("write failed")

func TestPartialWriteIsReported(t *testing.T) {
	d := &failingDestination{}
	if err := WriteBatch([]string{"a", "b", "c"}, d); err == nil { t.Fatal("partial write was reported as success") }
	if d.calls != 2 { t.Fatalf("calls=%d, want stop at failing write", d.calls) }
}
