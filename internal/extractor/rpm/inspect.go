package rpm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	rpmLeadMagic   = []byte{0xed, 0xab, 0xee, 0xdb}
	rpmHeaderMagic = []byte{0x8e, 0xad, 0xe8, 0x01}
)

const (
	TagRPMName        = 1000
	TagRPMVersion     = 1001
	TagRPMRelease     = 1002
	TagRPMSummary     = 1004
	TagRPMDescription = 1005
	TagRPMLicense     = 1014
	TagRPMArch        = 1022
)

type RPMMetadata struct {
	Name         string
	Version      string
	Release      string
	Architecture string
	Summary      string
	License      string
}

func InspectRPM(filePath string) (*RPMMetadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 1. Read and validate Lead (96 bytes)
	lead := make([]byte, 96)
	if _, err := io.ReadFull(f, lead); err != nil {
		return nil, fmt.Errorf("failed reading RPM lead: %w", err)
	}

	if !bytes.Equal(lead[:4], rpmLeadMagic) {
		return nil, fmt.Errorf("invalid RPM lead magic")
	}

	leadArch := parseLeadArch(binary.BigEndian.Uint16(lead[8:10]))
	leadName := strings.TrimRight(string(lead[10:76]), "\x00")

	meta := &RPMMetadata{
		Name:         leadName,
		Architecture: leadArch,
	}

	// 2. Read Signature Record Header (Preamble + Index + Data)
	sigEntries, sigDataSize, err := readHeaderPreamble(f)
	if err == nil {
		// Skip signature index entries (16 bytes each) and data store
		sigTotalSize := int64(sigEntries)*16 + int64(sigDataSize)
		// RPM rounds signature block to 8-byte alignment
		padding := (8 - (sigTotalSize % 8)) % 8
		if _, err := f.Seek(sigTotalSize+padding, io.SeekCurrent); err != nil {
			return meta, nil
		}

		// 3. Read Main Header Record
		hdrEntries, hdrDataSize, err := readHeaderPreamble(f)
		if err == nil && hdrEntries > 0 {
			parseHeaderTags(f, hdrEntries, hdrDataSize, meta)
		}
	}

	if meta.Architecture == "" {
		meta.Architecture = leadArch
	}
	if meta.Name == "" {
		meta.Name = leadName
	}

	return meta, nil
}

func readHeaderPreamble(r io.Reader) (uint32, uint32, error) {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, 0, err
	}

	if !bytes.Equal(buf[:4], rpmHeaderMagic) {
		return 0, 0, fmt.Errorf("invalid RPM header magic")
	}

	numEntries := binary.BigEndian.Uint32(buf[8:12])
	dataSize := binary.BigEndian.Uint32(buf[12:16])

	return numEntries, dataSize, nil
}

func parseHeaderTags(r io.Reader, numEntries, dataSize uint32, meta *RPMMetadata) {
	indexBuf := make([]byte, numEntries*16)
	if _, err := io.ReadFull(r, indexBuf); err != nil {
		return
	}

	dataStore := make([]byte, dataSize)
	if _, err := io.ReadFull(r, dataStore); err != nil {
		return
	}

	for i := uint32(0); i < numEntries; i++ {
		entryOffset := i * 16
		tag := binary.BigEndian.Uint32(indexBuf[entryOffset : entryOffset+4])
		dataOffset := binary.BigEndian.Uint32(indexBuf[entryOffset+8 : entryOffset+12])

		if dataOffset >= dataSize {
			continue
		}

		val := readNullTerminatedString(dataStore[dataOffset:])

		switch tag {
		case TagRPMName:
			if val != "" {
				meta.Name = val
			}
		case TagRPMVersion:
			meta.Version = val
		case TagRPMRelease:
			meta.Release = val
		case TagRPMArch:
			meta.Architecture = val
		case TagRPMSummary:
			meta.Summary = val
		case TagRPMLicense:
			meta.License = val
		}
	}
}

func readNullTerminatedString(data []byte) string {
	idx := bytes.IndexByte(data, 0)
	if idx == -1 {
		return string(data)
	}
	return string(data[:idx])
}

func parseLeadArch(archCode uint16) string {
	switch archCode {
	case 1:
		return "i386"
	case 12:
		return "arm"
	case 13:
		return "x86_64"
	case 14:
		return "aarch64"
	case 15:
		return "riscv64"
	default:
		return "x86_64 / RPM"
	}
}
