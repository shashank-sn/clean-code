package hosts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryHostWritesPortableInstructions(t *testing.T) {
	for _, host := range Catalog() {
		t.Run(host.ID, func(t *testing.T) {
			root := t.TempDir()
			path, err := WritePackage(root, host.ID)
			if err != nil {
				t.Fatal(err)
			}
			if relative, _ := filepath.Rel(root, path); relative != PackageTarget(host.ID) {
				t.Fatalf("unexpected package target %q", relative)
			}
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), "clean-code verify") || !strings.Contains(string(body), host.DisplayName) {
				t.Fatalf("package lacks canonical workflow: %s", body)
			}
			if _, err := WritePackage(root, host.ID); err == nil {
				t.Fatal("expected existing package rejection")
			}
		})
	}
}

func TestUnknownHostUsesGenericPackage(t *testing.T) {
	if PackageTarget("future-host") != "AGENTS.md" || !strings.Contains(Instructions("future-host"), "Generic coding environment") {
		t.Fatal("expected generic package fallback")
	}
}

func TestRuleHostsReceiveAlwaysOnFrontmatter(t *testing.T) {
	if !strings.HasPrefix(Instructions("cursor"), "---\ndescription:") || !strings.Contains(Instructions("cursor"), "alwaysApply: true") {
		t.Fatal("cursor package needs always-on MDC metadata")
	}
	if !strings.HasPrefix(Instructions("windsurf"), "---\ntrigger: always_on") {
		t.Fatal("windsurf package needs always-on rule metadata")
	}
}
