package valueobject

// Version represents a software version.
type Version struct {
	value string
}

// NewVersion creates a new Version from a string.
func NewVersion(v string) Version {
	return Version{value: v}
}

// String returns the string representation of the Version.
func (v Version) String() string {
	return v.value
}

// IsEmpty returns true if the Version is empty.
func (v Version) IsEmpty() bool {
	return v.value == ""
}

// Equals checks if two Versions are equal.
func (v Version) Equals(other Version) bool {
	return v.value == other.value
}
