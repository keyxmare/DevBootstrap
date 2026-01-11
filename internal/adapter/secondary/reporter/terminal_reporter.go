// Package reporter provides progress reporting adapters.
package reporter

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// Soft minimal design constants
const (
	minWidth     = 60
	maxWidth     = 100
	paddingRatio = 0.15 // 15% padding on each side
)

// Soft minimal icons (simple, clean)
const (
	IconCheck   = "●"
	IconEmpty   = "○"
	IconArrow   = "→"
	IconDot     = "·"
	IconSuccess = "✓"
	IconError   = "✕"
	IconWarning = "!"
	IconInfo    = "i"
)

// Soft minimal color palette - muted, elegant colors
var (
	// Primary accent - soft cyan/teal
	accent = color.New(color.FgCyan)

	// Text hierarchy
	textPrimary   = color.New(color.FgHiWhite)
	textSecondary = color.New(color.FgWhite)
	textMuted     = color.New(color.Faint)
	textDim       = color.New(color.FgHiBlack)

	// Status colors - softer versions
	statusSuccess = color.New(color.FgGreen)
	statusError   = color.New(color.FgRed)
	statusWarning = color.New(color.FgYellow)
	statusInfo    = color.New(color.FgBlue)

	// Decorative
	borderColor = color.New(color.FgHiBlack)
	labelColor  = color.New(color.FgCyan, color.Faint)
)

// TerminalReporter implements ProgressReporter for terminal output.
type TerminalReporter struct {
	dryRun        bool
	width         int
	contentWidth  int
	leftPadding   int
	spinnerActive bool
	spinnerStop   chan bool
}

// NewTerminalReporter creates a new TerminalReporter instance.
func NewTerminalReporter(dryRun bool) *TerminalReporter {
	initColors()
	r := &TerminalReporter{
		dryRun:      dryRun,
		spinnerStop: make(chan bool),
	}
	r.calculateDimensions()
	return r
}

