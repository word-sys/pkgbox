package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pkgbox/internal/desktop"
	"pkgbox/internal/detector"
)

type ProgressCallback func(stage string, fraction float64)

type InstallResult struct {
	AppID       string
	AppName     string
	InstallPath string
	DesktopPath string
	InstalledAt time.Time
}

func GetAppsRootDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	appsDir := filepath.Join(dataHome, "pkgbox", "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return "", err
	}
	return appsDir, nil
}

func GetAppInstallDir(appID string) (string, error) {
	rootDir, err := GetAppsRootDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(rootDir, sanitizeID(appID))
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return appDir, nil
}

func InstallUserSpaceApp(info *detector.FileInfo, progress ProgressCallback) (*InstallResult, error) {
	if info == nil {
		return nil, fmt.Errorf("file info is nil")
	}

	appID := sanitizeID(info.AppName)
	if appID == "" {
		appID = sanitizeID(info.FileName)
	}

	if progress != nil {
		progress("Preparing destination directory...", 0.2)
	}

	appDir, err := GetAppInstallDir(appID)
	if err != nil {
		return nil, fmt.Errorf("failed to create install directory: %w", err)
	}

	destBinary := filepath.Join(appDir, filepath.Base(info.Path))

	if progress != nil {
		progress("Copying application binary...", 0.5)
	}

	if err := copyExecutableFile(info.Path, destBinary); err != nil {
		return nil, fmt.Errorf("failed to copy executable: %w", err)
	}

	if progress != nil {
		progress("Registering desktop launcher...", 0.8)
	}

	cfg := desktop.EntryConfig{
		AppID:    appID,
		Name:     info.AppName,
		Comment:  fmt.Sprintf("Installed via PkgBox (%s)", info.Type),
		ExecPath: destBinary,
		ExecArgs: "%U",
		Terminal: false,
	}

	desktopPath, err := desktop.InstallDesktopEntry(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to register desktop entry: %w", err)
	}

	if progress != nil {
		progress("Installation complete!", 1.0)
	}

	return &InstallResult{
		AppID:       appID,
		AppName:     info.AppName,
		InstallPath: destBinary,
		DesktopPath: desktopPath,
		InstalledAt: time.Now(),
	}, nil
}

func copyExecutableFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

func sanitizeID(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.Trim(name, "-")
	if name == "" {
		return "pkgbox-app"
	}
	return name
}
