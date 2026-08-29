package downloader

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type RemotePackageInfo struct {
	URL           string
	FileName      string
	Size          int64
	FormattedSize string
	ContentType   string
}

func IsRemoteURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	// Translate flathub.org/apps/<app-id> to flathub flatpakref URL
	if strings.Contains(u.Host, "flathub.org") && strings.HasPrefix(u.Path, "/apps/") {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) >= 2 {
			appID := parts[len(parts)-1]
			if !strings.HasSuffix(appID, ".flatpakref") {
				return fmt.Sprintf("https://dl.flathub.org/repo/appstream/%s.flatpakref", appID)
			}
		}
	}

	return raw
}

func InspectRemoteURL(rawURL string) (*RemotePackageInfo, error) {
	targetURL := NormalizeURL(rawURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PkgBox/1.0 (Linux)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach remote URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	finalURL := resp.Request.URL.String()
	fileName := extractFilename(finalURL, resp.Header.Get("Content-Disposition"))
	size := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")

	return &RemotePackageInfo{
		URL:           finalURL,
		FileName:      fileName,
		Size:          size,
		FormattedSize: formatSize(size),
		ContentType:   contentType,
	}, nil
}

func extractFilename(finalURL, contentDisposition string) string {
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if filename, ok := params["filename"]; ok && filename != "" {
				return filename
			}
		}
	}

	u, err := url.Parse(finalURL)
	if err == nil && u.Path != "" {
		base := filepath.Base(u.Path)
		if base != "" && base != "." && base != "/" {
			return base
		}
	}

	return "downloaded-package"
}

func formatSize(bytes int64) string {
	if bytes <= 0 {
		return "Unknown size"
	}
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
