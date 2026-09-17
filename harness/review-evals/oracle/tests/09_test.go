package access

import "testing"
func TestCapabilityScope(t *testing.T) {
	c := Capability{Tenant: "a", Resource: "doc", Token: "secret"}
	if !CanRead(c, "a", "doc", "secret") { t.Fatal("valid capability denied") }
	if CanRead(c, "b", "doc", "secret") || CanRead(c, "a", "other", "secret") || CanRead(c, "a", "doc", "secrex") { t.Fatal("invalid capability accepted") }
}
