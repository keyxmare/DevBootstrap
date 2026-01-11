// Package runner provides command execution utilities.
package runner

// Result represents the outcome of a command execution.
type Result struct {
	Success    bool
	Stdout     string
	Stderr     string
	ExitCode   int
	Error      error
}

// NewResult creates a new Result with success status.
func NewResult(success bool, stdout, stderr string, exitCode int, err error) *Result {
	return &Result{
		Success:  success,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		Error:    err,
	}
}

// SuccessResult creates a successful result.
func SuccessResult(stdout string) *Result {
	return &Result{
		Success:  true,
		Stdout:   stdout,
		ExitCode: 0,
	}
}

// FailureResult creates a failed result.
func FailureResult(stderr string, exitCode int, err error) *Result {
	return &Result{
		Success:  false,
		Stderr:   stderr,
		ExitCode: exitCode,
		Error:    err,
	}
}
