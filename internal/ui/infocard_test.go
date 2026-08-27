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
