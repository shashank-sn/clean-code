package access

// Capabilities are loaded once at startup and are immutable for a request.
type Capability struct{ Tenant, Resource, Token string }
