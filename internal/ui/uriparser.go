package ui

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"pkgbox/internal/downloader"
)

type DropKind int

const (
	DropKindFile DropKind = iota
	DropKindURL
)

type DropItem struct {
	Kind  DropKind
	Value string
}

func ParseDropData(raw string) ([]DropItem, error) {
	lines := strings.Split(raw, "\n")
	var items []DropItem

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if downloader.IsRemoteURL(line) {
			items = append(items, DropItem{
				Kind:  DropKindURL,
				Value: downloader.NormalizeURL(line),
			})
			continue
		}

		u, err := url.Parse(line)
		if err != nil {
			continue
		}

		if u.Scheme == "http" || u.Scheme == "https" {
			items = append(items, DropItem{
				Kind:  DropKindURL,
				Value: downloader.NormalizeURL(line),
			})
			continue
		}

		if u.Scheme != "file" && u.Scheme != "" {
			continue
		}

		path := u.Path
		if path == "" {
			path = line
		}

		decodedPath, err := url.PathUnescape(path)
		if err != nil {
			decodedPath = path
		}

		if info, err := os.Stat(decodedPath); err == nil && !info.IsDir() {
			items = append(items, DropItem{
				Kind:  DropKindFile,
				Value: decodedPath,
			})
		}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no valid local files or remote URLs found in drop data")
	}

	return items, nil
}

func ParseURIList(raw string) ([]string, error) {
	items, err := ParseDropData(raw)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, item := range items {
		if item.Kind == DropKindFile {
			paths = append(paths, item.Value)
		}
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no valid local files found in drop data")
	}

	return paths, nil
}
