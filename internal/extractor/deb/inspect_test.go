package deb

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseControlFile(t *testing.T) {
	rawControl := `Package: ripgrep
Version: 14.1.0-1
Architecture: amd64
Maintainer: Debian Rust Maintainers <pkg-rust-maintainers@alioth-lists.debian.net>
Homepage: https://github.com/BurntSushi/ripgrep
Description: fast line-oriented search tool
 ripgrep is a line-oriented search tool that recursively searches
 your current directory for a regex pattern while respecting your
 gitignore rules.
`
	meta, err := parseControlFile(strings.NewReader(rawControl))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Package != "ripgrep" {
		t.Errorf("expected Package 'ripgrep', got %q", meta.Package)
	}
	if meta.Version != "14.1.0-1" {
		t.Errorf("expected Version '14.1.0-1', got %q", meta.Version)
	}
	if meta.Architecture != "amd64" {
		t.Errorf("expected Architecture 'amd64', got %q", meta.Architecture)
	}
	if !strings.Contains(meta.Description, "fast line-oriented search tool") {
		t.Errorf("expected Description to contain summary, got %q", meta.Description)
	}
}

func TestInspectDeb_Complete(t *testing.T) {
	tempDir := t.TempDir()
	debPath := filepath.Join(tempDir, "test.deb")

	// 1. Build control.tar.gz buffer
	controlContent := []byte("Package: myeditor\nVersion: 2.0.0\nArchitecture: all\nMaintainer: Editor Devs\nDescription: Minimal code editor\n")
	var controlTarGz bytes.Buffer
	gw := gzip.NewWriter(&controlTarGz)
	tw := tar.NewWriter(gw)
	tw.WriteHeader(&tar.Header{
		Name: "./control",
		Mode: 0644,
		Size: int64(len(controlContent)),
	})
	tw.Write(controlContent)
	tw.Close()
	gw.Close()

	// 2. Build .deb ar archive
	var debBuf bytes.Buffer
	debBuf.WriteString("!<arch>\n")

	// debian-binary entry (4 bytes: "2.0\n")
	debBinary := []byte("2.0\n")
	debBuf.WriteString(fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n", "debian-binary", 1600000000, 0, 0, 0644, len(debBinary)))
	debBuf.Write(debBinary)

	// control.tar.gz entry
	debBuf.WriteString(fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n", "control.tar.gz", 1600000000, 0, 0, 0644, controlTarGz.Len()))
	debBuf.Write(controlTarGz.Bytes())
	if controlTarGz.Len()%2 != 0 {
		debBuf.WriteByte('\n')
	}

	if err := os.WriteFile(debPath, debBuf.Bytes(), 0644); err != nil {
		t.Fatalf("failed writing test deb: %v", err)
	}

	meta, err := InspectDeb(debPath)
	if err != nil {
		t.Fatalf("unexpected InspectDeb error: %v", err)
	}

	if meta.Package != "myeditor" {
		t.Errorf("expected Package 'myeditor', got %q", meta.Package)
	}
	if meta.Version != "2.0.0" {
		t.Errorf("expected Version '2.0.0', got %q", meta.Version)
	}
	if meta.Architecture != "all" {
		t.Errorf("expected Architecture 'all', got %q", meta.Architecture)
	}
	if meta.Description != "Minimal code editor" {
		t.Errorf("expected Description 'Minimal code editor', got %q", meta.Description)
	}
}
