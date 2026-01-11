package valueobject

// OSType represents the operating system type.
type OSType int

const (
	// OSMacOS represents macOS.
	OSMacOS OSType = iota
	// OSUbuntu represents Ubuntu Linux.
	OSUbuntu
	// OSDebian represents Debian Linux.
	OSDebian
	// OSLinuxOther represents other Linux distributions.
	OSLinuxOther
	// OSUnsupported represents unsupported operating systems.
	OSUnsupported
)

// String returns the string representation of OSType.
func (o OSType) String() string {
	names := map[OSType]string{
		OSMacOS:       "macOS",
		OSUbuntu:      "Ubuntu",
		OSDebian:      "Debian",
		OSLinuxOther:  "Linux",
		OSUnsupported: "Unsupported",
	}
	if name, ok := names[o]; ok {
		return name
	}
	return "Unknown"
}

// IsLinux returns true if the OS is a Linux distribution.
func (o OSType) IsLinux() bool {
	return o == OSUbuntu || o == OSDebian || o == OSLinuxOther
}

// IsDebian returns true if the OS is Debian-based.
func (o OSType) IsDebian() bool {
	return o == OSUbuntu || o == OSDebian
}

// IsMacOS returns true if the OS is macOS.
func (o OSType) IsMacOS() bool {
	return o == OSMacOS
}
