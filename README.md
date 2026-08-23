# PkgBox

PkgBox is a universal Linux application installer and manager. It provides a drag-and-drop interface to inspect, install, and manage applications across multiple package formats.

## Supported Formats
- AppImage (.AppImage)
- Debian Packages (.deb)
- RPM Packages (.rpm)
- Flatpak (.flatpak, .flatpakref)
- Standalone binaries and scripts (will be developed in future)

## Requirements
- Go 1.18 or newer
- GTK 3 development libraries (`libgtk-3-dev`)
- pkg-config
- gcc

## Building
```bash
go build -o bin/pkgbox ./cmd/pkgbox
```

## Running
```bash
./bin/pkgbox
```

## Testing
```bash
./scripts/test_all.sh
```

## License
GPL-3.0
