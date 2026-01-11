// Package executor provides command execution adapters.
package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// ShellExecutor implements CommandExecutor using shell commands.
type ShellExecutor struct {
	dryRun      bool
	sudoAskpass string
	reporter    secondary.ProgressReporter
}

// NewShellExecutor creates a new ShellExecutor instance.
func NewShellExecutor(dryRun bool, reporter secondary.ProgressReporter) *ShellExecutor {
	return &ShellExecutor{
		dryRun:   dryRun,
		reporter: reporter,
	}
}

// SetSudoAskpass sets the path to a SUDO_ASKPASS script.
func (e *ShellExecutor) SetSudoAskpass(path string) {
	e.sudoAskpass = path
}

// Execute runs a command and returns the result.
func (e *ShellExecutor) Execute(ctx context.Context, args []string, opts ...secondary.CommandOption) *secondary.CommandResult {
	options := &secondary.CommandOptions{
		Timeout: 2 * time.Minute,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Add sudo if needed
	if options.Sudo && os.Geteuid() != 0 {
		if options.SudoNonInteractive {
			args = append([]string{"sudo", "-n"}, args...)
		} else if e.sudoAskpass != "" {
			args = append([]string{"sudo", "-A"}, args...)
		} else {
			args = append([]string{"sudo"}, args...)
		}
	}

	cmdStr := strings.Join(args, " ")

	if options.Description != "" && e.reporter != nil {
		e.reporter.Progress(options.Description)
	}

	if e.dryRun && !options.SkipDryRun {
		if e.reporter != nil {
			e.reporter.ClearProgress()
			e.reporter.Info(fmt.Sprintf("[DRY RUN] %s", cmdStr))
		}
		return &secondary.CommandResult{Success: true}
	}

	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// Set environment
	cmd.Env = os.Environ()
	if e.sudoAskpass != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SUDO_ASKPASS=%s", e.sudoAskpass))
	}
	for k, v := range options.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set working directory
	if options.WorkingDir != "" {
		cmd.Dir = options.WorkingDir
	}

	var stdout, stderr bytes.Buffer

	// For interactive or sudo commands without askpass, connect to terminal
	if options.Interactive || (options.Sudo && !options.SudoNonInteractive && e.sudoAskpass == "") {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if e.reporter != nil {
			e.reporter.ClearProgress()
		}

		err := cmd.Run()
		if ctx.Err() == context.DeadlineExceeded {
			return &secondary.CommandResult{
				Success:  false,
				Stderr:   "Command timed out",
				ExitCode: -1,
				Error:    ctx.Err(),
			}
		}
		if err != nil {
			exitCode := -1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
			return &secondary.CommandResult{
				Success:  false,
				Stderr:   err.Error(),
				ExitCode: exitCode,
				Error:    err,
			}
		}
		return &secondary.CommandResult{Success: true}
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if e.reporter != nil {
		e.reporter.ClearProgress()
	}

	if ctx.Err() == context.DeadlineExceeded {
		return &secondary.CommandResult{
			Success:  false,
			Stderr:   "Command timed out",
			ExitCode: -1,
			Error:    ctx.Err(),
		}
	}

	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return &secondary.CommandResult{
			Success:  false,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: exitCode,
			Error:    err,
		}
	}

	return &secondary.CommandResult{
		Success: true,
		Stdout:  strings.TrimSpace(stdout.String()),
	}
}

// CommandExists checks if a command is available in PATH.
func (e *ShellExecutor) CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetCommandPath returns the full path to a command.
func (e *ShellExecutor) GetCommandPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

// GetCommandVersion returns the version string of a command.
func (e *ShellExecutor) GetCommandVersion(name string, versionArg string) string {
	if versionArg == "" {
		versionArg = "--version"
	}

	result := e.Execute(context.Background(), []string{name, versionArg}, secondary.WithSkipDryRun())
	if result.Success {
		lines := strings.Split(result.Stdout, "\n")
		if len(lines) > 0 {
			return strings.TrimSpace(lines[0])
		}
	}
	return ""
}
