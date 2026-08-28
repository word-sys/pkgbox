package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type EntryConfig struct {
	AppID       string
	Name        string
	Comment     string
	ExecPath    string
	ExecArgs    string
	IconPath    string
	Categories  []string
	Terminal    bool
	ExtraFields map[string]string
}

func GenerateDesktopEntryContent(cfg EntryConfig) string {
	var sb strings.Builder

	sb.WriteString("[Desktop Entry]\n")
	sb.WriteString("Type=Application\n")
	sb.WriteString("Version=1.1\n")
	sb.WriteString(fmt.Sprintf("Name=%s\n", cfg.Name))

	if cfg.Comment != "" {
		sb.WriteString(fmt.Sprintf("Comment=%s\n", cfg.Comment))
	}

	execLine := cfg.ExecPath
	if cfg.ExecArgs != "" {
		execLine = fmt.Sprintf("%s %s", cfg.ExecPath, cfg.ExecArgs)
	}
	sb.WriteString(fmt.Sprintf("Exec=%s\n", execLine))

	if cfg.IconPath != "" {
		sb.WriteString(fmt.Sprintf("Icon=%s\n", cfg.IconPath))
	} else {
		sb.WriteString("Icon=application-x-executable\n")
	}

	if cfg.Terminal {
		sb.WriteString("Terminal=true\n")
	} else {
		sb.WriteString("Terminal=false\n")
	}

	if len(cfg.Categories) > 0 {
		cats := strings.Join(cfg.Categories, ";") + ";"
		sb.WriteString(fmt.Sprintf("Categories=%s\n", cats))
	} else {
		sb.WriteString("Categories=Utility;\n")
	}

	sb.WriteString("StartupNotify=true\n")
	sb.WriteString("X-PkgBox-Managed=true\n")

	for k, v := range cfg.ExtraFields {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	return sb.String()
}

func GetUserApplicationsDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	appDir := filepath.Join(dataHome, "applications")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return appDir, nil
}

func InstallDesktopEntry(cfg EntryConfig) (string, error) {
	appDir, err := GetUserApplicationsDir()
	if err != nil {
		return "", err
	}

	filename := sanitizeAppID(cfg.AppID)
	if !strings.HasSuffix(filename, ".desktop") {
		filename += ".desktop"
	}
	targetPath := filepath.Join(appDir, filename)

	content := GenerateDesktopEntryContent(cfg)
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return "", err
	}

	UpdateDesktopDatabase()
	return targetPath, nil
}

func RemoveDesktopEntry(appID string) error {
	appDir, err := GetUserApplicationsDir()
	if err != nil {
		return err
	}

	filename := sanitizeAppID(appID)
	if !strings.HasSuffix(filename, ".desktop") {
		filename += ".desktop"
	}
	targetPath := filepath.Join(appDir, filename)

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	UpdateDesktopDatabase()
	return nil
}

func UpdateDesktopDatabase() {
	appDir, err := GetUserApplicationsDir()
	if err != nil {
		return
	}
	cmd := exec.Command("update-desktop-database", appDir)
	_ = cmd.Run()
}

func sanitizeAppID(id string) string {
	id = strings.ToLower(id)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	id = strings.Trim(id, "-")
	if id == "" {
		return "pkgbox-app"
	}
	return id
}
