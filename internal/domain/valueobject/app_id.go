// Package valueobject contains domain value objects.
package valueobject

// AppID represents a unique application identifier.
type AppID struct {
	value string
}

// NewAppID creates a new AppID from a string.
func NewAppID(id string) AppID {
	return AppID{value: id}
}

// String returns the string representation of the AppID.
func (id AppID) String() string {
	return id.value
}

// IsEmpty returns true if the AppID is empty.
func (id AppID) IsEmpty() bool {
	return id.value == ""
}

// Equals checks if two AppIDs are equal.
func (id AppID) Equals(other AppID) bool {
	return id.value == other.value
}
