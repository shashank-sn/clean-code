package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"clean-code/internal/discover"
	"clean-code/internal/hosts"
)

func TestRunSetupUsesGenericFallback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"setup", "--host", "future-ide"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success, got %d: %s", code, stderr.String())
	}
	var result hosts.Capabilities
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.ID != "generic" || !result.CLI {
		t.Fatalf("unexpected fallback: %+v", result)
	}
}

func TestRunDiscoverWritesJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(root+"/go.mod", []byte("module sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"discover", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success, got %d: %s", code, stderr.String())
	}
	var result discover.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Languages) != 1 || result.Languages[0] != "go" {
		t.Fatalf("unexpected discovery result: %+v", result)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("expected usage error, got %d", code)
	}
}

func TestRunRejectsExtraArguments(t *testing.T) {
	tests := [][]string{
		{"version", "extra"},
		{"hosts", "extra"},
		{"setup", "extra"},
		{"discover", ".", "extra"},
	}
	for _, args := range tests {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Errorf("expected usage error for %v, got %d", args, code)
		}
	}
}
