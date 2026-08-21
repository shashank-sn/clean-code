package tests_test

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPackedNpmArtifactIncludesReferencedDocsAndBenchmark(t *testing.T) {
	repositoryRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	archivePath := packNpmArtifact(t, repositoryRoot, temp)
	packageRoot := extractNpmArtifact(t, archivePath, filepath.Join(temp, "extracted"))

	assertReadmeLinksResolve(t, packageRoot)
	for _, path := range []string{
		"skills/clean-build/SKILL.md",
		"skills/clean-build/agent.json",
		"examples/benchmark-flow/task.md",
		"examples/benchmark-flow/outcomes/ce/slug/slug.go",
		"examples/benchmark-flow/outcomes/cc/slug/slug.go",
		"harness/calibration/full-flow-manifest.json",
	} {
		if _, err := os.Stat(filepath.Join(packageRoot, path)); err != nil {
			t.Fatalf("packed artifact is missing %s: %v", path, err)
		}
	}

	command := exec.Command("node", "bin/clean-code.js", "agent", "validate")
	command.Dir = packageRoot
	command.Env = append(os.Environ(), "GOCACHE="+filepath.Join(temp, "go-cache"), "HOME="+filepath.Join(temp, "home"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("packed npm CLI agent validation failed: %v\n%s", err, output)
	}
	command = exec.Command("node", "bin/clean-code.js", "agent", "emit", "clean-build", "--mode", "prompt")
	command.Dir = packageRoot
	command.Env = append(os.Environ(), "GOCACHE="+filepath.Join(temp, "go-cache"), "HOME="+filepath.Join(temp, "home"))
	output, err = command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "# Clean Build") {
		t.Fatalf("packed npm CLI agent emit failed: %v\n%s", err, output)
	}
	command = exec.Command("node", "bin/clean-code.js", "benchmark-full-flow")
	command.Dir = packageRoot
	command.Env = append(os.Environ(), "GOCACHE="+filepath.Join(temp, "go-cache"), "HOME="+filepath.Join(temp, "home"))
	output, err = command.CombinedOutput()
	if err != nil {
		t.Fatalf("packed npm CLI benchmark-full-flow failed: %v\n%s", err, output)
	}
}

func packNpmArtifact(t *testing.T, repositoryRoot, temp string) string {
	t.Helper()
	command := exec.Command("npm", "pack", "--json", "--ignore-scripts", "--pack-destination", temp)
	command.Dir = repositoryRoot
	command.Env = append(os.Environ(), "npm_config_cache="+filepath.Join(temp, "npm-cache"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("npm pack failed: %v\n%s", err, output)
	}
	var packed []struct {
		Filename string `json:"filename"`
	}
	if err := json.Unmarshal(output, &packed); err != nil {
		t.Fatalf("parse npm pack output: %v\n%s", err, output)
	}
	if len(packed) != 1 || packed[0].Filename == "" {
		t.Fatalf("unexpected npm pack output: %s", output)
	}
	return filepath.Join(temp, packed[0].Filename)
}

func extractNpmArtifact(t *testing.T, archivePath, destination string) string {
	t.Helper()
	archive, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	compressed, err := gzip.NewReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()

	reader := tar.NewReader(compressed)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel("package", filepath.FromSlash(header.Name))
		if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Fatalf("unsafe package archive path %q", header.Name)
		}
		path := filepath.Join(destination, relative)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				t.Fatal(err)
			}
			_, copyErr := io.Copy(file, reader)
			closeErr := file.Close()
			if copyErr != nil {
				t.Fatal(copyErr)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
		default:
			t.Fatalf("unsupported package archive entry %q (%d)", header.Name, header.Typeflag)
		}
	}
	return destination
}

func assertReadmeLinksResolve(t *testing.T, packageRoot string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(packageRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	linkPattern := regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	for _, match := range linkPattern.FindAllStringSubmatch(string(body), -1) {
		target := strings.Split(match[1], "#")[0]
		if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		path := filepath.Join(packageRoot, filepath.FromSlash(target))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("packed README has broken link %q: %v", match[1], err)
		}
	}
}
