package tests_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRuntimeChecksum(t *testing.T) {
	repositoryRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "bin/runtime.checksum.test.js")
	command.Dir = repositoryRoot
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("runtime checksum test failed: %v\n%s", err, output)
	}
}
