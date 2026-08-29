package hasher

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateSHA256Reader(t *testing.T) {
	input := []byte("PkgBox Universal Installer")
	// SHA256 of "PkgBox Universal Installer"
	expected := "3193c4bc62e6d99d13fc4ef7d20ee31089b766ddd71de4bb9dd3b62a6431c120"

	hash, err := CalculateSHA256Reader(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestCalculateSHA256_File(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.bin")
	content := []byte("PkgBox Universal Installer")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	expected := "3193c4bc62e6d99d13fc4ef7d20ee31089b766ddd71de4bb9dd3b62a6431c120"
	hash, err := CalculateSHA256(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestCalculateSHA256_NonExistentFile(t *testing.T) {
	_, err := CalculateSHA256("/path/to/nonexistent/file/12345")
	if err == nil {
		t.Fatalf("expected error for non-existent file, got nil")
	}
}
