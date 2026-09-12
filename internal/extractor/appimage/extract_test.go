package appimage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEmbeddedDesktopFile(t *testing.T) {
	tempDir := t.TempDir()
	desktopPath := filepath.Join(tempDir, "vlc.desktop")

	content := `[Desktop Entry]
Version=1.1
Name=VLC Media Player
Comment=Read, capture, and stream your multimedia streams
Exec=vlc %U
Icon=vlc
Terminal=false
Type=Application
Categories=AudioVideo;Player;Application;Recorder;
`
	if err := os.WriteFile(desktopPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed creating test desktop file: %v", err)
	}

	meta := &ExtractedMetadata{}
	parseEmbeddedDesktopFile(desktopPath, meta)

	if meta.Name != "VLC Media Player" {
		t.Errorf("expected Name 'VLC Media Player', got %q", meta.Name)
	}
	if meta.Comment != "Read, capture, and stream your multimedia streams" {
		t.Errorf("expected Comment mismatch, got %q", meta.Comment)
	}
	if len(meta.Categories) != 3 {
		t.Fatalf("expected 3 categories (excluding 'Application'), got %d: %v", len(meta.Categories), meta.Categories)
	}
	if meta.Categories[0] != "AudioVideo" || meta.Categories[1] != "Player" || meta.Categories[2] != "Recorder" {
		t.Errorf("unexpected categories: %v", meta.Categories)
	}
}

func TestFindEmbeddedIcon_DirIcon(t *testing.T) {
	tempDir := t.TempDir()
	dirIconPath := filepath.Join(tempDir, ".DirIcon")
	dummyPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	if err := os.WriteFile(dirIconPath, dummyPNG, 0644); err != nil {
		t.Fatalf("failed creating dummy icon: %v", err)
	}

	data, ext := findEmbeddedIcon(tempDir)
	if len(data) != len(dummyPNG) {
		t.Errorf("expected icon data length %d, got %d", len(dummyPNG), len(data))
	}
	if ext != ".png" {
		t.Errorf("expected extension .png, got %q", ext)
	}
}
