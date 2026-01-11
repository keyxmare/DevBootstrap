// Package system provides OS and architecture detection utilities.
package system

// OSType represents the operating system type.
type OSType int

const (
	OSMacOS OSType = iota
	OSUbuntu
	OSDebian
	OSLinuxOther
	OSUnsupported
)

// String returns the string representation of the OS type.
func (o OSType) String() string {
	switch o {
	case OSMacOS:
		return "macOS"
	case OSUbuntu:
		return "Ubuntu"
	case OSDebian:
		return "Debian"
	case OSLinuxOther:
		return "Linux"
	case OSUnsupported:
		return "Unsupported"
	default:
		return "Unknown"
	}
}

// Architecture represents the CPU architecture.
type Architecture int

const (
	ArchARM64 Architecture = iota
	ArchAMD64
	ArchUnknown
)

// String returns the string representation of the architecture.
func (a Architecture) String() string {
	switch a {
	case ArchARM64:
		return "arm64"
	case ArchAMD64:
		return "amd64"
	default:
		return "unknown"
	}
}

// SystemInfo contains information about the current system.
type SystemInfo struct {
	OS        OSType
	OSName    string
	OSVersion string
	Arch      Architecture
	HomeDir   string
	IsRoot    bool
	HasSudo   bool
}

// IsSupported returns true if the OS is supported.
func (s *SystemInfo) IsSupported() bool {
	return s.OS == OSMacOS || s.OS == OSUbuntu || s.OS == OSDebian
}

// IsMacOS returns true if running on macOS.
func (s *SystemInfo) IsMacOS() bool {
	return s.OS == OSMacOS
}

// IsLinux returns true if running on any Linux distribution.
func (s *SystemInfo) IsLinux() bool {
	return s.OS == OSUbuntu || s.OS == OSDebian || s.OS == OSLinuxOther
}

// IsDebian returns true if running on Debian-based systems (Ubuntu or Debian).
func (s *SystemInfo) IsDebian() bool {
	return s.OS == OSUbuntu || s.OS == OSDebian
}