// calculateDimensions calculates the layout dimensions based on terminal size.
func (r *TerminalReporter) calculateDimensions() {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < minWidth {
		width = 80
	}

	r.width = width

	// Calculate content width (max 100 chars, with padding)
	r.contentWidth = width - int(float64(width)*paddingRatio*2)
	if r.contentWidth > maxWidth {
		r.contentWidth = maxWidth
	}
	if r.contentWidth < minWidth {
		r.contentWidth = minWidth
	}

	// Calculate left padding to center content
	r.leftPadding = (width - r.contentWidth) / 2
	if r.leftPadding < 2 {
		r.leftPadding = 2
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

// pad returns the left padding string.
func (r *TerminalReporter) pad() string {
	return strings.Repeat(" ", r.leftPadding)
}

// line creates a line of specific width with a character.
func (r *TerminalReporter) line(char string) string {
	return strings.Repeat(char, r.contentWidth)
}

// center centers text within the content width.
func (r *TerminalReporter) center(text string) string {
	textLen := len([]rune(text))
	if textLen >= r.contentWidth {
		return text
	}
	leftPad := (r.contentWidth - textLen) / 2
	return strings.Repeat(" ", leftPad) + text
}

// Header prints a clean, minimal header.
func (r *TerminalReporter) Header(title string) {
	// Clear screen effect - just add spacing
	fmt.Println()
	fmt.Println()

	// Subtle top border
	fmt.Print(r.pad())
	textDim.Println(r.line("─"))

	// Empty line
	fmt.Println()

	// Title centered
	fmt.Print(r.pad())
	textPrimary.Println(r.center(title))

	// Subtitle
	fmt.Print(r.pad())
	textMuted.Println(r.center("Configuration de votre environnement de développement"))

	// Empty line
	fmt.Println()

	// Subtle bottom border
	fmt.Print(r.pad())
	textDim.Println(r.line("─"))

	fmt.Println()
}

// Section starts a new section with minimal styling.
func (r *TerminalReporter) Section(title string) {
	fmt.Println()
	fmt.Print(r.pad())
	accent.Print("  ")
	textPrimary.Println(title)
	fmt.Print(r.pad())
	textDim.Println("  " + strings.Repeat("─", len(title)+4))
}

// Info reports an informational message.
func (r *TerminalReporter) Info(message string) {
	fmt.Print(r.pad())
	textDim.Print("  ")
	statusInfo.Print(IconInfo)
	textDim.Print("  ")
	textSecondary.Println(message)
}

// Success reports a success message.
func (r *TerminalReporter) Success(message string) {
	fmt.Print(r.pad())
	textDim.Print("  ")
	statusSuccess.Print(IconSuccess)
	textDim.Print("  ")
	textPrimary.Println(message)
}

// Warning reports a warning message.
func (r *TerminalReporter) Warning(message string) {
	fmt.Print(r.pad())
	textDim.Print("  ")
	statusWarning.Print(IconWarning)
	textDim.Print("  ")
	statusWarning.Println(message)
}

// Error reports an error message.
func (r *TerminalReporter) Error(message string) {
	fmt.Print(r.pad())
	textDim.Print("  ")
	statusError.Print(IconError)
	textDim.Print("  ")
	statusError.Println(message)
}

// Progress reports a progress message.
func (r *TerminalReporter) Progress(message string) {
	r.ClearProgress()
	fmt.Print(r.pad())
	textMuted.Printf("  %s  %s...", IconDot, message)
}

// ClearProgress clears any progress indicator.
func (r *TerminalReporter) ClearProgress() {
	fmt.Printf("\r%s\r", strings.Repeat(" ", r.width))
}

// Step reports a step in a multi-step process.
func (r *TerminalReporter) Step(current, total int, message string) {
	fmt.Print(r.pad())
	textMuted.Print("  ")
	accent.Printf("[%d/%d]", current, total)
	textMuted.Print("  ")
	textSecondary.Println(message)
}

// Summary prints a key-value summary with clean layout.
func (r *TerminalReporter) Summary(title string, items map[string]string) {
	r.Section(title)
	fmt.Println()

	// Find max key length for alignment
	maxLen := 0
	for key := range items {
		if len(key) > maxLen {
			maxLen = len(key)
		}
	}

	for key, value := range items {
		fmt.Print(r.pad())
		textMuted.Print("      ")
		labelColor.Printf("%-*s", maxLen, key)
		textDim.Print("  ")
		textSecondary.Println(value)
	}
}

// AppItem displays an application in a clean list format.
func (r *TerminalReporter) AppItem(name, description string, installed bool, version string) {
	fmt.Print(r.pad())

	if installed {
		statusSuccess.Print("  " + IconCheck + "  ")
		textPrimary.Print(name)
		if version != "" {
			textMuted.Printf("  %s", version)
		}
	} else {
		textDim.Print("  " + IconEmpty + "  ")
		textSecondary.Print(name)
	}
	fmt.Println()

	// Description on next line, indented
	fmt.Print(r.pad())
	textMuted.Printf("       %s\n", description)
}

// Divider prints a subtle divider.
func (r *TerminalReporter) Divider() {
	fmt.Println()
	fmt.Print(r.pad())
	textDim.Print("  ")
	for i := 0; i < r.contentWidth-4; i++ {
		if i%2 == 0 {
			textDim.Print(IconDot)
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Println()
}

// EmptyLine prints a blank line with proper padding.
func (r *TerminalReporter) EmptyLine() {
	fmt.Println()
}

// Text prints regular text with padding.
func (r *TerminalReporter) Text(message string) {
	fmt.Print(r.pad())
	textSecondary.Printf("  %s\n", message)
}

// Muted prints muted/dim text with padding.
func (r *TerminalReporter) Muted(message string) {
	fmt.Print(r.pad())
	textMuted.Printf("  %s\n", message)
}

// Accent prints accented text.
func (r *TerminalReporter) Accent(message string) {
	fmt.Print(r.pad())
	accent.Printf("  %s\n", message)
}

// ListItem prints a simple list item.
func (r *TerminalReporter) ListItem(text string, highlighted bool) {
	fmt.Print(r.pad())
	if highlighted {
		accent.Print("  " + IconArrow + "  ")
		textPrimary.Println(text)
	} else {
		textMuted.Print("     ")
		textSecondary.Println(text)
	}
}

// OptionItem prints an option for selection.
func (r *TerminalReporter) OptionItem(number int, title, description string) {
	fmt.Print(r.pad())
	textMuted.Print("  ")
	accent.Printf("[%d]", number)
	textMuted.Print("  ")
	textPrimary.Print(title)
	if description != "" {
		textMuted.Printf("  %s  %s", IconArrow, description)
	}
	fmt.Println()
}

// StartSpinner starts a spinner animation.
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
				fmt.Printf("\r%s", r.pad())
				accent.Printf("  %s  ", frames[i%len(frames)])
				textMuted.Printf("%s", message)
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

// Box prints content in a minimal box.
func (r *TerminalReporter) Box(lines []string) {
	fmt.Println()
	fmt.Print(r.pad())
	borderColor.Print("  ┌")
	borderColor.Print(strings.Repeat("─", r.contentWidth-6))
	borderColor.Println("┐")

	for _, line := range lines {
		fmt.Print(r.pad())
		borderColor.Print("  │")
		fmt.Print("  ")
		textSecondary.Print(line)
		// Calculate remaining space
		remaining := r.contentWidth - 10 - len([]rune(line))
		if remaining > 0 {
			fmt.Print(strings.Repeat(" ", remaining))
		}
		borderColor.Println("│")
	}

	fmt.Print(r.pad())
	borderColor.Print("  └")
	borderColor.Print(strings.Repeat("─", r.contentWidth-6))
	borderColor.Println("┘")
	fmt.Println()
}

// StatusBadge prints a status indicator.
func (r *TerminalReporter) StatusBadge(status string) {
	switch status {
	case "installed":
		statusSuccess.Print(" ● ")
	case "not_installed":
		textDim.Print(" ○ ")
	case "error":
		statusError.Print(" ✕ ")
	case "pending":
		textMuted.Print(" ◌ ")
	}
}

// GetPadding returns the left padding for external use.
func (r *TerminalReporter) GetPadding() string {
	return r.pad()
}

// GetContentWidth returns the content width.
func (r *TerminalReporter) GetContentWidth() int {
	return r.contentWidth
}
