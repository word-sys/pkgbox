package appimage

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

var (
	squashfsMagicLE = []byte{0x73, 0x71, 0x73, 0x68} // "sqsh"
	squashfsMagicBE = []byte{0x68, 0x73, 0x71, 0x73} // "hsqs"
	appimageType1   = []byte{0x41, 0x49, 0x01}
	appimageType2   = []byte{0x41, 0x49, 0x02}
)

type AppImageStructure struct {
	Type           int
	HasSquashFS    bool
	SquashFSOffset int64
}

func InspectAppImage(filePath string) (*AppImageStructure, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make([]byte, 16)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, fmt.Errorf("failed reading AppImage header: %w", err)
	}

	appType := 0
	if len(header) >= 11 {
		if bytes.Equal(header[8:11], appimageType1) {
			appType = 1
		} else if bytes.Equal(header[8:11], appimageType2) {
			appType = 2
		}
	}

	offset, err := FindSquashFSOffset(f)
	if err != nil {
		return &AppImageStructure{
			Type:           appType,
			HasSquashFS:    false,
			SquashFSOffset: -1,
		}, nil
	}

	return &AppImageStructure{
		Type:           appType,
		HasSquashFS:    true,
		SquashFSOffset: offset,
	}, nil
}

func FindSquashFSOffset(r io.ReadSeeker) (int64, error) {
	const scanLimit = 2 * 1024 * 1024 // Scan first 2MB for squashfs header
	const chunkSize = 64 * 1024

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return -1, err
	}

	buf := make([]byte, chunkSize)
	var currentOffset int64

	for currentOffset < scanLimit {
		n, err := r.Read(buf)
		if n <= 0 {
			break
		}

		for i := 0; i <= n-4; i++ {
			if bytes.Equal(buf[i:i+4], squashfsMagicLE) || bytes.Equal(buf[i:i+4], squashfsMagicBE) {
				return currentOffset + int64(i), nil
			}
		}

		currentOffset += int64(n)
		if err == io.EOF {
			break
		}
	}

	return -1, fmt.Errorf("squashfs signature not found within scan limit")
}
