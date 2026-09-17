package command

import (
	"reflect"
	"testing"
)
func TestCommaForwardCompatibility(t *testing.T) {
	got := append([]string{"child"}, DecodeFlags("--region=us, --verbose")...)
	want := []string{"child", "--region=us", "--verbose"}
	if !reflect.DeepEqual(got, want) { t.Fatalf("got %#v want %#v", got, want) }
}
