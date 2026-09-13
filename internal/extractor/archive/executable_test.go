package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPrimaryExecutable_RootMatchingName(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy ELF executable named 'mygame'
	exePath := filepath.Join(tempDir, "mygame")
	elfHeader := []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00}
	if err := os.WriteFile(exePath, elfHeader, 0755); err != nil {
		t.Fatalf("failed creating test executable: %v", err)
	}

	// Create a shared library that should be ignored
	soPath := filepath.Join(tempDir, "libengine.so")
	if err := os.WriteFile(soPath, elfHeader, 0755); err != nil {
		t.Fatalf("failed creating test library: %v", err)
	}

	found, err := FindPrimaryExecutable(tempDir, "mygame")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found != exePath {
		t.Errorf("expected %q, got %q", exePath, found)
	}
}

func TestFindPrimaryExecutable_InBinDirectory(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(binDir, 0755)

	scriptPath := filepath.Join(binDir, "run.sh")
	scriptContent := []byte("#!/bin/sh\necho 'running'\n")
	if err := os.WriteFile(scriptPath, scriptContent, 0755); err != nil {
		t.Fatalf("failed creating script: %v", err)
	}

	found, err := FindPrimaryExecutable(tempDir, "super-tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found != scriptPath {
		t.Errorf("expected %q, got %q", scriptPath, found)
	}
}

func TestFindPrimaryExecutable_NoneFound(t *testing.T) {
	tempDir := t.TempDir()
	txtFile := filepath.Join(tempDir, "readme.txt")
	_ = os.WriteFile(txtFile, []byte("plain text"), 0644)

	_, err := FindPrimaryExecutable(tempDir, "app")
	if err == nil {
		t.Fatalf("expected error when no executable exists, got nil")
	}
}
