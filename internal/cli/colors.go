// Package cli provides CLI utilities for colored output and user interaction.
package cli

import (
	"os"

	"github.com/fatih/color"
)

// Colors provides pre-configured color functions for consistent output.
var (
	// Text colors
	Red     = color.New(color.FgRed)
	Green   = color.New(color.FgGreen)
	Yellow  = color.New(color.FgYellow)
	Blue    = color.New(color.FgBlue)
	Magenta = color.New(color.FgMagenta)
	Cyan    = color.New(color.FgCyan)
	White   = color.New(color.FgWhite)

	// Styled colors
	Bold      = color.New(color.Bold)
	Dim       = color.New(color.Faint)
	BoldRed   = color.New(color.FgRed, color.Bold)
	BoldGreen = color.New(color.FgGreen, color.Bold)
	BoldBlue  = color.New(color.FgBlue, color.Bold)
	BoldCyan  = color.New(color.FgCyan, color.Bold)
)

// Icons for status messages
const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconPending = "○"
	IconArrow   = "▶"
)

// DisableColors disables all color output (for non-TTY output).
func DisableColors() {
	color.NoColor = true
}

// EnableColors enables color output.
func EnableColors() {
	color.NoColor = false
}

// IsTerminal returns true if stdout is a terminal.
func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// InitColors initializes color support based on terminal detection.
func InitColors() {
	if !IsTerminal() {
		DisableColors()
	}
}
