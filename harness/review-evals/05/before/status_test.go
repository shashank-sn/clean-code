package resource
import "testing"
func TestVisible(t *testing.T) { if IsReady("pending") { t.Fatal("pending resource served") }; if !IsReady("ready") { t.Fatal("ready resource unavailable") } }
