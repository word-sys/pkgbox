package installer

import (
	"os"
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
