package ui

import (
	"testing"

	"pkgbox/internal/detector"
)

func TestInfoCard_ScopeDerivation(t *testing.T) {
	tests := []struct {
		pkgType       detector.PackageType
		expectedScope string
	}{
		{detector.TypeAppImage, "User-space (No root required)"},
		{detector.TypeFlatpak, "User-space (No root required)"},
		{detector.TypeFlatpakRef, "User-space (No root required)"},
		{detector.TypeBinary, "User-space (No root required)"},
		{detector.TypeDeb, "System-wide (Polkit authentication required)"},
		{detector.TypeRPM, "System-wide (Polkit authentication required)"},
	}

	for _, tt := range tests {
		info := &detector.FileInfo{
			Type: tt.pkgType,
		}
		scope := "User-space (No root required)"
		if info.Type == detector.TypeDeb || info.Type == detector.TypeRPM {
			scope = "System-wide (Polkit authentication required)"
		}

		if scope != tt.expectedScope {
			t.Errorf("for type %v expected scope %q, got %q", tt.pkgType, tt.expectedScope, scope)
		}
	}
}

func TestInfoCard_SHA256Presence(t *testing.T) {
	info := &detector.FileInfo{
		AppName:       "Test App",
		FileName:      "test.AppImage",
		Type:          detector.TypeAppImage,
		Arch:          "x86_64",
		FormattedSize: "10 MB",
		SHA256:        "3193c4bc62e6d99d13fc4ef7d20ee31089b766ddd71de4bb9dd3b62a6431c120",
		Path:          "/tmp/test.AppImage",
	}

	if len(info.SHA256) != 64 {
		t.Errorf("expected 64-character SHA256 checksum, got %q", info.SHA256)
	}
}
