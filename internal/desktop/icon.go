package desktop

import (
	"os"
	"path/filepath"
)

func GetUserIconsDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	iconDir := filepath.Join(dataHome, "icons", "hicolor", "128x128", "apps")
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		return "", err
	}
	return iconDir, nil
}

func InstallIcon(appID string, iconData []byte, ext string) (string, error) {
	if len(iconData) == 0 {
		return "application-x-executable", nil
	}

	iconDir, err := GetUserIconsDir()
	if err != nil {
		return "", err
	}

	if ext == "" {
		ext = ".png"
	}
	targetFile := filepath.Join(iconDir, sanitizeAppID(appID)+ext)
	if err := os.WriteFile(targetFile, iconData, 0644); err != nil {
		return "", err
	}

	return targetFile, nil
}
