package appimage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectAppImage_Type2WithSquashFS(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "sample.AppImage")

	// Create 128KB dummy ELF header with Type 2 AppImage magic
	data := make([]byte, 131072+1024)
	copy(data[0:4], []byte{0x7f, 0x45, 0x4c, 0x46})
	data[8] = 0x41
	data[9] = 0x49
	data[10] = 0x02

	// Put "sqsh" at offset 131072
	copy(data[131072:131076], []byte("sqsh"))

	if err := os.WriteFile(testPath, data, 0755); err != nil {
		t.Fatalf("failed to create test AppImage: %v", err)
	}

	structure, err := InspectAppImage(testPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if structure.Type != 2 {
		t.Errorf("expected AppImage Type 2, got %d", structure.Type)
	}
	if !structure.HasSquashFS {
		t.Errorf("expected HasSquashFS true, got false")
	}
	if structure.SquashFSOffset != 131072 {
		t.Errorf("expected SquashFSOffset 131072, got %d", structure.SquashFSOffset)
	}
}

func TestFindSquashFSOffset_NotFound(t *testing.T) {
	r := bytes.NewReader([]byte("plain data without squashfs signature"))
	_, err := FindSquashFSOffset(r)
	if err == nil {
		t.Fatalf("expected error when squashfs signature is missing, got nil")
	}
}
