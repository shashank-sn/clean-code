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

func TestGeneratedAndCheckedInHostGuidanceCarryCanonicalProtocol(t *testing.T) {
	root := filepath.Join("..", "..")
	canonical := readFile(t, filepath.Join(root, "harness", "review", "protocol.md"))
	checkedIn := readFile(t, filepath.Join(root, "hosts", "generic", "AGENTS.md"))
	if section(checkedIn) != section(canonical) {
		t.Fatal("checked-in generic guidance diverges from canonical review protocol")
	}
	for _, hostID := range []string{"generic", "codex", "cursor", "windsurf"} {
		if section(Instructions(hostID)) != section(canonical) {
			t.Fatalf("generated %s guidance diverges from canonical review protocol", hostID)
		}
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func section(body string) string {
	const start = "<!-- review-protocol:start -->"
	const end = "<!-- review-protocol:end -->"
	from := strings.Index(body, start)
	to := strings.Index(body, end)
	if from < 0 || to < from {
		return ""
	}
	return body[from : to+len(end)]
}
