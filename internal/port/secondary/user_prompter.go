package secondary

// UserPrompter defines the interface for user interaction.
type UserPrompter interface {
	// Confirm asks a yes/no question.
	Confirm(question string, defaultValue bool) bool

	// Select asks user to select one option.
	Select(question string, options []string, defaultIndex int) int

	// MultiSelect asks user to select multiple options.
	MultiSelect(question string, options []string, defaultIndices []int) []int

	// Input asks user for text input.
	Input(question string, defaultValue string) string

	// Password asks user for password input (hidden).
	Password(question string) string
}
