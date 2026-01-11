// Package reporter provides progress reporting adapters.
package reporter

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Terminal formatting icons (Nerd Font compatible)
const (
	IconSuccess  = ""
	IconError    = ""
	IconWarning  = ""
	IconInfo     = ""
	IconArrow    = ""
	IconPackage  = ""
	IconCheck    = ""
	IconCross    = ""
	IconCircle   = ""
	IconDot      = ""
	IconRocket   = ""
	IconGear     = ""
	IconFolder   = ""
	IconTerminal = ""
	IconDocker   = ""
	IconCode     = ""
)

// Color palettes - Modern dark theme
var (
	// Primary colors
	primaryColor   = color.New(color.FgHiCyan)
	secondaryColor = color.New(color.FgHiMagenta)
	accentColor    = color.New(color.FgHiYellow)

	// Status colors
	successColor = color.New(color.FgHiGreen)
	errorColor   = color.New(color.FgHiRed)
	warningColor = color.New(color.FgHiYellow)
	infoColor    = color.New(color.FgHiBlue)

	// Text styles
	dimColor      = color.New(color.Faint)
	boldColor     = color.New(color.Bold)
	boldWhite     = color.New(color.FgHiWhite, color.Bold)
	boldCyan      = color.New(color.FgHiCyan, color.Bold)
	boldMagenta   = color.New(color.FgHiMagenta, color.Bold)
	boldGreen     = color.New(color.FgHiGreen, color.Bold)
	boldYellow    = color.New(color.FgHiYellow, color.Bold)
	boldRed       = color.New(color.FgHiRed, color.Bold)
	italicDim     = color.New(color.Faint, color.Italic)

	// Gradient-like effects
	gradientStart = color.New(color.FgHiCyan)
	gradientMid   = color.New(color.FgHiBlue)
	gradientEnd   = color.New(color.FgHiMagenta)

	// Background badges
	successBadge = color.New(color.BgGreen, color.FgHiWhite, color.Bold)
	errorBadge   = color.New(color.BgRed, color.FgHiWhite, color.Bold)
	warningBadge = color.New(color.BgYellow, color.FgBlack, color.Bold)
	infoBadge    = color.New(color.BgBlue, color.FgHiWhite, color.Bold)
)

// TerminalReporter implements ProgressReporter for terminal output.
type TerminalReporter struct {
	dryRun        bool
	spinnerActive bool
	spinnerStop   chan bool
}

// NewTerminalReporter creates a new TerminalReporter instance.
func NewTerminalReporter(dryRun bool) *TerminalReporter {
	initColors()
	return &TerminalReporter{
		dryRun:      dryRun,
		spinnerStop: make(chan bool),
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
	infoColor.Printf(" %s ", IconInfo)
	fmt.Println(message)
}

// Success reports a success message.
func (r *TerminalReporter) Success(message string) {
	successColor.Printf(" %s ", IconSuccess)
	boldWhite.Println(message)
}

// Warning reports a warning message.
func (r *TerminalReporter) Warning(message string) {
	warningColor.Printf(" %s ", IconWarning)
	boldYellow.Println(message)
}

// Error reports an error message.
func (r *TerminalReporter) Error(message string) {
	errorColor.Printf(" %s ", IconError)
	boldRed.Println(message)
}

// Progress reports a progress message with spinner animation.
func (r *TerminalReporter) Progress(message string) {
	r.ClearProgress()
	dimColor.Printf("   %s %s...", IconGear, message)
}

// ClearProgress clears any progress indicator.
func (r *TerminalReporter) ClearProgress() {
	fmt.Printf("\r%s\r", strings.Repeat(" ", 100))
}

// Section starts a new section with a styled header.
func (r *TerminalReporter) Section(title string) {
	fmt.Println()

	// Modern section header with gradient effect
	gradientStart.Print(" ┌")
	gradientMid.Print("─")
	gradientEnd.Print("─")
	boldCyan.Printf(" %s ", IconArrow)
	boldWhite.Print(title)
	fmt.Println()

	// Underline with gradient
	dimColor.Print(" │")
	fmt.Println()
}

// Step reports a step in a multi-step process.
func (r *TerminalReporter) Step(current, total int, message string) {
	progressBar := r.miniProgressBar(current, total)
	primaryColor.Printf(" %s ", progressBar)
	boldWhite.Printf("[%d/%d] ", current, total)
	fmt.Println(message)
}

// miniProgressBar creates a small visual progress indicator.
func (r *TerminalReporter) miniProgressBar(current, total int) string {
	filled := (current * 5) / total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 5-filled)
	return bar
}

// Header prints a modern ASCII art banner.
func (r *TerminalReporter) Header(title string) {
	fmt.Println()

	// Modern box with double lines and colors
	width := 62

	// Top border
	gradientStart.Print("  ╔")
	for i := 0; i < width; i++ {
		if i < width/3 {
			gradientStart.Print("═")
		} else if i < 2*width/3 {
			gradientMid.Print("═")
		} else {
			gradientEnd.Print("═")
		}
	}
	gradientEnd.Println("╗")

	// Empty line
	gradientStart.Print("  ║")
	fmt.Print(strings.Repeat(" ", width))
	gradientEnd.Println("║")

	// Title line with rocket icon
	gradientStart.Print("  ║")
	titleWithIcon := fmt.Sprintf(" %s  %s", IconRocket, title)
	padding := width - len([]rune(titleWithIcon))
	leftPad := padding / 2
	rightPad := padding - leftPad
	fmt.Print(strings.Repeat(" ", leftPad))
	accentColor.Print(IconRocket)
	boldWhite.Printf("  %s", title)
	fmt.Print(strings.Repeat(" ", rightPad))
	gradientEnd.Println("║")

	// Empty line
	gradientStart.Print("  ║")
	fmt.Print(strings.Repeat(" ", width))
	gradientEnd.Println("║")

	// Bottom border
	gradientStart.Print("  ╚")
	for i := 0; i < width; i++ {
		if i < width/3 {
			gradientStart.Print("═")
		} else if i < 2*width/3 {
			gradientMid.Print("═")
		} else {
			gradientEnd.Print("═")
		}
	}
	gradientEnd.Println("╝")

	fmt.Println()
}

