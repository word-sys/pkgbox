package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ExtractArchive(srcFile, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	lower := strings.ToLower(srcFile)
	if strings.HasSuffix(lower, ".zip") {
		return extractZip(srcFile, destDir)
	} else if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return extractTarGz(srcFile, destDir)
	} else if strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".tar") {
		return extractTarWithCommand(srcFile, destDir)
	}

	// Fallback detection using tar command
	return extractTarWithCommand(srcFile, destDir)
}

func extractZip(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		targetPath, err := safeJoin(destDir, f.Name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func extractTarGz(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath, err := safeJoin(destDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func extractTarWithCommand(src, destDir string) error {
	cmd := exec.Command("tar", "-xf", src, "-C", destDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar extraction failed: %w (output: %s)", err, string(output))
	}
	return nil
}

func safeJoin(baseDir, untrustedPath string) (string, error) {
	cleaned := filepath.Clean(untrustedPath)
	fullPath := filepath.Join(baseDir, cleaned)

	baseClean := filepath.Clean(baseDir) + string(os.PathSeparator)
	if !strings.HasPrefix(fullPath, baseClean) && fullPath != filepath.Clean(baseDir) {
		return "", fmt.Errorf("path traversal attempt blocked: %s", untrustedPath)
	}

	return fullPath, nil
}
