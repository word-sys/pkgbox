package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"pkgbox/internal/detector"
)

func TestInstallUserSpaceApp(t *testing.T) {
	tempDataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tempDataHome)

	// Create a dummy AppImage file
	sourceDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "calculator.AppImage")
	dummyContent := []byte("#!/bin/sh\necho 'running calculator'\n")
	if err := os.WriteFile(sourceFile, dummyContent, 0755); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	info := &detector.FileInfo{
		Path:          sourceFile,
		FileName:      "calculator.AppImage",
		AppName:       "Calculator",
		Size:          int64(len(dummyContent)),
		FormattedSize: "32 B",
		Type:          detector.TypeAppImage,
		Arch:          "x86_64 (64-bit)",
		IsExecutable:  true,
	}

	var progressStages []string
	result, err := InstallUserSpaceApp(info, func(stage string, fraction float64) {
		progressStages = append(progressStages, stage)
	})
	if err != nil {
		t.Fatalf("failed to install app: %v", err)
	}

	if result.AppName != "Calculator" {
		t.Errorf("expected app name Calculator, got %q", result.AppName)
	}

	// Verify installed binary exists and has execute bit
	destStat, err := os.Stat(result.InstallPath)
	if err != nil {
		t.Fatalf("installed binary does not exist at %q: %v", result.InstallPath, err)
	}
	if (destStat.Mode().Perm() & 0111) == 0 {
		t.Errorf("installed binary is missing executable permissions: %v", destStat.Mode())
	}

	// Verify desktop file exists
	if _, err := os.Stat(result.DesktopPath); err != nil {
		t.Fatalf("desktop file does not exist at %q: %v", result.DesktopPath, err)
	}

	if len(progressStages) == 0 {
		t.Errorf("expected progress callbacks, got none")
	}
}

func TestInstallUserSpaceApp_NilInfo(t *testing.T) {
	_, err := InstallUserSpaceApp(nil, nil)
	if err == nil {
		t.Fatalf("expected error for nil info, got nil")
	}
}

func TestInstallUserSpaceApp_Archive(t *testing.T) {
	tempDataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tempDataHome)

	sourceDir := t.TempDir()
	tarPath := filepath.Join(sourceDir, "portable-tool.tar.gz")

	elfHeader := []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00}
	archiveDir := filepath.Join(sourceDir, "raw")
	_ = os.MkdirAll(filepath.Join(archiveDir, "bin"), 0755)
	_ = os.WriteFile(filepath.Join(archiveDir, "bin", "portable-tool"), elfHeader, 0755)

	cmd := exec.Command("tar", "-czf", tarPath, "-C", archiveDir, ".")
	if err := cmd.Run(); err != nil {
		t.Skip("tar command not available for test archive creation")
	}

	info := &detector.FileInfo{
		Path:          tarPath,
		FileName:      "portable-tool.tar.gz",
		AppName:       "Portable Tool",
		Type:          detector.TypeArchive,
		IsExecutable:  true,
		FormattedSize: "1 KB",
	}

	result, err := InstallUserSpaceApp(info, nil)
	if err != nil {
		t.Fatalf("unexpected error installing archive: %v", err)
	}

	if result.AppName != "Portable Tool" {
		t.Errorf("expected AppName 'Portable Tool', got %q", result.AppName)
	}

	if _, err := os.Stat(result.InstallPath); err != nil {
		t.Errorf("expected installed binary at %q, err: %v", result.InstallPath, err)
	}

	if _, err := os.Stat(result.DesktopPath); err != nil {
		t.Errorf("expected desktop entry at %q, err: %v", result.DesktopPath, err)
	}
}
