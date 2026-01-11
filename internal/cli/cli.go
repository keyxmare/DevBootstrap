package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// CLI provides methods for user interaction and formatted output.
type CLI struct {
	NoInteraction bool
	DryRun        bool
}

// New creates a new CLI instance.
func New(noInteraction, dryRun bool) *CLI {
	InitColors()
	return &CLI{
		NoInteraction: noInteraction,
		DryRun:        dryRun,
	}
}

// Print prints a message without any formatting.
func (c *CLI) Print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Println prints a message with a newline.
func (c *CLI) Println(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// PrintHeader prints a styled header with borders.
func (c *CLI) PrintHeader(title string) {
	width := max(60, len(title)+4)
	border := strings.Repeat("═", width)

	fmt.Println()
	BoldCyan.Printf("╔%s╗\n", border)
	BoldCyan.Print("║ ")
	fmt.Print(centerString(title, width-2))
	BoldCyan.Println(" ║")
	BoldCyan.Printf("╚%s╝\n", border)
	fmt.Println()
}

// PrintSection prints a section header.
func (c *CLI) PrintSection(title string) {
	fmt.Println()
	BoldBlue.Printf("%s %s\n", IconArrow, title)
	Dim.Println(strings.Repeat("─", 50))
}

// PrintSuccess prints a success message.
func (c *CLI) PrintSuccess(message string) {
	Green.Printf("%s ", IconSuccess)
	fmt.Println(message)
}

// PrintError prints an error message.
func (c *CLI) PrintError(message string) {
	Red.Printf("%s ", IconError)
	fmt.Println(message)
}

// PrintWarning prints a warning message.
func (c *CLI) PrintWarning(message string) {
	Yellow.Printf("%s ", IconWarning)
	fmt.Println(message)
}

// PrintInfo prints an info message.
func (c *CLI) PrintInfo(message string) {
	Cyan.Printf("%s ", IconInfo)
	fmt.Println(message)
}

// PrintStep prints a step indicator (e.g., [1/5]).
func (c *CLI) PrintStep(current, total int, message string) {
	Cyan.Printf("[%d/%d] ", current, total)
	fmt.Println(message)
}

// PrintProgress prints a progress message (overwrites the current line).
func (c *CLI) PrintProgress(message string) {
	fmt.Printf("\r%s", strings.Repeat(" ", 80)) // Clear line
	fmt.Printf("\r")
	Dim.Printf("  %s...", message)
}

// ClearProgress clears the current progress line.
func (c *CLI) ClearProgress() {
	fmt.Printf("\r%s\r", strings.Repeat(" ", 80))
}

// PrintSummary prints a key-value summary.
func (c *CLI) PrintSummary(title string, items map[string]string) {
	c.PrintSection(title)
	for key, value := range items {
		Bold.Printf("  %s: ", key)
		fmt.Println(value)
	}
	fmt.Println()
}

// AskYesNo asks a yes/no question and returns the answer.
func (c *CLI) AskYesNo(question string, defaultValue bool) bool {
	if c.NoInteraction {
		answer := "non"
		if defaultValue {
			answer = "oui"
		}
		c.PrintInfo(fmt.Sprintf("%s → %s (auto)", question, answer))
		return defaultValue
	}

	result := defaultValue
	prompt := &survey.Confirm{
		Message: question,
		Default: defaultValue,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return defaultValue
	}

	return result
}

// AskChoice asks the user to select from a list of options.
func (c *CLI) AskChoice(question string, options []string, defaultIndex int) int {
	if c.NoInteraction {
		c.PrintInfo(fmt.Sprintf("%s → %s (auto)", question, options[defaultIndex]))
		return defaultIndex
	}

	var result string
	prompt := &survey.Select{
		Message: question,
		Options: options,
		Default: options[defaultIndex],
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return defaultIndex
	}

	// Find the selected index
	for i, opt := range options {
		if opt == result {
			return i
		}
	}

	return defaultIndex
}

// AskMultiSelect asks the user to select multiple options.
func (c *CLI) AskMultiSelect(question string, options []string, defaults []int) []int {
	if c.NoInteraction {
		c.PrintInfo(fmt.Sprintf("%s → selection automatique", question))
		return defaults
	}

	var results []string
	defaultStrs := make([]string, len(defaults))
	for i, idx := range defaults {
		if idx < len(options) {
			defaultStrs[i] = options[idx]
		}
	}

	prompt := &survey.MultiSelect{
		Message: question,
		Options: options,
		Default: defaultStrs,
	}

	if err := survey.AskOne(prompt, &results); err != nil {
		return defaults
	}

	// Convert selected strings back to indices
	indices := make([]int, 0, len(results))
	for _, result := range results {
		for i, opt := range options {
			if opt == result {
				indices = append(indices, i)
				break
			}
		}
	}

	return indices
}

// AskString asks the user for a string input.
func (c *CLI) AskString(question string, defaultValue string) string {
	if c.NoInteraction {
		c.PrintInfo(fmt.Sprintf("%s → %s (auto)", question, defaultValue))
		return defaultValue
	}

	var result string
	prompt := &survey.Input{
		Message: question,
		Default: defaultValue,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return defaultValue
	}

	return result
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

// Exit exits the program with the given code.
func Exit(code int) {
	os.Exit(code)
}
