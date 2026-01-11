package valueobject

// AppStatus represents the installation status of an application.
type AppStatus int

const (
	// StatusNotInstalled indicates the application is not installed.
	StatusNotInstalled AppStatus = iota
	// StatusInstalled indicates the application is installed.
	StatusInstalled
	// StatusUpdateAvailable indicates an update is available.
	StatusUpdateAvailable
)

// String returns the string representation of AppStatus.
func (s AppStatus) String() string {
	switch s {
	case StatusInstalled:
		return "installe"
	case StatusUpdateAvailable:
		return "mise a jour disponible"
	default:
		return "non installe"
	}
}

// IsInstalled returns true if the application is installed.
func (s AppStatus) IsInstalled() bool {
	return s == StatusInstalled || s == StatusUpdateAvailable
}
