package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type DownloadCallback func(downloaded int64, total int64, fraction float64)

func GetCacheDir() (string, error) {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheHome = filepath.Join(home, ".cache")
	}
	dir := filepath.Join(cacheHome, "pkgbox", "downloads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func DownloadPackage(ctx context.Context, rawURL string, callback DownloadCallback) (string, error) {
	targetURL := NormalizeURL(rawURL)

	info, err := InspectRemoteURL(targetURL)
	if err != nil {
		return "", err
	}

	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	fileName := info.FileName
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = fmt.Sprintf("package-%d", time.Now().Unix())
	}

	finalPath := filepath.Join(cacheDir, fileName)
	tmpPath := finalPath + ".tmp"

	req, err := http.NewRequestWithContext(ctx, "GET", info.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "PkgBox/1.0 (Linux)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("server returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	outFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary download file: %w", err)
	}

	cleanUp := func() {
		outFile.Close()
		os.Remove(tmpPath)
	}

	totalBytes := resp.ContentLength
	if totalBytes <= 0 && info.Size > 0 {
		totalBytes = info.Size
	}

	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		select {
		case <-ctx.Done():
			cleanUp()
			return "", ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
				cleanUp()
				return "", fmt.Errorf("failed writing download data: %w", writeErr)
			}
			downloaded += int64(n)

			if callback != nil {
				var fraction float64
				if totalBytes > 0 {
					fraction = float64(downloaded) / float64(totalBytes)
					if fraction > 1.0 {
						fraction = 1.0
					}
				}
				callback(downloaded, totalBytes, fraction)
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			cleanUp()
			return "", fmt.Errorf("error reading response stream: %w", readErr)
		}
	}

	if callback != nil {
		callback(downloaded, downloaded, 1.0)
	}

	if err := outFile.Sync(); err != nil {
		cleanUp()
		return "", err
	}
	outFile.Close()

	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed finalizing download file: %w", err)
	}

	return finalPath, nil
}
