package appimage

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pkgbox/internal/desktop"
)

type ExtractedMetadata struct {
	Name          string
	Comment       string
	Categories    []string
	IconPath      string
	HasCustomIcon bool
}

func ExtractAppImageMetadata(appImagePath, appID string) (*ExtractedMetadata, error) {
	tempDir, err := os.MkdirTemp("", "pkgbox-appimage-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// AppImageKit built-in extraction flag (extracts into squashfs-root in working directory)
	cmd := exec.Command(appImagePath, "--appimage-extract")
	cmd.Dir = tempDir
	_ = cmd.Run()

	squashRoot := filepath.Join(tempDir, "squashfs-root")
	if _, err := os.Stat(squashRoot); err != nil {
		return nil, err
	}

	meta := &ExtractedMetadata{
		Categories: []string{"Utility"},
	}

	// 1. Locate and parse .desktop file inside squashfs-root
	desktopFiles, _ := filepath.Glob(filepath.Join(squashRoot, "*.desktop"))
	if len(desktopFiles) > 0 {
		parseEmbeddedDesktopFile(desktopFiles[0], meta)
	}

	// 2. Locate and install icon (.DirIcon or *.png / *.svg)
	iconData, iconExt := findEmbeddedIcon(squashRoot)
	if len(iconData) > 0 {
		iconPath, err := desktop.InstallIcon(appID, iconData, iconExt)
		if err == nil {
			meta.IconPath = iconPath
			meta.HasCustomIcon = true
		}
	}

	return meta, nil
}

func parseEmbeddedDesktopFile(filePath string, meta *ExtractedMetadata) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inDesktopEntry := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[Desktop Entry]" {
			inDesktopEntry = true
			continue
		} else if strings.HasPrefix(line, "[") {
			inDesktopEntry = false
		}

		if !inDesktopEntry || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			if meta.Name == "" && val != "" {
				meta.Name = val
			}
		case "Comment":
			if meta.Comment == "" && val != "" {
				meta.Comment = val
			}
		case "Categories":
			if val != "" {
				cats := strings.Split(val, ";")
				var cleaned []string
				for _, c := range cats {
					c = strings.TrimSpace(c)
					if c != "" && c != "Application" {
						cleaned = append(cleaned, c)
					}
				}
				if len(cleaned) > 0 {
					meta.Categories = cleaned
				}
			}
		}
	}
}

func findEmbeddedIcon(squashRoot string) ([]byte, string) {
	// Check .DirIcon first
	dirIcon := filepath.Join(squashRoot, ".DirIcon")
	if data, err := os.ReadFile(dirIcon); err == nil && len(data) > 0 {
		return data, ".png"
	}

	// Check for any PNG in root or icons directory
	pngs, _ := filepath.Glob(filepath.Join(squashRoot, "*.png"))
	if len(pngs) > 0 {
		if data, err := os.ReadFile(pngs[0]); err == nil && len(data) > 0 {
			return data, ".png"
		}
	}

	// Check for SVG
	svgs, _ := filepath.Glob(filepath.Join(squashRoot, "*.svg"))
	if len(svgs) > 0 {
		if data, err := os.ReadFile(svgs[0]); err == nil && len(data) > 0 {
			return data, ".svg"
		}
	}

	return nil, ""
}
