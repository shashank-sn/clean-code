package resource
import "testing"
func TestVisible(t *testing.T) { if !IsReady("ready") { t.Fatal("ready resource unavailable") } }
