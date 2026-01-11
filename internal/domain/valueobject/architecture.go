package valueobject

// Architecture represents the CPU architecture.
type Architecture int

const (
	// ArchUnknown represents an unknown architecture.
	ArchUnknown Architecture = iota
	// ArchAMD64 represents the x86-64 architecture.
	ArchAMD64
	// ArchARM64 represents the ARM64 architecture.
	ArchARM64
)

// String returns the string representation of Architecture.
func (a Architecture) String() string {
	switch a {
	case ArchAMD64:
		return "amd64"
	case ArchARM64:
		return "arm64"
	default:
		return "unknown"
	}
}

// IsARM returns true if the architecture is ARM-based.
func (a Architecture) IsARM() bool {
	return a == ArchARM64
}

// IsX86 returns true if the architecture is x86-based.
func (a Architecture) IsX86() bool {
	return a == ArchAMD64
}
