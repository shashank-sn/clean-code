package access

import "crypto/subtle"
type Capability struct{ Tenant, Resource, Token string }
func CanRead(c Capability, tenant, resource, token string) bool {
	return c.Tenant == tenant && c.Resource == resource && subtle.ConstantTimeCompare([]byte(c.Token), []byte(token)) == 1
}
