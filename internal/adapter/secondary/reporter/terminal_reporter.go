// Package reporter provides progress reporting adapters.
package reporter

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Terminal formatting icons
const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconArrow   = "▶"
)

// Color functions
var (
	red       = color.New(color.FgRed)
	green     = color.New(color.FgGreen)
	yellow    = color.New(color.FgYellow)
	blue      = color.New(color.FgBlue)
	cyan      = color.New(color.FgCyan)
	dim       = color.New(color.Faint)
	bold      = color.New(color.Bold)
	boldBlue  = color.New(color.FgBlue, color.Bold)
	boldCyan  = color.New(color.FgCyan, color.Bold)
)

// TerminalReporter implements ProgressReporter for terminal output.
type TerminalReporter struct {
	dryRun bool
}

// NewTerminalReporter creates a new TerminalReporter instance.
func NewTerminalReporter(dryRun bool) *TerminalReporter {
	initColors()
	return &TerminalReporter{
		dryRun: dryRun,
	}
}

// initColors initializes color support based on terminal detection.
func initColors() {
	fi, err := os.Stdout.Stat()
	if err != nil {
		color.NoColor = true
		return
	}
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		color.NoColor = true
	}
}

// Info reports an informational message.
func (r *TerminalReporter) Info(message string) {
	cyan.Printf("%s ", IconInfo)
	fmt.Println(message)
}

// Success reports a success message.
func (r *TerminalReporter) Success(message string) {
	green.Printf("%s ", IconSuccess)
	fmt.Println(message)
}

// Warning reports a warning message.
func (r *TerminalReporter) Warning(message string) {
	yellow.Printf("%s ", IconWarning)
	fmt.Println(message)
}

// Error reports an error message.
func (r *TerminalReporter) Error(message string) {
	red.Printf("%s ", IconError)
	fmt.Println(message)
}

// Progress reports a progress message (may overwrite line).
func (r *TerminalReporter) Progress(message string) {
	fmt.Printf("\r%s", strings.Repeat(" ", 80))
	fmt.Printf("\r")
	dim.Printf("  %s...", message)
}

// ClearProgress clears any progress indicator.
func (r *TerminalReporter) ClearProgress() {
	fmt.Printf("\r%s\r", strings.Repeat(" ", 80))
}

// Section starts a new section with a header.
func (r *TerminalReporter) Section(title string) {
	fmt.Println()
	boldBlue.Printf("%s %s\n", IconArrow, title)
	dim.Println(strings.Repeat("─", 50))
}

// Step reports a step in a multi-step process.
func (r *TerminalReporter) Step(current, total int, message string) {
	cyan.Printf("[%d/%d] ", current, total)
	fmt.Println(message)
}

// Header prints a styled header.
func (r *TerminalReporter) Header(title string) {
	width := max(60, len(title)+4)
	border := strings.Repeat("═", width)

	fmt.Println()
	boldCyan.Printf("╔%s╗\n", border)
	boldCyan.Print("║ ")
	fmt.Print(centerString(title, width-2))
	boldCyan.Println(" ║")
	boldCyan.Printf("╚%s╝\n", border)
	fmt.Println()
}

// Summary prints a key-value summary.
func (r *TerminalReporter) Summary(title string, items map[string]string) {
	r.Section(title)
	for key, value := range items {
		bold.Printf("  %s: ", key)
		fmt.Println(value)
	}
	fmt.Println()
}

// centerString centers a string within a given width.
func centerString(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", width-len(s)-padding)
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
