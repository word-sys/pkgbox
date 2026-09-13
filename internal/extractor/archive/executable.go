package archive

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pkgbox/internal/desktop"
)

var elfMagic = []byte{0x7f, 0x45, 0x4c, 0x46}

func FindPrimaryExecutable(rootDir, appID string) (string, error) {
	appIDClean := strings.ToLower(sanitizeName(appID))

	// 1. Look for an embedded .desktop file first
	desktopFiles, _ := filepath.Glob(filepath.Join(rootDir, "*.desktop"))
	if len(desktopFiles) == 0 {
		desktopFiles, _ = filepath.Glob(filepath.Join(rootDir, "*", "*.desktop"))
	}
	if len(desktopFiles) > 0 {
		if execPath := parseDesktopExec(desktopFiles[0], rootDir); execPath != "" {
			if isRunnable(execPath) {
				_ = os.Chmod(execPath, 0755)
				return execPath, nil
			}
		}
	}

	var candidates []string

	// 2. Scan directory up to depth 3
	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}
		if strings.Count(rel, string(os.PathSeparator)) > 3 {
			return filepath.SkipDir
		}

		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, ".so") || strings.Contains(name, ".so.") || strings.HasSuffix(name, ".a") {
			return nil
		}

		if isRunnable(path) {
			candidates = append(candidates, path)
		}
		return nil
	})

	if len(candidates) == 0 {
		return "", fmt.Errorf("no executable binary or script found in extracted archive")
	}

	// 3. Rank candidates
	for _, c := range candidates {
		base := strings.ToLower(sanitizeName(filepath.Base(c)))
		if base == appIDClean {
			_ = os.Chmod(c, 0755)
			return c, nil
		}
	}

	for _, c := range candidates {
		if strings.Contains(c, "/bin/") || filepath.Dir(c) == filepath.Join(rootDir, "bin") {
			_ = os.Chmod(c, 0755)
			return c, nil
		}
	}

	// Pick first candidate
	_ = os.Chmod(candidates[0], 0755)
	return candidates[0], nil
}

func FindArchiveIcon(rootDir, appID string) string {
	var iconFiles []string
	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".svg") {
			iconFiles = append(iconFiles, path)
		}
		return nil
	})

	if len(iconFiles) == 0 {
		return ""
	}

	appIDClean := strings.ToLower(sanitizeName(appID))
	for _, iconPath := range iconFiles {
		base := strings.ToLower(sanitizeName(filepath.Base(iconPath)))
		if strings.Contains(base, appIDClean) {
			data, err := os.ReadFile(iconPath)
			if err == nil {
				installed, err := desktop.InstallIcon(appID, data, filepath.Ext(iconPath))
				if err == nil {
					return installed
				}
			}
		}
	}

	data, err := os.ReadFile(iconFiles[0])
	if err == nil {
		installed, err := desktop.InstallIcon(appID, data, filepath.Ext(iconFiles[0]))
		if err == nil {
			return installed
		}
	}

	return ""
}

func isRunnable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 4)
	n, err := io.ReadFull(f, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return false
	}

	if n >= 4 && bytes.Equal(header[:4], elfMagic) {
		return true
	}
	if n >= 2 && header[0] == '#' && header[1] == '!' {
		return true
	}
	return false
}

func parseDesktopExec(desktopPath, rootDir string) string {
	data, err := os.ReadFile(desktopPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Exec=") {
			rawExec := strings.TrimPrefix(line, "Exec=")
			parts := strings.Fields(rawExec)
			if len(parts) > 0 {
				execName := parts[0]
				if filepath.IsAbs(execName) && isFile(execName) {
					return execName
				}
				// Look for it relative to rootDir
				directRel := filepath.Join(rootDir, execName)
				if isFile(directRel) {
					return directRel
				}
				inBin := filepath.Join(rootDir, "bin", execName)
				if isFile(inBin) {
					return inBin
				}
			}
		}
	}
	return ""
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func sanitizeName(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}
