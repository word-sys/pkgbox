package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseURIList_SingleFileWithSpaces(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "my test application.AppImage")
	if err := os.WriteFile(tempFile, []byte("test-data"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	encodedURI := "file://" + filepath.ToSlash(tempDir) + "/my%20test%20application.AppImage"
	raw := "# GNOME Nautilus DnD\r\n" + encodedURI + "\r\n"

	paths, err := ParseURIList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != tempFile {
		t.Errorf("expected %q, got %v", tempFile, paths)
	}
}

func TestParseURIList_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "app1.deb")
	file2 := filepath.Join(tempDir, "app2.rpm")
	_ = os.WriteFile(file1, []byte("1"), 0644)
	_ = os.WriteFile(file2, []byte("2"), 0644)

	raw := "file://" + file1 + "\r\nfile://" + file2 + "\r\n"
	paths, err := ParseURIList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 files, got %d", len(paths))
	}
	if paths[0] != file1 || paths[1] != file2 {
		t.Errorf("path mismatch: got %v", paths)
	}
}

func TestParseURIList_NonExistentFile(t *testing.T) {
	raw := "file:///tmp/this_file_does_not_exist_12345.AppImage\r\n"
	_, err := ParseURIList(raw)
	if err == nil {
		t.Fatalf("expected error for non-existent file, got nil")
	}
}

func TestParseURIList_DirectoryRejection(t *testing.T) {
	tempDir := t.TempDir()
	raw := "file://" + tempDir + "\r\n"
	_, err := ParseURIList(raw)
	if err == nil {
		t.Fatalf("expected error when dropping a directory, got nil")
	}
}
