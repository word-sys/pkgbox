package rpm

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectRPM_LeadOnly(t *testing.T) {
	tempDir := t.TempDir()
	rpmPath := filepath.Join(tempDir, "minimal.rpm")

	lead := make([]byte, 96)
	copy(lead[0:4], []byte{0xed, 0xab, 0xee, 0xdb})
	lead[4] = 3 // Major version
	lead[5] = 0 // Minor version
	binary.BigEndian.PutUint16(lead[8:10], 13) // x86_64
	copy(lead[10:], []byte("htop\x00"))

	if err := os.WriteFile(rpmPath, lead, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	meta, err := InspectRPM(rpmPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Name != "htop" {
		t.Errorf("expected Name 'htop', got %q", meta.Name)
	}
	if meta.Architecture != "x86_64" {
		t.Errorf("expected Architecture 'x86_64', got %q", meta.Architecture)
	}
}

func TestInspectRPM_WithHeaderTags(t *testing.T) {
	tempDir := t.TempDir()
	rpmPath := filepath.Join(tempDir, "full.rpm")

	var buf bytes.Buffer

	// 1. Lead (96 bytes)
	lead := make([]byte, 96)
	copy(lead[0:4], []byte{0xed, 0xab, 0xee, 0xdb})
	binary.BigEndian.PutUint16(lead[8:10], 14) // aarch64
	copy(lead[10:], []byte("curl\x00"))
	buf.Write(lead)

	// 2. Signature record (empty preamble, 0 entries)
	sigPreamble := make([]byte, 16)
	copy(sigPreamble[0:4], []byte{0x8e, 0xad, 0xe8, 0x01})
	buf.Write(sigPreamble)

	// 3. Header record
	dataStore := []byte("curl\x008.5.0\x00Command line tool for transferring data\x00MIT\x00")
	nameOff := uint32(bytes.Index(dataStore, []byte("curl\x00")))
	verOff := uint32(bytes.Index(dataStore, []byte("8.5.0\x00")))
	sumOff := uint32(bytes.Index(dataStore, []byte("Command line tool")))
	licOff := uint32(bytes.Index(dataStore, []byte("MIT\x00")))

	numEntries := uint32(4)
	hdrPreamble := make([]byte, 16)
	copy(hdrPreamble[0:4], []byte{0x8e, 0xad, 0xe8, 0x01})
	binary.BigEndian.PutUint32(hdrPreamble[8:12], numEntries)
	binary.BigEndian.PutUint32(hdrPreamble[12:16], uint32(len(dataStore)))
	buf.Write(hdrPreamble)

	// 4 index entries (16 bytes each): Tag (4), Type (4), Offset (4), Count (4)
	tags := []struct {
		tag uint32
		off uint32
	}{
		{TagRPMName, nameOff},
		{TagRPMVersion, verOff},
		{TagRPMSummary, sumOff},
		{TagRPMLicense, licOff},
	}

	for _, entry := range tags {
		entryBuf := make([]byte, 16)
		binary.BigEndian.PutUint32(entryBuf[0:4], entry.tag)
		binary.BigEndian.PutUint32(entryBuf[4:8], 6) // string type
		binary.BigEndian.PutUint32(entryBuf[8:12], entry.off)
		binary.BigEndian.PutUint32(entryBuf[12:16], 1)
		buf.Write(entryBuf)
	}

	buf.Write(dataStore)

	if err := os.WriteFile(rpmPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write test rpm: %v", err)
	}

	meta, err := InspectRPM(rpmPath)
	if err != nil {
		t.Fatalf("unexpected InspectRPM error: %v", err)
	}

	if meta.Name != "curl" {
		t.Errorf("expected Name 'curl', got %q", meta.Name)
	}
	if meta.Version != "8.5.0" {
		t.Errorf("expected Version '8.5.0', got %q", meta.Version)
	}
	if meta.Summary != "Command line tool for transferring data" {
		t.Errorf("expected Summary mismatch, got %q", meta.Summary)
	}
	if meta.License != "MIT" {
		t.Errorf("expected License 'MIT', got %q", meta.License)
	}
	if meta.Architecture != "aarch64" {
		t.Errorf("expected Architecture 'aarch64', got %q", meta.Architecture)
	}
}
