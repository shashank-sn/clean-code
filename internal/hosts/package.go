package hosts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var packageTargets = map[string]string{
	"generic":     "AGENTS.md",
	"codex":       "AGENTS.md",
	"claude-code": "CLAUDE.md",
	"cursor":      filepath.Join(".cursor", "rules", "clean-code.mdc"),
	"copilot":     filepath.Join(".github", "copilot-instructions.md"),
	"gemini-cli":  "GEMINI.md",
	"windsurf":    filepath.Join(".windsurf", "rules", "clean-code.md"),
	"cline":       filepath.Join(".clinerules", "clean-code.md"),
	"roo-code":    filepath.Join(".roo", "rules", "clean-code.md"),
	"ide-agent":   "CLEAN_CODE.md",
}

func PackageTarget(id string) string {
	if target, ok := packageTargets[id]; ok {
		return target
	}
	return packageTargets["generic"]
}

func Instructions(id string) string {
	host := Resolve(id)
	prefix := ""
	switch host.ID {
	case "cursor":
		prefix = "---\ndescription: Apply the Clean Code evidence workflow to every change.\nalwaysApply: true\n---\n\n"
	case "windsurf":
		prefix = "---\ntrigger: always_on\n---\n\n"
	}
	return prefix + fmt.Sprintf(`# Clean Code workflow

Host: %s. Integration: %s.

1. Run clean-code discover before proposing repository checks. Discovery does not execute commands.
2. Record requirements, acceptance examples, and declared dependency boundaries before implementation.
3. Keep implementation, acceptance/UI testing, and independent review in separate contexts when possible.
4. Run clean-code verify against the final revision with an approved policy.
5. Preserve PASS, FAIL, NOT_AVAILABLE, NOT_CONFIGURED, NOT_RUN, and ERROR exactly.
6. Require evidence for review findings and permit a zero-finding review.
7. Record human requirement, acceptance, UI/QA, and code-sample checks before audit.

Use the standalone CLI whenever this host lacks a native capability. Host files may change invocation, never evidence or gate semantics.
`, host.DisplayName, host.Integration)
}

func WritePackage(root, id string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("host package output directory is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve host package output: %w", err)
	}
	if err := os.MkdirAll(absRoot, 0o755); err != nil {
		return "", fmt.Errorf("create host package output: %w", err)
	}
	if err := rejectSymlink(absRoot); err != nil {
		return "", err
	}
	target := filepath.Join(absRoot, PackageTarget(id))
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", fmt.Errorf("create host package directory: %w", err)
	}
	if err := rejectPathSymlinks(absRoot, parent); err != nil {
		return "", err
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return "", errors.New("host package target already exists")
		}
		return "", fmt.Errorf("create host package: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.Remove(target)
		}
	}()
	if _, err := file.WriteString(Instructions(id)); err != nil {
		file.Close()
		return "", fmt.Errorf("write host package: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close host package: %w", err)
	}
	complete = true
	return target, nil
}

func rejectPathSymlinks(root, target string) error {
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("host package target escapes output directory")
	}
	current := root
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		if err := rejectSymlink(current); err != nil {
			return err
		}
	}
	return nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect host package path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("host package paths cannot contain symlinks")
	}
	return nil
}
