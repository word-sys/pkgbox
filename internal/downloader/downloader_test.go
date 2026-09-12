package downloader

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadPackage_Success(t *testing.T) {
	tempCache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tempCache)

	testPayload := bytes.Repeat([]byte("PkgBoxDownloadDataChunk"), 1000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testPayload)))
		w.Header().Set("Content-Disposition", `attachment; filename="test-app.AppImage"`)
		w.WriteHeader(http.StatusOK)
		w.Write(testPayload)
	}))
	defer server.Close()

	var progressReports int
	var finalFraction float64

	downloadedFile, err := DownloadPackage(context.Background(), server.URL, func(downloaded, total int64, fraction float64) {
		progressReports++
		finalFraction = fraction
	})
	if err != nil {
		t.Fatalf("unexpected download error: %v", err)
	}

	if filepath.Base(downloadedFile) != "test-app.AppImage" {
		t.Errorf("expected filename test-app.AppImage, got %q", filepath.Base(downloadedFile))
	}

	content, err := os.ReadFile(downloadedFile)
	if err != nil {
		t.Fatalf("failed reading downloaded file: %v", err)
	}

	if !bytes.Equal(content, testPayload) {
		t.Errorf("downloaded content does not match payload")
	}

	if progressReports == 0 {
		t.Errorf("expected progress callbacks, got none")
	}
	if finalFraction != 1.0 {
		t.Errorf("expected final fraction 1.0, got %f", finalFraction)
	}
}

func TestDownloadPackage_Cancellation(t *testing.T) {
	tempCache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tempCache)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="canceled.AppImage"`)
		w.WriteHeader(http.StatusOK)
		for i := 0; i < 50; i++ {
			w.Write(bytes.Repeat([]byte("Chunk"), 1024))
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := DownloadPackage(ctx, server.URL, nil)
	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}

	tmpFile := filepath.Join(tempCache, "pkgbox", "downloads", "canceled.AppImage.tmp")
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Errorf("temporary download file was not cleaned up after cancellation")
	}
}
