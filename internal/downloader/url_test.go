package downloader

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsRemoteURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/app.AppImage", true},
		{"http://example.com/package.deb", true},
		{"file:///home/user/app.AppImage", false},
		{"/home/user/app.AppImage", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsRemoteURL(tt.input)
		if got != tt.expected {
			t.Errorf("IsRemoteURL(%q) = %v, expected %v", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizeURL_Flathub(t *testing.T) {
	input := "https://flathub.org/apps/org.videolan.VLC"
	expected := "https://dl.flathub.org/repo/appstream/org.videolan.VLC.flatpakref"

	got := NormalizeURL(input)
	if got != expected {
		t.Errorf("NormalizeURL(%q) = %q, expected %q", input, got, expected)
	}
}

func TestInspectRemoteURL_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("expected HEAD request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/vnd.debian.binary-package")
		w.Header().Set("Content-Length", "2048")
		w.Header().Set("Content-Disposition", `attachment; filename="mypackage_1.0_amd64.deb"`)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	info, err := InspectRemoteURL(server.URL + "/download")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.FileName != "mypackage_1.0_amd64.deb" {
		t.Errorf("expected filename mypackage_1.0_amd64.deb, got %q", info.FileName)
	}
	if info.Size != 2048 {
		t.Errorf("expected size 2048, got %d", info.Size)
	}
	if info.FormattedSize != "2.0 KB" {
		t.Errorf("expected FormattedSize '2.0 KB', got %q", info.FormattedSize)
	}
}

func TestInspectRemoteURL_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := InspectRemoteURL(server.URL + "/missing.deb")
	if err == nil {
		t.Fatalf("expected error for 404 response, got nil")
	}
}