// Summary prints a key-value summary with modern styling.
func (r *TerminalReporter) Summary(title string, items map[string]string) {
	r.Section(title)

	// Find max key length for alignment
	maxLen := 0
	for key := range items {
		if len(key) > maxLen {
			maxLen = len(key)
		}
	}

	for key, value := range items {
		dimColor.Print(" │  ")
		secondaryColor.Printf("%-*s", maxLen, key)
		dimColor.Print("  →  ")
		boldWhite.Println(value)
	}

	dimColor.Println(" │")
}

// AppCard displays an application with modern card styling.
func (r *TerminalReporter) AppCard(name, description, status string, installed bool, version string) {
	// Card border
	dimColor.Print(" ┌")
	dimColor.Println(strings.Repeat("─", 58) + "┐")

	// App name with icon and status badge
	dimColor.Print(" │ ")
	if installed {
		successColor.Printf("%s ", IconCheck)
		boldGreen.Printf("%-20s", name)
		successBadge.Print(" INSTALLED ")
	} else {
		dimColor.Printf("%s ", IconCircle)
		boldWhite.Printf("%-20s", name)
		dimColor.Print(" NOT INSTALLED ")
	}

	// Version if available
	if version != "" && installed {
		dimColor.Print(" ")
		italicDim.Printf("v%s", truncateVersion(version))
	}
	fmt.Println()

	// Description
	dimColor.Print(" │   ")
	dimColor.Println(description)

	// Bottom border
	dimColor.Print(" └")
	dimColor.Println(strings.Repeat("─", 58) + "┘")
}

// AppListItem displays an application as a list item (more compact).
func (r *TerminalReporter) AppListItem(name, description string, installed bool, version string) {
	if installed {
		successColor.Printf("  %s ", IconCheck)
		boldWhite.Print(name)
		successColor.Print(" ● ")
		dimColor.Print(description)
		if version != "" {
			dimColor.Print(" ")
			italicDim.Printf("(%s)", truncateVersion(version))
		}
	} else {
		dimColor.Printf("  %s ", IconCircle)
		fmt.Print(name)
		dimColor.Print(" ○ ")
		dimColor.Print(description)
	}
	fmt.Println()
}

// ProgressBar displays a progress bar.
func (r *TerminalReporter) ProgressBar(current, total int, label string) {
	width := 30
	filled := (current * width) / total
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	percent := (current * 100) / total

	fmt.Printf("\r  ")
	primaryColor.Printf("%s ", label)
	dimColor.Print("[")
	successColor.Print(bar[:filled])
	dimColor.Print(bar[filled:])
	dimColor.Print("] ")
	boldWhite.Printf("%3d%%", percent)
}

// Divider prints a styled divider line.
func (r *TerminalReporter) Divider() {
	fmt.Println()
	dimColor.Print("  ")
	for i := 0; i < 60; i++ {
		if i%3 == 0 {
			gradientStart.Print("·")
		} else if i%3 == 1 {
			gradientMid.Print("·")
		} else {
			gradientEnd.Print("·")
		}
	}
	fmt.Println()
}

// Badge prints a colored badge.
func (r *TerminalReporter) Badge(text string, badgeType string) {
	switch badgeType {
	case "success":
		successBadge.Printf(" %s ", text)
	case "error":
		errorBadge.Printf(" %s ", text)
	case "warning":
		warningBadge.Printf(" %s ", text)
	case "info":
		infoBadge.Printf(" %s ", text)
	default:
		dimColor.Printf(" %s ", text)
	}
}

// Spinner starts a spinner animation (call in goroutine).
func (r *TerminalReporter) StartSpinner(message string) {
	if r.spinnerActive {
		return
	}
	r.spinnerActive = true

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-r.spinnerStop:
				r.ClearProgress()
				r.spinnerActive = false
				return
			default:
				fmt.Printf("\r  ")
				primaryColor.Printf("%s ", frames[i%len(frames)])
				dimColor.Printf("%s...", message)
				time.Sleep(80 * time.Millisecond)
				i++
			}
		}
	}()
}

// StopSpinner stops the spinner animation.
func (r *TerminalReporter) StopSpinner() {
	if r.spinnerActive {
		r.spinnerStop <- true
	}
}

// truncateVersion truncates long version strings.
func truncateVersion(version string) string {
	// Remove common prefixes
	version = strings.TrimPrefix(version, "Docker version ")
	version = strings.TrimPrefix(version, "NVIM ")
	version = strings.TrimPrefix(version, "zsh ")

	// Truncate if too long
	if len(version) > 25 {
		return version[:22] + "..."
	}
	return version
}

// centerString centers a string within a given width.
func centerString(s string, width int) string {
	runeLen := len([]rune(s))
	if runeLen >= width {
		return s
	}
	padding := (width - runeLen) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", width-runeLen-padding)
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
