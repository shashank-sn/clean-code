package labels

import (
	"reflect"
	"testing"
)
func TestCanonicalIdempotence(t *testing.T) {
	in := []string{"z", "a", "z"}; got := Normalize(in)
	if !reflect.DeepEqual(got, []string{"a", "z"}) { t.Fatalf("got %#v", got) }
	if !reflect.DeepEqual(in, []string{"z", "a", "z"}) { t.Fatalf("input mutated: %#v", in) }
	if !reflect.DeepEqual(Normalize(got), got) { t.Fatal("second write changed representation") }
}
