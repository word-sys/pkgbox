package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectFile_AppImage(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "sample_app")

	header := make([]byte, 64)
	copy(header[0:4], []byte{0x7f, 0x45, 0x4c, 0x46})
	header[18] = 0x3e
	header[19] = 0x00
	header[8] = 0x41
	header[9] = 0x49
	header[10] = 0x02

	if err := os.WriteFile(path, header, 0755); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	info, err := InspectFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Type != TypeAppImage {
		t.Errorf("expected TypeAppImage, got %v", info.Type)
	}
	if info.Arch != "x86_64 (64-bit)" {
		t.Errorf("expected x86_64 (64-bit), got %v", info.Arch)
	}
	if !info.IsExecutable {
		t.Errorf("expected IsExecutable to be true")
	}
	if len(info.SHA256) != 64 {
		t.Errorf("expected 64-char SHA256 string, got %q", info.SHA256)
	}
}

func TestInspectFile_Deb(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "package.deb")

	data := []byte("!<arch>\ndebian-binary   1600000000  0     0     100644  4         `\n4\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	info, err := InspectFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Type != TypeDeb {
		t.Errorf("expected TypeDeb, got %v", info.Type)
	}
}

func TestInspectFile_RPM(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "package.rpm")

	data := []byte{0xed, 0xab, 0xee, 0xdb, 0x03, 0x00, 0x00, 0x00}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	info, err := InspectFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Type != TypeRPM {
		t.Errorf("expected TypeRPM, got %v", info.Type)
	}
}

func TestInspectFile_FlathubRef(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "org.mozilla.firefox.flatpakref")

	content := `[Flatpak Ref]
Name=org.mozilla.firefox
Branch=stable
Title=Firefox
Url=https://dl.flathub.org/repo/
SuggestRemoteName=flathub
RuntimeRepo=https://flathub.org/repo/flathub.flatpakrepo
IsRuntime=false
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	info, err := InspectFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Type != TypeFlatpakRef {
		t.Errorf("expected TypeFlatpakRef, got %v", info.Type)
	}
	if info.AppID != "org.mozilla.firefox" {
		t.Errorf("expected AppID org.mozilla.firefox, got %v", info.AppID)
	}
	if info.AppName != "Firefox" {
		t.Errorf("expected AppName Firefox, got %v", info.AppName)
	}
	if info.Arch != "Flatpak (stable)" {
		t.Errorf("expected Flatpak (stable), got %v", info.Arch)
	}
	if info.Extra["RuntimeRepo"] != "https://flathub.org/repo/flathub.flatpakrepo" {
		t.Errorf("missing RuntimeRepo in Extra map")
	}
}

func TestInspectFile_Script(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "install.sh")

	content := "#!/usr/bin/env bash\necho 'hello'\n"
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	info, err := InspectFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Type != TypeScript {
		t.Errorf("expected TypeScript, got %v", info.Type)
	}
}

func TestInspectFile_DirectoryError(t *testing.T) {
	tempDir := t.TempDir()
	_, err := InspectFile(tempDir)
	if err == nil {
		t.Fatalf("expected error when inspecting directory, got nil")
	}
}
