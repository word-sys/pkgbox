package deb

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var arMagic = []byte("!<arch>\n")

type DebMetadata struct {
	Package      string
	Version      string
	Architecture string
	Maintainer   string
	Homepage     string
	Description  string
}

func InspectDeb(debPath string) (*DebMetadata, error) {
	f, err := os.Open(debPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	magic := make([]byte, 8)
	if _, err := io.ReadFull(f, magic); err != nil || !bytes.Equal(magic, arMagic) {
		return nil, fmt.Errorf("not a valid debian ar archive")
	}

	headerBuf := make([]byte, 60)
	for {
		_, err := io.ReadFull(f, headerBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := strings.TrimSpace(string(headerBuf[0:16]))
		name = strings.TrimSuffix(name, "/")

		sizeStr := strings.TrimSpace(string(headerBuf[48:58]))
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ar file size: %w", err)
		}

		// Check if this entry is control.tar.gz or control.tar
		if strings.HasPrefix(name, "control.tar") {
			controlData := io.LimitReader(f, size)
			meta, err := parseControlArchive(name, controlData)
			if err == nil && meta != nil {
				return meta, nil
			}
		} else {
			// Skip payload
			if _, err := f.Seek(size, io.SeekCurrent); err != nil {
				return nil, err
			}
		}

		// ar archives pad odd file sizes with a newline byte
		if size%2 != 0 {
			if _, err := f.Seek(1, io.SeekCurrent); err != nil {
				return nil, err
			}
		}
	}

	return nil, fmt.Errorf("control archive not found in deb package")
}

func parseControlArchive(name string, r io.Reader) (*DebMetadata, error) {
	var tr *tar.Reader

	if strings.HasSuffix(name, ".gz") {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer gzr.Close()
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(r)
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		cleanName := strings.TrimPrefix(hdr.Name, "./")
		if cleanName == "control" {
			return parseControlFile(tr)
		}
	}

	return nil, fmt.Errorf("control file not found inside control archive")
}

func parseControlFile(r io.Reader) (*DebMetadata, error) {
	scanner := bufio.NewScanner(r)
	meta := &DebMetadata{}
	var currentField string

	for scanner.Scan() {
		line := scanner.Text()

		// Indented lines continue previous multiline field (e.g. Description)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if currentField == "Description" && meta.Description != "" {
				meta.Description += " " + strings.TrimSpace(line)
			}
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		currentField = key

		switch key {
		case "Package":
			meta.Package = val
		case "Version":
			meta.Version = val
		case "Architecture":
			meta.Architecture = val
		case "Maintainer":
			meta.Maintainer = val
		case "Homepage":
			meta.Homepage = val
		case "Description":
			meta.Description = val
		}
	}

	return meta, nil
}
