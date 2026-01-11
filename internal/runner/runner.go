package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/keyxmare/DevBootstrap/internal/cli"
)

// Runner executes shell commands with logging and error handling.
type Runner struct {
	CLI     *cli.CLI
	DryRun  bool
	UseSudo bool
}

// New creates a new Runner instance.
func New(c *cli.CLI, dryRun bool) *Runner {
	return &Runner{
		CLI:    c,
		DryRun: dryRun,
	}
}

// Run executes a command and captures its output.
func (r *Runner) Run(args []string, opts ...Option) *Result {
	options := &runOptions{
		timeout: 2 * time.Minute,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Add sudo if needed
	if options.sudo && os.Geteuid() != 0 {
		args = append([]string{"sudo"}, args...)
	}

	cmdStr := strings.Join(args, " ")

	if options.description != "" {
		r.CLI.PrintProgress(options.description)
	}

	if r.DryRun {
		r.CLI.ClearProgress()
		r.CLI.PrintInfo(fmt.Sprintf("[DRY RUN] %s", cmdStr))
		return SuccessResult("")
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// Set environment
	cmd.Env = os.Environ()
	if len(options.env) > 0 {
		for k, v := range options.env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Set working directory
	if options.cwd != "" {
		cmd.Dir = options.cwd
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	r.CLI.ClearProgress()

	if ctx.Err() == context.DeadlineExceeded {
		return FailureResult("Command timed out", -1, ctx.Err())
	}

	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return FailureResult(stderr.String(), exitCode, err)
	}

	return SuccessResult(strings.TrimSpace(stdout.String()))
}

// RunInteractive executes a command with real-time output to the terminal.
func (r *Runner) RunInteractive(args []string, opts ...Option) bool {
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Add sudo if needed
	if options.sudo && os.Geteuid() != 0 {
		args = append([]string{"sudo"}, args...)
	}

	if options.description != "" {
		r.CLI.PrintInfo(options.description)
	}

	if r.DryRun {
		r.CLI.PrintInfo(fmt.Sprintf("[DRY RUN] %s", strings.Join(args, " ")))
		return true
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment
	cmd.Env = os.Environ()
	if len(options.env) > 0 {
		for k, v := range options.env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Set working directory
	if options.cwd != "" {
		cmd.Dir = options.cwd
	}

	if err := cmd.Run(); err != nil {
		r.CLI.PrintError(fmt.Sprintf("Erreur: %v", err))
		return false
	}

	return true
}

// CommandExists checks if a command exists in PATH.
func (r *Runner) CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetCommandPath returns the full path to a command.
func (r *Runner) GetCommandPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

// GetCommandVersion returns the version of a command.
func (r *Runner) GetCommandVersion(name string, versionArg string) string {
	if versionArg == "" {
		versionArg = "--version"
	}

	result := r.Run([]string{name, versionArg})
	if result.Success {
		// Return first line
		lines := strings.Split(result.Stdout, "\n")
		if len(lines) > 0 {
			return strings.TrimSpace(lines[0])
		}
	}
	return ""
}

// EnsureDirectory creates a directory if it doesn't exist.
func (r *Runner) EnsureDirectory(path string, mode os.FileMode) error {
	if r.DryRun {
		r.CLI.PrintInfo(fmt.Sprintf("[DRY RUN] mkdir -p %s", path))
		return nil
	}
	return os.MkdirAll(path, mode)
}

// DownloadFile downloads a file from a URL.
func (r *Runner) DownloadFile(url, destination string, opts ...Option) bool {
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.description != "" {
		r.CLI.PrintProgress(options.description)
	}

	// Try curl first
	if r.CommandExists("curl") {
		result := r.Run([]string{"curl", "-fsSL", "-o", destination, url}, WithTimeout(10*time.Minute))
		if result.Success {
			r.CLI.ClearProgress()
			return true
		}
	}

	// Fall back to wget
	if r.CommandExists("wget") {
		result := r.Run([]string{"wget", "-q", "-O", destination, url}, WithTimeout(10*time.Minute))
		if result.Success {
			r.CLI.ClearProgress()
			return true
		}
	}

	r.CLI.ClearProgress()
	r.CLI.PrintError("Ni curl ni wget n'est disponible")
	return false
}

// CopyFile copies a file from src to dst.
func (r *Runner) CopyFile(src, dst string) error {
	if r.DryRun {
		r.CLI.PrintInfo(fmt.Sprintf("[DRY RUN] cp %s %s", src, dst))
		return nil
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// RemoveAll removes a file or directory recursively.
func (r *Runner) RemoveAll(path string) error {
	if r.DryRun {
		r.CLI.PrintInfo(fmt.Sprintf("[DRY RUN] rm -rf %s", path))
		return nil
	}
	return os.RemoveAll(path)
}

// runOptions holds options for running commands.
type runOptions struct {
	description string
	sudo        bool
	env         map[string]string
	cwd         string
	timeout     time.Duration
}

// Option is a functional option for Run.
type Option func(*runOptions)

// WithDescription sets the description for the command.
func WithDescription(desc string) Option {
	return func(o *runOptions) {
		o.description = desc
	}
}

// WithSudo runs the command with sudo.
func WithSudo() Option {
	return func(o *runOptions) {
		o.sudo = true
	}
}

// WithEnv sets environment variables for the command.
func WithEnv(env map[string]string) Option {
	return func(o *runOptions) {
		o.env = env
	}
}

// WithCwd sets the working directory for the command.
func WithCwd(cwd string) Option {
	return func(o *runOptions) {
		o.cwd = cwd
	}
}

// WithTimeout sets the timeout for the command.
func WithTimeout(timeout time.Duration) Option {
	return func(o *runOptions) {
		o.timeout = timeout
	}
}
