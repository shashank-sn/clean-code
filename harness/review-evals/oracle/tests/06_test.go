package rules

import (
	"reflect"
	"testing"
)
func TestEscapedSeparator(t *testing.T) {
	got := Parse(`name=alpha\;beta;enabled=true`)
	want := []string{"name=alpha;beta", "enabled=true"}
	if !reflect.DeepEqual(got, want) { t.Fatalf("got %#v want %#v", got, want) }
}
