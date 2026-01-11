// Package secondary contains driven/outbound port interfaces.
package secondary

import (
	"context"
	"time"
)

// CommandResult represents the result of command execution.
type CommandResult struct {
	Success  bool
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// CommandOptions configures command execution.
type CommandOptions struct {
	Description        string
	Sudo               bool
	SudoNonInteractive bool
	Env                map[string]string
	WorkingDir         string
	Timeout            time.Duration
	Interactive        bool
}

// CommandOption is a functional option for CommandOptions.
type CommandOption func(*CommandOptions)

// WithDescription sets the description for progress reporting.
func WithDescription(desc string) CommandOption {
	return func(o *CommandOptions) {
		o.Description = desc
	}
}

// WithSudo enables sudo for the command.
func WithSudo() CommandOption {
	return func(o *CommandOptions) {
		o.Sudo = true
	}
}

// WithSudoNonInteractive enables non-interactive sudo.
func WithSudoNonInteractive() CommandOption {
	return func(o *CommandOptions) {
		o.Sudo = true
		o.SudoNonInteractive = true
	}
}

// WithEnv sets environment variables for the command.
func WithEnv(env map[string]string) CommandOption {
	return func(o *CommandOptions) {
		o.Env = env
	}
}

// WithWorkingDir sets the working directory for the command.
func WithWorkingDir(dir string) CommandOption {
	return func(o *CommandOptions) {
		o.WorkingDir = dir
	}
}

// WithTimeout sets the timeout for the command.
func WithTimeout(timeout time.Duration) CommandOption {
	return func(o *CommandOptions) {
		o.Timeout = timeout
	}
}

// WithInteractive enables interactive mode (stdin/stdout connected to terminal).
func WithInteractive() CommandOption {
	return func(o *CommandOptions) {
		o.Interactive = true
	}
}

// CommandExecutor defines the interface for executing system commands.
type CommandExecutor interface {
	// Execute runs a command and returns the result.
	Execute(ctx context.Context, args []string, opts ...CommandOption) *CommandResult

	// CommandExists checks if a command is available in PATH.
	CommandExists(name string) bool

	// GetCommandPath returns the full path to a command.
	GetCommandPath(name string) string

	// GetCommandVersion returns the version string of a command.
	GetCommandVersion(name string, versionArg string) string

	// SetSudoAskpass sets the path to a SUDO_ASKPASS script.
	SetSudoAskpass(path string)
}
