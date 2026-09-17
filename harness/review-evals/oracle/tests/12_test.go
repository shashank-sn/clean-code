package config

import (
	"errors"
	"testing"
)
type fakeSource struct{ workspace, environment string; workspaceErr, environmentErr error }
func (f fakeSource) Workspace(string) (string, error) { return f.workspace, f.workspaceErr }
func (f fakeSource) Environment(string) (string, error) { return f.environment, f.environmentErr }
func TestFallbackBoundary(t *testing.T) {
	if got, err := Resolve(fakeSource{workspaceErr: ErrNotFound, environment: "env"}, "k"); err != nil || got != "env" { t.Fatalf("fallback: %q %v", got, err) }
	want := errors.New("workspace unavailable")
	if _, err := Resolve(fakeSource{workspaceErr: want, environment: "env"}, "k"); !errors.Is(err, want) { t.Fatalf("lost workspace error: %v", err) }
}
