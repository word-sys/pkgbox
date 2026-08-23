package ui

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func ParseURIList(raw string) ([]string, error) {
	lines := strings.Split(raw, "\n")
	var paths []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		u, err := url.Parse(line)
		if err != nil {
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
			paths = append(paths, decodedPath)
		}
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no valid local files found in drop data")
	}

	return paths, nil
}
