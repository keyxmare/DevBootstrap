package result

// UninstallResult represents the result of an uninstallation operation.
type UninstallResult struct {
	success  bool
	message  string
	errors   []string
	warnings []string
}

// NewUninstallSuccess creates a successful uninstall result.
func NewUninstallSuccess(message string) *UninstallResult {
	return &UninstallResult{
		success: true,
		message: message,
	}
}

// NewUninstallFailure creates a failed uninstall result.
func NewUninstallFailure(message string, errors ...string) *UninstallResult {
	return &UninstallResult{
		success: false,
		message: message,
		errors:  errors,
	}
}

// Success returns true if the operation was successful.
func (r *UninstallResult) Success() bool {
	return r.success
}

// Message returns the result message.
func (r *UninstallResult) Message() string {
	return r.message
}

// Errors returns all error messages.
func (r *UninstallResult) Errors() []string {
	return r.errors
}

// Warnings returns all warning messages.
func (r *UninstallResult) Warnings() []string {
	return r.warnings
}

// AddWarning adds a warning message.
func (r *UninstallResult) AddWarning(warning string) {
	r.warnings = append(r.warnings, warning)
}

// AddError adds an error message.
func (r *UninstallResult) AddError(err string) {
	r.errors = append(r.errors, err)
}
