package desktop

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDesktopEntryContent(t *testing.T) {
	cfg := EntryConfig{
		AppID:      "org.example.myapp",
		Name:       "My App",
		Comment:    "Test Application",
		ExecPath:   "/home/user/.local/share/pkgbox/apps/myapp/myapp.AppImage",
		ExecArgs:   "%U",
		IconPath:   "/home/user/.local/share/icons/myapp.png",
		Categories: []string{"Utility", "Development"},
		Terminal:   false,
	}

	content := GenerateDesktopEntryContent(cfg)

	requiredStrings := []string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=My App",
		"Comment=Test Application",
		"Exec=/home/user/.local/share/pkgbox/apps/myapp/myapp.AppImage %U",
		"Icon=/home/user/.local/share/icons/myapp.png",
		"Categories=Utility;Development;",
		"Terminal=false",
		"X-PkgBox-Managed=true",
	}

	for _, req := range requiredStrings {
		if !strings.Contains(content, req) {
			t.Errorf("missing expected string %q in generated desktop entry:\n%s", req, content)
		}
	}
}

func TestInstallAndRemoveDesktopEntry(t *testing.T) {
	tempDataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tempDataHome)

	cfg := EntryConfig{
		AppID:    "test-app",
		Name:     "Test App",
		ExecPath: "/bin/true",
	}

	installedPath, err := InstallDesktopEntry(cfg)
	if err != nil {
		t.Fatalf("failed to install desktop entry: %v", err)
	}

	expectedPath := filepath.Join(tempDataHome, "applications", "test-app.desktop")
	if installedPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, installedPath)
	}

	if _, err := os.Stat(installedPath); err != nil {
		t.Errorf("installed desktop entry does not exist on disk: %v", err)
	}

	// Validate using desktop-file-validate if available
	if validateCmd, err := exec.LookPath("desktop-file-validate"); err == nil {
		out, err := exec.Command(validateCmd, installedPath).CombinedOutput()
		if err != nil {
			t.Errorf("desktop-file-validate failed: %v\nOutput: %s", err, string(out))
		}
	}

	if err := RemoveDesktopEntry("test-app"); err != nil {
		t.Fatalf("failed to remove desktop entry: %v", err)
	}

	if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
		t.Errorf("desktop entry was not deleted")
	}
}

func TestInstallIcon(t *testing.T) {
	tempDataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tempDataHome)

	dummyIcon := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	iconPath, err := InstallIcon("sample-app", dummyIcon, ".png")
	if err != nil {
		t.Fatalf("failed to install icon: %v", err)
	}

	expectedIconPath := filepath.Join(tempDataHome, "icons", "hicolor", "128x128", "apps", "sample-app.png")
	if iconPath != expectedIconPath {
		t.Errorf("expected icon path %q, got %q", expectedIconPath, iconPath)
	}

	if _, err := os.Stat(iconPath); err != nil {
		t.Errorf("icon file does not exist: %v", err)
	}
}
