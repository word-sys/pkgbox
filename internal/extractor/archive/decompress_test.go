package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZip(t *testing.T) {
	tempDir := t.TempDir()
	zipFile := filepath.Join(tempDir, "test.zip")
	destDir := filepath.Join(tempDir, "extracted")

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("hello.txt")
	if err != nil {
		t.Fatalf("failed creating zip entry: %v", err)
	}
	w.Write([]byte("Hello from zip!"))
	zw.Close()

	if err := os.WriteFile(zipFile, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed writing zip: %v", err)
	}

	if err := ExtractArchive(zipFile, destDir); err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed reading extracted file: %v", err)
	}
	if string(content) != "Hello from zip!" {
		t.Errorf("expected 'Hello from zip!', got %q", string(content))
	}
}

func TestExtractTarGz(t *testing.T) {
	tempDir := t.TempDir()
	tarFile := filepath.Join(tempDir, "test.tar.gz")
	destDir := filepath.Join(tempDir, "extracted")

	buf := new(bytes.Buffer)
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)

	body := []byte("Hello from tar.gz!")
	hdr := &tar.Header{
		Name: "greeting.txt",
		Mode: 0644,
		Size: int64(len(body)),
	}
	tw.WriteHeader(hdr)
	tw.Write(body)
	tw.Close()
	gw.Close()

	if err := os.WriteFile(tarFile, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed writing tar.gz: %v", err)
	}

	if err := ExtractArchive(tarFile, destDir); err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "greeting.txt"))
	if err != nil {
		t.Fatalf("failed reading extracted file: %v", err)
	}
	if string(content) != "Hello from tar.gz!" {
		t.Errorf("expected 'Hello from tar.gz!', got %q", string(content))
	}
}

func TestSafeJoin_TraversalBlocked(t *testing.T) {
	baseDir := "/home/user/apps"
	maliciousPath := "../../../../etc/shadow"

	_, err := safeJoin(baseDir, maliciousPath)
	if err == nil {
		t.Fatalf("expected error for path traversal attempt, got nil")
	}
}
