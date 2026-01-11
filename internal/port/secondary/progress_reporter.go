package secondary

// ProgressReporter defines the interface for progress and status reporting.
type ProgressReporter interface {
	// Info reports an informational message.
	Info(message string)

	// Success reports a success message.
	Success(message string)

	// Warning reports a warning message.
	Warning(message string)

	// Error reports an error message.
	Error(message string)

	// Progress reports a progress message (may overwrite line).
	Progress(message string)

	// ClearProgress clears any progress indicator.
	ClearProgress()

	// Section starts a new section with a header.
	Section(title string)

	// Step reports a step in a multi-step process.
	Step(current, total int, message string)

	// Header prints a styled header.
	Header(title string)

	// Summary prints a key-value summary.
	Summary(title string, items map[string]string)
}
