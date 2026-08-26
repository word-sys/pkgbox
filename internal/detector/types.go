package detector

type PackageType string

const (
	TypeAppImage   PackageType = "AppImage"
	TypeDeb        PackageType = "Debian Package (.deb)"
	TypeRPM        PackageType = "RPM Package (.rpm)"
	TypeFlatpak    PackageType = "Flatpak Bundle (.flatpak)"
	TypeFlatpakRef PackageType = "Flatpak Reference (.flatpakref)"
	TypeFlatpakRepo PackageType = "Flatpak Repository (.flatpakrepo)"
	TypeBinary     PackageType = "Executable Binary (ELF)"
	TypeScript     PackageType = "Shell Script"
	TypeUnknown    PackageType = "Unknown / Unsupported"
)

type FileInfo struct {
	Path          string
	FileName      string
	AppName       string
	AppID         string
	Size          int64
	FormattedSize string
	Type          PackageType
	Arch          string
	IsExecutable  bool
	Extra         map[string]string
}
