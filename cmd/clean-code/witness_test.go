package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWitnessRoundTrip(t *testing.T) {
	dir := t.TempDir()
	receiptPath := filepath.Join(dir, "receipt with spaces.json")
	if err := os.WriteFile(receiptPath, []byte(`{"complete":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	witness := filepath.Join(dir, "witness.txt")
	if err := writeWitness(witness, receiptPath); err != nil {
		t.Fatalf("writeWitness: %v", err)
	}
	if err := verifyWitness(witness, receiptPath); err != nil {
		t.Fatalf("verifyWitness should pass for an unwitnessed-unchanged receipt: %v", err)
	}
}

func TestWitnessDetectsEditedReceipt(t *testing.T) {
	dir := t.TempDir()
	receiptPath := filepath.Join(dir, "receipt.json")
	if err := os.WriteFile(receiptPath, []byte(`{"complete":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	witness := filepath.Join(dir, "witness.txt")
	if err := writeWitness(witness, receiptPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(receiptPath, []byte(`{"complete":false}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyWitness(witness, receiptPath); err == nil {
		t.Fatal("expected witness verification to fail after the receipt was edited")
	} else if !strings.Contains(err.Error(), "regenerated or edited") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWitnessMissingRecordFails(t *testing.T) {
	dir := t.TempDir()
	receiptPath := filepath.Join(dir, "receipt.json")
	if err := os.WriteFile(receiptPath, []byte(`{"complete":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	witness := filepath.Join(dir, "empty-witness.txt")
	if err := os.WriteFile(witness, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyWitness(witness, receiptPath); err == nil {
		t.Fatal("expected witness verification to fail with no record")
	}
}
