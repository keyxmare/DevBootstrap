// Package result contains domain result types.
package result

// InstallResult represents the result of an installation operation.
type InstallResult struct {
	success  bool
	message  string
	version  string
	path     string
	errors   []string
	warnings []string
}

// NewSuccess creates a successful result.
func NewSuccess(message string) *InstallResult {
	return &InstallResult{
		success: true,
		message: message,
	}
}

// NewFailure creates a failed result.
func NewFailure(message string, errors ...string) *InstallResult {
	return &InstallResult{
		success: false,
		message: message,
		errors:  errors,
	}
}

// Success returns true if the operation was successful.
func (r *InstallResult) Success() bool {
	return r.success
}

// Message returns the result message.
func (r *InstallResult) Message() string {
	return r.message
}

// Version returns the installed version.
func (r *InstallResult) Version() string {
	return r.version
}

// Path returns the installation path.
func (r *InstallResult) Path() string {
	return r.path
}

// Errors returns all error messages.
func (r *InstallResult) Errors() []string {
	return r.errors
}

// Warnings returns all warning messages.
func (r *InstallResult) Warnings() []string {
	return r.warnings
}

// WithVersion sets the version and returns the result for chaining.
func (r *InstallResult) WithVersion(v string) *InstallResult {
	r.version = v
	return r
}

// WithPath sets the path and returns the result for chaining.
func (r *InstallResult) WithPath(p string) *InstallResult {
	r.path = p
	return r
}

// AddWarning adds a warning message.
func (r *InstallResult) AddWarning(warning string) {
	r.warnings = append(r.warnings, warning)
}

// AddError adds an error message.
func (r *InstallResult) AddError(err string) {
	r.errors = append(r.errors, err)
}
