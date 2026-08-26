package detector

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	elfMagic = []byte{0x7f, 0x45, 0x4c, 0x46}
	debMagic = []byte("!<arch>\n")
	rpmMagic = []byte{0xed, 0xab, 0xee, 0xdb}
)

func InspectFile(filePath string) (*FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not a package file", filePath)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make([]byte, 512)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}
	header = header[:n]

	pkgType, arch := detectTypeAndArch(filePath, header)
	isExec := (fileInfo.Mode().Perm() & 0111) != 0

	info := &FileInfo{
		Path:          filePath,
		FileName:      filepath.Base(filePath),
		AppName:       cleanBaseName(filepath.Base(filePath)),
		Size:          fileInfo.Size(),
		FormattedSize: formatSize(fileInfo.Size()),
		Type:          pkgType,
		Arch:          arch,
		IsExecutable:  isExec,
		Extra:         make(map[string]string),
	}

	if pkgType == TypeFlatpakRef || pkgType == TypeFlatpakRepo {
		parseFlatpakRefFile(filePath, info)
	}

	return info, nil
}

func parseFlatpakRefFile(filePath string, info *FileInfo) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			info.Extra[k] = v
		}
	}

	if appID, ok := info.Extra["Name"]; ok && appID != "" {
		info.AppID = appID
		if title, ok := info.Extra["Title"]; ok && title != "" {
			info.AppName = title
		} else {
			info.AppName = appID
		}
	}
	if branch, ok := info.Extra["Branch"]; ok && branch != "" {
		info.Arch = fmt.Sprintf("Flatpak (%s)", branch)
	}
}

func cleanBaseName(fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return strings.Title(strings.ToLower(name))
}

func detectTypeAndArch(path string, header []byte) (PackageType, string) {
	ext := strings.ToLower(filepath.Ext(path))

	if len(header) >= 4 && bytes.Equal(header[:4], rpmMagic) {
		return TypeRPM, "Any / RPM"
	}

	if len(header) >= 8 && bytes.Equal(header[:8], debMagic) {
		return TypeDeb, "Debian Architecture"
	}

	if ext == ".flatpakref" || (len(header) > 13 && strings.Contains(string(header), "[Flatpak Ref]")) {
		return TypeFlatpakRef, "Flathub / Flatpak"
	}

	if ext == ".flatpakrepo" || (len(header) > 14 && strings.Contains(string(header), "[Flatpak Repo]")) {
		return TypeFlatpakRepo, "Repository"
	}

	if ext == ".flatpak" {
		return TypeFlatpak, "Flatpak Bundle"
	}

	if len(header) >= 4 && bytes.Equal(header[:4], elfMagic) {
		arch := parseELFArch(header)
		if isAppImage(header, ext) {
			return TypeAppImage, arch
		}
		return TypeBinary, arch
	}

	if len(header) >= 2 && header[0] == '#' && header[1] == '!' {
		return TypeScript, "Script"
	}

	switch ext {
	case ".deb":
		return TypeDeb, "Debian Architecture"
	case ".rpm":
		return TypeRPM, "RPM Architecture"
	case ".appimage":
		return TypeAppImage, "x86_64 / Any"
	default:
		return TypeUnknown, "Unknown"
	}
}

func isAppImage(header []byte, ext string) bool {
	if ext == ".appimage" {
		return true
	}
	if len(header) >= 11 {
		if header[8] == 0x41 && header[9] == 0x49 && (header[10] == 0x01 || header[10] == 0x02) {
			return true
		}
	}
	return false
}

func parseELFArch(header []byte) string {
	if len(header) < 20 {
		return "ELF (Unknown)"
	}

	machine := uint16(header[18]) | (uint16(header[19]) << 8)
	switch machine {
	case 0x3e:
		return "x86_64 (64-bit)"
	case 0xb7:
		return "aarch64 (ARM 64-bit)"
	case 0x03:
		return "x86 (32-bit)"
	case 0x28:
		return "ARM (32-bit)"
	case 0xf3:
		return "RISC-V (64-bit)"
	default:
		return "ELF"
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
