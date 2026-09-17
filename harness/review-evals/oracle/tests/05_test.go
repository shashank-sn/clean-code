package resource

import "testing"
func TestOnlyReadyServed(t *testing.T) {
	for _, status := range []string{"pending", "failed", "disabled"} { if IsReady(status) { t.Fatalf("%s was served", status) } }
	if !IsReady("ready") { t.Fatal("ready was unavailable") }
}
